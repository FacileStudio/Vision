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

// ChangePasswordRequest carries the current and new passwords.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
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
