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
	"time"
)

type SubscriptionService interface {
	NewSubscription(ctx context.Context, dto *dto.SubscribeRequest) error

	NotifySubscribers(ctx context.Context, ad *domain.Advert) error
	HandleParsingResult(result *parser.ParseResult) error

	GetAllURLs(ctx context.Context) ([]string, error)
}

type subscriptionService struct {
	subscriptionRepo repositories.SubscriberRepository
	advertRepo       repositories.AdvertRepository
	notifier         notify.Notifier
	ringParser       *parser.RingParser
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

func (s *subscriptionService) NotifySubscribers(ctx context.Context, ad *domain.Advert) error {
	subscribers, err := s.subscriptionRepo.GetAdvertSubscribers(ctx, ad.URL)
	if err != nil {
		// TODO: wrap to internal
		return errors.WrapInternal(err, "subscriptionService.NotifySubscribers.GetAdvertSubscribers")
	}

	for _, subscriber := range subscribers {

		// hardcoded for now
		msg := fmt.Sprintf("Hey!\n%s is updated!\nNew price: %.2f\nPrev price: %.2f\n", ad.Title(), ad.CurrentPrice(), ad.LastPrice())

		// Notify actually
		// Imagine we've straightforwardly chosen telegram notifications
		// Otherwise we'd need to get user's wanted notification provider
		// and match arguments to specific notifier... see Notifier args...
		err := s.notifier.Notify(ad, subscriber.TelegramID, msg)
		if err != nil {
			// TODO: maybe some queue??
			return errors.WrapInternal(err, "subscriptionService.NotifySubscribers.Notify")
		}

	}

	return nil
}

func (s *subscriptionService) HandleParsingResult(result *parser.ParseResult) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	advert, err := s.advertRepo.GetByURL(ctx, result.URL())
	if err != nil {
		return errors.WrapInternal(err, "subscriptionService.HandleParsingResult.GetByURL")
	}

	hasTitle := advert.HasTitle()
	if !hasTitle {
		advert.UpdateTitle(result.Title())
	}

	didChange := advert.DidPriceChange(result.Price())
	if didChange {
		advert.UpdatePrice(result.Price())
	}

	// Update an advert
	err = s.advertRepo.Update(ctx, advert)
	if err != nil {
		return errors.WrapInternal(err, "subscriptionService.HandleParsingResult.Update")
	}

	// Notify
	err = s.NotifySubscribers(context.Background(), advert)
	if err != nil {
		return errors.WrapInternal(err, "subscriptionService.HandleParsingResult.NotifySubscribers")
	}

	return nil
}

// Looks for adverts that users are subscribed to and returns urls
func (s *subscriptionService) GetAllURLs(ctx context.Context) ([]string, error) {
	urls, err := s.subscriptionRepo.GetAllURLs(ctx)
	if err != nil {
		return nil, errors.WrapInternal(err, "subscriptionService.GetAllURLs.GetAllURLs")
	}

	return urls, nil
}
