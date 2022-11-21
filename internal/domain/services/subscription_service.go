package services

import (
	"context"
	"fmt"
	domain "parser/internal/domain/models"
	"parser/internal/domain/repositories"
	"parser/internal/notify"
)

type SubscriptionService interface {
	CompareWithLatestPrice(ctx context.Context, newPrice float64, url string) (bool, error)
	NotifySubscribers(ctx context.Context, target *domain.Advert) error
}

type subscriptionService struct {
	subscriptionRepo repositories.SubscriberRepository
	advertRepo       repositories.AdvertRepository
	notifier         notify.Notifier
}

func NewSubscriptionService(
	subscriptionRepo repositories.SubscriberRepository,
	advertRepo repositories.AdvertRepository,
	notifier notify.Notifier) SubscriptionService {
	return &subscriptionService{
		subscriptionRepo: subscriptionRepo,
		advertRepo:       advertRepo,
		notifier:         notifier,
	}
}

func (s *subscriptionService) NotifySubscribers(ctx context.Context, target *domain.Advert) error {
	subscribers, err := s.subscriptionRepo.GetAdvertSubscribers(ctx, target.URL)
	if err != nil {
		// TODO: wrap to internal
		return err
	}

	for _, subscriber := range subscribers {

		// hardcoded for now
		msg := fmt.Sprintf("Hey!\n%s is updated!\nNew price: %.2f", target.Title, target.LastPrice)

		// Notify actually
		// Imagine we've straightforwardly chosen telegram notifications
		// Otherwise we'd need to get user's wanted notification provider
		// and match arguments to specific notifier... see Notifier args...
		err := s.notifier.Notify(target, subscriber.TelegramID, msg)
		if err != nil {
			// TODO: maybe some queue??
			return fmt.Errorf("internal error: %w", err)
		}

	}

	return nil
}

func (s *subscriptionService) CompareWithLatestPrice(ctx context.Context, newPrice float64, url string) (bool, error) {
	advert, err := s.advertRepo.GetByURL(ctx, url)
	if err != nil {
		return false, err
	}

	return advert.DidPriceChange(newPrice), nil
}

func (s *subscriptionService) UpdateAdvertPrice(ctx context.Context, URL string, newPrice float64) error {
	return s.advertRepo.Update(ctx, URL, newPrice)
}
