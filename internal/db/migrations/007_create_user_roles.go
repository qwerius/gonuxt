package migrations

var Migration007CreateUserRoles = Migration{
	Version: 7,
	Name:    "create_user_roles",
	Up: `
CREATE TABLE IF NOT EXISTS user_roles (
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_user
      FOREIGN KEY (user_id) REFERENCES users(id)
      ON DELETE CASCADE,
    CONSTRAINT fk_role
      FOREIGN KEY (role_id) REFERENCES roles(id)
      ON DELETE CASCADE
);
`,
	Down: `
DROP TABLE IF EXISTS user_roles;
`,
}
