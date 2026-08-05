package apikeys

import documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "api-keys",
	Description: "API key management for programmatic access.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/api-keys",
			Summary:      "Create an API key",
			Description:  "Generates a new API key. The full key is returned only once.",
			Auth:         "bearer",
			RequestBody:  "CreateRequest",
			ResponseBody: "CreateResponse",
		},
		{
			Method:       "GET",
			Path:         "/api-keys",
			Summary:      "List API keys",
			Description:  "Returns all API keys for the authenticated user. Key values are not included.",
			Auth:         "bearer",
			ResponseBody: "[]APIKeyResponse",
		},
		{
			Method:      "DELETE",
			Path:        "/api-keys/{id}",
			Summary:     "Revoke an API key",
			Description: "Deactivates an API key immediately.",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "id", Type: "int", Description: "API Key ID"}},
		},
	},
}
