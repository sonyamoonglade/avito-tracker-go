package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrEmptyURL = errors.New("url must not be empty")
)

type Advert struct {
	AdvertID     string  `db:"advert_id"`
	URL          string  `db:"url"`
	Title        string  `db:"title"`
	CurrentPrice float64 `db:"current_price"`
	LastPrice    float64 `db:"last_price"`
}

func AdvertFromURL(URL string) (*Advert, error) {
	if URL == "" {
		return nil, ErrEmptyURL
	}

	return &Advert{
		AdvertID:     uuid.NewString(),
		URL:          URL,
		Title:        "",
		CurrentPrice: 0.0,
		LastPrice:    0.0,
	}, nil
}

func (ad *Advert) DidPriceChange(newPrice float64) bool {
	return ad.CurrentPrice == newPrice
}

func (ad *Advert) UpdatePrice(newPrice float64) {
	ad.CurrentPrice = newPrice
}
