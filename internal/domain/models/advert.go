package domain

type Advert struct {
	LastPrice float64 `db:"last_price"`
	URL       string  `db:"url"`
}

func (ad *Advert) DidPriceChange(newPrice float64) bool {
	return ad.LastPrice == newPrice
}

func (ad *Advert) UpdatePrice(newPrice float64) {
	ad.LastPrice = newPrice
}
