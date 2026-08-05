package wanted

import "veloham/backend/internal/handlers"

type Handler struct {
	handlers.WantedHandler
}

func NewHandler(service *Service) Handler {
	return Handler{WantedHandler: handlers.NewWantedHandler(service.db)}
}
