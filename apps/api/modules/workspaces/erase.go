package workspaces

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/gorm"
)

// successorOrder ranks the members who can inherit what a departing person authored:
// owners first, then admins, editors, and viewers last, the oldest membership first.
const successorOrder = `CASE m.role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 WHEN 'editor' THEN 2 ELSE 3 END, m.id`

// eraseSubject deletes the account behind a Registre subject and the personal data it
// holds, in one transaction. It is idempotent: a subject with no identity left is already
// erased and is a no-op.
//
// A workspace nobody else belongs to, and that Registre does not own, is the person's
// own: its sites and every pageview, event, goal, webhook and key under them go. Sites in
// a workspace others use stay, and so do their goals; their authorship moves to the next
// member, or to nobody when none is left. Vision stores no author name, so the "Deleted
// user" label is the absence of an owner. The person's own webhooks and API keys go,
// because the URL, the secret and the key are theirs.
func (d *Deletions) eraseSubject(ctx context.Context, provider, subject string) error {
	return d.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ids []int64
		err := tx.Raw(`SELECT user_id FROM porte_identities WHERE provider = ? AND subject = ?`,
			provider, subject).Scan(&ids).Error
		if err != nil || len(ids) == 0 {
			return err
		}
		userID := ids[0]
		if err := tx.Exec(`SELECT 1 FROM users WHERE id = ? FOR UPDATE`, userID).Error; err != nil {
			return err
		}
		steps := []func(*gorm.DB, int64) error{
			deletePersonalWorkspaces, handOverSharedAuthorship, deletePersonalConfig, deleteAccount,
		}
		for _, step := range steps {
			if err := step(tx, userID); err != nil {
				return fmt.Errorf("erase user %d: %w", userID, err)
			}
		}
		return nil
	})
}

// deletePersonalWorkspaces removes the unlinked workspaces the person is the only member
// of, with everything under their sites.
func deletePersonalWorkspaces(tx *gorm.DB, userID int64) error {
	var workspaceIDs []int64
	err := tx.Raw(`SELECT w.id FROM workspaces w
		JOIN workspace_members m ON m.workspace_id = w.id AND m.user_id = ?
		WHERE w.registre_space_id IS NULL
		  AND (SELECT count(*) FROM workspace_members c WHERE c.workspace_id = w.id) = 1`, userID).
		Scan(&workspaceIDs).Error
	if err != nil || len(workspaceIDs) == 0 {
		return err
	}
	siteScope := `site_id IN (SELECT id FROM sites WHERE workspace_id IN @ws)`
	statements := []string{
		`DELETE FROM pageviews WHERE ` + siteScope,
		`DELETE FROM custom_events WHERE ` + siteScope,
		`DELETE FROM visitor_sessions WHERE ` + siteScope,
		`DELETE FROM goals WHERE ` + siteScope,
		`DELETE FROM api_keys WHERE ` + siteScope + ` OR workspace_id IN @ws`,
		`DELETE FROM webhooks WHERE workspace_id IN @ws`,
		`DELETE FROM sites WHERE workspace_id IN @ws`,
		`DELETE FROM workspace_members WHERE workspace_id IN @ws`,
		`DELETE FROM workspaces WHERE id IN @ws`,
	}
	for _, statement := range statements {
		if err := tx.Exec(statement, sql.Named("ws", workspaceIDs)).Error; err != nil {
			return err
		}
	}
	return nil
}

// handOverSharedAuthorship gives the sites and goals the person authored in workspaces
// that stay to the next member, or to nobody (0) when none is left.
func handOverSharedAuthorship(tx *gorm.DB, userID int64) error {
	err := tx.Exec(`UPDATE sites s SET owner_id = COALESCE((
			SELECT m.user_id FROM workspace_members m
			 WHERE m.workspace_id = s.workspace_id AND m.user_id <> ?
			 ORDER BY `+successorOrder+` LIMIT 1), 0)
		WHERE s.owner_id = ?`, userID, userID).Error
	if err != nil {
		return err
	}
	err = tx.Exec(`UPDATE goals g SET owner_id = s.owner_id FROM sites s
		WHERE s.id = g.site_id AND g.owner_id = ?`, userID).Error
	if err != nil {
		return err
	}
	return tx.Exec(`UPDATE goals SET owner_id = 0 WHERE owner_id = ?`, userID).Error
}

// deletePersonalConfig removes the person's webhooks and API keys in the workspaces that
// stay: the delivery URL, its secret and the key hash are theirs.
func deletePersonalConfig(tx *gorm.DB, userID int64) error {
	if err := tx.Exec(`DELETE FROM webhooks WHERE owner_id = ?`, userID).Error; err != nil {
		return err
	}
	return tx.Exec(`DELETE FROM api_keys WHERE user_id = ?`, userID).Error
}

// deleteAccount removes the memberships and the user row. porte's identities, sessions,
// login codes and magic links cascade from it.
func deleteAccount(tx *gorm.DB, userID int64) error {
	if err := tx.Exec(`DELETE FROM workspace_members WHERE user_id = ?`, userID).Error; err != nil {
		return err
	}
	return tx.Exec(`DELETE FROM users WHERE id = ?`, userID).Error
}
