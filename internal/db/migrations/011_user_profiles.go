package migrations

var Migration011UserProfiles = Migration{
	Version: 11,
	Name:    "create_user_profiles_table",
	Up: `
CREATE TABLE IF NOT EXISTS user_profiles (
	id SERIAL PRIMARY KEY,
	user_id INT NOT NULL,
	profile_id INT NOT NULL,
	created_at TIMESTAMP DEFAULT NOW(),
	updated_at TIMESTAMP DEFAULT NOW(),

	CONSTRAINT fk_user
		FOREIGN KEY(user_id)
		REFERENCES users(id)
		ON DELETE CASCADE
		ON UPDATE CASCADE,

	CONSTRAINT fk_profile
		FOREIGN KEY(profile_id)
		REFERENCES profiles(id)
		ON DELETE CASCADE
		ON UPDATE CASCADE,

	CONSTRAINT unique_user_profile UNIQUE (user_id, profile_id)
);
`,
	Down: `
DROP TABLE IF EXISTS user_profiles;
`,
}
