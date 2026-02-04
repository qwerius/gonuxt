// Package migrations berisi definisi migrasi database untuk proyek ini.
package migrations

// Migration012CreateAuditLogs membuat tabel audit_logs untuk menyimpan aktivitas user
var Migration012CreateAuditLogs = Migration{
	Version: 12,
	Name:    "create_audit_logs",
	Up: `
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INT,
    method VARCHAR(10) NOT NULL,
    url TEXT NOT NULL,
    status INT NOT NULL,
    ip VARCHAR(45),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_user
      FOREIGN KEY (user_id) REFERENCES users(id)
      ON DELETE SET NULL
);
`,
	Down: `
DROP TABLE IF EXISTS audit_logs;
`,
}
