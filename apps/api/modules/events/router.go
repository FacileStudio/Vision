package events

import (
	"net/http"

	"api/internal/httpjson"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service) {
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
}
