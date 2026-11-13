package analytics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/FacileStudio/Vision/apps/api/internal/authcontext"
	"github.com/FacileStudio/Vision/apps/api/modules/events"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// parseSpace reads the summary's space filter: empty or "all" is every workspace (0),
// anything else must be a positive workspace id.
func parseSpace(raw string) (int64, error) {
	if raw == "" || raw == "all" {
		return 0, nil
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.Invalid("space must be all or a workspace id")
	}
	return id, nil
}

// parseSummaryRange reads from and to as RFC 3339 instants or YYYY-MM-DD dates, on top
// of the 30-day default the other analytics routes use.
func parseSummaryRange(request *http.Request) (time.Time, time.Time) {
	from, to := parseDateRange(request)
	if t, err := time.Parse(time.RFC3339, request.URL.Query().Get("from")); err == nil {
		from = t.UTC()
	}
	if t, err := time.Parse(time.RFC3339, request.URL.Query().Get("to")); err == nil {
		to = t.UTC()
	}
	return from, to
}

// summaryHandler answers GET /analytics/summary for the signed-in caller, adding each
// site's live visitor count from the in-memory tracker.
func summaryHandler(service *Service, tracker *events.ActiveTracker) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		userID, err := strconv.ParseInt(identity.UserID, 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Unauthorized("unauthorized"))
			return
		}
		workspaceID, err := parseSpace(request.URL.Query().Get("space"))
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		from, to := parseSummaryRange(request)
		resp, err := service.Summary(request.Context(), userID, workspaceID, from, to)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		for i := range resp.Sites {
			resp.Sites[i].Live = tracker.Count(resp.Sites[i].SiteID)
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	}
}
