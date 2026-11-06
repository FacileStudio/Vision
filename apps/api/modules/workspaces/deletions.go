package workspaces

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/FacileStudio/porte/directory"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// Deletions erases the people Registre erased: their personal data goes, their account
// goes, and the data a workspace shares with others stays.
//
// It wraps porte's receiver, which polls Registre's deletion list, accepts Registre's
// signed push and hands each person to eraseSubject once. Run it in exactly one replica;
// Vision is single-replica.
type Deletions struct {
	orm      *gorm.DB
	receiver *directory.Deletions
}

// NewDeletions returns the receiver for the Registre behind client. erase is porte's own
// erasure (pg.Store.EraseIdentity), which runs after Vision has removed its account.
func NewDeletions(orm *gorm.DB, client directory.Client, erase func(context.Context, string, string) error, logger *slog.Logger) *Deletions {
	d := &Deletions{orm: orm}
	d.receiver = &directory.Deletions{
		Client:          client,
		OnPersonDeleted: d.eraseSubject,
		Erase:           erase,
		State:           deletionState{orm: orm},
		Logger:          logger,
	}
	return d
}

// Handler answers Registre's signed user.deleted push. It is mounted outside the session
// and CSRF middleware because Registre holds no session; porte verifies the signature.
func (d *Deletions) Handler() http.Handler { return d.receiver.Handler() }

// Run polls Registre's deletion list until ctx ends.
func (d *Deletions) Run(ctx context.Context) { d.receiver.Run(ctx) }

// RegisterDeletions mounts the push on directory.DeletionsPath. The client proxy strips
// /api, so the API serves it without that prefix and Registre's deletion_uri keeps it.
func RegisterDeletions(router chi.Router, d *Deletions) {
	router.Method(http.MethodPost, strings.TrimPrefix(directory.DeletionsPath, "/api"), d.Handler())
}

// deletionState keeps the receiver's cursor and handled subjects in one row.
type deletionState struct{ orm *gorm.DB }

// Load returns the saved state, or nil before anything was saved.
func (s deletionState) Load(ctx context.Context) ([]byte, error) {
	var rows [][]byte
	err := s.orm.WithContext(ctx).Raw(`SELECT state FROM directory_deletion_state WHERE id = 1`).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0], nil
}

// Save replaces the saved state.
func (s deletionState) Save(ctx context.Context, state []byte) error {
	if len(state) == 0 {
		return errors.New("workspaces: refusing to save an empty deletion state")
	}
	return s.orm.WithContext(ctx).Exec(`INSERT INTO directory_deletion_state (id, state, updated_at) VALUES (1, ?, now())
		ON CONFLICT (id) DO UPDATE SET state = EXCLUDED.state, updated_at = now()`, state).Error
}
