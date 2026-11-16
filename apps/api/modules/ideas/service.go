package ideas

import (
	"context"
	"net/http"
	"strconv"

	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"gorm.io/gorm"
)

// Plans answers whether a person or a Registre space has ai.access.
type Plans interface {
	HasAI(ctx context.Context, userID int64, registreSpace string) (bool, error)
}

type scopeKey struct{}

type scope struct {
	space    string
	aiAccess bool
}

// Service resolves the space and the plan of every request that spends AI.
type Service struct {
	db  *gorm.DB
	cfg Config
}

// NewService builds the service over the database and the Bulle wiring.
func NewService(db *gorm.DB, cfg Config) *Service {
	return &Service{db: db, cfg: cfg}
}

func scopeOf(r *http.Request) scope {
	resolved, _ := r.Context().Value(scopeKey{}).(scope)
	return resolved
}

// resolve reads the active workspace from ?space_id=, a workspace the caller belongs to,
// and the plan behind it, and stores both on the request.
func (s *Service) resolve(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := kit.FromContext(r.Context())
		if !ok {
			httpjson.WriteError(w, errors.Unauthorized("missing auth"))
			return
		}
		registre, err := s.requestedSpace(r, identity.UserID)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		resolved, err := s.withPlan(r.Context(), identity.UserID, registre)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), scopeKey{}, resolved)))
	})
}

func (s *Service) requestedSpace(r *http.Request, userID int64) (string, error) {
	raw := r.URL.Query().Get("space_id")
	if raw == "" {
		return "", nil
	}
	workspaceID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || workspaceID <= 0 {
		return "", errors.Invalid("invalid space_id")
	}
	return s.registreSpace(r.Context(), userID, "w.id = ?", workspaceID)
}

// registreSpace is the Registre space id of the one workspace the where clause picks, for
// a person who belongs to it. A personal workspace, with one member and no link, answers
// the empty id: the person's own plan applies. A shared workspace not linked to Registre
// yet has no plan to read, so it is refused.
func (s *Service) registreSpace(ctx context.Context, userID int64, where string, arg any) (string, error) {
	var rows []struct {
		Registre *string
		Members  int64
	}
	err := s.db.WithContext(ctx).Table("workspaces w").
		Select("w.registre_space_id AS registre, (SELECT count(*) FROM workspace_members c WHERE c.workspace_id = w.id) AS members").
		Joins("JOIN workspace_members m ON m.workspace_id = w.id").
		Where(where+" AND m.user_id = ?", arg, userID).Limit(1).Scan(&rows).Error
	if err != nil {
		return "", errors.Internal("failed to load the space", err)
	}
	if len(rows) == 0 {
		return "", errors.NotFound("space not found")
	}
	if rows[0].Registre != nil && *rows[0].Registre != "" {
		return *rows[0].Registre, nil
	}
	if rows[0].Members > 1 {
		return "", errors.Failed("this space is not linked to Registre yet, so Bulle cannot work for it")
	}
	return "", nil
}

func (s *Service) withPlan(ctx context.Context, userID int64, registre string) (scope, error) {
	if s.cfg.Plans == nil {
		return scope{space: registre, aiAccess: true}, nil
	}
	allowed, err := s.cfg.Plans.HasAI(ctx, userID, registre)
	if err != nil {
		return scope{}, errors.Internal("failed to read the plan", err)
	}
	return scope{space: registre, aiAccess: allowed}, nil
}
