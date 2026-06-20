package model

import (
	"encoding/json"
	"testing"
)

func TestSubscriptionRequestValidate(t *testing.T) {
	valid := func() SubscriptionRequest {
		var start MonthYear
		_ = json.Unmarshal([]byte(`"07-2025"`), &start)
		return SubscriptionRequest{
			ServiceName: "Yandex Plus",
			Price:       400,
			UserID:      "60601fee-2bf1-4721-ae6f-7636e79a0cba",
			StartDate:   start,
		}
	}

	tests := []struct {
		name    string
		mutate  func(r *SubscriptionRequest)
		wantErr bool
	}{
		{"valid", func(*SubscriptionRequest) {}, false},
		{"empty service", func(r *SubscriptionRequest) { r.ServiceName = "" }, true},
		{"negative price", func(r *SubscriptionRequest) { r.Price = -1 }, true},
		{"bad uuid", func(r *SubscriptionRequest) { r.UserID = "not-a-uuid" }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := valid()
			tt.mutate(&r)
			err := r.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
