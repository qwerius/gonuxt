package migrations

func init() {
	Migrations = append(Migrations, Migration{
		Version: 2,
		Name:    "create_profiles_table",
		Up: `
		CREATE TABLE IF NOT EXISTS profiles (
			id SERIAL PRIMARY KEY,
			user_id INT UNIQUE NOT NULL,
			full_name TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			CONSTRAINT fk_profiles_user
				FOREIGN KEY (user_id) REFERENCES users(id)
				ON DELETE CASCADE
		);
		`,
		Down: `
		DROP TABLE IF EXISTS profiles;
		`,
	})
}
