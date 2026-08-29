package sites

import (
	"net/http"

	documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"
)

var Documentation = documentation.Module{
	Name:        "sites",
	Description: "Manage tracked websites.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/sites",
			Summary:      "Create a site",
			Description:  "Registers a new website for tracking by domain.",
			Auth:         "bearer",
			RequestBody:  CreateRequest{},
			ResponseBody: SiteResponse{},
			Status:       http.StatusCreated,
		},
		{
			Method:       "GET",
			Path:         "/sites",
			Summary:      "List sites",
			Description:  "Returns all sites owned by the authenticated user.",
			Auth:         "bearer",
			QueryParams:  []documentation.Field{{Name: "workspace_id", Type: "int", Description: "Workspace ID filter"}},
			ResponseBody: []SiteResponse{},
		},
		{
			Method:       "GET",
			Path:         "/sites/{id}",
			Summary:      "Get a site",
			Auth:         "bearer",
			ResponseBody: SiteResponse{},
			PathParams:   []documentation.Field{{Name: "id", Type: "string", Description: "Site ID"}},
		},
		{
			Method:       "PUT",
			Path:         "/sites/{id}",
			Summary:      "Update a site",
			Auth:         "bearer",
			RequestBody:  UpdateRequest{},
			ResponseBody: SiteResponse{},
			PathParams:   []documentation.Field{{Name: "id", Type: "string", Description: "Site ID"}},
		},
		{
			Method:     "DELETE",
			Path:       "/sites/{id}",
			Summary:    "Delete a site",
			Auth:       "bearer",
			PathParams: []documentation.Field{{Name: "id", Type: "string", Description: "Site ID"}},
			Status:     http.StatusNoContent,
		},
		{
			Method:       "POST",
			Path:         "/sites/{id}/share",
			Summary:      "Create public share token for site",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "string", Description: "Site ID"}},
			ResponseBody: map[string]string{"share_token": ""},
		},
		{
			Method:     "DELETE",
			Path:       "/sites/{id}/share",
			Summary:    "Revoke public share token for site",
			Auth:       "bearer",
			PathParams: []documentation.Field{{Name: "id", Type: "string", Description: "Site ID"}},
			Status:     http.StatusNoContent,
		},
	},
}
