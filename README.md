# OpenCentral

A self-hosted file manager with a Go backend and a React + TypeScript frontend. OpenCentral lets you browse, upload, organize, and manage files from any browser — with no cloud dependency.

---

## Features

- **File browsing** — Navigate directories, view file metadata, and open files directly in the browser
- **Upload** — Upload files via multipart form with configurable storage limits
- **Trash & restore** — Move files to trash instead of deleting them permanently; restore to original location
- **Starred files** — Mark files as starred for quick access
- **Batch download** — Select multiple files or folders and download them as a single ZIP archive
- **Rename & move** — Rename files and move them between directories
- **Directory creation** — Create new folders from the UI
- **Storage tracking** — Live storage usage synced to the database every hour
- **Auto trash cleanup** — A daily cron job permanently deletes trash items older than the configured retention period
- **Settings** — Configurable port, storage limit, and trash retention days stored in SQLite

---

## Tech Stack

| Layer    | Technology                              |
|----------|-----------------------------------------|
| Backend  | Go (stdlib `net/http`)                  |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no CGo) |
| Frontend | React 19 + TypeScript + Vite            |
| Styling  | Tailwind CSS v4 + shadcn/ui             |

---

## Project Structure

```
.
├── server/
│   ├── main.go          # Entry point; HTTP routes and server startup
│   ├── cron/
│   │   └── cron.go      # Background jobs: trash cleanup & storage sync
│   ├── db/
│   │   └── database.go  # SQLite init, schema, and query helpers
│   └── src/
│       ├── files.go     # All file operation handlers
│       ├── storage.go   # Storage root, path resolution, usage helpers
│       ├── settings.go  # Settings read/write handlers
│       └── handlers_test.go  # Integration tests
└── web/
    ├── src/             # React application source
    ├── vite.config.ts   # Vite config with /api proxy to :8080
    └── tsconfig*.json   # TypeScript project references
```

---

## Getting Started

### Prerequisites

- Go 1.22+
- Node.js 18+ and npm (or pnpm/yarn)

### Backend

```bash
cd server
go run main.go
```

The server starts on port `8080` by default. A `storage/` directory and `app.db` SQLite database are created automatically on first run.

### Frontend

```bash
cd web
npm install
npm run dev
```

Vite proxies all `/api` requests to `http://localhost:8080`, so no CORS configuration is needed during development.

---

## Docker

### Requirements

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/) installed

### Running with Docker

Build and start both the backend and frontend:
```bash
docker compose up --build
```

Then open `http://localhost`. To run in the background, add the `-d` flag:
```bash
docker compose up --build -d
```

### Stopping
```bash
docker compose down
```

This stops and removes the containers but **preserves your data** in the named volume.
To also delete all uploaded files and the database, add the `-v` flag:
```bash
docker compose down -v
```

### Data & Storage

Uploaded files and the SQLite database are stored in a Docker named volume called `storage`.
This volume persists across restarts and is unaffected by `docker compose down`.

To find where Docker stores the volume on your machine:
```bash
docker volume inspect opencentral_storage
```

---

## API Reference

| Method | Endpoint                | Description                          |
|--------|-------------------------|--------------------------------------|
| GET    | `/api/health`           | Health check                         |
| GET    | `/api/files`            | List files in a directory            |
| GET    | `/api/files/trash`      | List files in trash                  |
| GET    | `/api/file`             | Serve / download a single file       |
| POST   | `/api/files/star`       | Toggle starred status on a file      |
| POST   | `/api/upload`           | Upload one or more files             |
| POST   | `/api/delete`           | Permanently delete a file            |
| POST   | `/api/rename`           | Rename a file                        |
| POST   | `/api/move`             | Move a file to a new directory       |
| POST   | `/api/mkdir`            | Create a directory                   |
| POST   | `/api/trash`            | Move a file to trash                 |
| POST   | `/api/restore`          | Restore a file from trash            |
| GET    | `/api/storage/usage`    | Get current storage usage            |
| GET    | `/api/settings`         | Get all settings                     |
| POST   | `/api/settings/update`  | Update a setting                     |
| POST   | `/api/download/batch`   | Download multiple files as a ZIP     |

---

## Configuration

Settings are stored in the SQLite database and can be updated via the UI or the `/api/settings/update` endpoint.

| Key                   | Default | Description                                      |
|-----------------------|---------|--------------------------------------------------|
| `port`                | `8080`  | Port the HTTP server listens on                  |
| `storage_limit_bytes` | `0`     | Maximum storage in bytes (`0` = unlimited)       |
| `trash_retention_days`| `30`    | Days before trash items are permanently deleted  |

---

## Background Jobs

Two cron jobs run automatically at startup:

- **Storage sync** — Walks the storage directory and updates the `used_bytes` counter in the database every hour.
- **Trash cleanup** — Runs once per day. Permanently deletes any file in the trash older than `trash_retention_days`. Set to `0` to disable automatic cleanup.

---

## Running Tests

```bash
cd server
go test ./src/...
```

Tests use temporary directories and an in-memory SQLite database so they leave no state on disk.

---

## License

MIT
