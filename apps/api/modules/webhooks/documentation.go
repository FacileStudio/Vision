package webhooks

import (
	"net/http"

	documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"
)

var Documentation = documentation.Module{
	Name:        "webhooks",
	Description: "Periodic metric reports sent to external HTTP endpoints.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/webhooks",
			Summary:      "Create a report webhook",
			Auth:         "bearer",
			RequestBody:  CreateWebhookRequest{},
			ResponseBody: WebhookResponse{},
			Status:       http.StatusCreated,
		},
		{
			Method:       "GET",
			Path:         "/webhooks",
			Summary:      "List configured webhooks",
			Auth:         "bearer",
			QueryParams:  []documentation.Field{{Name: "workspace_id", Type: "int", Description: "Workspace ID"}},
			ResponseBody: []WebhookResponse{},
		},
		{
			Method:       "GET",
			Path:         "/webhooks/{id}",
			Summary:      "Get webhook details",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int", Description: "Webhook ID"}},
			ResponseBody: WebhookResponse{},
		},
		{
			Method:       "PUT",
			Path:         "/webhooks/{id}",
			Summary:      "Update webhook",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int", Description: "Webhook ID"}},
			RequestBody:  UpdateWebhookRequest{},
			ResponseBody: WebhookResponse{},
		},
		{
			Method:     "DELETE",
			Path:       "/webhooks/{id}",
			Summary:    "Delete webhook",
			Auth:       "bearer",
			PathParams: []documentation.Field{{Name: "id", Type: "int", Description: "Webhook ID"}},
			Status:     http.StatusNoContent,
		},
		{
			Method:       "POST",
			Path:         "/webhooks/{id}/test",
			Summary:      "Send test report webhook",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "int", Description: "Webhook ID"}},
			ResponseBody: map[string]string{"status": "sent"},
		},
	},
}
