package ratelimiter

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type SlaveManager struct {
	remoteAddr string
}

type SlaveRatelimiter struct {
	remoteAddr string
	tag        string
}

func (m *SlaveManager) GetOrCreate(tag string) (Ratelimiter, error) {
	url := fmt.Sprintf("http://%s/get_or_create?tag=%s", m.remoteAddr, tag)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to call get_or_create: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get_or_create failed with status: %s", resp.Status)
	}

	return &SlaveRatelimiter{remoteAddr: m.remoteAddr, tag: tag}, nil
}

func (rl *SlaveRatelimiter) SetRules(ruleStr string) error {
	ruleStrEscaped := url.QueryEscape(ruleStr)
	url := fmt.Sprintf("http://%s/set_rule?tag=%s&rule=%s", rl.remoteAddr, rl.tag, ruleStrEscaped)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to set rules: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("set_rule failed with status: %s", resp.Status)
	}

	return nil
}

func (rl *SlaveRatelimiter) Inc(key string) error {
	url := fmt.Sprintf("http://%s/inc?tag=%s&key=%s", rl.remoteAddr, rl.tag, key)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to increment: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("inc failed with status: %s", resp.Status)
	}

	return nil
}

func (rl *SlaveRatelimiter) Allow(key string) (bool, error) {
	url := fmt.Sprintf("http://%s/allow?tag=%s&key=%s", rl.remoteAddr, rl.tag, key)
	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to check allowance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("allow failed with status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) == 0 {
		return false, fmt.Errorf("empty response body")
	}

	return string(body) == "true", nil
}
