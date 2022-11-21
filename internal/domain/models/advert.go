package domain

type Advert struct {
	Title        string  `db:"title"`
	CurrentPrice float64 `db:"current_price"`
	LastPrice    float64 `db:"last_price"`
	URL          string  `db:"url"`
}

func (ad *Advert) DidPriceChange(newPrice float64) bool {
	return ad.CurrentPrice == newPrice
}

func (ad *Advert) UpdatePrice(newPrice float64) {
	ad.CurrentPrice = newPrice
}
