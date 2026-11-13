package auth

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"testing"

	"github.com/FacileStudio/Vision/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/testdb"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

const testIssuer = "https://sso.test"

var testConfig = testdb.Config{Prefix: "vision_auth_test", Migrate: schemas.Migrate}

func openStore(t *testing.T) (*kit.Store, *gorm.DB) {
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
		t.Fatal(err)
	}
	store, err := NewStore(t.Context(), sqlDB, nil)
	if err != nil {
		t.Fatalf("build the store: %v", err)
	}
	return store, db
}

func claims(subject, email string, verified bool) porte.Claims {
	return porte.Claims{
		Provider: testIssuer, Subject: subject, Email: email, EmailVerified: verified,
		Name: "Camille", Locale: "de",
	}
}

func linkIdentity(t *testing.T, db *gorm.DB, userID int64, subject string) {
	t.Helper()
	err := db.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)`,
		userID, testIssuer, subject).Error
	if err != nil {
		t.Fatalf("link identity: %v", err)
	}
}

func TestVerifiedEmailLinksAnExistingLocalAccountAndStoresTheLocale(t *testing.T) {
	store, db := openStore(t)
	local := schemas.User{Email: "camille@facile.studio", Name: "Old"}
	if err := db.Create(&local).Error; err != nil {
		t.Fatal(err)
	}

	id, err := store.UpsertFromOIDC(t.Context(), claims("sub-1", "Camille@Facile.Studio", true))
	if err != nil {
		t.Fatalf("link: %v", err)
	}
	if id != local.ID {
		t.Fatalf("linked to account %d, want the existing %d", id, local.ID)
	}
	var got schemas.User
	db.First(&got, id)
	if got.Locale != "de" || got.Name != "Camille" {
		t.Errorf("locale %q name %q, want de and Camille", got.Locale, got.Name)
	}
}

func TestUnverifiedEmailNeverLinks(t *testing.T) {
	store, db := openStore(t)
	local := schemas.User{Email: "camille@facile.studio"}
	if err := db.Create(&local).Error; err != nil {
		t.Fatal(err)
	}

	if _, err := store.UpsertFromOIDC(t.Context(), claims("sub-1", "camille@facile.studio", false)); err == nil {
		t.Fatal("an unverified address linked into an existing account")
	}
}

func TestSecondSubjectOfTheSameIssuerIsRefused(t *testing.T) {
	store, db := openStore(t)
	local := schemas.User{Email: "camille@facile.studio"}
	if err := db.Create(&local).Error; err != nil {
		t.Fatal(err)
	}
	linkIdentity(t, db, local.ID, "sub-1")

	if _, err := store.UpsertFromOIDC(t.Context(), claims("sub-2", "camille@facile.studio", true)); err == nil {
		t.Fatal("a recycled address was handed to a second subject")
	}
	if id, err := store.UpsertFromOIDC(t.Context(), claims("sub-1", "camille@facile.studio", true)); err != nil || id != local.ID {
		t.Fatalf("the owner no longer resolves: id %d err %v", id, err)
	}
}

func TestPasswordRoutesAreGone(t *testing.T) {
	router := chi.NewRouter()
	RegisterRoutes(router, &Service{})
	for _, route := range []struct{ method, path string }{
		{http.MethodPost, "/auth/register"}, {http.MethodPost, "/auth/login"}, {http.MethodPut, "/auth/password"},
	} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(route.method, route.path, nil))
		if recorder.Code != http.StatusNotFound && recorder.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s answered %d, want 404", route.method, route.path, recorder.Code)
		}
	}
}

func TestProfileUpdateCannotChangeTheEmail(t *testing.T) {
	_, db := openStore(t)
	user := schemas.User{Email: "mallory@example.com", Name: "Mallory"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	service := NewService(db, nil, nil, nil)

	got, err := service.controller.updateMe(t.Context(), strconv.FormatInt(user.ID, 10),
		&UpdateProfileRequest{Name: " Mal ", Email: "victim@facile.studio"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if got.Email != "mallory@example.com" || got.Name != "Mal" {
		t.Fatalf("email %q name %q, want the email untouched and the name saved", got.Email, got.Name)
	}
	var stored schemas.User
	db.First(&stored, user.ID)
	if stored.Email != "mallory@example.com" {
		t.Fatalf("the stored email became %q", stored.Email)
	}
}

func TestTheColumnsTheKitLeavesOutArePinned(t *testing.T) {
	store, _ := openStore(t)
	want := []string{"oidc_subject"}
	if got := slices.Sorted(slices.Values(store.Unwritten())); !slices.Equal(got, want) {
		t.Fatalf("unwritten = %v, want %v: fill a new column in fillNewAccount or accept it here", got, want)
	}
}

func TestANewAccountReadsLikeTheOldOnes(t *testing.T) {
	store, db := openStore(t)
	id, err := store.UpsertFromOIDC(t.Context(), claims("sub-new", "new@facile.studio", true))
	if err != nil {
		t.Fatalf("sign up: %v", err)
	}
	var nulls []string
	err = db.Raw(`SELECT key FROM users u, jsonb_each(to_jsonb(u)) WHERE u.id = ? AND value = 'null'::jsonb`, id).
		Scan(&nulls).Error
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(nulls, []string{"oidc_subject"}) {
		t.Fatalf("NULL columns on a new account: %v, want only oidc_subject", nulls)
	}
	var user schemas.User
	if err := db.First(&user, id).Error; err != nil {
		t.Fatalf("GORM cannot read the new account: %v", err)
	}
	if user.CreatedAt.IsZero() || user.Locale != "de" || user.Name != "Camille" {
		t.Errorf("created_at %v locale %q name %q", user.CreatedAt, user.Locale, user.Name)
	}
}

func TestAnUnverifiedAddressIsNotWrittenOverTheStoredOne(t *testing.T) {
	store, db := openStore(t)
	id, err := store.UpsertFromOIDC(t.Context(), claims("sub-1", "camille@facile.studio", true))
	if err != nil {
		t.Fatal(err)
	}
	linkIdentity(t, db, id, "sub-1")
	if _, err := store.UpsertFromOIDC(t.Context(), claims("sub-1", "other@facile.studio", false)); err != nil {
		t.Fatal(err)
	}
	var user schemas.User
	db.First(&user, id)
	if user.Email != "camille@facile.studio" {
		t.Fatalf("email became %q, want the verified one kept", user.Email)
	}
}

func TestBackfillFillsEmptyPicturesOnly(t *testing.T) {
	store, db := openStore(t)
	empty := schemas.User{Email: "a@facile.studio"}
	chosen := schemas.User{Email: "b@facile.studio", OIDCPictureURL: "https://elsewhere.test/b.png"}
	for _, user := range []*schemas.User{&empty, &chosen} {
		if err := db.Create(user).Error; err != nil {
			t.Fatal(err)
		}
	}
	linkIdentity(t, db, empty.ID, "sub-a")
	linkIdentity(t, db, chosen.ID, "sub-b")
	for range 2 {
		if _, err := store.BackfillPictures(t.Context(), testIssuer); err != nil {
			t.Fatal(err)
		}
	}
	want := map[int64]string{empty.ID: testIssuer + "/avatars/sub-a", chosen.ID: "https://elsewhere.test/b.png"}
	for id, picture := range want {
		var user schemas.User
		db.First(&user, id)
		if user.OIDCPictureURL != picture {
			t.Errorf("user %d picture %q, want %q", id, user.OIDCPictureURL, picture)
		}
	}
}
