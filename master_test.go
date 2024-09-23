package ratelimiter

import (
	"testing"
	"time"
)

func TestSetRules(t *testing.T) {
	rl := MasterRatelimiter{}
	rl.stats = make(map[string]*Stat)

	err := rl.SetRules("limit=10;reset=1m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rl.maxCount != 10 {
		t.Errorf("expected maxCount 10, got %d", rl.maxCount)
	}

	if rl.duration != time.Minute {
		t.Errorf("expected duration 1 minute, got %v", rl.duration)
	}

	if rl.isCalendar {
		t.Errorf("expected isCalendar to be false")
	}
}

func TestSetRulesCalendar(t *testing.T) {
	rl := &MasterRatelimiter{}
	rl.stats = make(map[string]*Stat)

	err := rl.SetRules("limit=5;reset=calendar@hour")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rl.maxCount != 5 {
		t.Errorf("expected maxCount 5, got %d", rl.maxCount)
	}

	if rl.calendarPeriod != "hour" {
		t.Errorf("expected calendarPeriod 'hour', got %s", rl.calendarPeriod)
	}

	if !rl.isCalendar {
		t.Errorf("expected isCalendar to be true")
	}
}

func TestSetRuleIncorrect(t *testing.T) {
	rl := &MasterRatelimiter{}
	rl.stats = make(map[string]*Stat)

	invalidRules := []string{
		"limit=;reset=calendar@hour",   // Invalid limit
		"limit=2;reset=invalid@period", // Invalid reset period
		"limit=-1;reset=calendar@day",  // Invalid limit (negative)
		"limit=2;reset=",               // Empty reset rule
		"unknown=key",                  // Unknown key
	}

	for _, rule := range invalidRules {
		err := rl.SetRules(rule)
		if err == nil {
			t.Errorf("expected error for rule: %s, but got none", rule)
		}
	}
}

func TestIncAllowWithHourReset(t *testing.T) {
	rl := &MasterRatelimiter{}
	rl.stats = make(map[string]*Stat)
	err := rl.SetRules("limit=2;reset=calendar@hour")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rl.Inc("user1")

	allow, err := rl.Allow("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allow {
		t.Errorf("expected Allow to return true after first increment")
	}

	rl.Inc("user1")

	allow, err = rl.Allow("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allow {
		t.Errorf("expected Allow to return true after second increment")
	}

	rl.Inc("user1")

	allow, err = rl.Allow("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allow {
		t.Errorf("expected Allow to return false after exceeding limit")
	}

	rl.stats["user1"].LastReset = rl.stats["user1"].LastReset.Add(-time.Hour)

	allow, err = rl.Allow("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allow {
		t.Errorf("expected Allow to return true after hour reset")
	}
}

func TestIncAllowWithDayReset(t *testing.T) {
	rl := &MasterRatelimiter{}
	rl.stats = make(map[string]*Stat)
	err := rl.SetRules("limit=2;reset=calendar@day")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rl.Inc("user1")

	allow, err := rl.Allow("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allow {
		t.Errorf("expected Allow to return true after first increment")
	}

	rl.Inc("user1")

	allow, err = rl.Allow("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allow {
		t.Errorf("expected Allow to return true after second increment")
	}

	rl.Inc("user1")

	allow, err = rl.Allow("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allow {
		t.Errorf("expected Allow to return false after exceeding limit")
	}

	rl.stats["user1"].LastReset = rl.stats["user1"].LastReset.Add(-24 * time.Hour)

	allow, err = rl.Allow("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allow {
		t.Errorf("expected Allow to return true after day reset")
	}
}

func TestIncResetByDuration(t *testing.T) {
	rl := &MasterRatelimiter{}
	rl.stats = make(map[string]*Stat)
	err := rl.SetRules("limit=2;reset=1s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rl.Inc("user1")

	allow, err := rl.Allow("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allow {
		t.Errorf("expected Allow to return true after first increment")
	}

	rl.Inc("user1")

	allow, err = rl.Allow("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allow {
		t.Errorf("expected Allow to return true after second increment")
	}

	rl.Inc("user1")

	allow, err = rl.Allow("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allow {
		t.Errorf("expected Allow to return false after exceeding limit")
	}

	time.Sleep(1 * time.Second)

	allow, err = rl.Allow("user1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allow {
		t.Errorf("expected Allow to return true after duration reset")
	}
}
