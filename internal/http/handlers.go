package http

import "parser/internal/domain/subscription"

type Handlers struct {
	subscription.Handler
}

func NewHttpHandlers(subscriptionHandler *subscription.Handler) *Handlers {
	return &Handlers{
		*subscriptionHandler,
	}
}
