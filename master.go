package ratelimiter

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MasterManager struct {
	ratelimiters map[string]*MasterRatelimiter // key=tag
	port         string
	mu           sync.RWMutex
}

type MasterRatelimiter struct {
	isActive       bool
	maxCount       int64
	isCalendar     bool
	duration       time.Duration
	calendarPeriod string // "hour", "day"
	stats          map[string]*Stat
	mu             sync.RWMutex
}

type Stat struct {
	Count     int64
	LastReset time.Time
}

func (m *MasterManager) GetOrCreate(tag string) (Ratelimiter, error) {
	// try to get existen ratelimiter
	m.mu.RLock()
	rl, isExist := m.ratelimiters[tag]
	m.mu.RUnlock()

	if isExist {
		return rl, nil
	}

	// create new ratelimiter
	m.mu.Lock()
	defer m.mu.Unlock()

	if rl, isExist = m.ratelimiters[tag]; !isExist {
		rl = &MasterRatelimiter{}
		rl.stats = make(map[string]*Stat)
		m.ratelimiters[tag] = rl
	}

	return rl, nil
}

func (m *MasterManager) Serve(port string) (err error) {
	http.HandleFunc("/get_or_create", m.getOrCreateHandler)
	http.HandleFunc("/set_rule", m.setRuleHandler)
	http.HandleFunc("/allow", m.allowHandler)
	http.HandleFunc("/inc", m.incHandler)

	http.ListenAndServe(":"+port, nil)
	err = http.ListenAndServe(":"+port, nil)

	return err
}

func (rl *MasterRatelimiter) SetRules(ruleStr string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	parts := strings.Split(ruleStr, ";")
	for _, part := range parts {
		keyValue := strings.Split(part, "=")
		if len(keyValue) != 2 {
			return fmt.Errorf("invalid rule format")
		}
		key, value := keyValue[0], keyValue[1]

		switch key {
		case "limit":
			maxCount, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid limit: %v", err)
			}
			if maxCount < 1 {
				return fmt.Errorf("invalid limit: value must be greater than 0")
			}

			rl.maxCount = maxCount

		case "reset":
			if strings.HasPrefix(value, "calendar@") {
				// reset by calendar period
				period := strings.TrimPrefix(value, "calendar@")
				if period != "hour" && period != "day" {
					return fmt.Errorf("invalid period")
				}
				rl.calendarPeriod = period
				rl.isCalendar = true
			} else {
				// reset by duration
				duration, err := time.ParseDuration(value)
				if err != nil {
					return fmt.Errorf("invalid duration: %v", err)
				}
				if duration <= 0 {
					return fmt.Errorf("invalid duration: value must be greater than 0")
				}
				rl.duration = duration
			}

		default:
			return fmt.Errorf("unknown key: %s", key)
		}

	}

	// rule activation
	if rl.maxCount <= 0 {
		return nil
	}
	if !rl.isCalendar && rl.duration <= 0 {
		return nil
	}
	rl.isActive = true

	return nil
}

func (rl *MasterRatelimiter) Inc(key string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	stat, isExist := rl.stats[key]
	if !isExist {
		stat = &Stat{}
		stat.LastReset = time.Now()
		rl.stats[key] = stat
	}

	// check for reset
	now := time.Now()
	if rl.isCalendar {
		switch rl.calendarPeriod {
		case "hour":
			if stat.LastReset.Hour() != now.Hour() || now.Sub(stat.LastReset) >= time.Hour {
				stat.Count = 0
				stat.LastReset = now
			}
		case "day":
			if stat.LastReset.Day() != now.Day() || now.Sub(stat.LastReset) >= 24*time.Hour {
				stat.Count = 0
				stat.LastReset = now
			}
		}
	} else {
		if time.Since(stat.LastReset) >= rl.duration {
			stat.Count = 0
			stat.LastReset = now
		}
	}

	// inc counter
	stat.Count++

	return nil
}

func (rl *MasterRatelimiter) Allow(key string) (bool, error) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stat, isExist := rl.stats[key]
	if !isExist {
		// for not existen key
		return true, nil
	}

	if rl.isCalendar {
		// calendar period rule
		now := time.Now()
		switch rl.calendarPeriod {
		case "hour":
			if stat.LastReset.Hour() != now.Hour() || now.Sub(stat.LastReset) >= time.Hour {
				return true, nil
			}
		case "day":
			if stat.LastReset.Day() != now.Day() || now.Sub(stat.LastReset) >= 24*time.Hour {
				return true, nil
			}
		}
	} else {
		// duration rule
		if time.Since(stat.LastReset) >= rl.duration {
			return true, nil
		}
	}

	// check limits
	if stat.Count <= rl.maxCount {
		return true, nil
	}

	return false, nil
}
