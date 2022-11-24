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
	title        string  `db:"title"`
	currentPrice float64 `db:"current_price"`
	lastPrice    float64 `db:"last_price"`
}

func AdvertFromURL(URL string) (*Advert, error) {
	if URL == "" {
		return nil, ErrEmptyURL
	}

	return &Advert{
		AdvertID:     uuid.NewString(),
		URL:          URL,
		title:        "",
		currentPrice: 0.0,
		lastPrice:    0.0,
	}, nil
}

func (ad *Advert) CurrentPrice() float64 {
	return ad.currentPrice
}

func (ad *Advert) LastPrice() float64 {
	return ad.lastPrice
}

func (ad *Advert) Title() string {
	return ad.title
}

func (ad *Advert) DidPriceChange(newPrice float64) bool {
	return ad.currentPrice == newPrice
}

func (ad *Advert) HasTitle() bool {
	return ad.title == ""
}

func (ad *Advert) UpdateTitle(title string) {
	ad.title = title
}

func (ad *Advert) UpdatePrice(price float64) {
	ad.lastPrice = ad.currentPrice
	ad.currentPrice = price
}
