package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNoSubscriptions = errors.New("empty subscriptions")
)

type Subscriber struct {
	SubscriberID string
	telegramID   int64

	subscriptions []*Subscription
}

func NewSubscriber(id string, telegramID int64) *Subscriber {
	return &Subscriber{SubscriberID: id, telegramID: telegramID}
}

func SubscriberFromTelegramID(telegramID int64) *Subscriber {
	return &Subscriber{
		SubscriberID:  uuid.NewString(),
		telegramID:    telegramID,
		subscriptions: nil,
	}
}

func (s *Subscriber) TelegramID() int64 {
	return s.telegramID
}

func (s *Subscriber) AddSubscription(subscriptions ...*Subscription) {
	s.subscriptions = append(s.subscriptions, subscriptions...)
}

func (s *Subscriber) Subscriptions() []*Subscription {
	return s.subscriptions
}

func (s *Subscriber) HasSubscriptions() bool {
	return s.subscriptions != nil
}
