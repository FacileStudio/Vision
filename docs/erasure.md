# Vision — Erasure

When Registre erases a person, Vision deletes their account and their personal data. Data a
workspace shares with other people stays, without the person's name on it.

## How the notice arrives

Vision runs porte/directory's deletion receiver when `OIDC_ISSUER` is set
(`startDirectory` in `apps/api/main.go`). It polls Registre's deletion list and accepts
Registre's signed push on `POST /api/registre/deletions`. Each erased person reaches porte/kit's
`EraseAccount` once. The receiver keeps its cursor and the subjects it already handled in
`directory_deletion_state`, so a notice delivered twice erases nothing the second time.

Registre holds a self-deletion for 30 days before it notifies apps.

## The rule

`EraseAccount` runs `workspaces.EraseData` (`apps/api/modules/workspaces/erase.go`) and then
deletes the `users` row, in one transaction, under the kit's account lock. The provider
matches with or without a trailing slash. Any error rolls the whole erasure back.

| What | What happens |
|---|---|
| The `users` row | Deleted |
| porte identities, sessions, login codes, magic links | Deleted with the user (foreign keys cascade) |
| A personal workspace: unlinked, and the person is its only member | Deleted with its sites, pageviews, custom events, visitor sessions, goals, API keys and webhooks |
| Sites and goals the person authored in a workspace that stays | Handed to the next member (owner, admin, editor, viewer, oldest membership first), or to nobody (`0`) when none is left |
| The person's own webhooks and API keys anywhere | Deleted: the URL, the secret and the key are theirs |
| Actions the person's MCP clients left waiting for confirmation (`mcp_actions`) | Deleted, open or resolved |
| Memberships | Deleted |
| A shared, unlinked workspace still named after the person (their name or address) | Renamed `Workspace` |
| A workspace linked to a Registre space | Kept, with its Registre name |

Vision stores no author name on sites or goals, so "Deleted user" is the absence of an owner.

## What stays, and why

- **Analytics rows** (`pageviews`, `custom_events`, `visitor_sessions`) under sites that stay.
  They describe visitors to a website, not the account holder, and carry a hashed visitor id.
- **`directory_deletion_state`** keeps the erased subject so the receiver can refuse to erase
  twice. It is an opaque Registre id with no email or name.
- **Logs** shipped to Journal follow Journal's own retention.

## The proof

`TestErasureLeavesNoTraceOfThePersonInAnyColumn` (`apps/api/modules/workspaces/erase_test.go`)
signs a person in through the kit, gives them rows in every table that references a user,
and scans every `text`, `varchar`, `jsonb` and `uuid` column for their email, name and
Registre subject, and every `user_id` and `owner_id` column for their id. The scan must find
each kind of row before the erasure and nothing after it. The erasure goes through the same
`OnPersonDeleted` callback the receiver calls, with the issuer spelled with a trailing slash.

Run it with a test database:

```sh
TEST_DATABASE_URL=postgres://... go test -run TestErasure ./modules/workspaces/
```

Back to the [documentation index](README.md).
