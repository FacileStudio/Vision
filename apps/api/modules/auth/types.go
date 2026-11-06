package auth

// AuthResponse carries the user id and token back from the CLI code exchange.
type AuthResponse struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

// UpdateProfileRequest is the payload for editing the profile.
type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ProfileResponse is a user profile safe to return to a client.
type ProfileResponse struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	Name         string `json:"name"`
	AvatarURL    string `json:"avatar_url"`
	AvatarSource string `json:"avatar_source"`
	Locale       string `json:"locale"`
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
