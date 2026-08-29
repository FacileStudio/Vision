package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/FacileStudio/Vision/apps/api/internal/env"
	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/local"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/testdb"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

var testConfig = testdb.Config{Prefix: "vision_auth_test", Migrate: schemas.Migrate}

// testAPI is the real /auth router over a real PostgreSQL. It is a type rather
// than a bag of free functions so a request helper keeps its parameter list
// inside filet's cap of five.
type testAPI struct {
	t       *testing.T
	handler http.Handler
	db      *gorm.DB
	service *Service
}

// newTestAPI mounts the module's own RegisterRoutes over porte's stores.
//
// The password paths cannot be exercised below this line. ChangePassword reads
// the caller's session id out of porte.From(ctx) and writes the rotated cookie
// through the ResponseWriter, so a service-level call holding a bare context
// touches neither and passes whatever the code does.
func newTestAPI(t *testing.T) *testAPI {
	t.Helper()
	url, ok := testdb.URL()
	if !ok {
		t.Skip(testdb.SkipReason("createdb vision_test"))
	}
	db, err := testdb.Open(url, testConfig)
	if err != nil {
		t.Fatalf("open the test database: %v", err)
	}
	if err := testdb.Truncate(db, testConfig); err != nil {
		t.Fatalf("truncate: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql handle: %v", err)
	}
	store := portepg.New(sqlDB)
	users := NewUserStore(db)
	logger := slog.New(slog.DiscardHandler)

	sessions, err := session.New(porte.Config{AcceptLegacyCookie: true}, session.Deps{
		Sessions: store.Sessions(),
		Logger:   logger,
	})
	if err != nil {
		t.Fatalf("session manager: %v", err)
	}
	passwords, err := local.New(local.Config{AllowRegistration: true, MinPasswordLength: 8}, local.Deps{
		Users:      users,
		Identities: store.Identities(),
		Sessions:   sessions,
		Logger:     logger,
		Count:      users.CountUsers,
	})
	if err != nil {
		t.Fatalf("local kit: %v", err)
	}

	service := NewService(db, sessions, passwords, logger)
	router := chi.NewRouter()
	RegisterRoutes(router, service, env.Config{})
	return &testAPI{t: t, handler: router, db: db, service: service}
}

// call sends one request as a bearer, which is the transport Vision's client
// uses. A cookie-authenticated mutating request would additionally need porte's
// CSRF header, so a bearer-only test would hide that rule rather than meet it.
func (a *testAPI) call(method, path, token string, body any) *httptest.ResponseRecorder {
	a.t.Helper()
	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			a.t.Fatalf("encode the body: %v", err)
		}
		payload = bytes.NewReader(encoded)
	}
	request := httptest.NewRequest(method, path, payload)
	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	a.handler.ServeHTTP(recorder, request)
	return recorder
}

// authenticates reports whether a token still opens a route behind RequireAuth.
func (a *testAPI) authenticates(token string) bool {
	a.t.Helper()
	return a.call(http.MethodGet, "/auth/me", token, nil).Code == http.StatusOK
}

// setPassword is PUT /auth/password. An empty current takes the first-password
// branch, which is the split porte v0.3.0 introduced.
func (a *testAPI) setPassword(token, current, next string) *httptest.ResponseRecorder {
	a.t.Helper()
	body := map[string]string{"new_password": next}
	if current != "" {
		body["current_password"] = current
	}
	return a.call(http.MethodPut, "/auth/password", token, body)
}

// register creates an account through the real route and returns its bearer
// with the account id. Seeding an identity row instead would decide the subject
// these tests exist to check.
func (a *testAPI) register(email, password string) (string, int64) {
	a.t.Helper()
	recorder := a.call(http.MethodPost, "/auth/register", "", map[string]string{
		"email":    email,
		"password": password,
	})
	if recorder.Code != http.StatusCreated {
		a.t.Fatalf("register %s: %d %s", email, recorder.Code, recorder.Body.String())
	}
	var response AuthResponse
	a.decode(recorder, &response)
	userID, err := strconv.ParseInt(response.UserID, 10, 64)
	if err != nil {
		a.t.Fatalf("parse the user id %q: %v", response.UserID, err)
	}
	return response.Token, userID
}

// login signs in and returns the bearer, so a test can hold a second browser.
func (a *testAPI) login(email, password string) string {
	a.t.Helper()
	recorder := a.call(http.MethodPost, "/auth/login", "", map[string]string{
		"email":    email,
		"password": password,
	})
	if recorder.Code != http.StatusOK {
		a.t.Fatalf("login %s: %d %s", email, recorder.Code, recorder.Body.String())
	}
	var response AuthResponse
	a.decode(recorder, &response)
	return response.Token
}

// federatedAccount is an SSO-only user: a row in users with no local identity
// behind it, holding an unlabelled session as if it had arrived through the
// OIDC callback. It is the only kind of account for which SetPassword is
// reachable.
func (a *testAPI) federatedAccount(email string) string {
	a.t.Helper()
	user := schemas.User{Email: email, Name: "Federated"}
	if err := a.db.Create(&user).Error; err != nil {
		a.t.Fatalf("create the federated account: %v", err)
	}
	token, _, err := a.service.Issue(a.t.Context(), user.ID, "")
	if err != nil {
		a.t.Fatalf("issue a session: %v", err)
	}
	return token
}

// errorCode reads the code tronc nests under "error", which is where the
// mapping this bump introduces is visible.
func (a *testAPI) errorCode(recorder *httptest.ResponseRecorder) string {
	a.t.Helper()
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	a.decode(recorder, &body)
	return body.Error.Code
}

func (a *testAPI) decode(recorder *httptest.ResponseRecorder, into any) {
	a.t.Helper()
	if err := json.Unmarshal(recorder.Body.Bytes(), into); err != nil {
		a.t.Fatalf("decode %s: %v", recorder.Body.String(), err)
	}
}
