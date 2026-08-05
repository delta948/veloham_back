package moderation

import "veloham/backend/internal/handlers"

type Handler struct {
	handlers.ReportAdminHandler
}

func NewHandler(service *Service) Handler {
	return Handler{ReportAdminHandler: handlers.NewReportAdminHandler(service.db)}
}
