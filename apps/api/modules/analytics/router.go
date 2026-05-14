package analytics

import (
	"net/http"
	"strconv"
	"time"

	"api/internal/authcontext"
	"api/internal/errors"
	"api/internal/httpjson"
	"api/internal/middleware"
	"api/modules/events"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func RegisterRoutes(router chi.Router, service *Service, tracker *events.ActiveTracker, orm *gorm.DB, authService middleware.Authenticator) {
	router.Route("/analytics", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Get("/{siteId}/overview", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			siteID, err := strconv.ParseInt(chi.URLParam(request, "siteId"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid site ID"))
				return
			}

			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)
			var count int64
			orm.Table("sites").Where("id = ? AND owner_id = ?", siteID, ownerID).Count(&count)
			if count == 0 {
				httpjson.WriteError(w, errors.NotFound("site not found"))
				return
			}

			from, to := parseDateRange(request)
			resp, err := service.overview(request.Context(), siteID, from, to)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Get("/{siteId}/realtime", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			siteID, err := strconv.ParseInt(chi.URLParam(request, "siteId"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid site ID"))
				return
			}

			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)
			var count int64
			orm.Table("sites").Where("id = ? AND owner_id = ?", siteID, ownerID).Count(&count)
			if count == 0 {
				httpjson.WriteError(w, errors.NotFound("site not found"))
				return
			}

			visitors := tracker.Count(siteID)
			httpjson.WriteJSON(w, http.StatusOK, map[string]int64{"visitors": visitors})
		})
	})
}

func parseDateRange(request *http.Request) (time.Time, time.Time) {
	now := time.Now().UTC()
	from := now.AddDate(0, 0, -30)
	to := now

	if v := request.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			from = t
		}
	}
	if v := request.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			to = t.Add(24*time.Hour - time.Nanosecond)
		}
	}
	return from, to
}
