package services

import (
	"parser/internal/domain/repositories"
	"parser/internal/notify"
	"parser/internal/parser"
)

type Services struct {
	SubscriptionService SubscriptionService
}

func NewServices(repos *repositories.Repositories, notifier notify.Notifier, ringParser *parser.RingParser) *Services {

	subscriptionService := NewSubscriptionService(repos.SubscriberRepo, repos.AdvertRepo, notifier, ringParser)

	return &Services{SubscriptionService: subscriptionService}

}
