package analytics

import (
	"net/http"

	"github.com/FacileStudio/Vision/apps/api/internal/siteaccess"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// registerBotRoutes mounts the bot traffic report under the authenticated /analytics
// router.
func registerBotRoutes(router chi.Router, service *Service, orm *gorm.DB) {
	guarded := router.With(siteaccess.KeyGuard(orm, "siteId"))
	guarded.Get("/{siteId}/bots", readableSite(orm, func(w http.ResponseWriter, request *http.Request, siteID int64) {
		from, to := parseDateRange(request)
		resp, err := service.Bots(request.Context(), siteID, from, to)
		if err != nil {
			httpjson.WriteError(w, err)
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, resp)
	}))
}
