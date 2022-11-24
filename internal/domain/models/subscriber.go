package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNoSubscriptions = errors.New("empty subscriptions")
)

type Subscriber struct {
	SubscriberID string `db:"subscriber_id"`
	TelegramID   int64  `db:"telegram_id"`
	// Readonly to prevent data corruption and bugs
	subscriptions []*Subscription
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

func SubscriberFromTelegramID(telegramID int64) *Subscriber {
	return &Subscriber{
		SubscriberID:  uuid.NewString(),
		TelegramID:    telegramID,
		subscriptions: nil,
	}
}
