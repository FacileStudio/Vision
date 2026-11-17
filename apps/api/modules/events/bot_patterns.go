package events

import (
	"strings"
	"time"
)

type botPattern struct {
	token    string
	name     string
	category string
}

var botPatterns = []botPattern{
	{"gptbot", "GPTBot", "ai"},
	{"oai-searchbot", "OAI-SearchBot", "ai"},
	{"chatgpt-user", "ChatGPT-User", "ai"},
	{"claudebot", "ClaudeBot", "ai"},
	{"claude-user", "Claude-User", "ai"},
	{"claude-searchbot", "Claude-SearchBot", "ai"},
	{"anthropic-ai", "anthropic-ai", "ai"},
	{"perplexitybot", "PerplexityBot", "ai"},
	{"perplexity-user", "Perplexity-User", "ai"},
	{"ccbot", "CCBot", "ai"},
	{"bytespider", "Bytespider", "ai"},
	{"google-cloudvertexbot", "Google-CloudVertexBot", "ai"},
	{"amazonbot", "Amazonbot", "ai"},
	{"meta-externalagent", "meta-externalagent", "ai"},
	{"meta-webindexer", "meta-webindexer", "search"},
	{"meta-externalads", "meta-externalads", "social"},
	{"cohere-ai", "cohere-ai", "ai"},
	{"diffbot", "Diffbot", "ai"},
	{"youbot", "YouBot", "ai"},
	{"mistralai-user", "MistralAI-User", "ai"},
	{"duckassistbot", "DuckAssistBot", "ai"},
	{"googlebot", "Googlebot", "search"},
	{"google-inspectiontool", "Google-InspectionTool", "search"},
	{"storebot-google", "Storebot-Google", "search"},
	{"bingbot", "bingbot", "search"},
	{"duckduckbot", "DuckDuckBot", "search"},
	{"yandexbot", "YandexBot", "search"},
	{"baiduspider", "Baiduspider", "search"},
	{"applebot", "Applebot", "search"},
	{"qwantbot", "Qwantbot", "search"},
	{"seznambot", "SeznamBot", "search"},
	{"sogou web spider", "Sogou", "search"},
	{"exabot", "Exabot", "search"},
	{"petalbot", "PetalBot", "search"},
	{"ahrefsbot", "AhrefsBot", "search"},
	{"semrushbot", "SemrushBot", "search"},
	{"mj12bot", "MJ12bot", "search"},
	{"dotbot", "DotBot", "search"},
	{"rogerbot", "rogerbot", "search"},
	{"blexbot", "BLEXBot", "search"},
	{"facebookexternalhit", "facebookexternalhit", "social"},
	{"facebot", "Facebot", "social"},
	{"twitterbot", "Twitterbot", "social"},
	{"linkedinbot", "LinkedInBot", "social"},
	{"slackbot", "Slackbot", "social"},
	{"discordbot", "Discordbot", "social"},
	{"telegrambot", "TelegramBot", "social"},
	{"whatsapp", "WhatsApp", "social"},
	{"pinterestbot", "Pinterestbot", "social"},
	{"redditbot", "redditbot", "social"},
	{"skypeuripreview", "SkypeUriPreview", "social"},
	{"embedly", "Embedly", "social"},
	{"iframely", "Iframely", "social"},
	{"mastodon", "Mastodon", "social"},
	{"uptimerobot", "UptimeRobot", "monitor"},
	{"pingdom", "Pingdom", "monitor"},
	{"statuscake", "StatusCake", "monitor"},
	{"site24x7", "Site24x7", "monitor"},
	{"better uptime", "Better Uptime", "monitor"},
	{"uptime-kuma", "Uptime Kuma", "monitor"},
	{"datadog", "Datadog", "monitor"},
	{"newrelicpinger", "NewRelicPinger", "monitor"},
	{"checkly", "Checkly", "monitor"},
	{"gtmetrix", "GTmetrix", "monitor"},
	{"curl/", "curl", "tool"},
	{"wget", "Wget", "tool"},
	{"python-requests", "python-requests", "tool"},
	{"python-urllib", "python-urllib", "tool"},
	{"aiohttp", "aiohttp", "tool"},
	{"httpx", "httpx", "tool"},
	{"go-http-client", "Go-http-client", "tool"},
	{"okhttp", "okhttp", "tool"},
	{"java/", "Java", "tool"},
	{"axios", "axios", "tool"},
	{"node-fetch", "node-fetch", "tool"},
	{"undici", "undici", "tool"},
	{"libwww-perl", "libwww-perl", "tool"},
	{"scrapy", "Scrapy", "tool"},
	{"postmanruntime", "PostmanRuntime", "tool"},
	{"insomnia", "Insomnia", "tool"},
	{"headlessbrowser", "HeadlessBrowser", "tool"},
	{"headlesschrome", "HeadlessChrome", "headless"},
	{"phantomjs", "PhantomJS", "headless"},
	{"lighthouse", "Lighthouse", "headless"},
	{"puppeteer", "Puppeteer", "headless"},
	{"playwright", "Playwright", "headless"},
	{"slimerjs", "SlimerJS", "headless"},
}

// knownBot names the first entry of botPatterns whose token the user agent contains, ignoring case.
func knownBot(userAgent string, _ int, _ time.Time) (category, name string) {
	lower := strings.ToLower(userAgent)
	for _, pattern := range botPatterns {
		if strings.Contains(lower, pattern.token) {
			return pattern.category, pattern.name
		}
	}
	return "", ""
}
