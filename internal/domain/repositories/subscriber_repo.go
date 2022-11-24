package repositories

import (
	"context"
	"fmt"
	domain "parser/internal/domain/models"
	"parser/internal/postgres"

	sq "github.com/Masterminds/squirrel"
)

type SubscriberRepository interface {
	// Inserts subscriber and underlying subscriptions
	InsertSubscriber(ctx context.Context, sub *domain.Subscriber) error
	// Expects to have at least one subscription
	InsertOnlySubscription(ctx context.Context, sub *domain.Subscriber) error

	GetSubscription(ctx context.Context, subscriberTelegramID int64, advertURL string) (*domain.Subscription, error)

	GetAdvertSubscribers(ctx context.Context, advertID string) ([]*domain.Subscriber, error)
	GetSubscriber(ctx context.Context, telegramID int64) (*domain.Subscriber, error)
}

type subscriberRepo struct {
	db *postgres.Postgres
}

func NewSubscriberRepo(db *postgres.Postgres) SubscriberRepository {
	return &subscriberRepo{db: db}
}

func (s *subscriberRepo) GetSubscription(ctx context.Context, subscriberTelegramID int64, advertURL string) (*domain.Subscription, error) {

	sql, args, err := sq.Select("sp.advert_id, sp.subscriber_id").
		From("subscriptions sp").
		Join("adverts ad on ad.advert_id = sp.advert_id").
		Join("subscribers sub on sp.subscriber_id = sub.subscriber_id").
		Where("sub.telegram_id = $1 and ad.url = $2", subscriberTelegramID, advertURL).
		ToSql()

	fmt.Println(sql)
	if err != nil {
		return nil, err
	}

	rows, release, err := s.db.Query(ctx, sql, args)
	if err != nil {
		return nil, err
	}
	defer release()

	var subscription domain.Subscription
	err = s.db.ScanOne(rows, &subscription)
	if err != nil {
		return nil, postgres.CheckEmptyRows(err)
	}

	return &subscription, nil
}

func (s *subscriberRepo) GetSubscriber(ctx context.Context, telegramID int64) (*domain.Subscriber, error) {

	sql, args, err := sq.Select("*").
		From("subscribers").
		Where("telegram_id = $1", telegramID).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	fmt.Println(sql)

	if err != nil {
		return nil, err
	}

	rows, release, err := s.db.Query(ctx, sql, args)
	if err != nil {
		return nil, err
	}
	defer release()

	var subscriber domain.Subscriber

	err = s.db.ScanOne(rows, &subscriber)
	if err != nil {
		return nil, postgres.CheckEmptyRows(err)
	}

	return &subscriber, nil
}

func (s *subscriberRepo) GetAdvertSubscribers(ctx context.Context, advertID string) ([]*domain.Subscriber, error) {

	sql, args, err := sq.Select("sub.subscriber_id, sub.telegram_id").
		From("subscriptions sp").
		Join("subscribers sub on sub.subscriber_id = sp.subscriber_id").
		Join("adverts ads on sp.advert_id = ads.advert_id").
		Where("ads.advert_id = $1", advertID).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	fmt.Println(sql)

	if err != nil {
		return nil, err
	}

	rows, release, err := s.db.Query(ctx, sql, args)
	if err != nil {
		return nil, err
	}

	defer release()

	var subscribers []*domain.Subscriber
	err = s.db.ScanAll(rows, &subscribers)
	if err != nil {
		return nil, postgres.CheckEmptyRows(err)
	}

	return subscribers, nil
}

func (s *subscriberRepo) InsertOnlySubscription(ctx context.Context, sub *domain.Subscriber) error {

	subscription := sub.Subscriptions()[0]

	sql, args, err := sq.Insert("subscriptions").
		Columns("advert_id", "subscriber_id").
		Values(subscription.AdvertID, subscription.SubscriberID).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	_, release, err := s.db.Query(ctx, sql, args)
	if err != nil {
		return err
	}

	defer release()

	return nil
}

func (s *subscriberRepo) InsertSubscriber(ctx context.Context, sub *domain.Subscriber) error {

	// Build sql for subscription insert
	sqlInsertSubscriber, argsInsertSubscriber, err := sq.Insert("subscribers").
		Columns("subscriber_id", "telegram_id").
		Values(sub.SubscriberID, sub.TelegramID).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	// Build sql for subscription insert
	subscription := sub.Subscriptions()[0]
	fmt.Println(subscription.AdvertID, subscription.SubscriberID)
	sqlInsertSubscription, argsInsertSubscription, err := sq.Insert("subscriptions").
		Columns("advert_id", "subscriber_id").
		Values(subscription.AdvertID, subscription.SubscriberID).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	fmt.Println(sqlInsertSubscriber, argsInsertSubscriber)
	fmt.Println(sqlInsertSubscription, argsInsertSubscription)
	conn, err := s.db.ConnAcquire(ctx)
	if err != nil {
		return err
	}

	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}

	// Executed within tx
	{
		// Insert subscriber
		_, err = tx.Exec(ctx, sqlInsertSubscriber, argsInsertSubscriber...)
		if err != nil {
			if txError := tx.Rollback(ctx); txError != nil {
				return fmt.Errorf("%v: %v", txError, err)
			}

			return err
		}

		// Insert subscription
		_, err = tx.Exec(ctx, sqlInsertSubscription, argsInsertSubscription...)
		if err != nil {
			if txError := tx.Rollback(ctx); txError != nil {
				return fmt.Errorf("%v: %v", txError, err)
			}

			return err
		}

	}

	if txError := tx.Commit(ctx); txError != nil {
		return txError
	}

	return nil
}
