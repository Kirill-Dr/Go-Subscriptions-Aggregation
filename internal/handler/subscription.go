package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"subscriptions/internal/model"
	"subscriptions/internal/repository"
)

type service interface {
	Create(ctx context.Context, sub model.Subscription) (model.Subscription, error)
	Get(ctx context.Context, id uuid.UUID) (model.Subscription, error)
	Update(ctx context.Context, sub model.Subscription) (model.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, f model.ListFilter) ([]model.Subscription, error)
	Summary(ctx context.Context, f model.SummaryFilter) (int64, error)
}

type SubscriptionHandler struct {
	svc service
	log *slog.Logger
}

func NewSubscriptionHandler(svc service, log *slog.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc, log: log}
}

func (h *SubscriptionHandler) RegisterRoutes(r chi.Router) {
	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/summary", h.Summary)
		r.Get("/{id}", h.Get)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	req, err := decodeRequest(r)
	if err != nil {
		h.log.Warn("invalid create request", "error", err)
		writeError(w, h.log, http.StatusBadRequest, err.Error())
		return
	}

	created, err := h.svc.Create(r.Context(), req.ToSubscription())
	if err != nil {
		writeError(w, h.log, http.StatusInternalServerError, "failed to create subscription")
		return
	}
	writeJSON(w, h.log, http.StatusCreated, model.NewSubscriptionResponse(created))
}

func (h *SubscriptionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, h.log, http.StatusBadRequest, err.Error())
		return
	}

	sub, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.writeRepoError(w, err)
		return
	}
	writeJSON(w, h.log, http.StatusOK, model.NewSubscriptionResponse(sub))
}

func (h *SubscriptionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, h.log, http.StatusBadRequest, err.Error())
		return
	}

	req, err := decodeRequest(r)
	if err != nil {
		h.log.Warn("invalid update request", "error", err)
		writeError(w, h.log, http.StatusBadRequest, err.Error())
		return
	}

	sub := req.ToSubscription()
	sub.ID = id

	updated, err := h.svc.Update(r.Context(), sub)
	if err != nil {
		h.writeRepoError(w, err)
		return
	}
	writeJSON(w, h.log, http.StatusOK, model.NewSubscriptionResponse(updated))
}

func (h *SubscriptionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, h.log, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.writeRepoError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SubscriptionHandler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseListFilter(r)
	if err != nil {
		writeError(w, h.log, http.StatusBadRequest, err.Error())
		return
	}

	subs, err := h.svc.List(r.Context(), filter)
	if err != nil {
		writeError(w, h.log, http.StatusInternalServerError, "failed to list subscriptions")
		return
	}

	resp := make([]model.SubscriptionResponse, 0, len(subs))
	for _, s := range subs {
		resp = append(resp, model.NewSubscriptionResponse(s))
	}
	writeJSON(w, h.log, http.StatusOK, resp)
}

func (h *SubscriptionHandler) Summary(w http.ResponseWriter, r *http.Request) {
	filter, err := parseSummaryFilter(r)
	if err != nil {
		writeError(w, h.log, http.StatusBadRequest, err.Error())
		return
	}

	total, err := h.svc.Summary(r.Context(), filter)
	if err != nil {
		writeError(w, h.log, http.StatusInternalServerError, "failed to calculate summary")
		return
	}
	writeJSON(w, h.log, http.StatusOK, model.SummaryResponse{TotalPrice: total})
}

func (h *SubscriptionHandler) writeRepoError(w http.ResponseWriter, err error) {
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, h.log, http.StatusNotFound, "subscription not found")
		return
	}
	writeError(w, h.log, http.StatusInternalServerError, "internal error")
}
