package events

import "strings"

func parseUserAgent(ua string) (browser, os, device string) {
	switch {
	case strings.Contains(ua, "Firefox/"):
		browser = "Firefox"
	case strings.Contains(ua, "Edg/"):
		browser = "Edge"
	case strings.Contains(ua, "OPR/") || strings.Contains(ua, "Opera/"):
		browser = "Opera"
	case strings.Contains(ua, "Chrome/"):
		browser = "Chrome"
	case strings.Contains(ua, "Safari/"):
		browser = "Safari"
	default:
		browser = "Other"
	}

	switch {
	case strings.Contains(ua, "Windows"):
		os = "Windows"
	case strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Mac OS"):
		os = "macOS"
	case strings.Contains(ua, "Android"):
		os = "Android"
	case strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") || strings.Contains(ua, "iPod"):
		os = "iOS"
	case strings.Contains(ua, "Linux"):
		os = "Linux"
	default:
		os = "Other"
	}

	switch {
	case strings.Contains(ua, "Mobile"):
		device = "Mobile"
	case strings.Contains(ua, "iPad") || strings.Contains(ua, "Tablet"):
		device = "Tablet"
	default:
		device = "Desktop"
	}

	return
}
