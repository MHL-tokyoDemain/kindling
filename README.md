# kindling

Batch-upload seed data into Firebase Firestore without touching the Firebase UI.

[![CI](https://github.com/MHL-tokyoDemain/kindling/actions/workflows/ci.yml/badge.svg)](https://github.com/MHL-tokyoDemain/kindling/actions/workflows/ci.yml)
[![Go version](https://img.shields.io/badge/Go-1.25-blue)](https://go.dev/doc/install)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

---

## What is Kindling?

Kindling is an open-source developer tool for batch-uploading seed data into Firebase Firestore. It runs either as an HTTP server (for IDE extension integration or tooling) or as a one-shot CLI command (for CI pipelines and scripts).

## Installation

### Homebrew (macOS/Linux)

_Coming soon — tap not yet created._

### GitHub Releases

Download the latest binary for your platform from the [releases page](https://github.com/MHL-tokyoDemain/kindling/releases).

### Go install

```bash
go install github.com/kindling/kindling/server@latest
```

### Uninstall

Delete the binary:

```bash
rm "$(which kindling)"
```

Remove any audit logs and configuration files manually.

### Build from source

```bash
git clone https://github.com/MHL-tokyoDemain/kindling.git
cd kindling
make build
```

The binary is written to `server/kindling`. Move it anywhere on your `$PATH`:

```bash
sudo mv server/kindling /usr/local/bin/
```

## Quick Start

### 1. Start the server

```bash
kindling server --creds /path/to/serviceAccountKey.json
```

### 2. Check health

```bash
curl localhost:9876/health
```

**Response (200):**

```json
{
  "status": "ok",
  "version": "0.1.0",
  "project_id": "my-project",
  "auth_mode": "service_account",
  "uptime_seconds": 42
}
```

### 3. Upload data

Create a JSON file with document content:

```json
[
  {"filename": "sensor.json", "content": {"temp": 22.5, "unit": "celsius"}, "content_type": "json"},
  {"filename": "note.txt", "content": "hello world", "content_type": "text"}
]
```

Upload it:

```bash
kindling upload \
  --collection projects/my-project/databases/(default)/documents/sensors \
  --files data.json
```

### 4. Upload via HTTP API

```bash
curl -X POST localhost:9876/upload \
  -H "Content-Type: application/json" \
  -d '{
    "collection": "projects/my-project/databases/(default)/documents/sensors",
    "documents": [
      {"filename": "sensor.json", "content": {"temp": 22.5}, "content_type": "json"}
    ]
  }'
```

**Response (200 — all succeeded):**

```json
{
  "success": true,
  "collection": "projects/my-project/databases/(default)/documents/sensors",
  "uploaded": [
    {"filename": "sensor.json", "document_id": "abc123", "status": "created"}
  ],
  "failed": []
}
```

**Response (207 — partial success):**

```json
{
  "success": true,
  "collection": "projects/my-project/databases/(default)/documents/sensors",
  "uploaded": [
    {"filename": "good.json", "document_id": "abc123", "status": "created"}
  ],
  "failed": [
    {"filename": "bad.json", "status": "failed", "code": "PARSE_001", "error": "failed to parse file"}
  ]
}
```

## CLI Reference

### `kindling server`

Start the HTTP server.

| Flag | Env var | Default | Description |
|------|---------|---------|-------------|
| `--port` | `KINDLING_PORT` | `9876` | Port to listen on |
| `--creds` | `GOOGLE_APPLICATION_CREDENTIALS` | `./serviceAccountKey.json` | Path to service account JSON |
| `--project` | `FIREBASE_PROJECT_ID` | — | Firebase project ID (overrides auto-detection) |
| `--max-file-size` | `KINDLING_MAX_FILE_SIZE` | `1048576` | Per-file size limit in bytes |

```bash
kindling server --port 9876 --creds ./serviceAccountKey.json
```

### `kindling upload`

One-shot batch upload without starting the server.

| Flag | Env var | Default | Description |
|------|---------|---------|-------------|
| `--collection` | — | — | Firestore collection path |
| `--files` | — | — | Comma-separated list of file paths |
| `--creds` | `GOOGLE_APPLICATION_CREDENTIALS` | `./serviceAccountKey.json` | Path to service account JSON |
| `--project` | `FIREBASE_PROJECT_ID` | — | Firebase project ID |
| `--max-file-size` | `KINDLING_MAX_FILE_SIZE` | `1048576` | Per-file size limit in bytes |
| `--concurrency` | — | `10` | Max parallel Firestore writes |

```bash
kindling upload \
  --collection projects/my-project/databases/(default)/documents/sensors \
  --files data.json,notes.txt
```

## API Endpoints

### `GET /health`

Returns server status, version, project ID, auth mode, and uptime.

**Response:** [`HealthResponse`](server/pkg/types/types.go#L31-L37)

### `POST /upload`

Upload one or more documents to Firestore.

**Request:** [`UploadRequest`](server/pkg/types/types.go#L11-L14)

| Field | Type | Description |
|-------|------|-------------|
| `collection` | `string` | Firestore collection path |
| `documents` | `[]Document` | Array of documents to upload |

**Document:**

| Field | Type | Description |
|-------|------|-------------|
| `filename` | `string` | Document filename (used for parsing) |
| `content` | `any` | Document content (string or object) |
| `content_type` | `string` | `"json"` or `"text"` |

**Response (200):** [`UploadResponse`](server/pkg/types/types.go#L24-L29)
**Response (207):** Same shape, with both `uploaded` and `failed` populated.
**Error codes:** `COLLECTION_REQUIRED`, `COLLECTION_PROJECT_MISMATCH`, `INVALID_CONTENT_TYPE`, `PARSE_001`, `EMPTY_JSON`, `PAYLOAD_TOO_LARGE`, `FIRESTORE_QUOTA`, `INTERNAL`

### `POST /shutdown`

Gracefully stops the server.

**Response (200):** `{"success": true, "status": "shutting_down"}`

## Configuration

| Setting | CLI flag | Env var | Default |
|---------|----------|---------|---------|
| Port | `--port` | `KINDLING_PORT` | `9876` |
| Credentials | `--creds` | `GOOGLE_APPLICATION_CREDENTIALS` | `./serviceAccountKey.json` |
| Project ID | `--project` | `FIREBASE_PROJECT_ID` | auto-detected from credentials |
| Max file size | `--max-file-size` | `KINDLING_MAX_FILE_SIZE` | `1048576` (1 MB) |
| Concurrency | `--concurrency` | — | `10` |
| Log level | — | — | `info` |
| Audit log path | — | — | `./kindling_audit.log` |

## VS Code Extension

A VS Code extension is planned for a future release. It will provide a UI for selecting files, configuring collection paths, and triggering uploads directly from the editor.

## Troubleshooting

### "no service account credentials found"

Kindling checks credential locations in this order:
1. `--creds` flag
2. `GOOGLE_APPLICATION_CREDENTIALS` environment variable
3. `./serviceAccountKey.json` in the working directory

Ensure one of these points to a valid Firebase service account key.

### "collection path must start with..."

Firestore collection paths must follow this format:

```
projects/<project-id>/databases/(default)/documents/<collection-name>
```

Replace `<project-id>` with your Firebase project ID and `<collection-name>` with the target collection.

### "per-file size limit exceeded"

The default limit is 1 MB per file. Increase it with `--max-file-size` (value in bytes).

### Error codes

| Code | Meaning | Status |
|------|---------|--------|
| `COLLECTION_REQUIRED` | Collection path is missing | 400 |
| `COLLECTION_PROJECT_MISMATCH` | Collection path doesn't match the bound project | 400 |
| `INVALID_CONTENT_TYPE` | Must be `"json"` or `"text"` | 400 |
| `PARSE_001` | Failed to parse file content | 400 |
| `EMPTY_JSON` | JSON content is empty | 400 |
| `PAYLOAD_TOO_LARGE` | Request body exceeds size limit | 413 |
| `FIRESTORE_QUOTA` | Firestore quota exceeded | 429 |
| `INTERNAL` | Internal server error | 500 |

## Privacy

Kindling does not collect telemetry, usage data, or crash reports. There are no background network requests, no analytics, and no update checks. The only outbound connections are the Firestore writes you explicitly configure.

## Verification

Kindling binaries are built with reproducible flags. To verify a downloaded binary:

**Checksum (SHA-256):**

```bash
sha256sum kindling
```

Compare the output against the checksum published in the release notes.

**Linux:**

```bash
file kindling
# Expected: ELF 64-bit LSB executable, x86-64, version 1 (SYSV), statically linked, stripped
```

**macOS:**

```bash
file kindling
# Expected: Mach-O 64-bit executable x86_64
```

**Windows:**

```bash
.\kindling.exe
# Expected: usage output (binary runs successfully)
```

## CI Example

Use `kindling upload` in GitHub Actions to seed test data before running integration tests:

```yaml
- name: Seed Firestore data
  run: |
    ./kindling upload \
      --collection projects/${{ vars.FIREBASE_PROJECT_ID }}/databases/(default)/documents/seeds \
      --files testdata/seed.json \
      --creds ${{ secrets.GCP_SA_KEY }}
```

## Development

### Prerequisites

- Go 1.25
- Make

### Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the binary |
| `make test` | Run unit tests with coverage |
| `make test-integration` | Run integration tests (requires Firestore emulator) |
| `make lint` | Run golangci-lint and gosec |
| `make clean` | Remove build artifacts |

## Contributing

Contributions are welcome. Please open an issue first to discuss the change.

## License

[MIT](LICENSE)
