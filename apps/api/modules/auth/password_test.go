package auth

import (
	"net/http"
	"testing"
)

// An account with no password gets one from new_password alone. This is the
// branch porte v0.3.0 split SetPassword down to, and before the split Vision
// could not reach it: the controller required current_password, so a federated
// user could never add a password and confirming one they never set answered
// "current password is incorrect".
func TestSettingAFirstPasswordOnAnSSOAccount(t *testing.T) {
	api := newTestAPI(t)
	token := api.federatedAccount("sso@facile.studio")

	recorder := api.setPassword(token, "", "correct-horse")
	if recorder.Code != http.StatusOK {
		t.Fatalf("set a first password: %d %s", recorder.Code, recorder.Body.String())
	}

	var response PasswordResponse
	api.decode(recorder, &response)
	if response.Token != "" {
		t.Error("a first password rotated the session, but there was no session minted by an older password to end")
	}
	if !api.authenticates(token) {
		t.Error("the caller was signed out by setting a first password")
	}
	if api.login("sso@facile.studio", "correct-horse") == "" {
		t.Fatal("the new password does not sign in")
	}
}

// Replacing a password confirms the current one, rotates the caller's session
// and hands back the new token. The identity it reads is keyed on the account
// id, so this also goes red if the re-key migration is dropped from porteSchema
// and an address-keyed row is all porte/local can find.
func TestChangingAPasswordRotatesTheCallersToken(t *testing.T) {
	api := newTestAPI(t)
	token, _ := api.register("camille@facile.studio", "first-password")

	recorder := api.setPassword(token, "first-password", "second-password")
	if recorder.Code != http.StatusOK {
		t.Fatalf("change the password: %d %s", recorder.Code, recorder.Body.String())
	}

	var response PasswordResponse
	api.decode(recorder, &response)
	if response.Token == "" {
		t.Fatal("no rotated token in the body, so the browser that changed the password keeps sending a revoked one")
	}
	if len(recorder.Result().Cookies()) == 0 {
		t.Error("porte set no session cookie, so a cookie-authenticated client is signed out")
	}

	if api.authenticates(token) {
		t.Error("the token that made the change still authenticates")
	}
	if !api.authenticates(response.Token) {
		t.Error("the rotated token does not authenticate")
	}
	if api.login("camille@facile.studio", "second-password") == "" {
		t.Fatal("the new password does not sign in")
	}
}

// The other browsers end and a named API token does not. porte's RevokeLogins
// spares a labelled session on purpose: replacing a password should not
// silently break every integration holding a key.
func TestChangingAPasswordEndsTheOtherLoginsButNotAPITokens(t *testing.T) {
	api := newTestAPI(t)
	caller, userID := api.register("noah@facile.studio", "first-password")
	other := api.login("noah@facile.studio", "first-password")

	named, _, err := api.service.Issue(t.Context(), userID, "ci")
	if err != nil {
		t.Fatalf("issue a named token: %v", err)
	}

	recorder := api.setPassword(caller, "first-password", "second-password")
	if recorder.Code != http.StatusOK {
		t.Fatalf("change the password: %d %s", recorder.Code, recorder.Body.String())
	}

	if api.authenticates(other) {
		t.Error("the other browser survived the password change")
	}
	if !api.authenticates(named) {
		t.Error("a named API token was revoked by a password change")
	}
}

// A wrong current password is 401 and changes nothing, including the caller's
// own session.
func TestChangingAPasswordWithTheWrongCurrentOne(t *testing.T) {
	api := newTestAPI(t)
	token, _ := api.register("wrong@facile.studio", "first-password")

	recorder := api.setPassword(token, "not-the-password", "second-password")
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d %s", recorder.Code, recorder.Body.String())
	}
	if code := api.errorCode(recorder); code != "unauthenticated" {
		t.Errorf("error code is %q", code)
	}
	if !api.authenticates(token) {
		t.Error("a refused change ended the caller's session anyway")
	}
	if api.login("wrong@facile.studio", "first-password") == "" {
		t.Fatal("the old password stopped working after a refused change")
	}
}

// new_password alone against an account that has one is 400 naming the missing
// field, not porte's 409. From inside the kit ErrPasswordSet is a race; from
// here it is a caller that left current_password out of the body.
func TestSettingAPasswordOnAnAccountThatHasOneNamesTheMissingField(t *testing.T) {
	api := newTestAPI(t)
	token, _ := api.register("taken@facile.studio", "first-password")

	recorder := api.setPassword(token, "", "second-password")
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d %s", recorder.Code, recorder.Body.String())
	}
	if code := api.errorCode(recorder); code != "invalid_argument" {
		t.Errorf("error code is %q", code)
	}
	if api.login("taken@facile.studio", "first-password") == "" {
		t.Fatal("a refused set changed the password anyway")
	}
}
