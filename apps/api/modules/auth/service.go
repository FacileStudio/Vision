package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Service is what is left of Vision's authentication after porte took the
// credential: the profile lookup the rest of the app reads.
type Service struct {
	orm        *gorm.DB
	sessions   *session.Manager
	users      kit.Resolver
	logger     *slog.Logger
	controller *Controller
}

// NewService wires a Service onto porte's session manager and the kit store that turns a
// session's user id into an identity.
func NewService(orm *gorm.DB, sessions *session.Manager, users kit.Resolver, logger *slog.Logger) *Service {
	service := &Service{orm: orm, sessions: sessions, users: users, logger: logger}
	service.controller = newController(service)
	return service
}

// RequireAuth is porte/kit's session middleware: porte authenticates the credential and the
// kit reads the account behind it. A session whose user is gone answers 401.
func (service *Service) RequireAuth(next http.Handler) http.Handler {
	return kit.RequireAuth(service.sessions, service.users)(next)
}

// AuthenticateRequest resolves the caller of a route that is not mounted
// behind RequireAuth — the inline-image endpoint, which a browser reaches with
// an <img src> and therefore with a cookie and no header.
func (service *Service) AuthenticateRequest(w http.ResponseWriter, r *http.Request) (int64, error) {
	identity, err := service.sessions.Authenticate(w, r)
	if err != nil {
		return 0, err
	}
	return identity.UserID, nil
}

// Sessions exposes the manager for the modules that list or revoke tokens.
func (service *Service) Sessions() *session.Manager { return service.sessions }

// AuthenticateToken resolves a credential this app received somewhere other
// than a header, and hands it to porte as the bearer token it is.
//
// GET /events/{siteId}/live is the one caller: it is an EventSource, and the
// browser's EventSource cannot set request headers, so the token arrives as a
// query parameter. porte's own middleware refuses a credential in the query
// string — there is a test asserting ?token= authenticates nobody — and that
// rule is right for every route that has a choice. This one does not, so the
// app opts in explicitly here rather than porte relaxing it for everybody, and
// the expiry and idle rules stay porte's instead of being reimplemented.
func (service *Service) AuthenticateToken(w http.ResponseWriter, r *http.Request, token string) (int64, error) {
	bearer := r.Clone(r.Context())
	bearer.Header.Set("Authorization", "Bearer "+token)
	return service.AuthenticateRequest(w, bearer)
}

// UpdateProfile changes the name. The address is Registre's: it is the key sign-in links
// accounts by, so a user who could rewrite it could squat another person's address and
// lock them out of linking.
func (service *Service) UpdateProfile(ctx context.Context, userID int64, name string) (*schemas.User, error) {
	var record schemas.User
	if err := service.orm.WithContext(ctx).First(&record, userID).Error; err != nil {
		return nil, errors.NotFound("user not found")
	}
	record.Name = name
	if err := service.orm.WithContext(ctx).Model(&record).Update("name", name).Error; err != nil {
		return nil, errors.Internal("failed to update user", err)
	}
	return &record, nil
}

// GetUser is the profile the /auth/me routes render.
func (service *Service) GetUser(ctx context.Context, userID int64) (*schemas.User, error) {
	var record schemas.User
	if err := service.orm.WithContext(ctx).First(&record, userID).Error; err != nil {
		return nil, errors.NotFound("user not found")
	}
	return &record, nil
}

// getUserByString and updateProfileByString adapt the decimal-string user id
// the controllers still carry to the int64 porte resolved. Keeping the string
// at the controller boundary leaves the response shapes untouched.
func (service *Service) getUserByString(ctx context.Context, userID string) (*schemas.User, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}
	return service.GetUser(ctx, id)
}

func (service *Service) updateProfileByString(ctx context.Context, userID, name string) (*schemas.User, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}
	return service.UpdateProfile(ctx, id, name)
}
