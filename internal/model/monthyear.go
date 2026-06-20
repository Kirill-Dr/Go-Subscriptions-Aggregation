package model

import (
	"encoding/json"
	"fmt"
	"time"
)

const MonthYearLayout = "01-2006"

type MonthYear struct {
	time.Time
}

func NewMonthYear(t time.Time) MonthYear {
	return MonthYear{Time: time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)}
}

func (m MonthYear) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.Format(MonthYearLayout))
}

func (m *MonthYear) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	t, err := time.Parse(MonthYearLayout, s)
	if err != nil {
		return fmt.Errorf("invalid date %q: expected format MM-YYYY (e.g. 07-2025)", s)
	}
	m.Time = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	return nil
}
