package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"subscriptions/internal/model"
	"subscriptions/internal/repository"
)

type stubService struct {
	createFn  func(ctx context.Context, sub model.Subscription) (model.Subscription, error)
	getFn     func(ctx context.Context, id uuid.UUID) (model.Subscription, error)
	summaryFn func(ctx context.Context, f model.SummaryFilter) (int64, error)
}

func (s *stubService) Create(ctx context.Context, sub model.Subscription) (model.Subscription, error) {
	return s.createFn(ctx, sub)
}
func (s *stubService) Get(ctx context.Context, id uuid.UUID) (model.Subscription, error) {
	return s.getFn(ctx, id)
}
func (s *stubService) Update(ctx context.Context, sub model.Subscription) (model.Subscription, error) {
	return sub, nil
}
func (s *stubService) Delete(ctx context.Context, id uuid.UUID) error { return nil }
func (s *stubService) List(ctx context.Context, f model.ListFilter) ([]model.Subscription, error) {
	return nil, nil
}
func (s *stubService) Summary(ctx context.Context, f model.SummaryFilter) (int64, error) {
	return s.summaryFn(ctx, f)
}

func newTestRouter(svc service) http.Handler {
	h := NewSubscriptionHandler(svc, slog.New(slog.NewTextHandler(io.Discard, nil)))
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return r
}

func TestCreate(t *testing.T) {
	id := uuid.New()
	svc := &stubService{
		createFn: func(_ context.Context, sub model.Subscription) (model.Subscription, error) {
			sub.ID = id
			return sub, nil
		},
	}
	body := `{"service_name":"Yandex Plus","price":400,"user_id":"60601fee-2bf1-4721-ae6f-7636e79a0cba","start_date":"07-2025"}`
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(body))
	rec := httptest.NewRecorder()

	newTestRouter(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body=%s", rec.Code, rec.Body.String())
	}
	var resp model.SubscriptionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ServiceName != "Yandex Plus" || resp.Price != 400 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestCreateValidationError(t *testing.T) {
	svc := &stubService{}
	body := `{"service_name":"","price":400,"user_id":"bad","start_date":"07-2025"}`
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", strings.NewReader(body))
	rec := httptest.NewRecorder()

	newTestRouter(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestGetNotFound(t *testing.T) {
	svc := &stubService{
		getFn: func(_ context.Context, _ uuid.UUID) (model.Subscription, error) {
			return model.Subscription{}, repository.ErrNotFound
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/subscriptions/"+uuid.New().String(), nil)
	rec := httptest.NewRecorder()

	newTestRouter(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestSummary(t *testing.T) {
	svc := &stubService{
		summaryFn: func(_ context.Context, f model.SummaryFilter) (int64, error) {
			return 2400, nil
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/subscriptions/summary?from=01-2025&to=12-2025", nil)
	rec := httptest.NewRecorder()

	newTestRouter(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var resp model.SummaryResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.TotalPrice != 2400 {
		t.Fatalf("total = %d, want 2400", resp.TotalPrice)
	}
}

func TestSummaryMissingParams(t *testing.T) {
	svc := &stubService{}
	req := httptest.NewRequest(http.MethodGet, "/subscriptions/summary", nil)
	rec := httptest.NewRecorder()

	newTestRouter(svc).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
