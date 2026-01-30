package migrations

var Migration006SeedRoles = Migration{
	Version: 6,
	Name:    "seed_roles",
	Up: `
INSERT INTO roles (name)
VALUES
  ('pelanggan'),
  ('admin')
ON CONFLICT (name) DO NOTHING;
`,
	Down: `
DELETE FROM roles
WHERE role IN ('pelanggan', 'admin');
`,
}
