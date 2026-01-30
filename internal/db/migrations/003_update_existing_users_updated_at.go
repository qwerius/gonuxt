package migrations

var Migration003UpdateUsersUpdatedAt = Migration{
	Version: 3,
	Name:    "set_updated_at_for_existing_users",
	Up: `
-- Set updated_at = created_at untuk semua user yang masih NULL
UPDATE users
SET updated_at = created_at
WHERE updated_at IS NULL;
`,
	Down: `
-- Tidak perlu rollback, biarkan NULL
UPDATE users
SET updated_at = NULL
WHERE updated_at = created_at;
`,
}
