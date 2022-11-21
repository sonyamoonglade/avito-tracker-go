package domain

type Subscriber struct {
	SubscriberID string `db:"subscriber_id"`
	TelegramID   string `db:"telegram_id"`
}
