package postgres

import (
	domain "parser/internal/domain/models"

	"github.com/google/uuid"
)

type AdvertDB struct {
	AdvertID     string  `db:"advert_id"`
	URL          string  `db:"url"`
	Title        string  `db:"title"`
	CurrentPrice float64 `db:"current_price"`
	LastPrice    float64 `db:"last_price"`
	IsParsed     bool    `db:"is_parsed"`
}

func (adb *AdvertDB) ToDomain() *domain.Advert {
	return domain.NewAdvert(adb.AdvertID, adb.URL, adb.Title, adb.CurrentPrice, adb.LastPrice, adb.IsParsed)
}

type SubscriberDB struct {
	SubscriberID uuid.UUID `db:"subscriber_id"`
	TelegramID   int64     `db:"telegram_id"`
}

func (sdb *SubscriberDB) ToDomain() *domain.Subscriber {
	return domain.NewSubscriber(sdb.SubscriberID.String(), sdb.TelegramID)
}
