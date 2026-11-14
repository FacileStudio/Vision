package siteaccess

import (
	"context"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Vision/apps/api/internal/authcontext"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// KeyAllowsWorkspace reports whether the API key on ctx may touch the workspace. A session
// and an unscoped key may. A site key may not: it sees one site, never a space.
func KeyAllowsWorkspace(ctx context.Context, workspaceID int64) bool {
	key := authcontext.KeyFromContext(ctx)
	if key == nil {
		return true
	}
	if key.SiteID != nil {
		return false
	}
	return key.WorkspaceID == nil || *key.WorkspaceID == workspaceID
}

// KeyAllowsSite reports whether the API key on ctx may touch the site, given the
// workspace that owns it.
func KeyAllowsSite(ctx context.Context, siteID int64, workspaceID int64) bool {
	key := authcontext.KeyFromContext(ctx)
	if key == nil {
		return true
	}
	if key.SiteID != nil && *key.SiteID != siteID {
		return false
	}
	return key.WorkspaceID == nil || *key.WorkspaceID == workspaceID
}

// ScopeSQL returns an AND clause and its arguments that confine a query over sites to
// the API key on ctx, and an empty clause for a session. siteColumn and workspaceColumn
// name the site id and workspace id columns as the query spells them.
func ScopeSQL(ctx context.Context, siteColumn string, workspaceColumn string) (string, []any) {
	key := authcontext.KeyFromContext(ctx)
	if key == nil {
		return "", nil
	}
	clause := ""
	var args []any
	if key.SiteID != nil {
		clause += " AND " + siteColumn + " = ?"
		args = append(args, *key.SiteID)
	}
	if key.WorkspaceID != nil {
		clause += " AND " + workspaceColumn + " = ?"
		args = append(args, *key.WorkspaceID)
	}
	return clause, args
}

// Restrict confines a query over the sites table to the API key on ctx.
func Restrict(ctx context.Context, query *gorm.DB) *gorm.DB {
	clause, args := ScopeSQL(ctx, "sites.id", "sites.workspace_id")
	if clause == "" {
		return query
	}
	return query.Where(clause[len(" AND "):], args...)
}

// KeyGuard refuses with 403 an API key that is not allowed to touch the site named by the
// URL parameter. Sessions pass untouched. A site that does not exist passes too, so the
// handler answers 404 as before.
func KeyGuard(orm *gorm.DB, param string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
			key := authcontext.KeyFromContext(request.Context())
			if key == nil {
				next.ServeHTTP(w, request)
				return
			}
			siteID, err := strconv.ParseInt(chi.URLParam(request, param), 10, 64)
			if err != nil {
				next.ServeHTTP(w, request)
				return
			}
			workspaceID := int64(-1)
			if key.WorkspaceID != nil {
				var found []int64
				orm.WithContext(request.Context()).Table("sites").Where("id = ?", siteID).Pluck("workspace_id", &found)
				if len(found) == 0 {
					next.ServeHTTP(w, request)
					return
				}
				workspaceID = found[0]
			}
			if !KeyAllowsSite(request.Context(), siteID, workspaceID) {
				httpjson.WriteError(w, errors.Forbidden("api key is not scoped to this site"))
				return
			}
			next.ServeHTTP(w, request)
		})
	}
}
