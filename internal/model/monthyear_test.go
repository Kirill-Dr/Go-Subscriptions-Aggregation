package model

import (
	"encoding/json"
	"testing"
)

func TestMonthYearMarshalUnmarshal(t *testing.T) {
	const raw = `"07-2025"`

	var my MonthYear
	if err := json.Unmarshal([]byte(raw), &my); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if my.Year() != 2025 || my.Month() != 7 || my.Day() != 1 {
		t.Fatalf("got %v, want 2025-07-01", my.Time)
	}

	out, err := json.Marshal(my)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != raw {
		t.Fatalf("marshal got %s, want %s", out, raw)
	}
}

func TestMonthYearUnmarshalInvalid(t *testing.T) {
	var my MonthYear
	if err := json.Unmarshal([]byte(`"2025-07"`), &my); err == nil {
		t.Fatal("expected error for wrong format, got nil")
	}
}
