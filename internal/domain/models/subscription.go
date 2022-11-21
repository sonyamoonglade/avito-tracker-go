package domain

type Subscription struct {
	SubscriberID string `db:"subscriber_id"`
	AdvertID     string `db:"advert_id"`
}
