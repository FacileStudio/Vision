package workspaces

import "testing"

func TestAssignableRoleRejectsOwner(t *testing.T) {
	for role, want := range map[string]bool{
		"owner":  false,
		"admin":  true,
		"editor": true,
		"viewer": true,
		"root":   false,
		"":       false,
	} {
		if got := assignableRole(role); got != want {
			t.Errorf("assignableRole(%q) = %v, want %v", role, got, want)
		}
	}
	if !validRoles["owner"] {
		t.Error("owner must be a recognized role")
	}
}
