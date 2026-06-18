# Kindling — Specification

## 1. Overview

Kindling is an open-source developer tool for batch-uploading seed data into Firebase Firestore. It targets developers who need to repeatedly reset and re-seed Firestore with test data during development, eliminating the manual effort of creating documents through the Firebase UI.

The project is delivered in two phases:

- **Phase 1:** A standalone Go CLI / HTTP server
- **Phase 2:** A VS Code extension that wraps the Phase 1 server in a graphical interface

The project also serves as a vehicle for evaluating AI-assisted software development (see [Academic Context](obsidian/Academic/Academic%20Context.md)).

---

## 2. Goals & Non-Goals

### Goals

- Upload plain text (`.txt`) and JSON (`.json`) files as Firestore documents
- Accept multiple files in a single batch operation
- Let the user define the target Firestore collection path interactively at upload time
- Support service-account authentication for CLI usage
- Support Firebase Auth (Google/email login) for the VS Code extension
- Provide a local HTTP server that the VS Code extension can communicate with
- Deliver a VS Code extension with sidebar, drag-and-drop, context menu, and command palette integration
- Automatically manage the server lifecycle within the extension

### Non-Goals

- Real-time Firestore sync or two-way editing
- Support for binary file types (images, PDFs, etc.)
- Complex data transformation or schema migration
- A web dashboard
- Multi-user collaboration
- Cloud-hosted service

---

## 3. Architecture

### 3.1 System Diagram

```
┌──────────────────────────────────┐
│          VS Code Extension       │  TypeScript, VS Code API
│  sidebar | drag-drop | menu      │
│  Firebase Auth login             │
└────────────┬─────────────────────┘
             │ HTTP (localhost)
┌────────────▼─────────────────────┐
│         Kindling Server          │  Go
│  /health  /upload  /auth         │
│  batch logic  file parsing       │
└────────────┬─────────────────────┘
             │ Firebase Admin SDK
┌────────────▼─────────────────────┐
│            Firebase              │
│     Firestore (documents)        │
└──────────────────────────────────┘
```

### 3.2 Local Server Pattern

The core application runs as a local HTTP server (Go binary). The VS Code extension spawns this server on activation and kills it on deactivation. The user never manually runs the server.

### 3.3 Port Strategy

- The server listens on `localhost` on a configurable port
- Default: `9876`
- The extension probes `localhost:9876/health` to confirm the server is ready
- Port conflicts are reported to the user with instructions

### 3.4 Extension Lifecycle

```
VS Code loads → extension activates
  → locates kindling binary (bundled or downloaded)
  → spawns `kindling server --port 9876`
  → polls GET /health until 200
  → registers sidebar, commands, context menus

User closes VS Code → extension deactivates
  → sends POST /shutdown (or SIGTERM)
  → kills process if it does not exit within 5 seconds
```

---

## 4. Phase 1 — CLI / Core Server (Go)

### 4.1 CLI Commands

```
kindling server           Start the HTTP server
  --port <int>            Port to listen on (default: 9876)
  --creds <path>          Path to service account JSON (default: ./serviceAccountKey.json)
  --project <id>          Firebase project ID (overrides SDK auto-detection)

kindling upload           (Future) Direct CLI upload without server
  --collection <path>     Firestore collection path
  --files <paths...>      Space-separated list of file paths
  --creds <path>          Path to service account JSON
```

The primary mode is `kindling server`; the `upload` subcommand is a convenience.

### 4.2 HTTP API Endpoints

#### `GET /health`

Returns server status. Used by the extension to verify the server is running.

**Response `200`:**
```json
{
  "status": "ok",
  "version": "0.1.0",
  "uptime_seconds": 42
}
```

#### `POST /upload`

Accepts a batch of files to upload to a Firestore collection.

**Request:**
```json
{
  "collection": "projects/my-app/data/sensors",
  "project_id": "my-firebase-project",
  "auth": {
    "type": "service_account",
    "credentials_path": "/path/to/serviceAccountKey.json"
  },
  "documents": [
    {
      "filename": "temperature-reading.json",
      "content": "{\"sensor\": \"temp-01\", \"value\": 23.5, \"unit\": \"celsius\"}",
      "content_type": "json"
    },
    {
      "filename": "log-entry.txt",
      "content": "Sensor calibration completed at 2026-05-21T10:00:00Z",
      "content_type": "text"
    }
  ]
}
```

- `content_type` determines how the document is stored: `json` content is parsed into a Firestore map, `text` content is stored under a `content` field as a string.
- Each document is written as a new document with an auto-generated ID.

**Response `200`:**
```json
{
  "success": true,
  "collection": "projects/my-app/data/sensors",
  "uploaded": [
    {
      "filename": "temperature-reading.json",
      "document_id": "abc123",
      "status": "created"
    },
    {
      "filename": "log-entry.txt",
      "document_id": "def456",
      "status": "created"
    }
  ],
  "failed": []
}
```

**Response `207` (Partial success):**
```json
{
  "success": true,
  "collection": "projects/my-app/data/sensors",
  "uploaded": [
    { "filename": "temp-01.json", "document_id": "abc123", "status": "created" }
  ],
  "failed": [
    {
      "filename": "invalid.json",
      "error": "invalid JSON: unexpected token at line 1"
    }
  ]
}
```

**Response `400`:**
```json
{
  "success": false,
  "error": "collection path is required"
}
```

#### `POST /shutdown`

Gracefully shuts down the server. Used by the extension on deactivation.

**Response `200`:**
```json
{
  "status": "shutting_down"
}
```

### 4.3 File Handling Rules

| Content Type | Extension | Storage Behaviour |
|---|---|---|
| `json` | `.json` | Parsed as JSON object → stored as Firestore document fields |
| `text` | `.txt` | Stored as `{ content: "<file contents>" }` |

- Non-JSON-parsable `.json` files are rejected with a parse error
- Empty files are uploaded as `{ content: "" }` (text) or rejected (JSON)
- File size limit: 1 MB per file (configurable)

### 4.4 Authentication — Service Account

- The CLI reads credentials from a `serviceAccountKey.json` file
- Default path: `./serviceAccountKey.json` (project root)
- Override via `--creds` flag or `GOOGLE_APPLICATION_CREDENTIALS` env var
- The Go server uses the Firebase Admin SDK to initialise with these credentials

---

## 5. Phase 2 — VS Code Extension (TypeScript)

### 5.1 Entry Points

| Trigger | Action |
|---|---|
| **Sidebar** | `kindling.uploadSidebar` — open sidebar with file list, collection input, upload button |
| **Command palette** | `Kindling: Upload Batch` — opens a file picker, then uploads |
| **Right-click (file)** | `Kindling: Upload to Firebase` — adds file to pending batch, shows sidebar |
| **Right-click (folder)** | `Kindling: Upload Folder to Firebase` — adds all supported files from folder |
| **Drag & drop** | Drop files onto sidebar panel to add them to the batch |

### 5.2 Sidebar Panel

The sidebar contains:

1. **Connection indicator** — green dot + "Server running" / red dot + "Server offline"
2. **Collection path input** — text field where user enters Firestore path (e.g. `projects/my-app/data/sensors`)
3. **File list** — table of files added to the current batch, each with filename, type, size, status
4. **Add files button** — opens system file picker
5. **Upload button** — sends the batch to the server
6. **History section** — list of recent uploads (filename, collection, timestamp, status)

### 5.3 Authentication — Firebase Auth

- When the extension starts without a `serviceAccountKey.json`, it falls back to Firebase Auth
- The extension initiates an OAuth flow (Google sign-in via Firebase Auth)
- The resulting token is passed to the server via a one-time `/auth` endpoint (or included in upload requests)
- The server exchanges the token for Firestore credentials using the Firebase Admin SDK

### 5.4 Server Binary Management

- The extension expects the `kindling` binary at a known location
- Bundling approach: included in the `.vsix` package (simplest for initial release)
- Future: download on first activation from GitHub Releases

---

## 6. Data Flow

### 6.1 Upload Flow (CLI)

```
User selects files
  → CLI parses each file (JSON → object, text → string)
  → CLI prompts for collection path
  → CLI authenticates via service account
  → CLI writes each document to Firestore via Admin SDK
  → CLI prints result summary
```

### 6.2 Upload Flow (Extension)

```
User drops files onto sidebar
  → Extension reads file contents (via VS Code API)
  → Extension builds /upload payload
  → User enters collection path in sidebar input
  → User clicks Upload
  → Extension POSTs to localhost:9876/upload
  → Server writes to Firestore via Admin SDK
  → Server returns result
  → Extension displays result in sidebar
```

---

## 7. Configuration

The server can be configured via:

| Setting | CLI Flag | Env Var | Default |
|---|---|---|---|
| Port | `--port` | `KINDLING_PORT` | `9876` |
| Service account path | `--creds` | `GOOGLE_APPLICATION_CREDENTIALS` | `./serviceAccountKey.json` |
| Firebase project ID | `--project` | `FIREBASE_PROJECT_ID` | (SDK auto-detect) |
| Max file size (bytes) | `--max-file-size` | `KINDLING_MAX_FILE_SIZE` | `1048576` |

The VS Code extension reads settings from `settings.json`:

```json
{
  "kindling.port": 9876,
  "kindling.serverPath": "/path/to/kindling",
  "kindling.maxFileSize": 1048576
}
```

---

## 8. Edge Cases & Error Handling

| Scenario | Behaviour |
|---|---|
| Server already in use (port conflict) | Server exits with error message; extension shows error notification |
| Server binary not found | Extension shows error with path to binary; offers to download |
| Network error during upload | Retry up to 3 times with exponential backoff; report failure |
| Invalid JSON file | Rejected with parse error message; other files in batch still upload |
| Empty batch (no files) | Upload button disabled until files are added |
| Collection path is empty | Upload button disabled; inline validation message |
| Service account file missing | CLI shows clear error; extension falls back to Firebase Auth |
| Firebase Auth token expired | Extension re-authenticates; user sees login prompt again |
| Firestore write quota exceeded | Server returns error; extension shows quota warning |

---

## 9. Project Structure

```
kindling/
├── SPEC.md                       # This document
├── LICENSE                       # MIT
├── README.md                     # Project overview
├── obsidian/                     # Design notes (Obsidian vault)
│   ├── Kindling - Project Overview.md
│   ├── Architecture/Architecture.md
│   ├── Development/Roadmap.md
│   └── Development/Tech Stack.md
├── server/                       # Go module root
│   ├── go.mod
│   ├── main.go                   # Entrypoint / CLI
│   ├── cmd/
│   │   └── server.go             # `kindling server` subcommand
│   ├── internal/
│   │   ├── server/               # HTTP server, handlers, middleware
│   │   │   ├── server.go
│   │   │   ├── health.go
│   │   │   └── upload.go
│   │   ├── firestore/            # Firebase Admin SDK wrapper
│   │   │   └── client.go
│   │   └── parser/               # File parsing logic
│   │       └── parser.go
│   └── pkg/
│       └── types/                # Shared types
│           └── types.go
├── extension/                    # VS Code extension root
│   ├── package.json
│   ├── tsconfig.json
│   ├── src/
│   │   ├── extension.ts          # Entrypoint: activate/deactivate
│   │   ├── serverManager.ts      # Spawn/manage/kill the Go server
│   │   ├── sidebarPanel.ts       # Webview/sidebar provider
│   │   ├── uploadManager.ts      # Build requests, handle responses
│   │   └── auth.ts               # Firebase Auth login flow
│   ├── media/                    # Icons, assets
│   └── test/
│       └── ...
└── scripts/                      # Build, package, release scripts
    └── ...
```

---

## 10. Testing Strategy

| Layer | Approach | Tools |
|---|---|---|
| **Go server (unit)** | Test handlers, parser, Firestore wrapper individually | `go test` + `testing` stdlib |
| **Go server (integration)** | Spin up server against Firestore emulator, send real HTTP requests | `go test`, Firestore Emulator |
| **Go server (mock)** | Mock Firestore client to test handler logic without network | `testify/mock` |
| **Extension (unit)** | Test upload manager, auth flow | Jest / Vitest |
| **Extension (e2e)** | Run extension against a real Go server + Firestore emulator | VS Code test runner |

---

## 11. Distribution & Packaging

| Phase | Distribution Method |
|---|---|
| **Phase 1 (CLI)** | Go cross-compile binaries for macOS (arm64 + amd64), Linux (amd64), Windows (amd64); published as GitHub Release assets; installable via `brew`, direct download, or `go install` |
| **Phase 2 (Extension)** | `.vsix` package distributed via VS Code Marketplace; `kindling` binary bundled inside the `.vsix` for initial release |
