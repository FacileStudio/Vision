package goals

import (
	"net/http"

	documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"
)

var Documentation = documentation.Module{
	Name:        "goals",
	Description: "Goal tracking and conversion rates.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/goals",
			Summary:      "Create a goal",
			Description:  "Creates a new goal for a site. Goal type can be pageview or event.",
			Auth:         "bearer",
			RequestBody:  CreateRequest{},
			ResponseBody: GoalResponse{},
			Status:       http.StatusCreated,
		},
		{
			Method:       "GET",
			Path:         "/goals",
			Summary:      "List goals for site",
			Description:  "Returns all goals for a site.",
			Auth:         "bearer",
			QueryParams:  []documentation.Field{{Name: "site_id", Type: "string", Description: "Site ID"}},
			ResponseBody: []GoalResponse{},
		},
		{
			Method:       "PUT",
			Path:         "/goals/{id}",
			Summary:      "Update a goal",
			Auth:         "bearer",
			RequestBody:  UpdateRequest{},
			ResponseBody: GoalResponse{},
			PathParams:   []documentation.Field{{Name: "id", Type: "string", Description: "Goal ID"}},
		},
		{
			Method:     "DELETE",
			Path:       "/goals/{id}",
			Summary:    "Delete a goal",
			Auth:       "bearer",
			PathParams: []documentation.Field{{Name: "id", Type: "string", Description: "Goal ID"}},
			Status:     http.StatusNoContent,
		},
		{
			Method:       "GET",
			Path:         "/goals/{siteId}/conversions",
			Summary:      "Get goal conversions",
			Description:  "Returns conversion counts and rates for all goals of a site within a date range.",
			Auth:         "bearer",
			QueryParams:  []documentation.Field{{Name: "from", Type: "string", Description: "Start date (YYYY-MM-DD)"}, {Name: "to", Type: "string", Description: "End date (YYYY-MM-DD)"}},
			ResponseBody: ConversionsResponse{},
			PathParams:   []documentation.Field{{Name: "siteId", Type: "string", Description: "Site ID"}},
		},
	},
}
