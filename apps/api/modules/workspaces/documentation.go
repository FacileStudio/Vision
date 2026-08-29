package workspaces

import (
	"net/http"

	documentation "github.com/FacileStudio/Vision/apps/api/internal/documentation"
)

var Documentation = documentation.Module{
	Name:        "workspaces",
	Description: "Workspace organization and team membership.",
	Routes: []documentation.Route{
		{
			Method:       "POST",
			Path:         "/workspaces",
			Summary:      "Create workspace",
			Auth:         "bearer",
			RequestBody:  CreateRequest{},
			ResponseBody: WorkspaceResponse{},
			Status:       http.StatusCreated,
		},
		{
			Method:       "GET",
			Path:         "/workspaces",
			Summary:      "List accessible workspaces",
			Auth:         "bearer",
			ResponseBody: []WorkspaceResponse{},
		},
		{
			Method:       "GET",
			Path:         "/workspaces/{id}",
			Summary:      "Get workspace by ID",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "string", Description: "Workspace ID"}},
			ResponseBody: WorkspaceResponse{},
		},
		{
			Method:       "PUT",
			Path:         "/workspaces/{id}",
			Summary:      "Update workspace",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "string", Description: "Workspace ID"}},
			RequestBody:  UpdateRequest{},
			ResponseBody: WorkspaceResponse{},
		},
		{
			Method:     "DELETE",
			Path:       "/workspaces/{id}",
			Summary:    "Delete workspace",
			Auth:       "bearer",
			PathParams: []documentation.Field{{Name: "id", Type: "string", Description: "Workspace ID"}},
			Status:     http.StatusNoContent,
		},
		{
			Method:       "GET",
			Path:         "/workspaces/{id}/members",
			Summary:      "List workspace members",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "string", Description: "Workspace ID"}},
			ResponseBody: []MemberResponse{},
		},
		{
			Method:       "POST",
			Path:         "/workspaces/{id}/members",
			Summary:      "Add member to workspace",
			Auth:         "bearer",
			PathParams:   []documentation.Field{{Name: "id", Type: "string", Description: "Workspace ID"}},
			RequestBody:  AddMemberRequest{},
			ResponseBody: MemberResponse{},
			Status:       http.StatusCreated,
		},
		{
			Method:      "PUT",
			Path:        "/workspaces/{id}/members/{userId}",
			Summary:     "Update workspace member role",
			Auth:        "bearer",
			PathParams:  []documentation.Field{{Name: "id", Type: "string", Description: "Workspace ID"}, {Name: "userId", Type: "string", Description: "User ID"}},
			RequestBody: UpdateMemberRequest{},
			Status:      http.StatusNoContent,
		},
		{
			Method:     "DELETE",
			Path:       "/workspaces/{id}/members/{userId}",
			Summary:    "Remove member from workspace",
			Auth:       "bearer",
			PathParams: []documentation.Field{{Name: "id", Type: "string", Description: "Workspace ID"}, {Name: "userId", Type: "string", Description: "User ID"}},
			Status:     http.StatusNoContent,
		},
		{
			Method:     "POST",
			Path:       "/workspaces/{id}/leave",
			Summary:    "Leave workspace",
			Auth:       "bearer",
			PathParams: []documentation.Field{{Name: "id", Type: "string", Description: "Workspace ID"}},
			Status:     http.StatusNoContent,
		},
	},
}
