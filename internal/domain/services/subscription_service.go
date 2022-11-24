package services

import (
	"context"
	"fmt"
	domain "parser/internal/domain/models"
	"parser/internal/domain/repositories"
	"parser/internal/errors"
	"parser/internal/http/dto"
	"parser/internal/notify"
	"parser/internal/parser"
)

type SubscriptionService interface {
	NewSubscription(ctx context.Context, dto *dto.SubscribeRequest) error
	CompareWithLatestPrice(ctx context.Context, newPrice float64, url string) (bool, error)
	NotifySubscribers(ctx context.Context, target *domain.Advert) error
}

type subscriptionService struct {
	subscriptionRepo repositories.SubscriberRepository
	advertRepo       repositories.AdvertRepository
	ringParser       *parser.RingParser
	notifier         notify.Notifier
}

func NewSubscriptionService(
	subscriptionRepo repositories.SubscriberRepository,
	advertRepo repositories.AdvertRepository,
	notifier notify.Notifier,
	ringParser *parser.RingParser) SubscriptionService {
	return &subscriptionService{
		subscriptionRepo: subscriptionRepo,
		advertRepo:       advertRepo,
		notifier:         notifier,
		ringParser:       ringParser,
	}
}

func (s *subscriptionService) NewSubscription(ctx context.Context, dto *dto.SubscribeRequest) error {

	// Before heavy buisiness logic perform quick check
	candidateSubscription, err := s.subscriptionRepo.GetSubscription(ctx, dto.TelegramID, dto.AdvertURL)
	if err != nil {
		return errors.WrapInternal(err, "subscriptionService.NewSubscription.GetSubscription")
	}

	fmt.Println(candidateSubscription)

	// Subscription already exists
	if candidateSubscription != nil {
		return domain.ErrSubscriptionExist
	}

	// Try get existing advert
	advert, err := s.advertRepo.GetByURL(ctx, dto.AdvertURL)
	if err != nil {
		return errors.WrapInternal(err, "subscriptionService.NewSubscription.GetByURL")
	}

	// No such advert so create one
	if advert == nil {
		fmt.Println("creating new advert")

		newAdvert, err := domain.AdvertFromURL(dto.AdvertURL)
		if err != nil {
			return errors.WrapDomain(err)
		}

		// TODO: parallel
		err = s.advertRepo.Insert(ctx, newAdvert)
		if err != nil {
			return errors.WrapInternal(err, "subscriptionService.NewSubscription.Insert")
		}

		advert = newAdvert
	}

	// Indicates if subscriber has already existed.
	var isNewSubscriber bool

	subscriber, err := s.subscriptionRepo.GetSubscriber(ctx, dto.TelegramID)
	if err != nil {
		return errors.WrapInternal(err, "subscriptionService.NewSubscription.GetSubscriber")
	}

	// No such subscriber so create one
	if subscriber == nil {
		fmt.Println("creating new subscriber")
		subscriber = domain.SubscriberFromTelegramID(dto.TelegramID)
		isNewSubscriber = true
	}

	subscription := domain.NewSubscription(subscriber.SubscriberID, advert.AdvertID)
	subscriber.AddSubscription(subscription)

	if isNewSubscriber {
		//TODO: parallel
		// Insert subscription (saves subscriptions automatically)
		err = s.subscriptionRepo.InsertSubscriber(ctx, subscriber)
		if err != nil {
			return errors.WrapInternal(err, "subscriptionService.NewSubscription.InsertSubscriber")
		}

	} else {
		//TODO: parallel
		// Just save subscription
		err = s.subscriptionRepo.InsertOnlySubscription(ctx, subscriber)
		if err != nil {
			return errors.WrapInternal(err, "subscriptionService.NewSubscription.InsertOnlySubscription")
		}

	}

	s.ringParser.AddTarget(advert.URL)
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
