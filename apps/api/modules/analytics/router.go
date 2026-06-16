package analytics

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"api/internal/authcontext"
	"api/internal/errors"
	"api/internal/httpjson"
	"api/internal/middleware"
	"api/internal/siteaccess"
	"api/modules/events"
	"api/schemas"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func RegisterRoutes(router chi.Router, service *Service, tracker *events.ActiveTracker, orm *gorm.DB, authService middleware.Authenticator) {
	router.Get("/share/{token}", func(w http.ResponseWriter, request *http.Request) {
		token := chi.URLParam(request, "token")

		var site schemas.Site
		if err := orm.Where("share_token = ?", token).First(&site).Error; err != nil {
			httpjson.WriteError(w, errors.NotFound("not found"))
			return
		}

		from, to := parseDateRange(request)
		granularity := request.URL.Query().Get("granularity")
		if granularity == "" {
			granularity = "day"
		}
		resp, err := service.Overview(request.Context(), site.ID, from, to, granularity, parseFilters(request))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}

		httpjson.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"site": map[string]interface{}{
				"name":   site.Name,
				"domain": site.Domain,
			},
			"overview": resp,
		})
	})

	router.Get("/share/{token}/realtime", func(w http.ResponseWriter, request *http.Request) {
		token := chi.URLParam(request, "token")

		var site schemas.Site
		if err := orm.Where("share_token = ?", token).First(&site).Error; err != nil {
			httpjson.WriteError(w, errors.NotFound("not found"))
			return
		}

		visitors := tracker.Count(site.ID)
		httpjson.WriteJSON(w, http.StatusOK, map[string]int64{"visitors": visitors})
	})

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
			if !siteaccess.CanAccess(request.Context(), orm, ownerID, siteID) {
				httpjson.WriteError(w, errors.NotFound("site not found"))
				return
			}

			from, to := parseDateRange(request)
			granularity := request.URL.Query().Get("granularity")
			if granularity == "" {
				granularity = "day"
			}
			resp, err := service.Overview(request.Context(), siteID, from, to, granularity, parseFilters(request))
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
			if !siteaccess.CanAccess(request.Context(), orm, ownerID, siteID) {
				httpjson.WriteError(w, errors.NotFound("site not found"))
				return
			}

			visitors := tracker.Count(siteID)
			httpjson.WriteJSON(w, http.StatusOK, map[string]int64{"visitors": visitors})
		})

		router.Get("/{siteId}/export", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			siteID, err := strconv.ParseInt(chi.URLParam(request, "siteId"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid site ID"))
				return
			}

			ownerID, _ := strconv.ParseInt(identity.UserID, 10, 64)
			if !siteaccess.CanAccess(request.Context(), orm, ownerID, siteID) {
				httpjson.WriteError(w, errors.NotFound("site not found"))
				return
			}

			from, to := parseDateRange(request)
			granularity := request.URL.Query().Get("granularity")
			if granularity == "" {
				granularity = "day"
			}
			resp, err := service.Overview(request.Context(), siteID, from, to, granularity, parseFilters(request))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}

			format := request.URL.Query().Get("format")
			if format == "csv" {
				w.Header().Set("Content-Type", "text/csv")
				w.Header().Set("Content-Disposition", "attachment; filename=vision-export.csv")
				writeCSV(w, resp)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", "attachment; filename=vision-export.json")
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
}

func writeCSV(w http.ResponseWriter, resp *OverviewResponse) {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	visitorsMap := make(map[string]int64)
	for _, v := range resp.UniqueVisitorsPerDay {
		visitorsMap[v.Date] = v.Count
	}

	writer.Write([]string{"Date", "Pageviews", "Unique Visitors"})
	for _, d := range resp.PageviewsPerDay {
		writer.Write([]string{
			d.Date,
			fmt.Sprintf("%d", d.Count),
			fmt.Sprintf("%d", visitorsMap[d.Date]),
		})
	}
}

func parseFilters(request *http.Request) Filters {
	return Filters{
		Country:  request.URL.Query().Get("country"),
		Browser:  request.URL.Query().Get("browser"),
		OS:       request.URL.Query().Get("os"),
		Device:   request.URL.Query().Get("device"),
		Path:     request.URL.Query().Get("path"),
		Referrer: request.URL.Query().Get("referrer"),
	}
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
