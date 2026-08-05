package goals

import documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"

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
			RequestBody:  "CreateRequest",
			ResponseBody: "GoalResponse",
		},
		{
			Method:       "GET",
			Path:         "/goals?site_id={siteId}",
			Summary:      "List goals",
			Description:  "Returns all goals for a site.",
			Auth:         "bearer",
			ResponseBody: "[]GoalResponse",
		},
		{
			Method:      "PUT",
			Path:        "/goals/{id}",
			Summary:     "Update a goal",
			Auth:        "bearer",
			RequestBody: "UpdateRequest",
			PathParams:  []documentation.Field{{Name: "id", Type: "int", Description: "Goal ID"}},
		},
		{
			Method:     "DELETE",
			Path:       "/goals/{id}",
			Summary:    "Delete a goal",
			Auth:       "bearer",
			PathParams: []documentation.Field{{Name: "id", Type: "int", Description: "Goal ID"}},
		},
		{
			Method:       "GET",
			Path:         "/goals/{siteId}/conversions",
			Summary:      "Get goal conversions",
			Description:  "Returns conversion counts and rates for all goals of a site within a date range.",
			Auth:         "bearer",
			ResponseBody: "ConversionsResponse",
			PathParams:   []documentation.Field{{Name: "siteId", Type: "int", Description: "Site ID"}},
		},
	},
}
