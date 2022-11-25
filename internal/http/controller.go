package http

import (
	"encoding/json"
	"net/http"
	"parser/internal/http/dto"
)

func (s *HTTPServer) Subscribe(w http.ResponseWriter, r *http.Request) {

	// whats input?
	// TelegramID
	// AdvertURL
	var inp dto.SubscribeRequest
	err := json.NewDecoder(r.Body).Decode(&inp)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	err = s.services.SubscriptionService.NewSubscription(r.Context(), &inp)
	if err != nil {
		// TODO: later add app error handling
		w.Write([]byte(err.Error()))
		return
	}

	w.Write([]byte("yahoo! New subscription is up"))
}
