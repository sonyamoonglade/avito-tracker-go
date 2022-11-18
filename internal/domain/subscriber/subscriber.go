package subscriber

import "parser/internal/domain/subscription"

type Subscriber struct {
	ID            string
	telegramID    string
	subscriptions []*subscription.Subscription
}

func (s *Subscriber) AddSubscription(subscriptions ...*subscription.Subscription) {
	s.subscriptions = append(s.subscriptions, subscriptions...)
}
