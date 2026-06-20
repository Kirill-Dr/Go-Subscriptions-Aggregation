package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"subscriptions/internal/model"
)

type repo interface {
	Create(ctx context.Context, s model.Subscription) (model.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.Subscription, error)
	Update(ctx context.Context, s model.Subscription) (model.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, f model.ListFilter) ([]model.Subscription, error)
	Summary(ctx context.Context, f model.SummaryFilter) (int64, error)
}

type SubscriptionService struct {
	repo repo
	log  *slog.Logger
}

func NewSubscriptionService(r repo, log *slog.Logger) *SubscriptionService {
	return &SubscriptionService{repo: r, log: log}
}

func (s *SubscriptionService) Create(ctx context.Context, sub model.Subscription) (model.Subscription, error) {
	created, err := s.repo.Create(ctx, sub)
	if err != nil {
		s.log.Error("failed to create subscription", "service_name", sub.ServiceName, "user_id", sub.UserID, "error", err)
		return model.Subscription{}, err
	}
	s.log.Info("subscription created", "id", created.ID, "user_id", created.UserID, "service_name", created.ServiceName)
	return created, nil
}

func (s *SubscriptionService) Get(ctx context.Context, id uuid.UUID) (model.Subscription, error) {
	sub, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.log.Warn("failed to get subscription", "id", id, "error", err)
		return model.Subscription{}, err
	}
	return sub, nil
}

func (s *SubscriptionService) Update(ctx context.Context, sub model.Subscription) (model.Subscription, error) {
	updated, err := s.repo.Update(ctx, sub)
	if err != nil {
		s.log.Error("failed to update subscription", "id", sub.ID, "error", err)
		return model.Subscription{}, err
	}
	s.log.Info("subscription updated", "id", updated.ID)
	return updated, nil
}

func (s *SubscriptionService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		s.log.Warn("failed to delete subscription", "id", id, "error", err)
		return err
	}
	s.log.Info("subscription deleted", "id", id)
	return nil
}

func (s *SubscriptionService) List(ctx context.Context, f model.ListFilter) ([]model.Subscription, error) {
	subs, err := s.repo.List(ctx, f)
	if err != nil {
		s.log.Error("failed to list subscriptions", "error", err)
		return nil, err
	}
	s.log.Debug("subscriptions listed", "count", len(subs))
	return subs, nil
}

func (s *SubscriptionService) Summary(ctx context.Context, f model.SummaryFilter) (int64, error) {
	total, err := s.repo.Summary(ctx, f)
	if err != nil {
		s.log.Error("failed to calculate summary", "error", err)
		return 0, err
	}
	s.log.Info("summary calculated", "from", f.From, "to", f.To, "total_price", total)
	return total, nil
}
