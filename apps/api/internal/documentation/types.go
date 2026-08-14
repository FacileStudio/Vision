// Package docs re-exports the shared API reference types so each module can
// declare its routes without importing tronc directly.
package docs

import "github.com/FacileStudio/tronc/apiref"

type (
	// Response is the shared API reference registry.
	Response = apiref.Registry
	// Module is a documented API module.
	Module = apiref.Module
	// Route is a documented API route.
	Route = apiref.Route
	// Field is a documented request or response field.
	Field = apiref.Field
	// Error is a documented API error response.
	Error = apiref.Error
)
