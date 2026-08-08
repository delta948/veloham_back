package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"veloham/backend/internal/middleware"
	"veloham/backend/internal/models"
	"veloham/backend/internal/services"
)

type ListingHandler struct {
	db        *gorm.DB
	uploadDir string
}

const (
	maxImageBytes         = 5 << 20
	maxImagesPerRequest   = 8
	maxImagesPerListing   = 12
	maxUploadRequestBytes = 42 << 20
)

var errImageTooLarge = errors.New("image exceeds the 5 MB limit")

func NewListingHandler(db *gorm.DB, uploadDir string) ListingHandler {
	return ListingHandler{db: db, uploadDir: uploadDir}
}

func (h ListingHandler) List(c *gin.Context) {
	var listings []models.Listing
	q := h.db.Preload("Images").Preload("User").Preload("BuildCard").Preload("MatchPref").Where("status NOT IN ?", []string{"hidden", "pending_payment"})
	if search := c.Query("search"); search != "" {
		like := "%" + search + "%"
		q = q.Where("title ILIKE ? OR description ILIKE ? OR city ILIKE ?", like, like, like)
	}
	for _, f := range []string{"category", "city", "condition", "status"} {
		if v := c.Query(f); v != "" {
			q = q.Where(f+" = ?", v)
		}
	}
	if brand := c.Query("brand"); brand != "" {
		q = q.Where("brand ILIKE ?", "%"+brand+"%")
	}
	if bikeType := c.Query("bike_type"); bikeType != "" {
		if normalized := normalizeBikeType(bikeType); normalized != "" {
			q = q.Where("bike_type = ?", normalized)
		}
	}
	if frameSizes := parseFrameSizeQuery(c); len(frameSizes) > 0 {
		var clauses []string
		var args []any
		for _, size := range frameSizes {
			clauses = append(clauses, "frame_size ILIKE ?")
			args = append(args, size)
		}
		q = q.Where("("+strings.Join(clauses, " OR ")+")", args...)
	}
	if riderHeight := parseInt(c.Query("rider_height")); riderHeight > 0 {
		q = q.Where("(rider_height_min = 0 OR rider_height_min <= ?) AND (rider_height_max = 0 OR rider_height_max >= ?)", riderHeight, riderHeight)
	}
	if categories := c.Query("categories"); categories != "" {
		q = q.Where("category IN ?", strings.Split(categories, ","))
	}
	if min := c.Query("min_price"); min != "" {
		q = q.Where("price >= ?", min)
	}
	if max := c.Query("max_price"); max != "" {
		q = q.Where("price <= ?", max)
	}
	if parseBool(c.Query("price_reduced")) {
		q = q.Where("price < initial_price")
	}
	if dealType := c.Query("deal_type"); dealType != "" {
		q = q.Where("deal_type = ?", dealType)
	}
	if labels := parseLabelsQuery(c.Query("labels")); len(labels) > 0 {
		for _, label := range labels {
			q = applyLabelFilter(q, label)
		}
	}
	if label := c.Query("label"); label != "" {
		if isAllowedListingLabel(label) {
			q = applyLabelFilter(q, label)
		}
	}
	for query, column := range map[string]string{
		"is_urgent":                "is_urgent",
		"is_bargain":               "is_bargain",
		"is_exchange":              "is_exchange",
		"extra_payment_from_me":    "extra_payment_from_me",
		"extra_payment_from_buyer": "extra_payment_from_buyer",
	} {
		if parseBool(c.Query(query)) {
			q = q.Where(column+" = ?", true)
		}
	}
	switch c.DefaultQuery("sort", "new") {
	case "price_asc":
		q = q.Order("price asc")
	case "price_desc":
		q = q.Order("price desc")
	case "popular":
		q = q.Order("views desc").Order("created_at desc")
	case "price_reduced_recently":
		q = q.Order("(SELECT max(changed_at) FROM listing_price_history ph WHERE ph.listing_id = listings.id AND ph.new_price < ph.old_price) desc nulls last")
	case "biggest_reduction":
		q = q.Order("(initial_price - price) desc")
	case "minimum_price":
		q = q.Order("price asc")
	case "maximum_price":
		q = q.Order("price desc")
	default:
		q = q.Order("created_at desc")
	}
	if err := q.Find(&listings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for i := range listings {
		hydrateListingContract(&listings[i])
		h.hydratePriceSummary(&listings[i])
	}
	c.JSON(http.StatusOK, listings)
}

func (h ListingHandler) Get(c *gin.Context) {
	var listing models.Listing
	if err := h.db.Preload("Images").Preload("User").Preload("BuildCard").Preload("MatchPref").First(&listing, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	if listing.Status == "pending_payment" {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	h.db.Model(&listing).UpdateColumn("views", gorm.Expr("views + ?", 1))
	hydrateListingContract(&listing)
	h.hydratePriceSummary(&listing)
	c.JSON(http.StatusOK, listing)
}

func (h ListingHandler) hydratePriceSummary(listing *models.Listing) {
	var last models.ListingPriceHistory
	result := h.db.Where("listing_id = ?", listing.ID).Order("changed_at desc").Limit(1).Find(&last)
	if result.Error == nil && result.RowsAffected > 0 {
		listing.PreviousPrice = &last.OldPrice
		listing.LastPriceChangeAt = &last.ChangedAt
	}
}

func (h ListingHandler) Create(c *gin.Context) {
	limitUploadBody(c)
	files, err := validateImageUploads(c)
	if err != nil {
		writeUploadError(c, err)
		return
	}
	listing, ok := h.bindListing(c)
	if !ok {
		return
	}
	listing.UserID = middleware.CurrentUserID(c)
	if listing.Status == "" {
		listing.Status = "active"
	}
	if listing.DealType == "" {
		listing.DealType = "продажа"
	}
	listing.DealType = normalizeDealType(listing.DealType)
	normalizeListingContract(&listing)
	if listing.Price <= 0 || listing.Price > services.MaxListingPrice {
		c.JSON(http.StatusBadRequest, gin.H{"error": services.ErrInvalidPrice.Error()})
		return
	}
	listing.InitialPrice = listing.Price
	targetStatus := listing.Status
	if targetStatus != "active" && targetStatus != "hidden" {
		targetStatus = "active"
	}
	placement := models.ListingPlacement{UserID: listing.UserID, TargetStatus: targetStatus, Currency: "KGS"}
	paymentRequired := false
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&user, "id = ?", listing.UserID).Error; err != nil {
			return err
		}
		var freeUsed int64
		if err := tx.Model(&models.ListingPlacement{}).Where("user_id = ? AND kind = 'free' AND status = 'paid'", listing.UserID).Count(&freeUsed).Error; err != nil {
			return err
		}
		if freeUsed < 3 {
			placement.Kind, placement.Status, placement.Amount = "free", "paid", 0
			now := time.Now()
			placement.PaidAt = &now
			listing.Status = targetStatus
		} else {
			var pending int64
			if err := tx.Model(&models.ListingPlacement{}).Where("user_id = ? AND kind = 'paid' AND status = 'pending'", listing.UserID).Count(&pending).Error; err != nil {
				return err
			}
			if pending >= 3 {
				return errors.New("too many pending listing payments")
			}
			placement.Kind, placement.Status, placement.Amount = "paid", "pending", 30
			listing.Status, paymentRequired = "pending_payment", true
		}
		if err := tx.Create(&listing).Error; err != nil {
			return err
		}
		placement.ListingID = &listing.ID
		return tx.Create(&placement).Error
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.saveImages(files, listing.ID); err != nil {
		h.db.Delete(&placement)
		h.db.Delete(&listing)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save images"})
		return
	}
	h.saveBuildCard(c, listing.ID)
	h.saveMatchPreference(c, listing.ID)
	h.db.Preload("Images").Preload("User").Preload("BuildCard").Preload("MatchPref").First(&listing, "id = ?", listing.ID)
	hydrateListingContract(&listing)
	if paymentRequired {
		c.JSON(http.StatusAccepted, gin.H{"listing": listing, "payment_required": true, "payment_id": placement.ID, "amount": placement.Amount, "currency": placement.Currency})
		return
	}
	c.JSON(http.StatusCreated, listing)
}

func (h ListingHandler) Update(c *gin.Context) {
	limitUploadBody(c)
	var listing models.Listing
	if err := h.db.First(&listing, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	actorID := middleware.CurrentUserID(c)
	isAdmin := actorIsAdmin(h.db, actorID)
	if listing.UserID != actorID && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": services.ErrListingForbidden.Error()})
		return
	}
	files, err := validateImageUploads(c)
	if err != nil {
		writeUploadError(c, err)
		return
	}
	if err := h.ensureImageCapacity(listing.ID, len(files)); err != nil {
		writeUploadError(c, err)
		return
	}
	patch, ok := h.bindListing(c)
	if !ok {
		return
	}
	patch.DealType = normalizeDealType(patch.DealType)
	normalizeListingContract(&patch)
	if listing.Status == "pending_payment" {
		patch.Status = "pending_payment"
	}
	priceResult, err := (services.PriceHistoryService{DB: h.db}).ChangePrice(listing.ID, actorID, isAdmin, patch.Price, c.ClientIP())
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, services.ErrListingForbidden) {
			status = http.StatusForbidden
		}
		if errors.Is(err, services.ErrPriceRateLimited) {
			status = http.StatusTooManyRequests
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Model(&listing).Updates(map[string]any{
		"title": patch.Title, "description": patch.Description,
		"city": patch.City, "brand": patch.Brand, "category": patch.Category, "bike_type": patch.BikeType, "condition": patch.Condition,
		"frame_size": patch.FrameSize, "frame_size_text": patch.FrameSizeText,
		"rider_height_min": patch.RiderMin, "rider_height_max": patch.RiderMax,
		"recommended_height_min": patch.RecommendedHeightMin, "recommended_height_max": patch.RecommendedHeightMax,
		"deal_type": patch.DealType, "labels": patch.Labels,
		"is_urgent": patch.IsUrgent, "is_bargain": patch.IsBargain, "is_exchange": patch.IsExchange,
		"extra_payment_from_me": patch.ExtraPaymentFromMe, "extra_payment_from_buyer": patch.ExtraPaymentFromBuyer,
		"status": patch.Status,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.saveImages(files, listing.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save images"})
		return
	}
	h.saveBuildCard(c, listing.ID)
	h.saveMatchPreference(c, listing.ID)
	h.db.Preload("Images").Preload("User").Preload("BuildCard").Preload("MatchPref").First(&listing, "id = ?", listing.ID)
	hydrateListingContract(&listing)
	c.JSON(http.StatusOK, gin.H{"listing": listing, "price_change": priceResult.Change})
}

func (h ListingHandler) Delete(c *gin.Context) {
	var listing models.Listing
	if err := h.db.First(&listing, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	if listing.UserID != middleware.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can delete listing"})
		return
	}
	if listing.Status == "pending_payment" {
		var placement models.ListingPlacement
		if err := h.db.Where("listing_id = ?", listing.ID).First(&placement).Error; err == nil && placement.CheckoutURL != "" {
			c.JSON(http.StatusConflict, gin.H{"error": "payment has already been initialized"})
			return
		}
		h.db.Where("listing_id = ?", listing.ID).Delete(&models.ListingPlacement{})
	}
	if err := deleteListingCascade(h.db, listing.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h ListingHandler) PatchStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}
	var listing models.Listing
	if err := h.db.First(&listing, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	if listing.UserID != middleware.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can change status"})
		return
	}
	if listing.Status == "pending_payment" {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "payment is required before publishing"})
		return
	}
	h.db.Model(&listing).Update("status", req.Status)
	c.JSON(http.StatusOK, listing)
}

func (h ListingHandler) Archive(c *gin.Context) {
	var listing models.Listing
	if err := h.db.First(&listing, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	if listing.UserID != middleware.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can archive listing"})
		return
	}
	if listing.Status == "pending_payment" {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "payment is required before publishing"})
		return
	}
	h.db.Model(&listing).Update("status", "hidden")
	c.JSON(http.StatusOK, listing)
}

func (h ListingHandler) Bump(c *gin.Context) {
	var listing models.Listing
	if err := h.db.First(&listing, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	if listing.UserID != middleware.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can bump listing"})
		return
	}
	if listing.Status == "pending_payment" {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "payment is required before publishing"})
		return
	}
	h.db.Model(&listing).Updates(map[string]any{"created_at": gorm.Expr("now()"), "updated_at": gorm.Expr("now()")})
	c.JSON(http.StatusOK, listing)
}

func (h ListingHandler) AddImages(c *gin.Context) {
	limitUploadBody(c)
	var listing models.Listing
	if err := h.db.First(&listing, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	files, err := validateImageUploads(c)
	if err != nil {
		writeUploadError(c, err)
		return
	}
	if err := h.ensureImageCapacity(listing.ID, len(files)); err != nil {
		writeUploadError(c, err)
		return
	}
	if listing.UserID != middleware.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can upload images"})
		return
	}
	if err := h.saveImages(files, listing.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save images"})
		return
	}
	h.db.Preload("Images").First(&listing, "id = ?", listing.ID)
	c.JSON(http.StatusOK, listing.Images)
}

func (h ListingHandler) DeleteImage(c *gin.Context) {
	var image models.ListingImage
	if err := h.db.First(&image, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "image not found"})
		return
	}
	var listing models.Listing
	h.db.First(&listing, "id = ?", image.ListingID)
	if listing.UserID != middleware.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can delete image"})
		return
	}
	h.db.Delete(&image)
	if name := filepath.Base(image.ImageURL); name != "." && name != "/" {
		_ = os.Remove(filepath.Join(h.uploadDir, name))
	}
	c.Status(http.StatusNoContent)
}

func (h ListingHandler) Matches(c *gin.Context) {
	var listing models.Listing
	if err := h.db.Preload("MatchPref").First(&listing, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	pref := listing.MatchPref
	if !pref.ExchangeEnabled {
		c.JSON(http.StatusOK, []models.Listing{})
		return
	}
	var categories []string
	for _, item := range strings.Split(pref.Categories, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			categories = append(categories, trimmed)
		}
	}
	q := h.db.Preload("Images").Preload("User").Preload("BuildCard").Preload("MatchPref").
		Where("id <> ? AND status = ?", listing.ID, "active").
		Where("deal_type <> ?", "продажа")
	if len(categories) > 0 {
		q = q.Where("category IN ?", categories)
	}
	if pref.MinPrice > 0 {
		q = q.Where("price >= ?", pref.MinPrice)
	}
	if pref.MaxPrice > 0 {
		q = q.Where("price <= ?", pref.MaxPrice)
	}
	if pref.SameCityOnly {
		q = q.Where("city = ?", listing.City)
	}
	var matches []models.Listing
	if err := q.Order("created_at desc").Find(&matches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, matches)
}

func (h ListingHandler) SaveMatchPreference(c *gin.Context) {
	var listing models.Listing
	if err := h.db.First(&listing, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "listing not found"})
		return
	}
	if listing.UserID != middleware.CurrentUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "only owner can edit match preferences"})
		return
	}
	var req models.MatchPreference
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ListingID = listing.ID
	h.db.Where("listing_id = ?", listing.ID).Delete(&models.MatchPreference{})
	h.db.Create(&req)
	c.JSON(http.StatusOK, req)
}

func (h ListingHandler) bindListing(c *gin.Context) (models.Listing, bool) {
	var listing models.Listing
	listing.Title = c.PostForm("title")
	listing.Description = c.PostForm("description")
	listing.City = c.PostForm("city")
	listing.Brand = c.PostForm("brand")
	listing.Category = c.PostForm("category")
	listing.BikeType = normalizeBikeType(c.PostForm("bike_type"))
	listing.Condition = c.PostForm("condition")
	listing.FrameSize = firstPostForm(c, "frame_size", "frame_size_text")
	listing.FrameSizeText = firstPostForm(c, "frame_size_text", "frame_size")
	listing.RiderMin = parseFirstInt(c, "rider_height_min", "recommended_height_min")
	listing.RiderMax = parseFirstInt(c, "rider_height_max", "recommended_height_max")
	listing.RecommendedHeightMin = parseFirstInt(c, "recommended_height_min", "rider_height_min")
	listing.RecommendedHeightMax = parseFirstInt(c, "recommended_height_max", "rider_height_max")
	listing.DealType = c.DefaultPostForm("deal_type", "продажа")
	listing.Labels = parseListingLabels(c)
	listing.IsUrgent = parseBool(c.PostForm("is_urgent"))
	listing.IsBargain = parseBool(c.PostForm("is_bargain"))
	listing.IsExchange = parseBool(c.PostForm("is_exchange"))
	listing.ExtraPaymentFromMe = parseBool(c.PostForm("extra_payment_from_me"))
	listing.ExtraPaymentFromBuyer = parseBool(c.PostForm("extra_payment_from_buyer"))
	listing.Status = c.DefaultPostForm("status", "active")
	if _, err := fmt.Sscanf(c.PostForm("price"), "%d", &listing.Price); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price is required"})
		return listing, false
	}
	if listing.Title == "" || listing.Description == "" || listing.City == "" || listing.Category == "" || listing.Condition == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required listing fields"})
		return listing, false
	}
	if listing.RiderMin > 0 && listing.RiderMax > 0 && listing.RiderMin > listing.RiderMax {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recommended height min must be less than max"})
		return listing, false
	}
	return listing, true
}

func (h ListingHandler) saveImages(files []*multipart.FileHeader, listingID string) error {
	for i, file := range files {
		ext, err := imageExtension(file)
		if err != nil {
			return err
		}
		suffix, err := randomHex(12)
		if err != nil {
			return err
		}
		name := fmt.Sprintf("%s-%s%s", listingID, suffix, ext)
		path := filepath.Join(h.uploadDir, name)
		if err := saveUploadedFile(file, path); err != nil {
			return err
		}
		if err := h.db.Create(&models.ListingImage{ListingID: listingID, ImageURL: "/uploads/" + name, SortOrder: i}).Error; err != nil {
			_ = os.Remove(path)
			return err
		}
	}
	return nil
}

func (h ListingHandler) ensureImageCapacity(listingID string, additional int) error {
	if additional == 0 {
		return nil
	}
	var count int64
	if err := h.db.Model(&models.ListingImage{}).Where("listing_id = ?", listingID).Count(&count).Error; err != nil {
		return err
	}
	if count+int64(additional) > maxImagesPerListing {
		return fmt.Errorf("a listing may contain at most %d images", maxImagesPerListing)
	}
	return nil
}

func limitUploadBody(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadRequestBytes)
}

func validateImageUploads(c *gin.Context) ([]*multipart.FileHeader, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("invalid multipart request: %w", err)
	}
	files := form.File["images"]
	if len(files) > maxImagesPerRequest {
		return nil, fmt.Errorf("a maximum of %d images is allowed", maxImagesPerRequest)
	}
	for _, file := range files {
		if file.Size > maxImageBytes {
			return nil, errImageTooLarge
		}
		if _, err := imageExtension(file); err != nil {
			return nil, err
		}
	}
	return files, nil
}

func imageExtension(file *multipart.FileHeader) (string, error) {
	source, err := file.Open()
	if err != nil {
		return "", err
	}
	defer source.Close()
	buffer := make([]byte, 512)
	n, err := source.Read(buffer)
	if err != nil && n == 0 {
		return "", errors.New("empty image file")
	}
	switch http.DetectContentType(buffer[:n]) {
	case "image/jpeg":
		return ".jpg", nil
	case "image/png":
		return ".png", nil
	case "image/webp":
		return ".webp", nil
	default:
		return "", errors.New("only JPEG, PNG, and WebP images are allowed")
	}
}

func saveUploadedFile(file *multipart.FileHeader, destination string) error {
	source, err := file.Open()
	if err != nil {
		return err
	}
	defer source.Close()
	target, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer target.Close()
	_, err = io.Copy(target, source)
	if err != nil {
		_ = os.Remove(destination)
	}
	return err
}

func randomHex(bytes int) (string, error) {
	buffer := make([]byte, bytes)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func writeUploadError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, errImageTooLarge) || strings.Contains(err.Error(), "request body too large") {
		status = http.StatusRequestEntityTooLarge
	}
	c.JSON(status, gin.H{"error": err.Error()})
}

func (h ListingHandler) saveBuildCard(c *gin.Context, listingID string) {
	card := models.BuildCard{
		ListingID: listingID,
		Frame:     c.PostForm("build_frame"), Size: c.PostForm("build_size"), Fork: c.PostForm("build_fork"),
		Wheels: c.PostForm("build_wheels"), Hubs: c.PostForm("build_hubs"), Tires: c.PostForm("build_tires"),
		Handlebar: c.PostForm("build_handlebar"), Stem: c.PostForm("build_stem"), Saddle: c.PostForm("build_saddle"),
		Cranks: c.PostForm("build_cranks"), BottomBracket: c.PostForm("build_bottom_bracket"), Chain: c.PostForm("build_chain"),
		Cog: c.PostForm("build_cog"), Brakes: c.PostForm("build_brakes"), Weight: c.PostForm("build_weight"),
		FrameCondition: c.PostForm("build_frame_condition"), Defects: c.PostForm("build_defects"),
		Documents: parseBool(c.PostForm("build_documents")),
	}
	if card.Frame == "" && card.Wheels == "" && card.Fork == "" && card.Cranks == "" {
		return
	}
	h.db.Where("listing_id = ?", listingID).Delete(&models.BuildCard{})
	h.db.Create(&card)
}

func (h ListingHandler) saveMatchPreference(c *gin.Context, listingID string) {
	if c.PostForm("exchange_enabled") == "" {
		return
	}
	pref := models.MatchPreference{
		ListingID: listingID, ExchangeEnabled: parseBool(c.PostForm("exchange_enabled")),
		Wants: c.PostForm("match_wants"), Categories: c.PostForm("match_categories"),
		MinPrice: parseInt(c.PostForm("match_min_price")), MaxPrice: parseInt(c.PostForm("match_max_price")),
		CanAddCash: parseBool(c.PostForm("match_can_add_cash")), MaxCashAdd: parseInt(c.PostForm("match_max_cash_add")),
		SameCityOnly: parseBool(c.PostForm("match_same_city_only")),
	}
	h.db.Where("listing_id = ?", listingID).Delete(&models.MatchPreference{})
	h.db.Create(&pref)
}

func parseBool(v string) bool {
	return v == "true" || v == "1" || v == "on"
}

func parseInt(v string) int {
	i, _ := strconv.Atoi(v)
	return i
}

func parseFirstInt(c *gin.Context, keys ...string) int {
	for _, key := range keys {
		if value := c.PostForm(key); value != "" {
			return parseInt(value)
		}
	}
	return 0
}

func firstPostForm(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(c.PostForm(key)); value != "" {
			return value
		}
	}
	return ""
}

func parseListingLabels(c *gin.Context) []string {
	values := c.PostFormArray("labels")
	if len(values) == 0 {
		values = parseLabelsQuery(c.PostForm("labels"))
	}
	return normalizeListingLabels(values)
}

func parseLabelsQuery(raw string) []string {
	if raw == "" {
		return nil
	}
	var labels []string
	for _, item := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			labels = append(labels, trimmed)
		}
	}
	return labels
}

func parseFrameSizeQuery(c *gin.Context) []string {
	values := c.QueryArray("frame_size")
	if len(values) == 0 {
		values = c.QueryArray("frame_size[]")
	}
	if len(values) == 0 {
		values = parseLabelsQuery(c.Query("frame_size"))
	}
	seen := map[string]bool{}
	var sizes []string
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			size := strings.TrimSpace(part)
			if size == "" || seen[size] {
				continue
			}
			seen[size] = true
			sizes = append(sizes, size)
		}
	}
	return sizes
}

func normalizeBikeType(value string) string {
	switch strings.TrimSpace(value) {
	case "fixed", "road", "mtb", "bmx", "city":
		return strings.TrimSpace(value)
	default:
		return ""
	}
}

func normalizeDealType(value string) string {
	switch strings.TrimSpace(value) {
	case "обмен":
		return "обмен"
	case "продажа или обмен":
		return "продажа или обмен"
	default:
		return "продажа"
	}
}

func normalizeListingLabels(values []string) []string {
	seen := map[string]bool{}
	var labels []string
	for _, value := range values {
		label := strings.TrimSpace(value)
		if !isAllowedListingLabel(label) || seen[label] {
			continue
		}
		seen[label] = true
		labels = append(labels, label)
	}
	if labels == nil {
		return []string{}
	}
	return labels
}

func isAllowedListingLabel(label string) bool {
	switch label {
	case "срочно", "торг", "обмен", "с моей доплатой", "с вашей доплатой":
		return true
	default:
		return false
	}
}

func normalizeListingContract(listing *models.Listing) {
	if listing.FrameSizeText == "" {
		listing.FrameSizeText = listing.FrameSize
	}
	if listing.FrameSize == "" {
		listing.FrameSize = listing.FrameSizeText
	}
	if listing.RecommendedHeightMin == 0 {
		listing.RecommendedHeightMin = listing.RiderMin
	}
	if listing.RecommendedHeightMax == 0 {
		listing.RecommendedHeightMax = listing.RiderMax
	}
	if listing.RiderMin == 0 {
		listing.RiderMin = listing.RecommendedHeightMin
	}
	if listing.RiderMax == 0 {
		listing.RiderMax = listing.RecommendedHeightMax
	}
	if listing.DealType == "обмен" || listing.DealType == "продажа или обмен" {
		listing.IsExchange = true
	}
	listing.Labels = normalizeListingLabels(append(listing.Labels, listingLabelsFromBooleans(*listing)...))
}

func hydrateListingContract(listing *models.Listing) {
	normalizeListingContract(listing)
	for _, label := range listing.Labels {
		switch label {
		case "срочно":
			listing.IsUrgent = true
		case "торг":
			listing.IsBargain = true
		case "обмен":
			listing.IsExchange = true
		case "с моей доплатой":
			listing.ExtraPaymentFromMe = true
		case "с вашей доплатой":
			listing.ExtraPaymentFromBuyer = true
		}
	}
}

func listingLabelsFromBooleans(listing models.Listing) []string {
	var labels []string
	if listing.IsUrgent {
		labels = append(labels, "срочно")
	}
	if listing.IsBargain {
		labels = append(labels, "торг")
	}
	if listing.IsExchange {
		labels = append(labels, "обмен")
	}
	if listing.ExtraPaymentFromMe {
		labels = append(labels, "с моей доплатой")
	}
	if listing.ExtraPaymentFromBuyer {
		labels = append(labels, "с вашей доплатой")
	}
	return labels
}

func applyLabelFilter(q *gorm.DB, label string) *gorm.DB {
	switch label {
	case "срочно":
		return q.Where("(is_urgent = ? OR labels @> ?::jsonb)", true, mustJSONLabel(label))
	case "торг":
		return q.Where("(is_bargain = ? OR labels @> ?::jsonb)", true, mustJSONLabel(label))
	case "обмен":
		return q.Where("(is_exchange = ? OR labels @> ?::jsonb)", true, mustJSONLabel(label))
	case "с моей доплатой":
		return q.Where("(extra_payment_from_me = ? OR labels @> ?::jsonb)", true, mustJSONLabel(label))
	case "с вашей доплатой":
		return q.Where("(extra_payment_from_buyer = ? OR labels @> ?::jsonb)", true, mustJSONLabel(label))
	default:
		return q
	}
}

func mustJSONLabel(label string) string {
	encoded, _ := json.Marshal([]string{label})
	return string(encoded)
}
