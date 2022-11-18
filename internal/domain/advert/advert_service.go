package advert

import (
	"context"
	"fmt"
	"parser/pkg/postgres"

	sq "github.com/Masterminds/squirrel"
)

type Service struct {
	db *postgres.Postgres
}

func NewService(db *postgres.Postgres) *Service {
	return &Service{db: db}
}

func (s *Service) Insert(ctx context.Context, ad *Advert) error {
	sql, args, err := sq.Insert("adverts").
		Columns("last_price", "url").
		Values(ad.LastPrice, ad.URL).
		ToSql()
	fmt.Println(sql)

	if err != nil {
		return err
	}

	_, release, err := s.db.Exec(ctx, sql, args)
	if err != nil {
		return fmt.Errorf("internal error: %w", err)
	}

	defer release()

	return nil
}

func (s *Service) Update(ctx context.Context, URL string, newPrice float64) error {
	sql, args, err := sq.Update("adverts").
		Set("last_price", newPrice).
		Where("url = $1", URL).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	fmt.Println(sql)
	if err != nil {
		return err
	}

	_, release, err := s.db.Exec(ctx, sql, args)
	if err != nil {
		return fmt.Errorf("internal error: %w", err)
	}
	defer release()

	return nil
}
