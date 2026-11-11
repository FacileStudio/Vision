package events

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Vision/apps/api/internal/middleware"
	"github.com/FacileStudio/Vision/apps/api/internal/siteaccess"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"

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

// TrackerRate and TrackerBurst bound the public /e/* routes per client address: a token
// bucket refilling TrackerRate requests a second, holding TrackerBurst. Production's peak
// was about 5.5 tracker requests a second and 24 in one second, all of it then arriving from
// one address, so each address gets ten times the whole service's peak.
const (
	TrackerRate  = 60
	TrackerBurst = 300
)

// trackerLimiter refuses a client address past the tracker budget with 429. The key is the
// host of RemoteAddr as tronc's real-IP middleware rewrote it, so a forwarded header the
// client wrote never mints a fresh bucket.
func trackerLimiter() func(http.Handler) http.Handler {
	return troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{
		Limit:   TrackerRate,
		Burst:   TrackerBurst,
		MaxKeys: 50000,
		OnRefuse: func(w http.ResponseWriter, _ *http.Request) {
			httpjson.WriteError(w, errors.RateLimited("too many tracker requests"))
		},
	})
}

// trackerHeaders gives every /e/* answer the headers a third-party page needs: any origin
// may read it, nothing caches it, and an answer with no body of its own is JSON.
func trackerHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		header := w.Header()
		header.Set("Access-Control-Allow-Origin", "*")
		header.Set("Cache-Control", "no-cache, no-store")
		header.Set("Content-Type", "application/json")
		next.ServeHTTP(w, request)
	})
}

// answerPreflight allows any origin to send the tracker's GET and POST requests for a day.
func answerPreflight(w http.ResponseWriter, _ *http.Request) {
	header := w.Header()
	header.Set("Access-Control-Allow-Origin", "*")
	header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	header.Set("Access-Control-Allow-Headers", "Content-Type")
	header.Set("Access-Control-Max-Age", "86400")
	w.WriteHeader(http.StatusNoContent)
}

// Preflight answers OPTIONS for every path under prefix before next runs. The router's
// CORS middleware refuses a preflight from an origin it does not list with 403, and every
// site Vision tracks is such an origin.
func Preflight(next http.Handler, prefix string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodOptions && strings.HasPrefix(request.URL.Path, prefix) {
			answerPreflight(w, request)
			return
		}
		next.ServeHTTP(w, request)
	})
}

// RegisterRoutes mounts the event ingestion and live-stream routes.
//
// The live stream accepts the credential as ?token= instead of a header:
// EventSource cannot set request headers. porte refuses a query-string
// credential itself, deliberately, so the exception is opted into here rather
// than relaxed for every route.
func RegisterRoutes(router chi.Router, service *Service, hub *Hub, tracker *ActiveTracker, authService middleware.Authenticator, orm *gorm.DB) {
	router.Route("/e", func(router chi.Router) {
		router.Use(trackerHeaders, trackerLimiter())
		router.Post("/p", func(w http.ResponseWriter, request *http.Request) {
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

			if request.URL.Query().Get("type") == "perf" {
				host := req.Hostname
				if host == "" {
					host = extractHost(request.Header.Get("Referer"))
				}
				if host != "" {
					site, err := service.resolveSiteByOrigin(request.Context(), "", "https://"+host)
					if err == nil {
						service.updatePerformance(request.Context(), site, req.VisitorID, req.Path, req.Performance)
					}
				}
				w.Header().Set("Content-Type", "image/gif")
				w.Write(pixel)
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

		router.Get("/t", func(w http.ResponseWriter, request *http.Request) {

			dataParam := request.URL.Query().Get("data")
			if dataParam == "" {
				w.Header().Set("Content-Type", "image/gif")
				w.Write(pixel)
				return
			}

			var req CustomEventRequest
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
					_ = service.recordCustomEvent(request.Context(), site, &req)
				}
			}

			w.Header().Set("Content-Type", "image/gif")
			w.Write(pixel)
		})

		router.Options("/p", answerPreflight)

		router.Get("/h", func(w http.ResponseWriter, request *http.Request) {

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
			id, err := authService.AuthenticateRequest(w, request)
			if err != nil {
				if raw := request.URL.Query().Get("token"); raw != "" {
					id, err = authService.AuthenticateToken(w, request, raw)
				}
			}
			if err != nil {
				httpjson.WriteError(w, errors.Unauthorized("missing auth"))
				return
			}
			userID := strconv.FormatInt(id, 10)

			siteID, err := strconv.ParseInt(chi.URLParam(request, "siteId"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid site ID"))
				return
			}

			ownerID, _ := strconv.ParseInt(userID, 10, 64)
			if !siteaccess.CanAccess(request.Context(), orm, ownerID, siteID) {
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
