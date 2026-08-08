package payments

import (
	"encoding/xml"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"veloham/backend/internal/config"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
)

const ListingPlacementPrice = 30

type Handler struct {
	db     *gorm.DB
	cfg    config.Config
	client FreedomPayClient
}

func NewHandler(db *gorm.DB, cfg config.Config) Handler {
	return Handler{db: db, cfg: cfg, client: FreedomPayClient{BaseURL: cfg.FreedomPayAPIBase, MerchantID: cfg.FreedomPayMerchantID, Secret: cfg.FreedomPaySecret}}
}

func (h Handler) Quota(c *gin.Context) {
	var used int64
	h.db.Model(&models.ListingPlacement{}).Where("user_id = ? AND kind = 'free' AND status = 'paid'", middleware.CurrentUserID(c)).Count(&used)
	c.JSON(http.StatusOK, gin.H{"free_limit": 3, "free_used": used, "free_remaining": max(0, 3-used), "next_price": ListingPlacementPrice, "currency": "KGS"})
}

func (h Handler) Checkout(c *gin.Context) {
	if h.cfg.FreedomPayMerchantID == "" || h.cfg.FreedomPaySecret == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "payment provider is not configured"})
		return
	}
	var placement models.ListingPlacement
	if err := h.db.Where("id = ? AND user_id = ? AND kind = 'paid'", c.Param("id"), middleware.CurrentUserID(c)).First(&placement).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	if placement.Status == "paid" {
		c.JSON(http.StatusOK, placement)
		return
	}
	if placement.Status == "failed" {
		h.db.Model(&placement).Updates(map[string]any{"status": "pending", "provider_payment_id": "", "checkout_url": ""})
		placement.Status, placement.ProviderPaymentID, placement.CheckoutURL = "pending", "", ""
	}
	if placement.CheckoutURL != "" {
		c.JSON(http.StatusOK, placement)
		return
	}
	claimed := h.db.Model(&models.ListingPlacement{}).
		Where("id = ? AND status = 'pending' AND checkout_url = '' AND (provider IS NULL OR provider = '' OR provider = 'freedompay')", placement.ID).
		Update("provider", "freedompay_initializing")
	if claimed.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize payment"})
		return
	}
	if claimed.RowsAffected == 0 {
		h.db.First(&placement, "id = ?", placement.ID)
		if placement.CheckoutURL != "" {
			c.JSON(http.StatusOK, placement)
			return
		}
		c.JSON(http.StatusConflict, gin.H{"error": "payment initialization is already in progress"})
		return
	}
	var user models.User
	h.db.First(&user, "id = ?", placement.UserID)
	contactEmail := user.Email
	if strings.HasSuffix(contactEmail, "@phone.veloham.local") {
		contactEmail = ""
	}
	base := strings.TrimRight(h.cfg.PublicBaseURL, "/")
	paymentID, checkoutURL, err := h.client.Init(c.Request.Context(), placement.ID, placement.UserID, contactEmail, base+"/api/v1/payments/freedompay/result", base+"/payment?order="+placement.ID)
	if err != nil {
		h.db.Model(&placement).Update("provider", "freedompay")
		log.Printf("initialize FreedomPay payment: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to initialize payment"})
		return
	}
	h.db.Model(&placement).Updates(map[string]any{"provider": "freedompay", "provider_payment_id": paymentID, "checkout_url": checkoutURL})
	placement.Provider, placement.ProviderPaymentID, placement.CheckoutURL = "freedompay", paymentID, checkoutURL
	c.JSON(http.StatusOK, placement)
}

func (h Handler) Status(c *gin.Context) {
	var placement models.ListingPlacement
	if err := h.db.Where("id = ? AND user_id = ?", c.Param("id"), middleware.CurrentUserID(c)).First(&placement).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}
	if placement.Status == "pending" && placement.ProviderPaymentID != "" {
		if paid, err := h.client.Status(c.Request.Context(), placement.ProviderPaymentID); err == nil && paid {
			_ = h.markPaid(placement.ID, placement.ProviderPaymentID)
			h.db.First(&placement, "id = ?", placement.ID)
		}
	}
	c.JSON(http.StatusOK, placement)
}

func (h Handler) Result(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")
	var parseErr error
	if strings.HasPrefix(contentType, "multipart/form-data") {
		parseErr = c.Request.ParseMultipartForm(1 << 20)
	} else {
		parseErr = c.Request.ParseForm()
	}
	if parseErr != nil || !h.client.VerifyCallback(c.Request.Form) {
		c.Status(http.StatusBadRequest)
		return
	}
	values := c.Request.Form
	var placement models.ListingPlacement
	if err := h.db.First(&placement, "id = ?", values.Get("pg_order_id")).Error; err != nil ||
		placement.Amount != ListingPlacementPrice || placement.Currency != "KGS" ||
		!validListingFee(values.Get("pg_amount")) || values.Get("pg_currency") != "KGS" ||
		values.Get("pg_merchant_id") != h.cfg.FreedomPayMerchantID {
		h.writeResult(c, "rejected", "Платеж не найден")
		return
	}
	paymentID := values.Get("pg_payment_id")
	if paymentID == "" || placement.ProviderPaymentID != "" && placement.ProviderPaymentID != paymentID {
		h.writeResult(c, "rejected", "Неверный платеж")
		return
	}
	if placement.Status == "paid" {
		h.writeResult(c, "ok", "Заказ уже оплачен")
		return
	}
	if values.Get("pg_result") == "2" {
		h.writeResult(c, "ok", "Платеж обрабатывается")
		return
	}
	if values.Get("pg_result") != "1" {
		h.db.Model(&placement).Update("status", "failed")
		h.writeResult(c, "ok", "Результат принят")
		return
	}
	paid, err := h.client.Status(c.Request.Context(), paymentID)
	if err != nil || !paid {
		h.writeResult(c, "error", "Статус платежа не подтвержден")
		return
	}
	if err := h.markPaid(placement.ID, paymentID); err != nil {
		log.Printf("activate paid listing: %v", err)
		h.writeResult(c, "error", "Внутренняя ошибка")
		return
	}
	h.writeResult(c, "ok", "Заказ оплачен")
}

func (h Handler) markPaid(placementID, paymentID string) error {
	return h.db.Transaction(func(tx *gorm.DB) error {
		var placement models.ListingPlacement
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&placement, "id = ?", placementID).Error; err != nil {
			return err
		}
		if placement.Status == "paid" {
			return nil
		}
		now := time.Now()
		if err := tx.Model(&placement).Updates(map[string]any{"status": "paid", "paid_at": now, "provider_payment_id": paymentID}).Error; err != nil {
			return err
		}
		if placement.ListingID != nil {
			return tx.Model(&models.Listing{}).Where("id = ? AND status = 'pending_payment'", *placement.ListingID).Update("status", placement.TargetStatus).Error
		}
		return nil
	})
}

type resultXML struct {
	XMLName     xml.Name `xml:"response"`
	Status      string   `xml:"pg_status"`
	Description string   `xml:"pg_description"`
	Salt        string   `xml:"pg_salt"`
	Signature   string   `xml:"pg_sig"`
}

func (h Handler) writeResult(c *gin.Context, status, description string) {
	salt := randomSalt()
	params := map[string]string{"pg_status": status, "pg_description": description, "pg_salt": salt}
	c.XML(http.StatusOK, resultXML{Status: status, Description: description, Salt: salt, Signature: sign("result", params, h.cfg.FreedomPaySecret)})
}
