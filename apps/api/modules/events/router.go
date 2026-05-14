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

var pixel = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00,
	0x80, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x21,
	0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
	0x01, 0x00, 0x3b,
}

func RegisterRoutes(router chi.Router, service *Service, hub *Hub, tracker *ActiveTracker, authService middleware.Authenticator, orm *gorm.DB) {
	router.Route("/e", func(router chi.Router) {
		router.Post("/p", func(w http.ResponseWriter, request *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")

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
			tracker.Touch(site.ID, req.VisitorID)

			httpjson.WriteJSON(w, http.StatusNoContent, nil)
		})

		router.Get("/p", func(w http.ResponseWriter, request *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Cache-Control", "no-cache, no-store")

			dataParam := request.URL.Query().Get("data")
			if dataParam == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var req PageviewRequest
			if err := json.Unmarshal([]byte(dataParam), &req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			origin := req.Hostname
			if origin == "" {
				origin = extractHost(request.Header.Get("Referer"))
			}
			site, err := service.resolveSiteByOrigin(request.Context(), "", "https://"+origin)
			if err != nil {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			userAgent := request.UserAgent()
			country := request.Header.Get("CF-IPCountry")
			_ = service.recordPageview(request.Context(), site, &req, userAgent, country)
			tracker.Touch(site.ID, req.VisitorID)

			w.Header().Set("Content-Type", "image/gif")
			w.Write(pixel)
		})

		router.Options("/p", func(w http.ResponseWriter, request *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusNoContent)
		})

		router.Get("/h", func(w http.ResponseWriter, request *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Cache-Control", "no-cache, no-store")

			dataParam := request.URL.Query().Get("data")
			if dataParam == "" {
				w.Header().Set("Content-Type", "image/gif")
				w.Write(pixel)
				return
			}

			var req struct {
				Hostname  string `json:"hostname"`
				VisitorID string `json:"visitor_id"`
			}
			if err := json.Unmarshal([]byte(dataParam), &req); err != nil {
				w.Header().Set("Content-Type", "image/gif")
				w.Write(pixel)
				return
			}

			host := req.Hostname
			if host == "" {
				host = extractHost(request.Header.Get("Referer"))
			}
			if host != "" {
				site, err := service.resolveSiteByOrigin(request.Context(), "", "https://"+host)
				if err == nil {
					tracker.Touch(site.ID, req.VisitorID)
				}
			}

			w.Header().Set("Content-Type", "image/gif")
			w.Write(pixel)
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

			rc := http.NewResponseController(w)
			_ = rc.SetWriteDeadline(time.Time{})

			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")
			w.Header().Set("X-Accel-Buffering", "no")
			if err := rc.Flush(); err != nil {
				httpjson.WriteError(w, errors.Internal("streaming not supported", err))
				return
			}

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
					_ = rc.Flush()
				case <-keepalive.C:
					fmt.Fprintf(w, ": keepalive\n\n")
					_ = rc.Flush()
				}
			}
		})
	})
}
