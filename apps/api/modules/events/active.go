package events

import (
	"sync"
	"time"
)

const activeWindow = 60 * time.Second

// ActiveTracker tracks which visitors hit a site within the active window.
type ActiveTracker struct {
	mu       sync.RWMutex
	visitors map[int64]map[string]time.Time
}

// NewActiveTracker returns an ActiveTracker and starts its background
// expiration sweep.
func NewActiveTracker() *ActiveTracker {
	t := &ActiveTracker{
		visitors: make(map[int64]map[string]time.Time),
	}
	go t.cleanup()
	return t
}

func (t *ActiveTracker) Touch(siteID int64, visitorID string) {
	if visitorID == "" {
		return
	}
	t.mu.Lock()
	if t.visitors[siteID] == nil {
		t.visitors[siteID] = make(map[string]time.Time)
	}
	t.visitors[siteID][visitorID] = time.Now()
	t.mu.Unlock()
}

func (t *ActiveTracker) Count(siteID int64) int64 {
	cutoff := time.Now().Add(-activeWindow)
	t.mu.RLock()
	defer t.mu.RUnlock()
	var count int64
	for _, ts := range t.visitors[siteID] {
		if ts.After(cutoff) {
			count++
		}
	}
	return count
}

func (t *ActiveTracker) cleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		cutoff := time.Now().Add(-activeWindow)
		t.mu.Lock()
		for siteID, visitors := range t.visitors {
			for vid, ts := range visitors {
				if ts.Before(cutoff) {
					delete(visitors, vid)
				}
			}
			if len(visitors) == 0 {
				delete(t.visitors, siteID)
			}
		}
		t.mu.Unlock()
	}
}
