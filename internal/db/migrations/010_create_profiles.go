package migrations

var Migration010Profiles = Migration{
	Version: 10,
	Name:    "create_profiles_table",
	Up: `
CREATE TABLE IF NOT EXISTS profiles (
	id SERIAL PRIMARY KEY,
	nama VARCHAR(255) NOT NULL,
	nama_belakang VARCHAR(255),
	tanggal_lahir DATE NOT NULL,
	avatar TEXT,
	is_verified BOOLEAN DEFAULT FALSE,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW()
);
`,
	Down: `
DROP TABLE IF EXISTS profiles;
`,
}
