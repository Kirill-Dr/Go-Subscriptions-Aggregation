package model

import (
	"fmt"

	"github.com/google/uuid"
)

type SubscriptionRequest struct {
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      string     `json:"user_id"`
	StartDate   MonthYear  `json:"start_date"`
	EndDate     *MonthYear `json:"end_date,omitempty"`
}

func (r SubscriptionRequest) Validate() error {
	if r.ServiceName == "" {
		return fmt.Errorf("service_name is required")
	}
	if r.Price < 0 {
		return fmt.Errorf("price must be a non-negative integer")
	}
	if _, err := uuid.Parse(r.UserID); err != nil {
		return fmt.Errorf("user_id must be a valid UUID")
	}
	if r.StartDate.IsZero() {
		return fmt.Errorf("start_date is required (format MM-YYYY)")
	}
	if r.EndDate != nil && r.EndDate.Before(r.StartDate.Time) {
		return fmt.Errorf("end_date must not be earlier than start_date")
	}
	return nil
}

func (r SubscriptionRequest) ToSubscription() Subscription {
	sub := Subscription{
		ServiceName: r.ServiceName,
		Price:       r.Price,
		UserID:      uuid.MustParse(r.UserID),
		StartDate:   r.StartDate.Time,
	}
	if r.EndDate != nil {
		end := r.EndDate.Time
		sub.EndDate = &end
	}
	return sub
}

type SubscriptionResponse struct {
	ID          string     `json:"id"`
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserID      string     `json:"user_id"`
	StartDate   MonthYear  `json:"start_date"`
	EndDate     *MonthYear `json:"end_date,omitempty"`
}

func NewSubscriptionResponse(s Subscription) SubscriptionResponse {
	resp := SubscriptionResponse{
		ID:          s.ID.String(),
		ServiceName: s.ServiceName,
		Price:       s.Price,
		UserID:      s.UserID.String(),
		StartDate:   NewMonthYear(s.StartDate),
	}
	if s.EndDate != nil {
		end := NewMonthYear(*s.EndDate)
		resp.EndDate = &end
	}
	return resp
}

type SummaryResponse struct {
	TotalPrice int64 `json:"total_price"`
}
