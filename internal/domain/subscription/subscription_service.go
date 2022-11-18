package subscription

import (
	"context"
	"fmt"
	"parser/pkg/postgres"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type Service struct {
	db *postgres.Postgres
}

func NewService(db *postgres.Postgres) *Service {
	return &Service{db: db}
}

func (s *Service) Insert(ctx context.Context, sub *Subscription) error {

	sub.ID = uuid.NewString()

	sql, args, err := sq.Insert("subscriptions").
		Columns("id", "advert_id", "subscriber_id").
		Values(sub.ID, sub.AdvertID, sub.SubscriberID).
		ToSql()

	if err != nil {
		return err
	}
	_, release, err := s.db.Query(ctx, sql, args)
	if err != nil {
		return fmt.Errorf("internal error: %w", err)
	}
	defer release()

	return nil
}

func (s *Service) GetOne(ctx context.Context, ID string) (*Subscription, error) {

	sql, args, err := sq.Select("*").
		From("subscriptions").
		Where("id = $1", ID).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return nil, err
	}

	rows, release, err := s.db.Query(ctx, sql, args)
	if err != nil {
		return nil, fmt.Errorf("internal error: %w", err)
	}
	defer release()

	var sub Subscription

	err = s.db.ScanOne(rows, &sub)
	if err != nil {
		return nil, fmt.Errorf("scanning error: %w", err)
	}

	return &sub, nil
}
