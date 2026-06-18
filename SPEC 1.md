# Kindling — Specification

## 1. Overview

Kindling is an open-source developer tool for batch-uploading seed data into Firebase Firestore. It targets developers who need to repeatedly reset and re-seed Firestore with test data during development, eliminating the manual effort of creating documents through the Firebase UI.

The project is delivered in two phases:

- **Phase 1:** A standalone Go CLI / HTTP server
- **Phase 2:** A VS Code extension that wraps the Phase 1 server in a graphical interface

The project also serves as a vehicle for evaluating AI-assisted software development (see [Academic Context](obsidian/Academic/Academic%20Context.md)).

### 1.1 Project Context

- **Sponsor:** Hart Digital
- **Constraint:** 16-week academic window, June – September 2026
- **Phase 1 ship target:** End of window
- **Sign-off authority:** Hart Digital (product acceptance); academic institution (course completion)
- **Decision log:** See DR-01 through DR-04 (inline in this spec; consolidated at end of document)
- **Window policy:** If a feature threatens the Phase 1 ship target, it is cut or deferred — not slipped. The spec's acceptance criteria (§4.5) are the gate; partial releases are not permitted.
- **Post-window track:** Multi-window VS Code server lifecycle, telemetry, i18n, a11y audit, and any other non-Phase-1 features are explicitly deferred to a post-academic release track and are not in scope during June – September 2026.

### 1.2 Milestone Schedule (16 weeks, 2 June → 22 September 2026)

| # | Milestone | Weeks | Deliverable | §4.5 gate it satisfies |
|---|---|---|---|---|
| M1 | CLI skeleton + parser | 1–2 | `kindling server` boots, `kindling upload` parses `.txt`/`.json`, unit tests for parser pass | Quality (parser coverage) |
| M2 | HTTP server + `/upload` | 3–5 | Server accepts JSON/text uploads, writes to Firestore emulator, integration tests green | Functional (core upload) |
| M3 | Authentication (service account + Firebase Auth) | 6–8 | `/auth` flow, SecretStorage integration, bearer-token session model | Functional (auth), Security |
| M4 | Hardening + error envelopes | 9–10 | All error codes documented in §4.2 implemented; rate-limit and quota handling; structured logging | Functional (error paths), Quality |
| M5 | Distribution + code signing | 11–12 | macOS/Windows/Linux installers built, signed, and verified; `brew install` works | Distribution |
| M6 | Documentation + demo + sign-off | 13–16 | README, CI example, demo script against live Firebase, Hart Digital sponsor sign-off, academic submission | Documentation, Demoable, Sign-off |

Buffer policy: each milestone has 0.5–1 week of slack absorbed into the boundary with the next milestone. If a milestone slips, the next milestone compresses before any Phase 1 ship date moves.

**Checkpoint reviews:**
- End of M2 (week 5): mid-term check-in with Hart Digital — show end-to-end upload working
- End of M4 (week 10): hardening review — show signed builds, full §4.5 item walkthrough
- End of M6 (week 16): sponsor sign-off + academic submission

---

## 2. Goals & Non-Goals

### Goals

- Upload plain text (`.txt`) and JSON (`.json`) files as Firestore documents
- Accept multiple files in a single batch operation (supports up to 10 parallel uploads for dev workloads)
- Let the user define the target Firestore collection path interactively at upload time
- Support service-account authentication for CLI usage
- Support Firebase Auth (Google/email login) for the VS Code extension (authenticated tokens stored locally to minimize timeouts, mirroring industry standards like GitHub Copilot)
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
│          VS Code Extension        │  TypeScript, VS Code API
│  sidebar | drag-drop | menu      │
│  Firebase Auth login             │
└────────────┬─────────────────────┘
             │ HTTP (localhost)
┌────────────▼─────────────────────┐
│         Kindling Server           │  Go
│  /health  /upload  /auth         │
│  batch logic  file parsing       │
└────────────┬─────────────────────┘
             │ Firebase Admin SDK
┌────────────▼─────────────────────┐
│            Firebase               │
│     Firestore (documents)         │
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
kindling server           Start the HTTP server (used by the VS Code extension and for tooling integration)
  --port <int>            Port to listen on (default: 9876)
  --creds <path>          Path to service account JSON (default: ./serviceAccountKey.json)
  --project <id>          Firebase project ID (overrides SDK auto-detection; see resolution order in §4.2)
  --max-file-size <bytes> Per-file size limit in bytes (default: 1048576)

kindling upload           One-shot batch upload without starting the server. Intended for CI pipelines, scripts, and developer terminals.
  --collection <path>     Firestore collection path (required)
  --files <paths...>      Space-separated list of file paths (required; minimum 1)
  --creds <path>          Path to service account JSON (default: ./serviceAccountKey.json)
  --project <id>          Firebase project ID (overrides SDK auto-detection; see resolution order in §4.2)
  --max-file-size <bytes> Per-file size limit in bytes (default: 1048576)
  --concurrency <int>     Max parallel Firestore writes (default: 10)
```

**When to use which:**

- `kindling server` — the default for the VS Code extension and for any HTTP-based integration. Long-lived, exposes the documented API in §4.2.
- `kindling upload` — for ad-hoc terminal use and CI jobs. Spawns no server, parses files, writes them, exits. Reuses the same parser and Firestore client code paths as the `/upload` HTTP handler (no logic duplication).

Both subcommands share the same `project_id` resolution order (flag → env → service-account JSON → SDK auto-detect) and the same collection-path binding rule (path must start with `projects/<bound-project-id>/...`). A request that would fail validation on `/upload` fails identically on the CLI.

Exit codes for `kindling upload`:

| Code | Meaning |
|---|---|
| 0 | All files uploaded successfully |
| 1 | One or more files failed (partial success still returns 0 only if every file succeeded) |
| 2 | Invalid arguments (missing flags, unreadable files, etc.) |
| 3 | Authentication failure (service account file missing or invalid) |
| 4 | Firestore-side error (quota, permission, network) |

### 4.2 HTTP API Endpoints

All endpoints return JSON. Errors use a consistent envelope:

```json
{ "success": false, "code": "STRING_CODE", "error": "human-readable message" }
```

#### `GET /health`

Returns server status. Used by the extension to verify the server is running. No authentication required.

**Response `200`:**
```json
{
  "status": "ok",
  "version": "0.1.0",
  "project_id": "my-firebase-project",
  "auth_mode": "service_account",
  "uptime_seconds": 42
}
```

- `auth_mode` is `"service_account"` (CLI on localhost) or `"firebase_auth"` (extension session).

#### `POST /auth`

Exchanges a Firebase Auth ID token for a short-lived server session. Used by the VS Code extension to authenticate end-users. The server is bound to a single project at startup (see §4.1); the `project_id` in the request is validated against that bound project.

**Request:**
```json
{
  "id_token": "eyJhbGciOi...",
  "project_id": "my-firebase-project"
}
```

**Response `200`:**
```json
{
  "success": true,
  "session_token": "sess_8f3a2b1c...",
  "expires_in": 3600,
  "uid": "user@example.com"
}
```

**Response `401`:**
```json
{ "success": false, "code": "AUTH_001", "error": "invalid or expired ID token" }
```

**Rules:**
- The server verifies the `id_token` against the Firebase Auth JWKS (cached, refreshed per Firebase key rotation schedule).
- The `session_token` is opaque (32 random bytes, base64url-encoded) and held **in server memory only** — never written to disk.
- Sessions expire after `expires_in` seconds (default 3600); the extension must re-authenticate after expiry.
- Rate-limited to 10 `/auth` attempts per source IP per minute.

#### `POST /upload`

Accepts a batch of files to upload to a Firestore collection. Authentication is via a bearer session token (extension) or implicit localhost service-account trust (CLI).

**Headers (extension):**
```
Authorization: Bearer sess_8f3a2b1c...
Content-Type: application/json
```

**Request:**
```json
{
  "collection": "projects/my-app/data/sensors",
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

**Rules:**
- The server is bound to a single `project_id` at startup (resolution order: `--project` flag → `FIREBASE_PROJECT_ID` env var → `project_id` field in the service-account JSON → SDK auto-detect). The `collection` path MUST start with `projects/<bound-project-id>/...`; any other value returns `400 COLLECTION_PROJECT_MISMATCH`.
- `content_type` is required and must be `"json"` or `"text"`. Missing or invalid values return `400 INVALID_CONTENT_TYPE`.
- `json` content is parsed into a Firestore map; `text` content is stored under a `content` field as a string.
- Each document is written with an auto-generated ID. Existing documents are never overwritten (Append-Only policy).
- Maximum request body size: 10 MB.
- Request timeout: 60 seconds.
- Concurrent uploads per request: 10 (configurable via `KINDLING_UPLOAD_CONCURRENCY`).

**Response `200` (all succeeded):**
```json
{
  "success": true,
  "collection": "projects/my-app/data/sensors",
  "uploaded": [
    { "filename": "temperature-reading.json", "document_id": "abc123", "status": "created" },
    { "filename": "log-entry.txt", "document_id": "def456", "status": "created" }
  ],
  "failed": []
}
```

**Response `207` (partial success):**
```json
{
  "success": true,
  "collection": "projects/my-app/data/sensors",
  "uploaded": [
    { "filename": "temp-01.json", "document_id": "abc123", "status": "created" }
  ],
  "failed": [
    { "filename": "invalid.json", "code": "PARSE_001", "error": "invalid JSON: unexpected token at line 1" }
  ]
}
```

**Response `400`:**
```json
{ "success": false, "code": "COLLECTION_REQUIRED", "error": "collection path is required" }
```

**Response `401`:**
```json
{ "success": false, "code": "AUTH_002", "error": "missing or invalid session token" }
```

**Response `413`:**
```json
{ "success": false, "code": "PAYLOAD_TOO_LARGE", "error": "request body exceeds 10 MB" }
```

**Error code reference:**

| Code | HTTP | Meaning |
|---|---|---|
| `COLLECTION_REQUIRED` | 400 | `collection` field missing or empty |
| `COLLECTION_PROJECT_MISMATCH` | 400 | Collection path does not start with the bound project ID |
| `INVALID_CONTENT_TYPE` | 400 | `content_type` is not `json` or `text` |
| `PARSE_001` | 207 (in `failed[]`) | File content is not valid JSON |
| `EMPTY_JSON` | 207 (in `failed[]`) | JSON file has zero bytes |
| `PAYLOAD_TOO_LARGE` | 413 | Request body exceeds configured limit |
| `AUTH_001` | 401 | Firebase ID token invalid or expired |
| `AUTH_002` | 401 | Session token missing, invalid, or expired |
| `FIRESTORE_QUOTA` | 429 | Firestore write quota exceeded; honor `Retry-After` header |
| `INTERNAL` | 500 | Unhandled server error |

#### `POST /shutdown`

Gracefully shuts down the server. Used by the extension on deactivation.

**Response `200`:**
```json
{ "success": true, "status": "shutting_down" }
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

### 4.5 Phase 1 Acceptance Criteria

Phase 1 ships when **all** of the following gates are green. Partial completion is not releasable.

**Functional**
- [ ] `kindling server` starts and serves `GET /health`, `POST /upload`, `POST /shutdown` per §4.2
- [ ] Service-account auth path works via `--creds` flag, `GOOGLE_APPLICATION_CREDENTIALS` env var, and `./serviceAccountKey.json` default
- [ ] `project_id` resolution order per §4.2 works (flag → env → service-account JSON → SDK auto-detect)
- [ ] `.json` uploads are parsed into Firestore maps; `.txt` uploads are stored as `{ content: "..." }` strings
- [ ] Firestore emulator integration tests pass (see §10)
- [ ] Append-Only policy enforced: existing documents are never overwritten
- [ ] Collection-path project binding enforced: paths not under the bound project return `COLLECTION_PROJECT_MISMATCH`
- [ ] All §8 "Edge Cases" rows have a corresponding automated test that passes
- [ ] `kindling upload` subcommand works end-to-end against the Firestore emulator per §4.1, with documented exit codes and a CI-pipeline example in the README

**Quality**
- [ ] Go server unit-test coverage ≥ 80% (`go test -cover ./...`)
- [ ] `golangci-lint run` reports zero issues
- [ ] `gosec` reports zero HIGH or CRITICAL findings
- [ ] `govulncheck` reports no known vulnerabilities in the dependency set
- [ ] `make release` (or equivalent) produces a clean build from a fresh clone in under 10 minutes

**Distribution**
- [ ] Cross-compiled binaries for the matrix in §12.1 build reproducibly via CI
- [ ] All binaries are code-signed and notarised per §12.1
- [ ] GitHub Release contains signed binaries, `SHA256SUMS.txt`, and `SHA256SUMS.minisig`
- [ ] Homebrew formula published at `kindling/tap`; `brew install kindling/tap/kindling` works on macOS amd64 + arm64
- [ ] `LICENSE` (MIT) present with named copyright holder and current year
- [ ] `CHANGELOG.md` contains a `v0.1.0` entry

**Documentation**
- [ ] `README.md` covers: install, configure, run, troubleshoot, uninstall
- [ ] All CLI flags documented in `--help` output
- [ ] All §4.2 endpoints have at least one example request and one example response in this spec
- [ ] Verification commands for each platform's signature (codesign / signtool / minisign) are in the README

**Demoable**
- [ ] Recorded walkthrough (screencast or written) showing: clone → build → configure credentials → upload a `.json` and a `.txt` to the Firestore emulator → both documents appear
- [ ] Same walkthrough repeated against a real (non-emulator) Firebase project on a throwaway dataset

**Sign-off**
- [ ] Hart Digital sponsor sign-off recorded in `RELEASES/v0.1.0.md` (product acceptance) before the release tag is pushed
- [ ] Academic submission requirements satisfied (per §1.1)

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

The sidebar (`kindling.uploadSidebar`) is the primary UI surface. It is a `WebviewView` rendered in the VS Code activity bar. It has six regions, each with explicit states:

| Region | Purpose | States |
|---|---|---|
| Connection indicator | Shows whether the bundled server is reachable | green "Server running" / red "Server offline" / amber "Authenticating" / grey "Starting" |
| Project indicator | Shows the bound Firebase project (from server startup) | "Project: my-app-dev" with refresh; "(unknown)" until server has bound |
| Collection path input | Text field for `projects/<bound-project>/databases/(default)/documents/...` | empty / valid / invalid (inline error using codes from §4.2) / "stale" if server restarted with a different binding |
| File batch | Table of files in the current batch | empty / adding / ready / uploading (per-file progress) / partial failure (failed rows highlighted, per-row retry) |
| Action bar | Upload / Cancel / Clear buttons | Upload enabled only when path valid and batch non-empty; Cancel only during active upload; Clear only when batch non-empty |
| History | Last 20 uploads (see §5.9) | per-row: file count, collection, timestamp, status; click to re-populate the batch from that upload |

All copy is in English-only for v0.1.0 (i18n deferred per §1.1). All regions are keyboard-navigable in source order (Tab); the action bar is reachable by `Tab` + `Enter`; the file batch table supports arrow-key row selection.

### 5.3 Authentication — Firebase Auth

When the extension starts without a `serviceAccountKey.json` available to the bundled server, it falls back to Firebase Auth.

**Flow:**

1. The extension opens a VS Code authentication input ("Sign in with Google"). Additional providers (email/password, anonymous) may be added in a later release; Google is the Phase 2 default.
2. The user completes Firebase Auth. The resulting Firebase ID token is sent to the server via `POST /auth` (see §4.2).
3. The server verifies the token against the Firebase Auth JWKS, creates a short-lived in-memory session, and returns an opaque `session_token`.
4. The extension stores the `session_token` (never the ID token) in the **VS Code SecretStorage API** — a per-extension, OS-encrypted store scoped to the user's VS Code profile. (The `keytar` library is intentionally not used; it is unmaintained and has known vulnerabilities.)
5. On every subsequent `/upload` call, the extension sends the `session_token` in an `Authorization: Bearer ...` header.
6. A background task in `src/auth.ts` refreshes the ID token 5 minutes before expiry, re-exchanges it at `/auth`, and updates the stored `session_token`. If refresh fails (revoked, network down, user signed out elsewhere), the user is prompted to re-authenticate — mirroring the "sign in once" UX of extensions like GitHub Copilot.

**Token lifecycle:**

| Token | Lifetime | Storage | Refresh |
|---|---|---|---|
| Firebase ID token | ~1 hour, Firebase-controlled | Not stored by client | Silent; 5 min before expiry |
| Server `session_token` | 1 hour, server-controlled | VS Code SecretStorage | On expiry → re-`POST /auth` |

**Cleanup:**

- Stored `session_token` is deleted on extension uninstall (VS Code handles SecretStorage lifecycle).
- Stored `session_token` is deleted on explicit sign-out (`Kindling: Sign Out` command).
- The server's in-memory session is destroyed on `/auth` re-issue, server restart, or 1-hour TTL expiry.

### 5.4 Server Binary Management

- The extension expects the `kindling` binary at a known location
- Bundling approach: included in the `.vsix` package (simplest for initial release)
- Future: download on first activation from GitHub Releases

### 5.5 File Batch Management

- **Add sources**: sidebar Add button, right-click menu (per §5.1), drag-and-drop onto the sidebar, command palette
- **Supported extensions**: `.json`, `.txt` only. Unsupported files are silently filtered at add time; a small "N files skipped" notice appears in the batch footer (no error toast — the user didn't do anything wrong, they just dropped a `.png`)
- **Per-batch limits**: 100 files max, 100 MB total. The UI prevents the user from queueing a batch that would be rejected server-side (matches §4.2 limits)
- **Unsaved buffers**: cannot be added. Files must exist on disk to avoid a race between save and upload
- **Hidden / system files**: filtered out (respects VS Code's `files.exclude` setting)
- **Duplicate filenames**: allowed; treated as separate documents; UI shows `(2)` suffix when names collide
- **Removal**: per-row X button, or "Clear all" from the action bar

### 5.6 Collection Path Input

- **Expected format**: `projects/<bound-project-id>/databases/(default)/documents/<collection-path>` (full Firestore resource path)
- **Validation feedback**: live, as the user types
  - Empty → no error, Upload disabled
  - Missing `projects/<bound-project-id>/` prefix → inline error "Path must start with the bound project" (path validation per §4.2, `code: COLLECTION_PATH_INVALID`)
  - Invalid characters → inline error using the matching code from §4.2
  - Valid → green checkmark, Upload enabled
- **History dropdown**: last 10 paths used, stored in workspace state (not SecretStorage — not secret)
- **Scope**: paths are remembered per workspace, not globally
- **Autocomplete**: deferred to a future release. Would require fetching the collection list from Firestore per keystroke; expensive and out of scope for the 16-week window

### 5.7 Error Surfacing & Notifications

Four channels, used by rule:

| Channel | Use for |
|---|---|
| Inline (in the relevant sidebar region) | Validation errors, per-file parse failures, connection-state changes |
| Status bar (bottom-left of VS Code) | Long-running operation progress — "Uploading 47/100…" |
| VS Code notification (toast) | Terminal outcomes: "Uploaded 100 files to test/docs in 4.2s", "Upload failed: <error code from §4.2>", "Session expired — please sign in" |
| Output channel (`Kindling` log) | Verbose detail when the user clicks "Show logs" on a toast; never the default channel |

Errors always include the machine-readable `code` from §4.2 so users can grep the docs or the source. No raw stack traces in the UI; stack traces go to the output channel.

### 5.8 Cancellation

- During an active upload, the Cancel button is enabled
- Cancel is **client-side only** in v0.1.0: the client closes the HTTP request. The server may continue the writes to completion in the background, but the response is discarded
- Documents already written before cancellation remain in Firestore (the operation is append-only and not transactional — there is no rollback)
- After cancellation, the batch UI shows three groups: **uploaded** (committed before cancel), **cancelled** (in-flight when cancel happened), **not attempted** (still in queue)
- "Best-effort" wording is shown next to the cancelled group so users understand the data state
- A real server-side cancellation hook (a `/cancel` endpoint, a `context.Done()`-aware handler, or a job-id model) is a v0.2.0 item and is explicitly out of scope for the 16-week window

### 5.9 Upload History & Retention

- Stores the last **20 uploads** in workspace state (per-workspace, per-VS-Code-profile)
- Each entry: timestamp, collection path, file count, total bytes, duration, success/failure, error code (if any)
- "Click to re-run" action: re-populates the file list from a snapshot of the original paths and the collection path, then prompts to Upload
- Files that no longer exist on disk are silently dropped, with a "N files missing" notice on the history row
- Retention: FIFO at 20 entries. Explicit `Kindling: Clear History` command available in the command palette
- No cloud sync, no telemetry. History stays on the local machine

### 5.10 Multi-Window Server Lifecycle

**Status: deferred to post-window track per §1.1.**

Multiple VS Code windows with the Kindling extension active will all attempt to bind the server to port 9876. v0.1.0 ships with a "first window wins" model:

- The first window starts the server and binds the port
- Subsequent windows see a red "Server offline (owned by another window)" indicator and use a file lock on the binary path to detect the existing instance
- They do **not** start a second server
- They cannot upload from the second window in v0.1.0 (this is the known limitation; tracked in `KNOWN_ISSUES.md`)

This is acknowledged in `README.md` and tracked as a v0.2.0 item. A real solution requires one of: (a) per-workspace server instances on different ports, (b) a single shared server with a control protocol for multiple UIs, or (c) a daemon process. None are in the 16-week window.

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

### 8.1 Audit Logging (UPDATED)

To ensure data integrity and provide an audit trail for debugging, the server writes to a local structured log file.

*   **Log File Path:** `./kindling_audit.log` (project root) or the system-specific user config directory if bundled in the `.vsix`.
*   **Format:** Structured JSON or key-value pairs.
*   **Levels:** `INFO` (successful writes), `WARN` (retries/skipped), `ERROR` (validation failures, quota exceeded).
*   **Schema:** Each log entry must capture the `timestamp`, `transaction_id` (a UUID generated per batch upload), `filename`, `collection_path`, and `outcome` (Document ID or error message).
*   **UI Integration:** The VS Code extension's error notification must include a "View Error Log" action that opens the local log file directly in the editor for easier debugging.

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

## 10. Testing & Observability Strategy

### 10.1 Test Pyramid

| Layer | Scope | Tools | When it runs |
|---|---|---|---|
| Go unit | Handlers, parser, validation, error mapping, auth token store | `go test`, `testing` stdlib, `testify` | Every commit |
| Go integration | Full HTTP server → Firestore emulator, real HTTP requests | `go test`, Firestore Emulator | Every PR |
| Go fuzz | Parser with random inputs to surface panics and bad encodings | `go test -fuzz` | Nightly |
| Extension unit | Upload manager, auth flow, SecretStorage wrapper, error display | Jest / Vitest | Every commit |
| Extension e2e | Full extension activated against a real Go server + Firestore emulator | VS Code test runner (`@vscode/test-electron`) | Pre-release |
| Manual demo | End-to-end walkthrough against emulator and real Firebase | Runbook (see §10.7) | Pre-release, M2 check-in, M6 sign-off |

### 10.2 Coverage Targets

- **Minimum line coverage on Go packages: 80%** (gates CI; matches §4.5)
- **Critical packages: 90%** — `internal/parser`, `internal/auth`, `internal/errors`
- Tracked per-package via `go test -coverprofile`; CI fails on any package falling below its target
- Extension coverage target: 70% (lower because UI is hard to cover meaningfully; integration with VS Code is the real test)

### 10.3 Emulator Setup

- **`firebase-tools` emulator suite**: Firestore + Auth emulators
- `make test-integration` boots the emulator, waits for ready, runs the test binary, tears down
- Emulator UI on `localhost:4000` for manual debugging during development
- Test fixtures in `testdata/`:
  - Small JSON files (single doc, array, nested objects)
  - Plain text files (single line, multi-line, empty, UTF-8 with BOM)
  - Malformed inputs (invalid JSON, oversized, wrong extension) — used for negative tests
  - All fixtures are synthetic; no real user data anywhere in the repo

### 10.4 Performance Budget

Measured against the Firestore emulator (real Firebase will be slower; emulator sets the lower bound):

| Operation | Budget | Test |
|---|---|---|
| `kindling server` cold start | ≤ 500ms to "listening" | `time kindling server &` until `/health` returns 200 |
| 100 × 1 KB files via `/upload` | ≤ 30s end-to-end | Integration test, single batch |
| `kindling upload` for 10 × 1 KB files | ≤ 5s parse + write | CLI integration test |
| Token verification (`/upload` with `Authorization: Bearer`) | ≤ 1ms p99 | Unit test on session-store lookup |
| Memory at idle | ≤ 30 MB RSS | Smoke test in CI |

Budgets are enforced in CI; regression > 20% fails the build. Re-baselining a budget is a code-review-level change, not a personal one.

### 10.5 Security Testing

| Threat | Test |
|---|---|
| Token replay (expired or foreign token) | Auth tests assert 401 with `code: AUTH_TOKEN_INVALID` |
| Path traversal in collection path | Parser rejects `..`, leading `/`, and non-`projects/...` prefixes (per §4.2) |
| Oversized files | `> 1 MB` (configurable via `--max-file-size`) rejected with 413 |
| Concurrent uploads | 100 simultaneous `/upload` requests don't corrupt the session store or write to wrong projects |
| Dependency vulnerabilities | `govulncheck check` in CI; `npm audit --audit-level=high` in extension CI |
| Static analysis | `gosec` (Go) and ESLint `security` plugin (TS) block PRs on new findings |
| Secret leakage | `gitleaks` in CI; no service-account JSON or signing keys ever in the repo |
| Fuzzing the parser | Nightly `go test -fuzz` run for 10 minutes; panics or `os.Exit` failures fail the job |

### 10.6 CI Matrix

Per-platform per-stage:

| Job | ubuntu-latest | macos-latest | windows-latest |
|---|---|---|---|
| Go unit + integration | ✓ | ✓ | ✓ (cross-compile check) |
| Go fuzz (nightly) | ✓ | — | — |
| `golangci-lint` | ✓ | — | — |
| `gosec` | ✓ | — | — |
| `govulncheck` | ✓ | — | — |
| Extension unit (Jest) | ✓ | — | — |
| macOS signing (codesign + notarytool) | — | ✓ | — |
| Windows signing (Azure Trusted Signing) | — | — | ✓ |
| Linux minisign | ✓ | — | — |
| Homebrew formula update | ✓ (post-`SHA256SUMS.txt` upload) | — | — |

### 10.7 Manual Release Checklist

Drives the §4.5 Demoable and Sign-off gates. Run on a clean clone of the repo at `v0.1.0` candidate:

- [ ] `make build` produces binaries for all three OSes without error
- [ ] `kindling server` boots; `curl localhost:9876/health` returns `{"status":"ok"}`
- [ ] `kindling upload --collection test/docs --files testdata/single.json --emulator` writes a document to the emulator (document visible at `localhost:4000`)
- [ ] `kindling upload` exit codes match §4.1 (covered by CLI integration tests; manual spot-check on codes 0, 2, 3)
- [ ] `POST /auth` + `POST /upload` with `Authorization: Bearer` works against the emulator
- [ ] `README.md` quick-start commands run green on a fresh Mac
- [ ] `brew install kindling/tap/kindling` installs cleanly on a clean Mac
- [ ] `codesign -dv` and `spctl -a` verify the macOS binary
- [ ] `minisign -Vm` verifies the Linux binary against the pinned public key
- [ ] Windows installer Properties → Digital Signatures shows a valid Authenticode signature
- [ ] Demo walkthrough (M6 deliverable) runs end-to-end against a throwaway real Firebase project

### 10.8 Observability & Logging

**Logging library.** `log/slog` (Go stdlib, structured JSON output).

**Log levels and when to use them.**

| Level | When |
|---|---|
| `ERROR` | Auth failures, Firestore write errors, unrecoverable per-request failures (with the error code from §4.2) |
| `WARN` | Rate-limit responses from Firestore, transient network errors, retries, deprecated-flag usage |
| `INFO` | Server start/stop, auth events (`/auth` call result, session expiry), upload summaries (file count, byte count, duration) |
| `DEBUG` | Per-request bodies, token contents, file contents — **opt-in only** via `--log-level debug`; never on by default |

**No PII rule.** Logs must never contain: file contents, document bodies, full request/response payloads, raw tokens, or service-account key material. Tests assert this with a log-capture helper.

**Log output and retention.**
- Dev: stderr (so `kindling server 2> server.log` works)
- Production: stderr → captured by the host's log collector (systemd `journalctl`, macOS unified log, Windows Event Log)
- No built-in log shipper, no remote endpoint, no analytics

**Metrics.** None in v0.1.0. The `/health` endpoint exposes process uptime and a monotonic request counter for smoke checks. Anything more (request duration histogram, error rate, etc.) is deferred to the post-window track per §1.1.

**Crash reporting.** None. The server exits with a non-zero code on unhandled panic (recovered by `slog`); no crash dump, no remote report. Post-window track per §1.1.

**Telemetry.** Explicit "no telemetry" decision for v0.1.0. Documented in `README.md` under a "Privacy" heading so users can see it without reading the spec. No usage beacons, no anonymous pings, no update checks.

---

## 11. Governance & Community

Hart Digital is putting its name on a public open-source repo. The governance doc set is the difference between "side project that might die" and "tool the community can rely on." Everything below is M6 work (weeks 13–16) but is specified now so the build can be designed to comply with it.

### 11.1 Code of Conduct

- File: `CODE_OF_CONDUCT.md` at repo root
- Content: Contributor Covenant v2.1 (https://www.contributor-covenant.org/version/2/1/code_of_conduct/)
- Enforcement contact: a Hart Digital email address (TBD by sponsor) — not the personal email of any individual contributor
- Reporting channel: the published email only — no DMs, no social media

### 11.2 Contributing Guide

- File: `CONTRIBUTING.md` at repo root
- **Dev setup**: `git clone`, `go` version pin (matches `go.mod`), `make setup` (installs `golangci-lint`, `gosec`, `firebase-tools`, `node`, `pnpm`)
- **Run the tests**: `make test` (unit), `make test-integration` (boots emulator, per §10.3), `make test-fuzz` (nightly, 10-min timeout)
- **Lint and static analysis**: `make lint` runs `golangci-lint`, `gosec`, `staticcheck`; `pnpm lint` for the extension
- **PR template**: `.github/PULL_REQUEST_TEMPLATE.md` with sections — Summary, Linked issue, Test plan (paste command output), §4.5 gates affected, Screenshots (for UI changes)
- **Commit format**: Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`); squash-merge on landing
- **DCO sign-off**: `git commit -s` (Developer Certificate of Origin 1.1) required on every commit. CLA not required.

### 11.3 Security Policy

- File: `SECURITY.md` at repo root
- **Reporting**: email to `security@kindling.dev` (or a Hart Digital alias TBD by sponsor); GPG key fingerprint published in the file
- **Response SLA**: acknowledge within 3 business days; status update within 7; coordinated disclosure window of 90 days
- **Scope**: the `kindling` binary, the `kindling-vscode` extension, and any release artefact under `github.com/kindling/...`
- **Out of scope**: Firebase SDK, Firestore emulator, third-party dependencies (file issues upstream)
- **Recognition**: reporters credited in `RELEASES/<version>.md` unless they request anonymity
- **CVEs**: published via GitHub Security Advisories on the relevant repo; CVE IDs from MITRE if the issue crosses an embargo threshold

### 11.4 Threat Model

A short, honest document, in two halves.

**What this protects against:**

- Unauthenticated writes — bearer-token model in §5.3 / DR-01
- Long-lived token leakage — opaque session tokens with 1h TTL, no ID tokens at rest
- Token theft from disk — VS Code SecretStorage is OS-keychain backed (DR-02)
- Service-account credential leakage — `gitleaks` in CI per §10.5; no service-account JSON in the repo
- Path traversal in collection paths — §4.2 validation, §10.5 tests
- Oversized upload DoS — 1 MB per file, 100 MB per batch, server-enforced
- Concurrent upload corruption — Firestore SDK is thread-safe; verified in §10.5

**What this does NOT protect against (explicit non-goals):**

- A compromised developer workstation (the OS keychain is the trust boundary; if it's broken, we are too)
- A malicious Firestore project owner (we trust whoever holds the service-account key)
- Network MITM between extension and server (loopback only in v0.1.0)
- A user pasting sensitive data into a seed file (we are a developer tool, not a secrets manager)
- Denial-of-service from the Firebase side (Firestore quota limits us; users must respect them)

This is a living document. Any change that adds a network surface, a new auth provider, or a new storage backend must update the threat model as part of the PR.

### 11.5 License

- **Repo license: Apache License 2.0**
  - Rationale: explicit patent grant, attribution clauses, well-understood by enterprise sponsors; aligns with the Firebase tooling ecosystem
  - File: `LICENSE` (full text) at repo root
- **Third-party dep policy**:
  - Go: dependencies must be Apache-2.0, MIT, BSD-2/3-Clause, MPL-2.0, or Unlicense
  - TypeScript/Node: same list, plus ISC and 0BSD
  - License check in CI: `go-licenses` (Go) and `license-checker` (Node); PR fails on GPL/AGPL/LGPL/SSPL/Commons-Clause
- **Trademark**: "Kindling" name and logo are reserved by Hart Digital; not granted by the Apache license. `TRADEMARKS.md` notes this. Sponsor must approve any use of the name in derived works or marketing

### 11.6 Known Issues

- File: `KNOWN_ISSUES.md` at repo root
- Tracked at v0.1.0:
  - **Multi-window server lifecycle** (§5.10) — only the first window can upload; subsequent windows see "Server offline (owned by another window)". v0.2.0
  - **Server-side cancellation** (§5.8) — Cancel is client-side only; documents written before cancel remain in Firestore. v0.2.0
  - **No server-side metrics** (§10.8) — only `/health` counters; no request duration or error rate. Post-window
  - **Windows ARM64** (§12.1) — not in the v0.1.0 build matrix. v0.2.0
  - **i18n, telemetry, crash reporting, full a11y audit** (§1.1) — explicitly deferred to post-academic release track
- Each entry: symptom, workaround, target version, linked issue

---

## 12. Distribution & Packaging

### 12.1 Phase 1 (CLI / Server)

**Build matrix:**

| OS | Arch | Binary |
|---|---|---|
| macOS | arm64 (Apple Silicon) | `kindling-darwin-arm64` |
| macOS | amd64 (Intel) | `kindling-darwin-amd64` |
| Linux | amd64 | `kindling-linux-amd64` |
| Windows | amd64 | `kindling-windows-amd64.exe` |

Windows ARM64 is deferred to a follow-up release; tracked separately.

**Build flags:**

```
go build -tags netgo,osusergo -trimpath \
  -ldflags="-s -w -buildid=" \
  -o kindling-<os>-<arch>
```

`-tags netgo,osusergo` and `-trimpath` produce static, location-independent binaries with zero external runtime dependencies. `-s -w` strips debug info to reduce size. An empty `-buildid` improves build reproducibility.

**Release channel:** GitHub Releases. Direct download of the signed artefact is the recommended installation path.

**Package managers:**

- **macOS / Linux:** `brew install kindling/tap/kindling` (Homebrew tap hosted at `github.com/kindling/homebrew-tap`)
- **Advanced / development:** `go install github.com/kindling/kindling/cmd/kindling@latest` — not the primary user path; corporate proxy and firewall issues make this unreliable for non-developers

**Code signing (required for all release binaries):**

| Platform | Tool | Cost | Notes |
|---|---|---|---|
| macOS | Apple Developer ID + `codesign --deep --options=runtime --timestamp` + `notarytool` | $99/yr Apple Developer Program | Signing cert stored in CI secret `MACOS_SIGNING_CERT_P12`; notarytool API key in `MACOS_NOTARY_KEY` |
| Windows | Azure Trusted Signing (free for OSS) or SignPath.io | Free (Trusted Signing) or sponsor-covered | Authenticode signature with timestamp; certificate SHA-256 thumbprint published in release notes |
| Linux | Minisign `.sig` files + SHA-256 `SHA256SUMS.txt` | Free | Public key pinned in repo and published to a public keyserver |

**CI pipeline (GitHub Actions, triggers on tag push `v*.*.*`):**

1. Build all four matrix entries on `ubuntu-latest`, `macos-latest`, and `windows-latest` runners
2. Sign each binary with the platform-appropriate tool
3. Generate `SHA256SUMS.txt` containing all four binaries
4. Sign `SHA256SUMS.txt` with the release minisign key → `SHA256SUMS.minisig`
5. Attach all artefacts to the GitHub Release
6. Trigger the `homebrew-tap` repo workflow to update the formula with new URLs and `SHA256SUMS.txt` values

**User verification (documented in `README.md`):**

- **macOS:** `codesign -dv --verbose=4 kindling-darwin-arm64` should show the Developer Team ID; `spctl -a -v kindling-darwin-arm64` should return `accepted`
- **Windows:** right-click → Properties → Digital Signatures shows a valid Authenticode signature from the publisher
- **All platforms:** `minisign -Vm kindling-<os>-<arch> -P <published-pubkey>` verifies against the pinned public key

**Why minisign for Linux?** GPG signatures are harder to verify from the command line and require a configured keyring. Minisign is a single static binary, produces deterministic signatures, and the verification command is a one-liner that works in CI and on developer workstations alike.

### 12.2 Phase 2 (Extension)

*   `.vsix` package distributed via the VS Code Marketplace.
*   The `kindling` binary is bundled inside the `.vsix` for the initial release to provide a zero-config experience.
*   Future iterations may download the appropriate binary from GitHub Releases upon first activation to reduce the marketplace package size.

---

## 13. Decision Log

This section records the non-trivial decisions made during specification. Each record captures the question, the options considered, the trade-offs, and the rationale. Decision-makers are the Hart Digital sponsor and the project lead.

### DR-01 — HTTP API authentication model

| Field | Value |
|---|---|
| **Date raised** | 2026-06-09 |
| **Decision-maker** | Project lead |
| **Status** | Resolved — Option A |
| **Implemented in** | §4.2, §5.3 |

**Question.** How should the HTTP API authenticate requests from the Phase 2 VS Code extension?

**Context.** The original spec sent Firebase ID tokens in the request body of every `/upload` call, which (a) leaked the token into logs, (b) required the extension to manage Firebase SDK calls before every request, and (c) made the server's auth surface dependent on Firebase's auth flow.

**Options considered.**

| | **A. Bearer-token session (chosen)** | **B. Firebase ID token per request** | **C. mTLS** |
|---|---|---|---|
| Server coupling | Low — opaque tokens, server doesn't know Firebase | High — server must verify Firebase ID tokens | Low — server handles certs only |
| Token lifetime | 1h, in-memory store | Matches Firebase (1h) | Cert-based |
| Compromise blast radius | Session token only | Long-lived ID token | Cert |
| Implementation cost | Medium — need `/auth` + session store | Low — reuse Firebase verifier | High — cert provisioning, rotation |

**Decision.** Option A. The extension calls `POST /auth` once with a Firebase ID token, receives an opaque `session_token` (32 random bytes, base64url), and sends it as `Authorization: Bearer <token>` on subsequent calls. The server stores sessions in memory keyed by token with a 1-hour TTL; on expiry the extension re-authenticates transparently. Server code never imports the Firebase Auth SDK; revoking access is a server-side concern.

**What it unblocks.** Clean separation between Firebase auth and server auth. Easier testing (no Firebase dependency in server tests). Better security posture (no long-lived tokens in flight).

---

### DR-02 — Token storage on the client

| Field | Value |
|---|---|
| **Date raised** | 2026-06-09 |
| **Decision-maker** | Project lead |
| **Status** | Resolved — Option B |
| **Implemented in** | §5.3 |

**Question.** Where should the VS Code extension persist the Firebase ID token and session token between sessions?

**Context.** The original spec used `keytar` (OS keychain wrapper). `keytar` is unmaintained as of 2023, has known unpatched CVEs, and is no longer recommended for new VS Code extensions.

**Options considered.**

| | **A. `keytar` (rejected)** | **B. VS Code SecretStorage API (chosen)** | **C. Encrypted local file** |
|---|---|---|---|
| Maintenance | Unmaintained | First-party, supported | DIY — must implement and audit |
| Security backing | OS keychain (when available) | OS keychain via VS Code | File encryption key must be managed |
| VS Code integration | None | Native `ExtensionContext.secrets` | None |
| Migration cost | — | Low — one API swap | High — implement crypto, key rotation |

**Decision.** Option B. The extension uses `ExtensionContext.secrets.store()` / `.get()` / `.delete()` from the VS Code API. Backed by the OS keychain on macOS/Windows/Linux. First-party supported, no third-party dependency, no `node-gyp` build, no deprecation risk during the 16-week window.

**What it unblocks.** Removal of a known-vulnerable dependency. Cleaner extension packaging (no native modules).

---

### DR-03 — `project_id` resolution and binding

| Field | Value |
|---|---|
| **Date raised** | 2026-06-09 |
| **Decision-maker** | Project lead |
| **Status** | Resolved — Option B |
| **Implemented in** | §4.1, §4.2 |

**Question.** How should `project_id` be specified, and where should the binding live?

**Context.** The original spec left `project_id` implicit (resolved per-request from the request body) and didn't specify how the CLI/server and extension should agree on which Firebase project to write to. This created a path where a single instance could write to multiple projects — a privilege-escalation risk and a UX cliff.

**Options considered.**

| | **A. Per-request `project_id` in payload (rejected)** | **B. Bound at server startup (chosen)** | **C. URL-scoped only (e.g. `/projects/:id/upload`)** |
|---|---|---|---|
| Privilege boundary | Per-request — any caller can target any project | One project per server process | One project per URL prefix |
| Spec coverage | Low (caller controls it) | High (validated server-side) | Medium (path controls it) |
| CLI ergonomics | Awkward (must pass every call) | Clean (flag/env at startup) | Awkward (per-call URL) |
| Multi-tenant risk | High | Low | Medium |

**Decision.** Option B. `project_id` is resolved once at server startup using the order: `--project` flag → `KINDLING_PROJECT` env var → `project_id` field in the service-account JSON → Firebase Admin SDK auto-detection. The resolved value is bound for the lifetime of the process; the server rejects any collection path that doesn't begin with `projects/<bound-id>/...`. The `/upload` payload and the `kindling upload` CLI both omit `project_id` from their inputs.

**What it unblocks.** Predictable authorisation model, cleaner UI (extension shows the bound project, no project picker per call), testable collection-path validation, single-tenant safety.

---

### DR-04 — Scope of `kindling upload` subcommand

| Field | Value |
|---|---|
| **Date raised** | 2026-06-09 |
| **Decision-maker** | Hart Digital sponsor |
| **Status** | Resolved — Option A |
| **Implemented in** | §4.1, §4.5 |

**Question.** Should `kindling upload` ship as a Phase 1 deliverable alongside `kindling server`, or be deferred to v0.2.0?

**Context.** Phase 1's primary deliverable is the HTTP server consumed by the Phase 2 extension. The CLI subcommand would let the same upload logic run without a server — useful for terminal use, CI jobs, and developer-side testing. The 16-week window forces an honest trade-off: every additional surface is more code, tests, documentation, and sign-off work.

**Options considered.**

| | **A. Ship in Phase 1 (chosen)** | **B. Defer to v0.2.0** |
|---|---|---|
| LOC added | ~150 (CLI wiring, flag parsing, exit codes, tests) | 0 |
| Test surface added | CLI integration tests, flag-validation tests, exit-code tests | 0 |
| Demo value | End-to-end "user pastes a file, data lands in Firestore" without the extension | Server only |
| Risk to Sept deadline | Small but additive | Removes one variable |
| Long-term value | Real — Hart Digital engineers will use it in CI | Real but later |

**Decision.** Option A. Reasons:

1. The parser, Firestore client, and validation rules already exist (or are being built) for the HTTP handler. The CLI is a thin wrapper that reuses them — no logic duplication, so the real LOC cost is closer to 100 than 150.
2. The CLI is the simplest possible end-to-end demo for Hart Digital's sponsor review. It works without VS Code, without the extension, without a running server — the lowest-friction way to show progress at the mid-term check-in (week 5).
3. The CI use case is real for a dev-tools sponsor: Hart Digital will plausibly want a scriptable interface from day one, not in a follow-up release.

**Risk if wrong.** Low. The CLI is a thin wrapper; if it slips, the server (the Phase 2 dependency) is unaffected. We can still ship Phase 1 in September with Option A's contract in §4.1 unimplemented by gutting it back to Option B at a cost of one day, not the project.

---
