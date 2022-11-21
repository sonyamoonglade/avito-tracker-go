package services

import (
	"context"
	domain "parser/internal/domain/models"
	"parser/internal/domain/repositories"
)

type SubscriptionService interface {
	CompareWithLatestPrice(ctx context.Context, newPrice float64, url string) (bool, error)
	NotifySubscribers(ctx context.Context, target *domain.Advert) error
}

type subscriptionService struct {
	subscriptionRepo repositories.SubscriberRepository
	advertRepo       repositories.AdvertRepository
}

func (s *subscriptionService) NotifySubscribers(ctx context.Context, target *domain.Advert) error {
	subscribers, err := s.subscriptionRepo.GetAdvertSubscribers(ctx, target.URL)
	if err != nil {
		// wrap
		return err
	}

	for _, subscriber := range subscribers {
		// Notify actually

	}
}

func (s *subscriptionService) CompareWithLatestPrice(ctx context.Context, newPrice float64, url string) (bool, error) {
	advert, err := s.advertRepo.GetByURL(ctx, url)
	if err != nil {
		return false, err
	}

	didChange := advert.DidPriceChange(newPrice)

	return didChange, nil
}

func NewSubscriptionService(
	subscriptionRepo repositories.SubscriberRepository,
	advertRepo repositories.AdvertRepository) SubscriptionService {
	return &subscriptionService{
		subscriptionRepo: subscriptionRepo,
		advertRepo:       advertRepo,
	}
}
