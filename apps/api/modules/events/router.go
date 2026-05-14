package events

import (
	"net/http"
	"strings"

	"api/internal/httpjson"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
	router.Route("/event", func(router chi.Router) {
		router.Post("/pageview", func(w http.ResponseWriter, request *http.Request) {
			apiKey := extractAPIKey(request)
			site, err := service.resolveSite(request.Context(), apiKey)
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
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			httpjson.WriteJSON(w, http.StatusNoContent, nil)
		})

		router.Options("/pageview", func(w http.ResponseWriter, request *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusNoContent)
		})
	})
}

func extractAPIKey(request *http.Request) string {
	auth := request.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return request.URL.Query().Get("key")
}
