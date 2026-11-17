package events

import (
	"testing"
	"time"
)

var classifyNow = time.Date(2026, time.November, 17, 12, 0, 0, 0, time.UTC)

func TestClassifyBotLeavesRealBrowsersHuman(t *testing.T) {
	humans := map[string]string{
		"chrome desktop":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36",
		"firefox desktop":  "Mozilla/5.0 (X11; Linux x86_64; rv:145.0) Gecko/20100101 Firefox/145.0",
		"safari desktop":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/26.1 Safari/605.1.15",
		"safari ios":       "Mozilla/5.0 (iPhone; CPU iPhone OS 18_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.6 Mobile/15E148 Safari/604.1",
		"chrome android":   "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Mobile Safari/537.36",
		"samsung internet": "Mozilla/5.0 (Linux; Android 14; SAMSUNG SM-S921B) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/28.0 Chrome/130.0.0.0 Mobile Safari/537.36",
		"edge":             "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36 Edg/142.0.0.0",
		"cubot phone":      "Mozilla/5.0 (Linux; Android 11; CUBOT P50 Build/RP1A.200720.011) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/141.0.0.0 Mobile Safari/537.36",
		"chrome canary":    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/161.0.0.0 Safari/537.36",
	}
	for label, userAgent := range humans {
		if category, name := classifyBot(userAgent, 0, classifyNow); category != "" {
			t.Errorf("%s classified as %s/%s", label, category, name)
		}
	}
}

func TestClassifyBotNamesBots(t *testing.T) {
	cases := []struct {
		label     string
		userAgent string
		flags     int
		category  string
	}{
		{"googlebot", "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)", 0, "search"},
		{"gptbot", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; GPTBot/1.2; +https://openai.com/gptbot)", 0, "ai"},
		{"chatgpt-user", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko); compatible; ChatGPT-User/1.0; +https://openai.com/bot", 0, "ai"},
		{"claudebot", "Mozilla/5.0 AppleWebKit/537.36 (KHTML, like Gecko; compatible; ClaudeBot/1.0; +claudebot@anthropic.com)", 0, "ai"},
		{"facebook", "facebookexternalhit/1.1 (+http://www.facebook.com/externalhit_uatext.php)", 0, "social"},
		{"curl", "curl/8.5.0", 0, "tool"},
		{"empty", "  ", 0, "tool"},
		{"headless chrome", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) HeadlessChrome/142.0.0.0 Safari/537.36", 0, "headless"},
		{"webdriver", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36", 1, "automation"},
		{"zero window", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36", 2, "headless"},
		{"future chrome", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/170.0.0.0 Safari/537.36", 0, "spoofed"},
		{"chrome on windows 7", "Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36", 0, "spoofed"},
		{"generic crawler", "Mozilla/5.0 (compatible; FooCrawler/1.0)", 0, "tool"},
	}
	for _, c := range cases {
		if category, name := classifyBot(c.userAgent, c.flags, classifyNow); category != c.category || name == "" {
			t.Errorf("%s: got %q/%q, want category %q", c.label, category, name, c.category)
		}
	}
}

func TestClassifyBotNamesTheGenericProduct(t *testing.T) {
	if _, name := classifyBot("Mozilla/5.0 (compatible; FooCrawler/1.0)", 0, classifyNow); name != "FooCrawler" {
		t.Errorf("generic name %q", name)
	}
	if _, name := classifyBot("", 0, classifyNow); name != "Empty user agent" {
		t.Errorf("empty name %q", name)
	}
}
