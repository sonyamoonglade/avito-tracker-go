package services

import (
	"parser/internal/domain/repositories"
	"parser/internal/notify"
)

type Services struct {
	SubscriptionService SubscriptionService
}

func NewServices(repos *repositories.Repositories, notifier notify.Notifier) *Services {

	subscriptionService := NewSubscriptionService(repos.SubscriberRepo, repos.AdvertRepo, notifier)

	return &Services{SubscriptionService: subscriptionService}

}
