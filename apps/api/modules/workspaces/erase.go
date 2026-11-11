package workspaces

import (
	"context"
	"database/sql"
	"fmt"
)

// successorOrder ranks the members who can inherit what a departing person authored:
// owners first, then admins, editors, and viewers last, the oldest membership first.
const successorOrder = `CASE m.role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 WHEN 'editor' THEN 2 ELSE 3 END, m.id`

// personalWorkspaces are the unlinked workspaces the person is the only member of.
const personalWorkspaces = `SELECT w.id FROM workspaces w
	JOIN workspace_members m ON m.workspace_id = w.id AND m.user_id = $1
	WHERE w.registre_space_id IS NULL
	  AND (SELECT count(*) FROM workspace_members c WHERE c.workspace_id = w.id) = 1`

const siteScope = `site_id IN (SELECT id FROM sites WHERE workspace_id = ANY($1))`

// workspaceCascade empties the person's own workspaces: every pageview, event, session,
// goal, key and webhook under them, then the sites, the memberships and the workspaces.
func workspaceCascade() []string {
	return []string{
		`DELETE FROM pageviews WHERE ` + siteScope,
		`DELETE FROM custom_events WHERE ` + siteScope,
		`DELETE FROM visitor_sessions WHERE ` + siteScope,
		`DELETE FROM goals WHERE ` + siteScope,
		`DELETE FROM api_keys WHERE ` + siteScope + ` OR workspace_id = ANY($1)`,
		`DELETE FROM webhooks WHERE workspace_id = ANY($1)`,
		`DELETE FROM sites WHERE workspace_id = ANY($1)`,
		`DELETE FROM workspace_members WHERE workspace_id = ANY($1)`,
		`DELETE FROM workspaces WHERE id = ANY($1)`,
	}
}

// personCascade hands over what the person authored in workspaces that stay, to the next
// member or to nobody (0), then removes their webhooks, keys and memberships. A shared,
// unlinked workspace still carrying the name Vision gave it after the person (their name
// or address) is renamed, because that name is theirs.
func personCascade() []string {
	return []string{
		`UPDATE workspaces w SET name = 'Workspace' FROM users u
		 WHERE u.id = $1 AND w.registre_space_id IS NULL
		   AND w.id IN (SELECT workspace_id FROM workspace_members WHERE user_id = $1)
		   AND ((u.name <> '' AND w.name = u.name) OR lower(w.name) = lower(u.email))`,
		`UPDATE sites s SET owner_id = COALESCE((
		SELECT m.user_id FROM workspace_members m
		 WHERE m.workspace_id = s.workspace_id AND m.user_id <> $1
		 ORDER BY ` + successorOrder + ` LIMIT 1), 0)
	 WHERE s.owner_id = $1`,
		`UPDATE goals g SET owner_id = s.owner_id FROM sites s WHERE s.id = g.site_id AND g.owner_id = $1`,
		`UPDATE goals SET owner_id = 0 WHERE owner_id = $1`,
		`DELETE FROM webhooks WHERE owner_id = $1`,
		`DELETE FROM api_keys WHERE user_id = $1`,
		`DELETE FROM workspace_members WHERE user_id = $1`,
	}
}

// EraseData deletes the personal data of an account porte/kit is erasing, in the kit's
// transaction. The kit deletes the users row afterwards, and porte's identities,
// sessions, login codes and magic links cascade from it.
//
// A workspace nobody else belongs to, and that Registre does not own, is the person's
// own: its sites and everything under them go. Sites in a workspace others use stay, and
// so do their goals; their authorship moves to the next member, or to nobody when none is
// left. Vision stores no author name, so the "Deleted user" label is the absence of an
// owner. The person's own webhooks and API keys go, because the URL, the secret and the
// key are theirs.
func EraseData(ctx context.Context, tx *sql.Tx, userID int64) error {
	ids, err := personalWorkspaceIDs(ctx, tx, userID)
	if err != nil {
		return fmt.Errorf("erase user %d: %w", userID, err)
	}
	if len(ids) > 0 {
		if err := execAll(ctx, tx, workspaceCascade(), ids); err != nil {
			return fmt.Errorf("erase user %d: %w", userID, err)
		}
	}
	if err := execAll(ctx, tx, personCascade(), userID); err != nil {
		return fmt.Errorf("erase user %d: %w", userID, err)
	}
	return nil
}

func personalWorkspaceIDs(ctx context.Context, tx *sql.Tx, userID int64) ([]int64, error) {
	rows, err := tx.QueryContext(ctx, personalWorkspaces, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func execAll(ctx context.Context, tx *sql.Tx, statements []string, arg any) error {
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement, arg); err != nil {
			return err
		}
	}
	return nil
}
