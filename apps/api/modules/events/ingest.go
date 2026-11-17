package events

import (
	"net/http"
	"time"

	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/tronc/httpjson"
)

type ingestor struct {
	service *Service
	tracker *ActiveTracker
}

// botOf classifies the request's user agent together with the tracker's flags.
func botOf(request *http.Request, flags int) (category, name string) {
	return classifyBot(request.UserAgent(), flags, time.Now())
}

// recordVisit records a human pageview and marks the visitor active, or counts a bot hit
// instead when classifyBot names the sender.
func (in ingestor) recordVisit(request *http.Request, site *schemas.Site, req *PageviewRequest) error {
	if category, name := botOf(request, req.Flags); category != "" {
		return in.service.recordBotHit(request.Context(), site, req.Path, category, name)
	}
	if err := in.service.recordPageview(request.Context(), site, req, request.UserAgent(), request.Header.Get("CF-IPCountry")); err != nil {
		return err
	}
	in.tracker.Touch(site.ID, req.VisitorID)
	return nil
}

// postPageview answers POST /e/p: the site comes from the Origin or Referer header.
func (in ingestor) postPageview(w http.ResponseWriter, request *http.Request) {
	site, err := in.service.resolveSiteByOrigin(request.Context(), request.Header.Get("Origin"), request.Header.Get("Referer"))
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	var req PageviewRequest
	if err := httpjson.DecodeJSON(w, request, &req); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if err := in.recordVisit(request, site, &req); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusNoContent, nil)
}

// pixelPageview answers GET /e/p: a pageview, or with type=perf the timings of the last one.
func (in ingestor) pixelPageview(w http.ResponseWriter, request *http.Request) {
	var req PageviewRequest
	if !decodePixelData(request, &req) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if request.URL.Query().Get("type") == "perf" {
		in.pixelPerformance(request, &req)
		writePixel(w)
		return
	}
	site, err := in.siteForPixel(request, req.Hostname)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	_ = in.recordVisit(request, site, &req)
	writePixel(w)
}

// pixelPerformance stores the timings a human's page reported after load. Bots report none.
func (in ingestor) pixelPerformance(request *http.Request, req *PageviewRequest) {
	if category, _ := botOf(request, req.Flags); category != "" {
		return
	}
	if site, err := in.siteForPixel(request, req.Hostname); err == nil {
		in.service.updatePerformance(request.Context(), site, req.VisitorID, req.Path, req.Performance)
	}
}

// pixelEvent answers GET /e/t: a custom event from a human. A bot's event is dropped, and
// every answer is the pixel.
func (in ingestor) pixelEvent(w http.ResponseWriter, request *http.Request) {
	defer writePixel(w)
	var req CustomEventRequest
	if !decodePixelData(request, &req) {
		return
	}
	if category, _ := botOf(request, req.Flags); category != "" {
		return
	}
	if site, err := in.siteForPixel(request, req.Hostname); err == nil {
		_ = in.service.recordCustomEvent(request.Context(), site, &req)
	}
}

// heartbeat answers GET /e/h: it keeps a human visitor active and records the first real
// gesture on their session. A bot's heartbeat is dropped, and every answer is the pixel.
func (in ingestor) heartbeat(w http.ResponseWriter, request *http.Request) {
	defer writePixel(w)
	var req HeartbeatRequest
	if !decodePixelData(request, &req) {
		return
	}
	if category, _ := botOf(request, req.Flags); category != "" {
		return
	}
	site, err := in.siteForPixel(request, req.Hostname)
	if err != nil {
		return
	}
	in.tracker.Touch(site.ID, req.VisitorID)
	if req.Interacted && req.VisitorID != "" {
		_ = in.service.markInteracted(request.Context(), site.ID, req.VisitorID)
	}
}
