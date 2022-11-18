package subscription

import "net/http"

type Handler struct {
	subscriptionService *Service // TODO: replace with interface
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		subscriptionService: service,
	}
}

func (h *Handler) NewSubscription(w http.ResponseWriter, r *http.Request) {

}
