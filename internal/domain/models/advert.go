package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrEmptyURL = errors.New("url must not be empty")
)

type Advert struct {
	AdvertID     string
	url          string
	title        string
	currentPrice float64
	lastPrice    float64
	isParsed     bool
}

func NewAdvert(id, url, title string, currentPrice, lastPrice float64, isParsed bool) *Advert {
	return &Advert{
		AdvertID:     id,
		url:          url,
		title:        title,
		currentPrice: currentPrice,
		lastPrice:    lastPrice,
		isParsed:     isParsed,
	}
}

func NewEmptyAdvert(URL string) (*Advert, error) {
	if URL == "" {
		return nil, ErrEmptyURL
	}

	return &Advert{
		AdvertID:     uuid.NewString(),
		url:          URL,
		title:        "",
		currentPrice: 0.0,
		lastPrice:    0.0,
		isParsed:     false,
	}, nil
}

func (ad *Advert) URL() string {
	return ad.url
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

func (ad *Advert) IsParsed() bool {
	return ad.isParsed
}

func (ad *Advert) DidPriceChange(newPrice float64) bool {
	return ad.currentPrice != newPrice
}

func (ad *Advert) HasTitle() bool {
	return ad.title != ""
}

func (ad *Advert) UpdateTitle(title string) {
	ad.title = title
}

// Updates isParsed to TRUE
func (ad *Advert) Parsed() {
	ad.isParsed = true
}

func (ad *Advert) UpdatePrice(price float64) {
	// Advert is just created
	if ad.lastPrice == 0 {
		ad.lastPrice = price
		ad.currentPrice = price
		return
	}

	// General case
	ad.lastPrice = ad.currentPrice
	ad.currentPrice = price
}
