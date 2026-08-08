package notifications

import "veloham/backend/internal/models"

const notificationListLimit = 100

type Service interface {
	List(userID string) ([]models.Notification, error)
	MarkRead(userID, notificationID string) (bool, error)
	MarkAllRead(userID string) error
}

type service struct{ repository Repository }

func NewService(repository Repository) Service { return service{repository: repository} }

func (s service) List(userID string) ([]models.Notification, error) {
	return s.repository.List(userID, notificationListLimit)
}

func (s service) MarkRead(userID, notificationID string) (bool, error) {
	return s.repository.MarkRead(userID, notificationID)
}

func (s service) MarkAllRead(userID string) error { return s.repository.MarkAllRead(userID) }
