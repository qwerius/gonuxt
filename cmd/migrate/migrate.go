package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/qwerius/gonuxt/internal/config"
	"github.com/qwerius/gonuxt/internal/db"
	"github.com/qwerius/gonuxt/internal/db/migrations"
)

func main() {
	config.Load()

	if len(os.Args) < 2 {
		log.Fatal("gunakan: up [version] | down | status")
	}

	command := os.Args[1]

	conn, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ensureMigrationTable(conn)

	switch command {
	case "up":
		var version int
		if len(os.Args) == 3 {
			version, _ = strconv.Atoi(os.Args[2])
		}
		migrateUp(conn, version)
	case "down":
		migrateDown(conn)
	case "status":
		printStatus(conn)
	default:
		log.Fatal("command tidak dikenal")
	}
}

/* =======================
   FUNGSI MIGRASI
   ======================= */

func ensureMigrationTable(dbConn *sql.DB) {
	_, err := dbConn.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
	version INT PRIMARY KEY,
	applied_at TIMESTAMP DEFAULT NOW()
);
`)
	if err != nil {
		log.Fatal(err)
	}
}

// migrateUp hanya jalankan migration tertentu jika version > 0
func migrateUp(dbConn *sql.DB, version int) {
	for _, m := range migrations.Migrations {
		if version > 0 && m.Version != version {
			continue // skip yang bukan target
		}

		if isApplied(dbConn, m.Version) {
			fmt.Println("sudah diterapkan:", m.Name)
			continue
		}

		fmt.Println("↑ apply:", m.Name)

		tx, err := dbConn.Begin()
		if err != nil {
			log.Fatal(err)
		}

		if _, err := tx.Exec(m.Up); err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version) VALUES ($1)",
			m.Version,
		); err != nil {
			tx.Rollback()
			log.Fatal(err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatal(err)
		}
	}
}

// migrateDown rollback migration terakhir
func migrateDown(dbConn *sql.DB) {
	var version int
	err := dbConn.QueryRow(
		"SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1",
	).Scan(&version)
	if err != nil {
		fmt.Println("tidak ada migrasi untuk rollback")
		return
	}

	for i := len(migrations.Migrations) - 1; i >= 0; i-- {
		m := migrations.Migrations[i]
		if m.Version == version {
			fmt.Println("↓ rollback:", m.Name)

			tx, err := dbConn.Begin()
			if err != nil {
				log.Fatal(err)
			}

			if _, err := tx.Exec(m.Down); err != nil {
				tx.Rollback()
				log.Fatal(err)
			}

			if _, err := tx.Exec(
				"DELETE FROM schema_migrations WHERE version = $1",
				version,
			); err != nil {
				tx.Rollback()
				log.Fatal(err)
			}

			if err := tx.Commit(); err != nil {
				log.Fatal(err)
			}
			return
		}
	}
}

// printStatus tampilkan migration yang sudah dan belum diterapkan
func printStatus(dbConn *sql.DB) {
	fmt.Println("Status migration:")
	for _, m := range migrations.Migrations {
		if isApplied(dbConn, m.Version) {
			fmt.Printf("[✓] %d - %s\n", m.Version, m.Name)
		} else {
			fmt.Printf("[ ] %d - %s\n", m.Version, m.Name)
		}
	}
}

func isApplied(dbConn *sql.DB, version int) bool {
	var v int
	err := dbConn.QueryRow(
		"SELECT version FROM schema_migrations WHERE version = $1",
		version,
	).Scan(&v)
	return err == nil
}
