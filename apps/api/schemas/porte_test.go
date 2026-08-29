package schemas

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

const testIssuer = "https://porte.test/application/o/vision/"

// seedPrePorte rebuilds the shape production is in before this deploy: the old
// sessions table, a federated identity recorded on the user row, and a local
// password hash in users.password_hash.
//
// The legacy sessions table is created here in SQL because the model is gone.
// That is the point: after this migration it exists only in databases that
// predate it, and the only thing that still has to understand it is AdoptPorte.
//
// api_keys is created too, empty, so the test can assert the migration leaves
// it standing — it is a scoped analytics credential, not a session.
func seedPrePorte(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`DROP TABLE IF EXISTS sessions`,
		`CREATE TABLE sessions (token text PRIMARY KEY, user_id bigint NOT NULL, expires_at timestamptz, created_at timestamptz)`,
		`CREATE TABLE IF NOT EXISTS api_keys (id bigserial PRIMARY KEY, user_id bigint NOT NULL, site_id bigint, scopes text)`,
		`INSERT INTO users (id, email, name, oidc_subject, oidc_access_token, oidc_refresh_token, profile_synced_at, password_hash, created_at)
		 VALUES (1, 'camille@facile.studio', 'Camille', 'sub-1', 'ciphertext', 'ciphertext', now(), '', now())`,
		`INSERT INTO users (id, email, name, oidc_subject, password_hash, created_at)
		 VALUES (2, 'Noah@Facile.Studio', 'Noah', NULL, '$argon2id$fake', now())`,
		`INSERT INTO sessions (token, user_id, expires_at, created_at) VALUES
			('live', 1, now() + interval '10 days', now() - interval '40 days'),
			('dead', 1, now() - interval '1 day', now() - interval '31 days')`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("seed: %v\n%s", err, statement)
		}
	}
	t.Cleanup(func() {
		db.Exec(`DROP TABLE IF EXISTS sessions`)
		db.Exec(`DROP TABLE IF EXISTS api_keys`)
		db.Exec(`DELETE FROM users`)
	})
}

// Nobody may be signed out by this deploy. Both tables store the SHA-256 hex of
// a token and nothing else, which is exactly what porte stores, so the rows
// move and the cookie already in somebody's browser keeps authenticating.
//
// created_at is deliberately not carried: copying it would put a session 40
// days into the seven-day idle window and sign the user out on the deploy meant
// to keep them.
//
// api_keys is left alone: it is not a session with a name on it but a scoped
// analytics credential carrying a site_id, a workspace_id and a scopes column,
// so folding one into porte_sessions would turn a read-only key for one site
// into a full user session.
func TestAdoptPorteKeepsEverybodySignedIn(t *testing.T) {
	db := openTestDatabase(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	var carried struct {
		UserID     int64
		Label      string
		LastUsedAt time.Time
	}
	if err := db.Raw(`SELECT user_id, label, last_used_at FROM porte_sessions WHERE token_hash = 'live'`).Scan(&carried).Error; err != nil {
		t.Fatalf("read the carried session: %v", err)
	}
	if carried.UserID != 1 || carried.Label != "" {
		t.Fatalf("the browser session did not survive as an unlabelled session: %+v", carried)
	}
	if time.Since(carried.LastUsedAt) > time.Hour {
		t.Fatalf("last_used_at was copied instead of stamped: %v", carried.LastUsedAt)
	}

	var expired int64
	if err := db.Raw(`SELECT count(*) FROM porte_sessions WHERE token_hash = 'dead'`).Scan(&expired).Error; err != nil {
		t.Fatalf("count expired: %v", err)
	}
	if expired != 0 {
		t.Fatal("an already-expired session was carried over")
	}

	var remaining *string
	if err := db.Raw(`SELECT to_regclass('sessions')::text`).Scan(&remaining).Error; err != nil {
		t.Fatalf("check sessions: %v", err)
	}
	if remaining != nil {
		t.Fatal("the legacy sessions table survived")
	}

	var keys *string
	if err := db.Raw(`SELECT to_regclass('api_keys')::text`).Scan(&keys).Error; err != nil {
		t.Fatalf("check api_keys: %v", err)
	}
	if keys == nil {
		t.Fatal("the migration dropped api_keys, which are scoped analytics credentials and not sessions")
	}

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt is not idempotent: %v", err)
	}
}

// The password hash moves into the identity row porte/local reads. Without it
// the login form answers "invalid credentials" to a correct password, with the
// hash still sitting in the users table and no error anywhere.
//
// The identity is keyed on the account id, which is what porte.LocalSubject
// returns since v0.3.0. Keying it on the address instead is a row porte/local
// never looks for, so the login answers 401 to a correct password.
func TestAdoptPorteMovesThePasswords(t *testing.T) {
	db := openTestDatabase(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	var identity struct {
		UserID       int64
		PasswordHash string
	}
	err := db.Raw(
		`SELECT user_id, password_hash FROM porte_identities WHERE provider = 'local' AND subject = '2'`,
	).Scan(&identity).Error
	if err != nil {
		t.Fatalf("read the local identity: %v", err)
	}
	if identity.UserID != 2 || identity.PasswordHash != "$argon2id$fake" {
		t.Fatalf("the password did not move: %+v", identity)
	}

	var withoutPassword int64
	if err := db.Raw(`SELECT count(*) FROM porte_identities WHERE provider = 'local' AND user_id = 1`).Scan(&withoutPassword).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if withoutPassword != 0 {
		t.Fatal("an account with no password gained a local identity, which is a login that cannot be used and an account that cannot be registered")
	}
}

// A database that adopted porte at v0.2.x holds local identities keyed on the
// address, and the re-key UPDATE is the only statement that reaches them.
// Without it the version bump compiles, boots, and answers 401 to every correct
// password.
//
// users.password_hash is blanked here rather than left alone, and that is the
// whole point of the case. porte's CreateFromPassword never writes that column,
// so every account registered since the v0.2 adoption sits outside
// adoptExistingPasswords' filter and the INSERT cannot re-create its row. With
// the hash still present the INSERT writes an id-keyed row of its own and the
// test passes with the migration deleted — which is what it did before this
// comment was written.
//
// The federated row is asserted untouched: the UPDATE is filtered on
// provider = 'local', and sweeping an OIDC subject into an account id would
// unlink every SSO user instead.
func TestAdoptPorteRekeysIdentitiesLeftOnTheAddress(t *testing.T) {
	db := openTestDatabase(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt: %v", err)
	}
	rollback := []string{
		`UPDATE porte_identities SET subject = lower(btrim(u.email))
		   FROM users u WHERE u.id = porte_identities.user_id AND provider = 'local'`,
		`UPDATE users SET password_hash = '' WHERE id = 2`,
	}
	for _, statement := range rollback {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("rebuild the v0.2 shape: %v\n%s", err, statement)
		}
	}

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("re-adopt: %v", err)
	}

	var subjects []string
	if err := db.Raw(`SELECT subject FROM porte_identities WHERE provider = 'local' AND user_id = 2`).Scan(&subjects).Error; err != nil {
		t.Fatalf("read the local identity: %v", err)
	}
	if len(subjects) != 1 || subjects[0] != "2" {
		t.Fatalf("expected one local identity keyed on the account id, got %q", subjects)
	}

	var federated string
	if err := db.Raw(`SELECT subject FROM porte_identities WHERE provider = ?`, testIssuer).Scan(&federated).Error; err != nil {
		t.Fatalf("read the federated identity: %v", err)
	}
	if federated != "sub-1" {
		t.Fatalf("the re-key swept a federated identity to %q", federated)
	}
}

// The federated identity moves off the user row. Without it porte finds no
// identity, falls back to matching the verified email and relinks on the next
// login — which works, but leans the whole existing user base on the weaker of
// the two matching paths, on the one deploy where nobody would notice.
//
// Vision encrypts the provider tokens with ENCRYPTION_KEY and porte stores them
// as it will send them, so the ciphertext deliberately stays behind: handing
// porte a refresh token that is not one makes the first profile sync fail and
// look like the provider revoked it.
func TestAdoptPorteMovesTheOIDCSubject(t *testing.T) {
	db := openTestDatabase(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	var identity struct {
		UserID       int64
		AccessToken  string
		RefreshToken string
	}
	err := db.Raw(
		`SELECT user_id, access_token, refresh_token FROM porte_identities WHERE provider = ? AND subject = 'sub-1'`,
		testIssuer,
	).Scan(&identity).Error
	if err != nil {
		t.Fatalf("read the identity: %v", err)
	}
	if identity.UserID != 1 {
		t.Fatal("the oidc subject was not adopted")
	}
	if identity.AccessToken != "" || identity.RefreshToken != "" {
		t.Fatalf("encrypted provider tokens were carried across: %+v", identity)
	}

	var rows int64
	if err := db.Raw(`SELECT count(*) FROM porte_identities WHERE provider = ?`, testIssuer).Scan(&rows).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if rows != 1 {
		t.Fatalf("expected exactly one federated identity, got %d", rows)
	}
}

// An empty issuer is a deployment with SSO switched off. The sessions and the
// passwords still have to move — they are what keeps people
// signed in and able to sign in — but there is no provider to key a federated
// identity against.
func TestAdoptPorteWithoutAnIssuerStillMovesTheCredentials(t *testing.T) {
	db := openTestDatabase(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, ""); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	var sessions, federated int64
	if err := db.Raw(`SELECT count(*) FROM porte_sessions`).Scan(&sessions).Error; err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if err := db.Raw(`SELECT count(*) FROM porte_identities WHERE provider <> 'local'`).Scan(&federated).Error; err != nil {
		t.Fatalf("count identities: %v", err)
	}
	if sessions != 1 {
		t.Fatalf("expected the one live session, got %d", sessions)
	}
	if federated != 0 {
		t.Fatalf("an identity was keyed against no provider: %d rows", federated)
	}
}
