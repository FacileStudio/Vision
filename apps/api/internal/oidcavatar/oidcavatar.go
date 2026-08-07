package oidcavatar

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Profile struct {
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	Picture           string `json:"picture"`
}

func (p Profile) DisplayName() string {
	if p.Name != "" {
		return p.Name
	}
	if p.PreferredUsername != "" {
		return p.PreferredUsername
	}
	full := strings.TrimSpace(p.GivenName + " " + p.FamilyName)
	if full != "" {
		return full
	}
	return ""
}

const maxAvatarSize = 5 << 20

func FetchAvatar(pictureURL, storageDir string, userID int64, logger *slog.Logger) (string, error) {
	if !strings.HasPrefix(pictureURL, "https://") {
		return "", fmt.Errorf("avatar URL must be HTTPS")
	}

	parsed, err := net.ResolveTCPAddr("tcp", hostFromURL(pictureURL))
	if err == nil && isPrivateIP(parsed.IP) {
		return "", fmt.Errorf("avatar URL resolves to a private IP")
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pictureURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch avatar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("avatar fetch returned status %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	ext, ok := avatarExtension(ct)
	if !ok {
		return "", fmt.Errorf("unsupported avatar content-type: %s", ct)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxAvatarSize+1))
	if err != nil {
		return "", fmt.Errorf("failed to read avatar body: %w", err)
	}
	if len(body) > maxAvatarSize {
		return "", fmt.Errorf("avatar exceeds 5MB limit")
	}

	avatarDir := filepath.Join(storageDir, "avatars")
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create avatar directory: %w", err)
	}

	filename := fmt.Sprintf("oidc-%d-%d.%s", userID, time.Now().UnixNano(), ext)
	fullPath := filepath.Join(avatarDir, filename)
	if err := os.WriteFile(fullPath, body, 0o644); err != nil {
		return "", fmt.Errorf("failed to write avatar file: %w", err)
	}

	relativePath := filepath.Join("avatars", filename)
	logger.Debug("fetched OIDC avatar", slog.String("path", relativePath))
	return relativePath, nil
}

func RemoveFile(storageDir, relativePath string) {
	if relativePath == "" {
		return
	}
	os.Remove(filepath.Join(storageDir, relativePath))
}

// Missing reports whether an avatar the database still points at is absent from disk.
// The row and the file have independent lifetimes — a container rebuilt over an unmounted
// storage dir takes the files and leaves the rows — and without this check the refetch
// condition reads "picture unchanged, path non-empty" and never repairs itself, so the
// avatar 404s forever. An empty path counts as missing: there is nothing to serve.
func Missing(storageDir, relativePath string) bool {
	if relativePath == "" {
		return true
	}
	_, err := os.Stat(filepath.Join(storageDir, relativePath))
	return err != nil
}

func avatarExtension(ct string) (string, bool) {
	ct = strings.ToLower(strings.TrimSpace(ct))
	if idx := strings.Index(ct, ";"); idx != -1 {
		ct = strings.TrimSpace(ct[:idx])
	}
	switch ct {
	case "image/png":
		return "png", true
	case "image/jpeg":
		return "jpg", true
	case "image/gif":
		return "gif", true
	case "image/webp":
		return "webp", true
	default:
		return "", false
	}
}

func hostFromURL(rawURL string) string {
	trimmed := strings.TrimPrefix(rawURL, "https://")
	if idx := strings.Index(trimmed, "/"); idx != -1 {
		trimmed = trimmed[:idx]
	}
	if !strings.Contains(trimmed, ":") {
		trimmed += ":443"
	}
	return trimmed
}

func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	return false
}
