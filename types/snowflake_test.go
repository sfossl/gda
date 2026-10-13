package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestParseSnowflake(t *testing.T) {
	s, err := ParseSnowflake("175928847299117063")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.String() != "175928847299117063" {
		t.Fatalf("expected 175928847299117063, got %s", s.String())
	}
	if !s.IsValid() {
		t.Fatalf("expected valid snowflake")
	}

	_, err = ParseSnowflake("invalid")
	if err == nil {
		t.Fatalf("expected error for invalid string")
	}
}

func TestSnowflakeTime(t *testing.T) {
	s := Snowflake(175928847299117063)
	expectedTime := time.Date(2016, 4, 30, 11, 18, 25, 796000000, time.UTC)
	actualTime := s.Time().UTC()
	if !actualTime.Equal(expectedTime) {
		t.Fatalf("expected %v, got %v", expectedTime, actualTime)
	}

	zero := Snowflake(0)
	if !zero.Time().IsZero() {
		t.Fatalf("expected zero snowflake to return IsZero() == true")
	}
}

func TestSnowflakeJSON(t *testing.T) {
	type wrapper struct {
		ID Snowflake `json:"id"`
	}

	w := wrapper{ID: Snowflake(175928847299117063)}
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if string(data) != `{"id":"175928847299117063"}` {
		t.Fatalf("unexpected json: %s", string(data))
	}

	var parsed wrapper
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if parsed.ID != w.ID {
		t.Fatalf("expected %d, got %d", w.ID, parsed.ID)
	}

	var nullWrapper wrapper
	if err := json.Unmarshal([]byte(`{"id":null}`), &nullWrapper); err != nil {
		t.Fatalf("unmarshal null error: %v", err)
	}
	if nullWrapper.ID != 0 {
		t.Fatalf("expected 0, got %d", nullWrapper.ID)
	}
}

