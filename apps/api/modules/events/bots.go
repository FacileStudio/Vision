package events

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	flagWebDriver  = 1
	flagZeroWindow = 2
)

var (
	chromeVersion      = regexp.MustCompile(`Chrome/(\d+)`)
	legacyWindows      = regexp.MustCompile(`Windows NT 6\.[123]`)
	genericBotPattern  = regexp.MustCompile(`bot(?:[^a-z]|$)|crawl|spider|slurp|scrape|fetcher|preview|monitor|checker|http-client|headless`)
	tokenSeparators    = strings.NewReplacer(";", " ", "(", " ", ")", " ", ",", " ")
	chromeReleaseStart = time.Date(2026, time.November, 17, 0, 0, 0, 0, time.UTC)
)

var botChecks = []func(userAgent string, flags int, now time.Time) (string, string){
	emptyAgent,
	knownBot,
	genericBot,
	flaggedBrowser,
	spoofedChrome,
}

// classifyBot sorts a tracker request into a bot category and a bot name from its user agent
// and the tracker's flags (bit 1 navigator.webdriver, bit 2 a 0x0 outer window). An empty
// category means a human. A named bot from the pattern table wins over the flags.
func classifyBot(userAgent string, flags int, now time.Time) (category, name string) {
	for _, check := range botChecks {
		if category, name = check(userAgent, flags, now); category != "" {
			return category, name
		}
	}
	return "", ""
}

// emptyAgent flags a request with no user agent at all, which no browser sends.
func emptyAgent(userAgent string, _ int, _ time.Time) (category, name string) {
	if strings.TrimSpace(userAgent) == "" {
		return "tool", "Empty user agent"
	}
	return "", ""
}

// genericBot catches the unlisted bots the way isbot does: a bare single-token user agent,
// or a product token carrying a bot word. Tokens naming a Cubot phone are skipped.
func genericBot(userAgent string, _ int, _ time.Time) (category, name string) {
	trimmed := strings.TrimSpace(userAgent)
	if !strings.Contains(trimmed, " ") {
		return "tool", productName(trimmed)
	}
	for _, token := range strings.Fields(tokenSeparators.Replace(trimmed)) {
		lower := strings.ToLower(token)
		if !strings.Contains(lower, "cubot") && genericBotPattern.MatchString(lower) {
			return "tool", productName(token)
		}
	}
	return "", ""
}

// productName returns the product part of a name/version token, or "Unknown bot" when the
// token is a URL or the bare Mozilla prefix.
func productName(token string) string {
	name, _, _ := strings.Cut(strings.TrimLeft(token, "+"), "/")
	if name == "" || strings.Contains(token, "://") || strings.EqualFold(name, "mozilla") {
		return "Unknown bot"
	}
	return name
}

// flaggedBrowser trusts the tracker's own probes: a WebDriver-driven browser, or a window
// with no outer size, which only a headless browser has.
func flaggedBrowser(_ string, flags int, _ time.Time) (category, name string) {
	if flags&flagWebDriver != 0 {
		return "automation", "WebDriver"
	}
	if flags&flagZeroWindow != 0 {
		return "headless", "Zero-size window"
	}
	return "", ""
}

// spoofedChrome flags a Chrome user agent no real Chrome sends: a major version not yet
// released, or a recent major on Windows 7 or 8, which Chrome dropped after 109.
func spoofedChrome(userAgent string, _ int, now time.Time) (category, name string) {
	match := chromeVersion.FindStringSubmatch(userAgent)
	if match == nil {
		return "", ""
	}
	major, _ := strconv.Atoi(match[1])
	if name = spoofedChromeName(userAgent, major, now); name == "" {
		return "", ""
	}
	return "spoofed", name
}

// spoofedChromeName names the impossible Chrome a user agent claims, or returns "" when the
// claim is plausible. Stable Chrome ships a major at most every 14 days from 158 on
// 2026-11-17, so the estimate never runs behind a real release, and three majors of headroom
// cover the beta, dev and canary channels.
func spoofedChromeName(userAgent string, major int, now time.Time) string {
	days := math.Floor(now.Sub(chromeReleaseStart).Hours() / 24)
	if expected := 158 + int(math.Floor(days/14)); major > expected+3 {
		return fmt.Sprintf("Chrome %d", major)
	}
	if major > 109 && legacyWindows.MatchString(userAgent) {
		return fmt.Sprintf("Chrome %d on Windows 7/8", major)
	}
	return ""
}
