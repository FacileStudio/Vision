package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCallbackAcceptsOnlyTheMatchingState(t *testing.T) {
	h := &oidcHandler{successURL: "https://vision.example.com"}
	call := func(cookie, query string) string {
		r := httptest.NewRequest(http.MethodGet, "/api/auth/oidc/callback?code=abc&state="+query, nil)
		r.AddCookie(&http.Cookie{Name: oidcStateCookie, Value: cookie})
		w := httptest.NewRecorder()
		h.callback(w, r)
		return w.Header().Get("Location")
	}

	if got := call("expected-state", "forged-state"); !strings.Contains(got, "login+session+expired") {
		t.Errorf("mismatched state was not rejected, redirected to %q", got)
	}
	if got := call("expected-state", "expected-state"); strings.Contains(got, "login+session+expired") {
		t.Errorf("matching state was rejected, redirected to %q", got)
	}
}

func TestEmailClaimTrustedRefusesOnlyAnExplicitDenial(t *testing.T) {
	cases := []struct {
		name    string
		claim   any
		trusted bool
	}{
		{"absent claim is trusted, or every pre-subject account is stranded", nil, true},
		{"verified", true, true},
		{"unverified", false, false},
		{"verified as a string", "true", true},
		{"unverified as a string", "false", false},
		{"unverified as a string, mixed case", "False", false},
		{"unexpected type", 1, false},
	}
	for _, c := range cases {
		if got := emailClaimTrusted(c.claim); got != c.trusted {
			t.Errorf("%s: emailClaimTrusted(%#v) = %v, want %v", c.name, c.claim, got, c.trusted)
		}
	}
}
