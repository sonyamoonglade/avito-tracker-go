package repositories

import (
	"context"
	"fmt"
	domain "parser/internal/domain/models"
	"parser/pkg/postgres"

	sq "github.com/Masterminds/squirrel"
)

type AdvertRepository interface {
	Insert(ctx context.Context, ad *domain.Advert) error
	Update(ctx context.Context, URL string, newPrice float64) error
	GetByURL(ctx context.Context, url string) (*domain.Advert, error)
}

type advertRepo struct {
	db *postgres.Postgres
}

func NewAdvertRepo(db *postgres.Postgres) AdvertRepository {
	return &advertRepo{db: db}
}

func (s *advertRepo) GetByURL(ctx context.Context, url string) (*domain.Advert, error) {
	sql, args, err := sq.Select("*").
		From("adverts").
		Where("url = $1", url).
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

	var ad domain.Advert

	err = s.db.ScanOne(rows, &ad)
	if err != nil {
		return nil, fmt.Errorf("internal error: %w", err)
	}

	return &ad, nil
}

func (s *advertRepo) Insert(ctx context.Context, ad *domain.Advert) error {
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

func (s *advertRepo) Update(ctx context.Context, URL string, newPrice float64) error {
	sql, args, err := sq.Update("adverts").
		Set("last_price", newPrice).
		Where("url = $1", URL).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return err
	}

	fmt.Println(sql)

	_, release, err := s.db.Exec(ctx, sql, args)
	if err != nil {
		return fmt.Errorf("internal error: %w", err)
	}

	defer release()

	return nil
}
