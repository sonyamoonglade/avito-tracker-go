package domain

import "errors"

var (
	ErrSubscriptionExist = errors.New("subscription already exists")
)

type Subscription struct {
	SubscriberID string `db:"subscriber_id"`
	AdvertID     string `db:"advert_id"`
}

func NewSubscription(subscriberID, advertID string) *Subscription {
	return &Subscription{
		SubscriberID: subscriberID,
		AdvertID:     advertID,
	}
}
