package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"subscriptions/internal/model"
)

var ErrNotFound = errors.New("subscription not found")

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, s model.Subscription) (model.Subscription, error) {
	const query = `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate)
	created, err := scanSubscription(row)
	if err != nil {
		return model.Subscription{}, fmt.Errorf("create subscription: %w", err)
	}
	return created, nil
}

func (r *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (model.Subscription, error) {
	const query = `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE id = $1`

	row := r.pool.QueryRow(ctx, query, id)
	sub, err := scanSubscription(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Subscription{}, ErrNotFound
		}
		return model.Subscription{}, fmt.Errorf("get subscription: %w", err)
	}
	return sub, nil
}

func (r *SubscriptionRepository) Update(ctx context.Context, s model.Subscription) (model.Subscription, error) {
	const query = `
		UPDATE subscriptions
		SET service_name = $2, price = $3, user_id = $4, start_date = $5, end_date = $6, updated_at = now()
		WHERE id = $1
		RETURNING id, service_name, price, user_id, start_date, end_date, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, s.ID, s.ServiceName, s.Price, s.UserID, s.StartDate, s.EndDate)
	updated, err := scanSubscription(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Subscription{}, ErrNotFound
		}
		return model.Subscription{}, fmt.Errorf("update subscription: %w", err)
	}
	return updated, nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM subscriptions WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SubscriptionRepository) List(ctx context.Context, f model.ListFilter) ([]model.Subscription, error) {
	const query = `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE ($1::uuid IS NULL OR user_id = $1)
		  AND ($2::text IS NULL OR service_name = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, query, f.UserID, f.ServiceName, f.Limit, f.Offset)
	if err != nil {
		return nil, fmt.Errorf("list subscriptions: %w", err)
	}
	defer rows.Close()

	subs := make([]model.Subscription, 0)
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, fmt.Errorf("scan subscription: %w", err)
		}
		subs = append(subs, sub)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate subscriptions: %w", err)
	}
	return subs, nil
}

func (r *SubscriptionRepository) Summary(ctx context.Context, f model.SummaryFilter) (int64, error) {
	const query = `
		SELECT COALESCE(SUM(
			price * (
				  (EXTRACT(YEAR FROM LEAST(COALESCE(end_date, $2), $2)) * 12
				 + EXTRACT(MONTH FROM LEAST(COALESCE(end_date, $2), $2)))
				- (EXTRACT(YEAR FROM GREATEST(start_date, $1)) * 12
				 + EXTRACT(MONTH FROM GREATEST(start_date, $1)))
				+ 1
			)
		), 0)::bigint AS total
		FROM subscriptions
		WHERE start_date <= $2
		  AND (end_date IS NULL OR end_date >= $1)
		  AND ($3::uuid IS NULL OR user_id = $3)
		  AND ($4::text IS NULL OR service_name = $4)`

	var total int64
	err := r.pool.QueryRow(ctx, query, f.From, f.To, f.UserID, f.ServiceName).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("summary subscriptions: %w", err)
	}
	return total, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanSubscription(row rowScanner) (model.Subscription, error) {
	var s model.Subscription
	err := row.Scan(
		&s.ID,
		&s.ServiceName,
		&s.Price,
		&s.UserID,
		&s.StartDate,
		&s.EndDate,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	return s, err
}
