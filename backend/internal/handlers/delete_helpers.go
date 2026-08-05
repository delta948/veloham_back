package handlers

import (
	"gorm.io/gorm"
	"veloham/backend/internal/models"
)

func deleteListingCascade(db *gorm.DB, listingID string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var chats []models.Chat
		if err := tx.Where("listing_id = ?", listingID).Find(&chats).Error; err != nil {
			return err
		}
		for _, chat := range chats {
			if err := tx.Where("chat_id = ?", chat.ID).Delete(&models.Message{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("listing_id = ?", listingID).Delete(&models.Chat{}).Error; err != nil {
			return err
		}
		if err := tx.Where("listing_id = ?", listingID).Delete(&models.ListingImage{}).Error; err != nil {
			return err
		}
		if err := tx.Where("listing_id = ?", listingID).Delete(&models.Favorite{}).Error; err != nil {
			return err
		}
		if err := tx.Where("listing_id = ?", listingID).Delete(&models.BuildCard{}).Error; err != nil {
			return err
		}
		if err := tx.Where("listing_id = ?", listingID).Delete(&models.MatchPreference{}).Error; err != nil {
			return err
		}
		if err := tx.Where("listing_id = ?", listingID).Delete(&models.WantedOffer{}).Error; err != nil {
			return err
		}
		if err := tx.Where("listing_id = ?", listingID).Delete(&models.Review{}).Error; err != nil {
			return err
		}
		if err := tx.Where("listing_id = ?", listingID).Delete(&models.Report{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.Listing{}, "id = ?", listingID).Error
	})
}
