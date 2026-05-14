package events

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"api/internal/errors"
	"api/internal/httpjson"
	"api/internal/middleware"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func RegisterRoutes(router chi.Router, service *Service, hub *Hub, authService middleware.Authenticator, orm *gorm.DB) {
	router.Route("/event", func(router chi.Router) {
		router.Post("/pageview", func(w http.ResponseWriter, request *http.Request) {
			origin := request.Header.Get("Origin")
			referer := request.Header.Get("Referer")
			site, err := service.resolveSiteByOrigin(request.Context(), origin, referer)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}

			var req PageviewRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			userAgent := request.UserAgent()
			country := request.Header.Get("CF-IPCountry")

			if err := service.recordPageview(request.Context(), site, &req, userAgent, country); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			httpjson.WriteJSON(w, http.StatusNoContent, nil)
		})

		router.Options("/pageview", func(w http.ResponseWriter, request *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusNoContent)
		})
	})

	router.Route("/events", func(router chi.Router) {
		router.Get("/{siteId}/live", func(w http.ResponseWriter, request *http.Request) {
			authorization := request.Header.Get("Authorization")
			if authorization == "" {
				if t := request.URL.Query().Get("token"); t != "" {
					authorization = "Bearer " + t
				}
			}

			userID, rawData, err := authService.Authenticate(request.Context(), authorization)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			data, ok := rawData.(interface{ GetEmail() string })
			if !ok || data == nil {
				httpjson.WriteError(w, errors.Unauthorized("missing auth"))
				return
			}
			_ = data

			siteID, err := strconv.ParseInt(chi.URLParam(request, "siteId"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid site ID"))
				return
			}

			ownerID, _ := strconv.ParseInt(userID, 10, 64)
			var count int64
			orm.Table("sites").Where("id = ? AND owner_id = ?", siteID, ownerID).Count(&count)
			if count == 0 {
				httpjson.WriteError(w, errors.NotFound("site not found"))
				return
			}

			flusher, ok := w.(http.Flusher)
			if !ok {
				httpjson.WriteError(w, errors.Internal("streaming not supported", nil))
				return
			}

			rc := http.NewResponseController(w)
			_ = rc.SetWriteDeadline(time.Time{})

			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("X-Accel-Buffering", "no")
			flusher.Flush()

			ch := hub.Subscribe(siteID)
			defer hub.Unsubscribe(siteID, ch)

			keepalive := time.NewTicker(30 * time.Second)
			defer keepalive.Stop()

			ctx := request.Context()
			for {
				select {
				case <-ctx.Done():
					return
				case event := <-ch:
					data, _ := json.Marshal(event)
					fmt.Fprintf(w, "data: %s\n\n", data)
					flusher.Flush()
				case <-keepalive.C:
					fmt.Fprintf(w, ": keepalive\n\n")
					flusher.Flush()
				}
			}
		})
	})
}
