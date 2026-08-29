package auth

// RegisterRequest is the payload for creating a local account.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest is the payload for signing in with a password.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse carries the user id and token back from register/login.
type AuthResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

// UpdateProfileRequest is the payload for editing the profile.
type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ChangePasswordRequest carries the new password and, when the account already
// has one, the current password confirming it. An account with no password —
// an SSO-only user adding one — sends new_password alone.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// PasswordResponse carries the rotated session token.
//
// Replacing a password ends the account's other logins and rotates the
// caller's own, and porte puts the new session in the cookie. Vision's client
// authenticates with a bearer out of localStorage as well, so without the
// token here the browser that changed the password would keep sending one
// porte revoked mid-request. Setting a first password rotates nothing and
// therefore omits it.
type PasswordResponse struct {
	Status string `json:"status"`
	Token  string `json:"token,omitempty"`
}

// ProfileResponse is a user profile safe to return to a client.
type ProfileResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	AvatarURL    string `json:"avatar_url"`
	AvatarSource string `json:"avatar_source"`
	CreatedAt    string `json:"created_at"`
}

// Data carries the login email used by porte's callback resolution.
type Data struct {
	Email string `json:"email"`
}

func (d *Data) GetEmail() string {
	if d == nil {
		return ""
	}
	return d.Email
}
