package analytics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/FacileStudio/Vision/apps/api/internal/authcontext"
	"github.com/FacileStudio/Vision/apps/api/internal/siteaccess"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

const compareDays = 7

// AnomaliesResponse is GET /analytics/{siteId}/anomalies.
type AnomaliesResponse struct {
	Anomalies []Anomaly `json:"anomalies"`
}

// registerInsightRoutes mounts the anomaly list and the period comparison under the
// authenticated /analytics router.
func registerInsightRoutes(router chi.Router, service *Service, orm *gorm.DB) {
	guarded := router.With(siteaccess.KeyGuard(orm, "siteId"))
	guarded.Get("/{siteId}/anomalies", readableSite(orm, func(w http.ResponseWriter, request *http.Request, siteID int64) {
		query, err := anomalyQueryOf(request)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		list, err := service.ListAnomalies(request.Context(), siteID, query)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, AnomaliesResponse{Anomalies: list})
	}))
	guarded.Get("/{siteId}/compare", readableSite(orm, func(w http.ResponseWriter, request *http.Request, siteID int64) {
		current, previous, limit, err := compareParams(request, time.Now())
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		resp, err := service.Compare(request.Context(), siteID, current, previous, limit)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	}))
}

// readableSite runs next for a site the caller can read, and answers 404 otherwise.
func readableSite(orm *gorm.DB, next func(http.ResponseWriter, *http.Request, int64)) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		identity, _ := authcontext.IdentityFromContext(request.Context())
		siteID, err := strconv.ParseInt(chi.URLParam(request, "siteId"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid site ID"))
			return
		}
		userID, _ := strconv.ParseInt(identity.UserID, 10, 64)
		if !siteaccess.CanAccess(request.Context(), orm, userID, siteID) {
			httpjson.WriteError(w, errors.NotFound("site not found"))
			return
		}
		next(w, request, siteID)
	}
}

// ParseDay reads a YYYY-MM-DD day in UTC.
func ParseDay(raw string, name string) (time.Time, error) {
	day, err := time.Parse(dayKeyLayout, raw)
	if err != nil {
		return time.Time{}, errors.Invalid(name + " must be a date such as 2026-11-01")
	}
	return day, nil
}

func anomalyQueryOf(request *http.Request) (AnomalyQuery, error) {
	values := request.URL.Query()
	query := AnomalyQuery{Metric: values.Get("metric")}
	var err error
	if raw := values.Get("from"); raw != "" {
		if query.From, err = ParseDay(raw, "from"); err != nil {
			return query, err
		}
	}
	if raw := values.Get("to"); raw != "" {
		if query.To, err = ParseDay(raw, "to"); err != nil {
			return query, err
		}
	}
	if raw := values.Get("limit"); raw != "" {
		if query.Limit, err = strconv.Atoi(raw); err != nil || query.Limit < 1 {
			return query, errors.Invalid("limit must be a positive number")
		}
	}
	return query, nil
}

// compareParams reads from and to (default the last seven days, today included),
// compare_from and compare_to (default the period of equal length right before), and
// limit.
func compareParams(request *http.Request, now time.Time) (Period, Period, int, error) {
	values := request.URL.Query()
	current := Period{From: utcDay(now).AddDate(0, 0, 1-compareDays), To: utcDay(now)}
	if err := readPeriod(values.Get("from"), values.Get("to"), "from", "to", &current); err != nil {
		return current, current, 0, err
	}
	previous := PreviousPeriod(current)
	if err := readPeriod(values.Get("compare_from"), values.Get("compare_to"), "compare_from", "compare_to", &previous); err != nil {
		return current, previous, 0, err
	}
	limit := 0
	if raw := values.Get("limit"); raw != "" {
		var err error
		if limit, err = strconv.Atoi(raw); err != nil || limit < 1 {
			return current, previous, 0, errors.Invalid("limit must be a positive number")
		}
	}
	return current, previous, limit, nil
}

// readPeriod overrides the period with whichever of from and to is given.
func readPeriod(from string, to string, fromName string, toName string, into *Period) error {
	if from != "" {
		day, err := ParseDay(from, fromName)
		if err != nil {
			return err
		}
		into.From = day
	}
	if to != "" {
		day, err := ParseDay(to, toName)
		if err != nil {
			return err
		}
		into.To = day
	}
	return nil
}
