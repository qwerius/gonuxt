package migrations

func init() {
	Migrations = append(Migrations, Migration{
		Version: 3,
		Name:    "create_posts_table",
		Up: `
		CREATE TABLE IF NOT EXISTS posts (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			title TEXT,
			body TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			CONSTRAINT fk_posts_user
				FOREIGN KEY (user_id) REFERENCES users(id)
				ON DELETE CASCADE
		);
		`,
		Down: `
		DROP TABLE IF EXISTS posts;
		`,
	})
}
