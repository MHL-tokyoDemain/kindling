# kindling

Batch-upload seed data into Firebase Firestore without touching the Firebase UI.

## Prerequisites

- Go 1.22+

## Build

```bash
cd server
go build -o kindling .
```

## Usage

```
kindling <command> [flags]
```

### `kindling server`

Start the HTTP server (used by the VS Code extension and for tooling integration).

| Flag | Default | Description |
|---|---|---|
| `--port` | `9876` | Port to listen on |
| `--creds` | `./serviceAccountKey.json` | Path to service account JSON |
| `--project` | — | Firebase project ID (overrides SDK auto-detection) |
| `--max-file-size` | `1048576` | Per-file size limit in bytes |

```bash
kindling server --port 9876 --creds ./serviceAccountKey.json
```

### `kindling upload`

One-shot batch upload without starting the server. Intended for CI pipelines, scripts, and developer terminals.

| Flag | Default | Description |
|---|---|---|
| `--collection` | — | Firestore collection path (required) |
| `--creds` | `./serviceAccountKey.json` | Path to service account JSON |
| `--project` | — | Firebase project ID |
| `--max-file-size` | `1048576` | Per-file size limit in bytes |
| `--concurrency` | `10` | Max parallel Firestore writes |

```bash
kindling upload --collection projects/my-app/data/sensors --files data.json
```

## Status

Phase 1 — initial CLI skeleton complete. Business logic, HTTP handlers, and Firestore integration are under development.
