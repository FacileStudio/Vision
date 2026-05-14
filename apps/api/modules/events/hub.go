package events

import (
	"sync"
)

type PageviewEvent struct {
	SiteID    int64  `json:"site_id"`
	Path      string `json:"path"`
	Referrer  string `json:"referrer"`
	Country   string `json:"country"`
	VisitorID string `json:"visitor_id"`
	Timestamp string `json:"timestamp"`
}

type Hub struct {
	mu          sync.RWMutex
	subscribers map[int64]map[chan PageviewEvent]struct{}
}

func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[int64]map[chan PageviewEvent]struct{}),
	}
}

func (h *Hub) Subscribe(siteID int64) chan PageviewEvent {
	ch := make(chan PageviewEvent, 64)
	h.mu.Lock()
	if h.subscribers[siteID] == nil {
		h.subscribers[siteID] = make(map[chan PageviewEvent]struct{})
	}
	h.subscribers[siteID][ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(siteID int64, ch chan PageviewEvent) {
	h.mu.Lock()
	if subs, ok := h.subscribers[siteID]; ok {
		delete(subs, ch)
		if len(subs) == 0 {
			delete(h.subscribers, siteID)
		}
	}
	h.mu.Unlock()
	close(ch)
}

func (h *Hub) Broadcast(event PageviewEvent) {
	h.mu.RLock()
	subs := h.subscribers[event.SiteID]
	for ch := range subs {
		select {
		case ch <- event:
		default:
		}
	}
	h.mu.RUnlock()
}
