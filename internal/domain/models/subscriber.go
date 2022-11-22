package domain

import "github.com/google/uuid"

type Subscriber struct {
	SubscriberID string `db:"subscriber_id"`
	TelegramID   int64  `db:"telegram_id"`
}

func SubscriberFromTelegramID(telegramID int64) *Subscriber {
	return &Subscriber{
		SubscriberID: uuid.NewString(),
		TelegramID:   telegramID,
	}
}
