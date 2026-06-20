package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"subscriptions/internal/model"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

func decodeRequest(r *http.Request) (model.SubscriptionRequest, error) {
	var req model.SubscriptionRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		return model.SubscriptionRequest{}, err
	}
	if err := req.Validate(); err != nil {
		return model.SubscriptionRequest{}, err
	}
	return req, nil
}

func parseID(r *http.Request) (uuid.UUID, error) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid id: must be a valid UUID")
	}
	return id, nil
}

func parseListFilter(r *http.Request) (model.ListFilter, error) {
	q := r.URL.Query()
	filter := model.ListFilter{Limit: defaultLimit}

	if v := q.Get("user_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return model.ListFilter{}, errors.New("invalid user_id: must be a valid UUID")
		}
		filter.UserID = &id
	}
	if v := q.Get("service_name"); v != "" {
		filter.ServiceName = &v
	}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return model.ListFilter{}, errors.New("invalid limit: must be a positive integer")
		}
		if n > maxLimit {
			n = maxLimit
		}
		filter.Limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return model.ListFilter{}, errors.New("invalid offset: must be a non-negative integer")
		}
		filter.Offset = n
	}
	return filter, nil
}

func parseSummaryFilter(r *http.Request) (model.SummaryFilter, error) {
	q := r.URL.Query()

	from, err := parseMonthYear(q.Get("from"))
	if err != nil {
		return model.SummaryFilter{}, errors.New("invalid 'from': expected format MM-YYYY")
	}
	to, err := parseMonthYear(q.Get("to"))
	if err != nil {
		return model.SummaryFilter{}, errors.New("invalid 'to': expected format MM-YYYY")
	}
	if to.Before(from) {
		return model.SummaryFilter{}, errors.New("'to' must not be earlier than 'from'")
	}

	filter := model.SummaryFilter{From: from, To: to}
	if v := q.Get("user_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return model.SummaryFilter{}, errors.New("invalid user_id: must be a valid UUID")
		}
		filter.UserID = &id
	}
	if v := q.Get("service_name"); v != "" {
		filter.ServiceName = &v
	}
	return filter, nil
}

func parseMonthYear(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, errors.New("empty")
	}
	t, err := time.Parse(model.MonthYearLayout, s)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC), nil
}
