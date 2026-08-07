// Package oidcavatar reads the profile claims Authentik returns at login.
//
// It used to download the picture and keep a copy on disk. It no longer does. The URL
// Authentik hands over is public, cacheable and served by Porte, so the copy was a cache
// of something Vision is not the source of — one that a container rebuild silently emptied
// while the database went on pointing at it, which is exactly how every avatar here came
// to 404. Storing the URL removes the copy, the volume it needed, the file-serving route,
// and the SSRF surface that fetching an ID-token-supplied URL from inside our own network
// opened.
//
// Vision has no avatar upload, so the URL is the only source there is.
package oidcavatar

import "strings"

type Profile struct {
	Name              string
	PreferredUsername string
	GivenName         string
	FamilyName        string
	Picture           string
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

// PhotoURL returns the picture claim when it is a photo somebody actually chose, and ""
// when it is not.
//
// Authentik never omits the claim: a user with no photo gets `data:image/svg+xml;base64,…`,
// its own rendering of their initials. Testing `picture != ""` — which every app in the
// suite did — therefore reads "has an avatar" as always true, which is why the old fetch
// logged a failed HTTPS check every five minutes for those users and why no fallback could
// ever be reached. An https URL is a file in Porte's media store; anything else is a
// placeholder the client draws better itself.
func PhotoURL(pictureClaim string) string {
	if strings.HasPrefix(strings.ToLower(pictureClaim), "https://") {
		return pictureClaim
	}
	return ""
}
