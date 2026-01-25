package migrations

var Migration001Users = Migration{
	Version: 1,
	Name:    "create_users_table",
	Up: `
CREATE TABLE users (
	id SERIAL PRIMARY KEY,
	email TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT NOW()
);
`,
	Down: `
DROP TABLE IF EXISTS users;
`,
}
