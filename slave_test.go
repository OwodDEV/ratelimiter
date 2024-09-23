package ratelimiter

import (
	"testing"
)

var (
	masterManager *MasterManager
)

func startMasterServer() {
}

func TestMain(m *testing.M) {
	masterManager = &MasterManager{ratelimiters: make(map[string]*MasterRatelimiter)}
	go masterManager.Serve("8080")
	m.Run()
}

func TestSlaveManager_GetOrCreate(t *testing.T) {
	slave := &SlaveManager{remoteAddr: "localhost:8080"}
	rl, err := slave.GetOrCreate("my_tag")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if rl == nil {
		t.Fatalf("expected a rate limiter, got nil")
	}
}

func TestSlaveRatelimiter_SetRules(t *testing.T) {
	rl := &SlaveRatelimiter{remoteAddr: "localhost:8080", tag: "my_tag"}
	err := rl.SetRules("limit=5;reset=10s")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSlaveRatelimiter_Inc(t *testing.T) {
	rl := &SlaveRatelimiter{remoteAddr: "localhost:8080", tag: "my_tag"}
	err := rl.Inc("my_key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSlaveRatelimiter_Allow(t *testing.T) {
	rl := &SlaveRatelimiter{remoteAddr: "localhost:8080", tag: "my_tag"}
	allowed, err := rl.Allow("my_key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !allowed {
		t.Fatalf("expected true, got false")
	}
}
