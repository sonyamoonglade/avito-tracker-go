package repositories

import "parser/internal/postgres"

type Repositories struct {
	AdvertRepo     AdvertRepository
	SubscriberRepo SubscriberRepository
}

func NewRepositories(pg *postgres.Postgres) *Repositories {

	advertRepo := NewAdvertRepo(pg)
	subscriberRepo := NewSubscriberRepo(pg)

	return &Repositories{
		AdvertRepo:     advertRepo,
		SubscriberRepo: subscriberRepo,
	}
}
