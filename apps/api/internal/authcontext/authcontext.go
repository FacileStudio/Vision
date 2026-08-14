package authcontext

import "context"

// Identity is the resolved caller: the id and email the rest of Vision reads
// once authentication has happened.
type Identity struct {
	UserID string
	Email  string
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
