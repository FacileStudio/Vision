package sites

import documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"

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
			RequestBody:  "CreateRequest",
			ResponseBody: "SiteResponse",
		},
		{
			Method:       "GET",
			Path:         "/sites",
			Summary:      "List sites",
			Description:  "Returns all sites owned by the authenticated user.",
			Auth:         "bearer",
			ResponseBody: "[]SiteResponse",
		},
		{
			Method:       "GET",
			Path:         "/sites/{id}",
			Summary:      "Get a site",
			Auth:         "bearer",
			ResponseBody: "SiteResponse",
			PathParams:   []documentation.Field{{Name: "id", Type: "int", Description: "Site ID"}},
		},
		{
			Method:      "PUT",
			Path:        "/sites/{id}",
			Summary:     "Update a site",
			Auth:        "bearer",
			RequestBody: "UpdateRequest",
			PathParams:  []documentation.Field{{Name: "id", Type: "int", Description: "Site ID"}},
		},
		{
			Method:     "DELETE",
			Path:       "/sites/{id}",
			Summary:    "Delete a site",
			Auth:       "bearer",
			PathParams: []documentation.Field{{Name: "id", Type: "int", Description: "Site ID"}},
		},
	},
}
