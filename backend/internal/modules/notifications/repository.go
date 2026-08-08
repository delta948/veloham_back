package notifications

import (
	"gorm.io/gorm"
	"veloham/backend/internal/models"
)

type Repository interface {
	List(userID string, limit int) ([]models.Notification, error)
	MarkRead(userID, notificationID string) (bool, error)
	MarkAllRead(userID string) error
}

type repository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return repository{db: db} }

func (r repository) List(userID string, limit int) ([]models.Notification, error) {
	rows := make([]models.Notification, 0)
	err := r.db.Where("user_id = ?", userID).Order("created_at desc").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r repository) MarkRead(userID, notificationID string) (bool, error) {
	result := r.db.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true)
	return result.RowsAffected > 0, result.Error
}

func (r repository) MarkAllRead(userID string) error {
	return r.db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Update("is_read", true).Error
}
