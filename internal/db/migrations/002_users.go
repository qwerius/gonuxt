package migrations

var Migration002AddUpdatedAt = Migration{
	Version: 2,
	Name:    "add_updated_at_column_to_users",
	Up: `
ALTER TABLE users ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT NOW();
`,
	Down: `
ALTER TABLE users DROP COLUMN IF EXISTS updated_at;
`,
}
