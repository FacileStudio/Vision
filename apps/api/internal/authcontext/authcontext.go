package authcontext

import "context"

// KeyScope is what an API key is limited to. A nil SiteID and WorkspaceID is a key for
// everything its owner can read.
type KeyScope struct {
	SiteID      *int64
	WorkspaceID *int64
	Write       bool
}

// Identity is the resolved caller: the id and email the rest of Vision reads
// once authentication has happened. Key is nil for a cookie or bearer session and
// set when the caller presented an API key.
type Identity struct {
	UserID string
	Email  string
	Key    *KeyScope
}

type contextKey struct{}

// WithIdentity stores the resolved identity on the request context.
func WithIdentity(parentContext context.Context, identity Identity) context.Context {
	return context.WithValue(parentContext, contextKey{}, identity)
}

// IdentityFromContext reads the identity previously stored by WithIdentity,
// reporting whether one was present.
func IdentityFromContext(parentContext context.Context) (Identity, bool) {
	identity, ok := parentContext.Value(contextKey{}).(Identity)
	return identity, ok
}

// KeyFromContext returns the scope of the API key that authenticated the request, or nil
// for a session.
func KeyFromContext(parentContext context.Context) *KeyScope {
	identity, _ := IdentityFromContext(parentContext)
	return identity.Key
}
