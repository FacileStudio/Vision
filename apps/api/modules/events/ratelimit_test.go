package events

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func heartbeatFrom(handler http.Handler, remoteAddr string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/e/h", nil)
	request.RemoteAddr = remoteAddr
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func TestTrackerRefusesAnAddressPastItsBurstWith429(t *testing.T) {
	router := chi.NewRouter()
	RegisterRoutes(router, nil, nil, nil, nil, nil)

	var refused *httptest.ResponseRecorder
	admitted := 0
	for range 10 * TrackerBurst {
		recorder := heartbeatFrom(router, "203.0.113.7:40000")
		if recorder.Code != http.StatusOK {
			refused = recorder
			break
		}
		admitted++
	}

	if refused == nil || refused.Code != http.StatusTooManyRequests {
		t.Fatalf("no 429 after %d requests from one address", admitted)
	}
	if admitted < TrackerBurst {
		t.Fatalf("refused after %d requests, inside the burst of %d", admitted, TrackerBurst)
	}
	if refused.Header().Get("Retry-After") == "" {
		t.Error("the refusal carries no Retry-After")
	}
	if refused.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("a tracked site cannot read the refusal")
	}

	if other := heartbeatFrom(router, "198.51.100.9:40000"); other.Code != http.StatusOK {
		t.Errorf("another address shared the refused bucket: %d", other.Code)
	}
}
