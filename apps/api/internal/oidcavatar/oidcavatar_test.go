package oidcavatar

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMissing(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "avatars"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	present := filepath.Join("avatars", "oidc-1-123.png")
	if err := os.WriteFile(filepath.Join(dir, present), []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	cases := []struct {
		name string
		rel  string
		want bool
	}{
		{"file on disk", present, false},
		{"row points at a file a rebuild took", filepath.Join("avatars", "oidc-1-999.png"), true},
		{"no avatar recorded", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Missing(dir, tc.rel); got != tc.want {
				t.Errorf("Missing(%q) = %v, want %v", tc.rel, got, tc.want)
			}
		})
	}
}
