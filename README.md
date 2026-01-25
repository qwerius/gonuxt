
```md
# myGoproject

API server menggunakan Go (Fiber) dan PostgreSQL.

## Tech Stack
- Go
- Fiber
- PostgreSQL
- database/sql
- godotenv

## Struktur Project
```

```
cmd/
 └─ server/
     └─ main.go
internal/
 ├─ api/
 ├─ config/
 ├─ db/
go.mod
go.sum
```

## Konfigurasi Environment

Buat file `.env`:

```env
APP_PORT=8080

DB_HOST=localhost
DB_PORT=5432
DB_USER=user
DB_PASSWORD=password
DB_NAME=mydb
DB_SSLMODE=disable
```

## Menjalankan Aplikasi

```bash
go run cmd/server/main.go
```

## Endpoint Contoh

```
GET /hello
```

Response:

```
Hello Fiber
```

## Lisensi

MIT

```

---

## 4️⃣ KENAPA FILE-NYA HARUS `README.md`?
- `README.md` **dibaca otomatis oleh GitHub**
- `.md` = **Markdown**
- standar industri (Go, Python, Rust, dll)

---

## 5️⃣ ATURAN EMAS README (INGAT INI)
README **bukan dokumentasi teknis panjang**, tapi:
- apa ini?
- cara jalanin?
- config apa perlu?
- endpoint apa ada?

---

## 6️⃣ KAPAN README DIUPDATE?
- tambah fitur
- tambah env baru
- ubah cara run
- deploy

---


