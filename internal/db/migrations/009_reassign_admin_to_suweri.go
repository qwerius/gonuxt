package migrations

var Migration009ReassignAdminToSuweri = Migration{
	Version: 9,
	Name:    "assign_admin_to_suweri",
	Up: `
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id
FROM users u
JOIN roles r ON r.name = 'admin'
WHERE u.email = 'suweri@gmail.com'
ON CONFLICT (user_id, role_id) DO NOTHING;
`,
	Down: `
DELETE FROM user_roles
USING users u, roles r
WHERE user_roles.user_id = u.id
  AND user_roles.role_id = r.id
  AND u.email = 'suweri@gmail.com'
  AND r.name = 'admin';
`,
}
