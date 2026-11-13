package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/tronc/httpx"
)

func clientHeaders(t *testing.T, target string) http.Header {
	t.Helper()
	router := httpx.NewRouter(httpx.Config{})
	mountClient(router, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, target, nil))
	return recorder.Header()
}

func TestOnlySharePagesCanBeFramedByDocs(t *testing.T) {
	share := clientHeaders(t, "/share/any-token?embed=1")
	if csp := share.Get("Content-Security-Policy"); !strings.Contains(csp, "frame-ancestors 'self' https://docs.facile.studio;") {
		t.Errorf("share CSP = %q", csp)
	}
	if xfo := share.Get("X-Frame-Options"); xfo != "" {
		t.Errorf("share page still sends X-Frame-Options %q", xfo)
	}

	for _, target := range []string{"/", "/login", "/sites/1", "/shared/x"} {
		header := clientHeaders(t, target)
		if csp := header.Get("Content-Security-Policy"); !strings.Contains(csp, "frame-ancestors 'none'") {
			t.Errorf("%s CSP = %q", target, csp)
		}
		if xfo := header.Get("X-Frame-Options"); xfo != "DENY" {
			t.Errorf("%s X-Frame-Options = %q", target, xfo)
		}
	}
}
