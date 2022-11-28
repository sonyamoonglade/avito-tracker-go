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

type UpdateHandler func(result *parser.ParseResult) error

type SubscriptionService interface {
	NewSubscription(ctx context.Context, dto *dto.SubscribeRequest) error

	NotifySubscribers(ctx context.Context, ad *domain.Advert) error

	GetUpdateHandler() UpdateHandler

	GetURLFetcher() func(ctx context.Context) ([]string, error)
}

type subscriptionService struct {
	subscriptionRepo repositories.SubscriberRepository
	advertRepo       repositories.AdvertRepository
	notifier         notify.Notifier
	targetAdder      parser.TargetAdder
}

func NewSubscriptionService(
	subscriptionRepo repositories.SubscriberRepository,
	advertRepo repositories.AdvertRepository,
	notifier notify.Notifier,
	targetAdder parser.TargetAdder) SubscriptionService {
	return &subscriptionService{
		subscriptionRepo: subscriptionRepo,
		advertRepo:       advertRepo,
		notifier:         notifier,
		targetAdder:      targetAdder,
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

		newAdvert, err := domain.NewEmptyAdvert(dto.AdvertURL)
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

	s.targetAdder.AddTarget(advert.URL())
	return nil
}

func (s *subscriptionService) NotifySubscribers(ctx context.Context, ad *domain.Advert) error {
	subscribers, err := s.subscriptionRepo.GetAdvertSubscribers(ctx, ad.AdvertID)
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
		err := s.notifier.Notify(ad, subscriber.TelegramID(), msg)
		if err != nil {
			// TODO: maybe some queue??
			return errors.WrapInternal(err, "subscriptionService.NotifySubscribers.Notify")
		}
	}

	return nil
}

func (s *subscriptionService) GetUpdateHandler() UpdateHandler {
	return s.handleUpdate
}

// Returns helper func to fetch all URLs that subscribers are subscribed to
func (s *subscriptionService) GetURLFetcher() func(ctx context.Context) ([]string, error) {
	return s.getAllURLs
}

func (s *subscriptionService) handleUpdate(update *parser.ParseResult) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	advert, err := s.advertRepo.GetByURL(ctx, update.URL())
	if err != nil {
		return errors.WrapInternal(err, "subscriptionService.handleUpdate.GetByURL")
	}

	// Indicates if title of advert has updated (from empty title to normal)
	var titleChanged bool

	hasTitle := advert.HasTitle()
	if !hasTitle {
		advert.UpdateTitle(update.Title())
		titleChanged = true
	}

	priceChanged := advert.DidPriceChange(update.Price())
	if priceChanged {
		advert.UpdatePrice(update.Price())
	}

	fmt.Printf("update status:\n\tprice-[%t]\t\ntitle-[%t]\n", priceChanged, titleChanged)

	// If nothing has changed - ignore
	if !priceChanged && !titleChanged {
		return nil
	}

	// Important note:
	// If advert is parsed for the first time (lastprice=0, currentprice=0)
	// Program will notify change of the price from 0(currentprice) to parsed(e.g. 800).
	// To avoid that: check if it's parsing for the first time.
	// Only do an update in storage (advertRepo.Update).
	// Do not notify.
	// Duplicate advertRepo.Update is mandatory for calling advert.Parsed()
	if !advert.IsParsed() {
		advert.Parsed()

		err = s.advertRepo.Update(ctx, advert)
		if err != nil {
			return errors.WrapInternal(err, "subscriptionService.handleUpdate.Update")
		}

		return nil
	}

	err = s.advertRepo.Update(ctx, advert)
	if err != nil {
		return errors.WrapInternal(err, "subscriptionService.handleUpdate.Update")
	}

	err = s.NotifySubscribers(context.Background(), advert)
	if err != nil {
		// NotifySubscribers is method that returns an ApplicationError
		// so call errors.ChainInternal it for full errortrace
		return errors.ChainInternal(err, "handleUpdate.NotifySubscribers")
	}

	return nil
}

func (s *subscriptionService) getAllURLs(ctx context.Context) ([]string, error) {
	urls, err := s.subscriptionRepo.GetAllURLs(ctx)
	if err != nil {
		return nil, errors.WrapInternal(err, "subscriptionService.getAllURLs.GetAllURLs")
	}

	return urls, nil
}
