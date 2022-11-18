package subscriber

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

func (s *Service) Insert(ctx context.Context, sub *Subscriber) error {

	sub.ID = uuid.NewString()

	sql, args, err := sq.Insert("subscriptions").
		Columns("id", "telegram_id").
		Values(sub.ID, sub.telegramID).
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
