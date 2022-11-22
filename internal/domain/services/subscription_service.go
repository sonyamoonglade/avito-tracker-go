package services

import (
	"context"
	"fmt"
	domain "parser/internal/domain/models"
	"parser/internal/domain/repositories"
	"parser/internal/errors"
	"parser/internal/http/dto"
	"parser/internal/notify"
)

type SubscriptionService interface {
	NewSubscription(ctx context.Context, dto *dto.SubscribeRequest) error
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

func (s *subscriptionService) NewSubscription(ctx context.Context, dto *dto.SubscribeRequest) error {

	var err error

	emptyAdvert, err := domain.AdvertFromURL(dto.AdvertURL)
	if err != nil {
		return errors.WrapDomain(err)
	}

	// TODO: parallel
	err = s.advertRepo.Insert(ctx, emptyAdvert)
	if err != nil {
		return errors.WrapInternal(err, "subscriptionService.NewSubscription.Insert")
	}

	//TODO: parallel
	subscriber := domain.SubscriberFromTelegramID(dto.TelegramID)
	err = s.subscriptionRepo.InsertSubscriber(ctx, subscriber)
	if err != nil {
		return errors.WrapInternal(err, "subscriptionService.NewSubscription.InsertSubscriber")
	}

	//TODO: parallel
	subscription := domain.NewSubscription(subscriber.SubscriberID, emptyAdvert.AdvertID)
	err = s.subscriptionRepo.InsertSubscription(ctx, subscription)
	if err != nil {
		return errors.WrapInternal(err, "subscriptionService.NewSubscription.InsertSubscription")
	}

	return nil
}

func (s *subscriptionService) NotifySubscribers(ctx context.Context, target *domain.Advert) error {
	subscribers, err := s.subscriptionRepo.GetAdvertSubscribers(ctx, target.URL)
	if err != nil {
		// TODO: wrap to internal
		return errors.WrapInternal(err, "subscriptionService.NotifySubscribers.GetAdvertSubscribers")
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
			return errors.WrapInternal(err, "subscriptionService.NotifySubscribers.Notify")
		}

	}

	return nil
}

func (s *subscriptionService) CompareWithLatestPrice(ctx context.Context, newPrice float64, url string) (bool, error) {
	advert, err := s.advertRepo.GetByURL(ctx, url)
	if err != nil {
		return false, errors.WrapInternal(err, "subscriptionService.CompareWithLatestPrice.GetByURL")
	}

	return advert.DidPriceChange(newPrice), nil
}

func (s *subscriptionService) UpdateAdvertPrice(ctx context.Context, URL string, newPrice float64) error {
	err := s.advertRepo.Update(ctx, URL, newPrice)
	if err != nil {
		return errors.WrapInternal(err, "subscriptionService.UpdateAdvertPrice.Update")
	}

	return nil
}
