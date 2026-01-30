package migrations

var Migration005CreateRole = Migration{
	Version: 5,
	Name:    "create_roles_table",
	Up: `
CREATE TABLE roles (
	id SERIAL PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,	
	created_at TIMESTAMP DEFAULT NOW()
	
);
`,
	Down: `
DROP TABLE IF EXISTS roles;
`,
}
