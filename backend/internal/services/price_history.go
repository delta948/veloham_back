package services

import (
	"errors"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"veloham/backend/internal/models"
)

const MaxListingPrice = 2_000_000_000

var (
	ErrPriceRateLimited = errors.New("Вы недавно изменяли цену. Повторите попытку позже")
	ErrInvalidPrice     = errors.New("цена должна быть целым числом от 1 до 2 000 000 000 сом")
	ErrListingForbidden = errors.New("изменять цену может только владелец объявления или администратор")
)

type PriceChangeResult struct {
	Listing models.Listing              `json:"listing"`
	Change  *models.ListingPriceHistory `json:"price_change,omitempty"`
}

type PriceHistoryItem struct {
	OldPrice      int       `json:"old_price"`
	NewPrice      int       `json:"new_price"`
	ChangeAmount  int       `json:"change_amount"`
	ChangePercent float64   `json:"change_percent"`
	ChangedAt     time.Time `json:"changed_at"`
}

type PriceHistoryResponse struct {
	ListingID          string             `json:"listing_id"`
	InitialPrice       int                `json:"initial_price"`
	CurrentPrice       int                `json:"current_price"`
	MinimumPrice       int                `json:"minimum_price"`
	MaximumPrice       int                `json:"maximum_price"`
	MinimumPrice30Days int                `json:"minimum_price_30_days"`
	TotalChange        int                `json:"total_change"`
	TotalChangePercent float64            `json:"total_change_percent"`
	History            []PriceHistoryItem `json:"history"`
}

type PriceHistoryService struct{ DB *gorm.DB }

func percent(change, base int) float64 {
	if base == 0 {
		return 0
	}
	return math.Round((float64(change)/float64(base))*10000) / 100
}

func (s PriceHistoryService) ChangePrice(listingID, actorID string, isAdmin bool, newPrice int, ip string) (PriceChangeResult, error) {
	if newPrice <= 0 || newPrice > MaxListingPrice {
		return PriceChangeResult{}, ErrInvalidPrice
	}
	var result PriceChangeResult
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var listing models.Listing
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&listing, "id = ?", listingID).Error; err != nil {
			return err
		}
		if listing.UserID != actorID && !isAdmin {
			return ErrListingForbidden
		}
		if listing.Price == newPrice {
			result.Listing = listing
			return nil
		}

		var last models.ListingPriceHistory
		if err := tx.Where("listing_id = ?", listingID).Order("changed_at desc").First(&last).Error; err == nil && time.Since(last.ChangedAt) < 10*time.Minute {
			return ErrPriceRateLimited
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		suspicious, reason := detectSuspicious(tx, listingID, listing.Price, newPrice)
		change := models.ListingPriceHistory{ListingID: listing.ID, OldPrice: listing.Price, NewPrice: newPrice, ChangedBy: actorID, IPAddress: ip, Suspicious: suspicious, SuspiciousReason: reason}
		if err := tx.Create(&change).Error; err != nil {
			return err
		}
		if err := tx.Model(&listing).Updates(map[string]any{"price": newPrice, "updated_at": gorm.Expr("now()")}).Error; err != nil {
			return err
		}

		if newPrice < listing.Price {
			message := fmt.Sprintf("Цена на объявление “%s” снизилась с %s до %s сом", listing.Title, groupNumber(listing.Price), groupNumber(newPrice))
			if err := tx.Exec(`INSERT INTO notifications (user_id, listing_id, price_history_id, type, message, link)
				SELECT user_id, ?, ?, 'price_drop', ?, ? FROM favorites WHERE listing_id = ? AND user_id <> ?
				ON CONFLICT (user_id, price_history_id) DO NOTHING`, listing.ID, change.ID, message, "/listing/"+listing.ID, listing.ID, listing.UserID).Error; err != nil {
				return err
			}
		}
		listing.Price = newPrice
		result = PriceChangeResult{Listing: listing, Change: &change}
		return nil
	})
	return result, err
}

func detectSuspicious(tx *gorm.DB, listingID string, oldPrice, newPrice int) (bool, string) {
	if newPrice > oldPrice && percent(newPrice-oldPrice, oldPrice) > 30 {
		return true, "повышение более чем на 30%"
	}
	var last models.ListingPriceHistory
	if newPrice < oldPrice && tx.Where("listing_id = ? AND new_price > old_price", listingID).Order("changed_at desc").First(&last).Error == nil && time.Since(last.ChangedAt) < 24*time.Hour {
		return true, "снижение вскоре после повышения"
	}
	var count int64
	tx.Model(&models.ListingPriceHistory{}).Where("listing_id = ? AND changed_at > ?", listingID, time.Now().Add(-time.Hour)).Count(&count)
	if count >= 3 {
		return true, "много изменений за короткий период"
	}
	return false, ""
}

func (s PriceHistoryService) Get(listingID string) (PriceHistoryResponse, error) {
	var listing models.Listing
	if err := s.DB.First(&listing, "id = ?", listingID).Error; err != nil {
		return PriceHistoryResponse{}, err
	}
	var rows []models.ListingPriceHistory
	if err := s.DB.Where("listing_id = ?", listingID).Order("changed_at asc").Find(&rows).Error; err != nil {
		return PriceHistoryResponse{}, err
	}
	minPrice, maxPrice, min30 := listing.InitialPrice, listing.InitialPrice, listing.Price
	items := make([]PriceHistoryItem, 0, len(rows))
	for _, row := range rows {
		if row.NewPrice < minPrice {
			minPrice = row.NewPrice
		}
		if row.NewPrice > maxPrice {
			maxPrice = row.NewPrice
		}
		if row.ChangedAt.After(time.Now().AddDate(0, 0, -30)) && row.NewPrice < min30 {
			min30 = row.NewPrice
		}
		items = append(items, PriceHistoryItem{row.OldPrice, row.NewPrice, row.NewPrice - row.OldPrice, percent(row.NewPrice-row.OldPrice, row.OldPrice), row.ChangedAt})
	}
	change := listing.Price - listing.InitialPrice
	return PriceHistoryResponse{listing.ID, listing.InitialPrice, listing.Price, minPrice, maxPrice, min30, change, percent(change, listing.InitialPrice), items}, nil
}

func groupNumber(n int) string {
	s := fmt.Sprintf("%d", n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + " " + s[i:]
	}
	return s
}
