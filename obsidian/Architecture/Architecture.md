---
tags: [architecture, design]
created: 2026-05-21
status: decided
---

# 🏗️ Architecture

## Overview

Kindling uses a **local server pattern**: the core application runs as a local HTTP server, and the VS Code extension acts as a thin UI shell that communicates with it. The extension automatically starts and stops the server as part of its lifecycle — the user never has to manually run anything.

```
┌─────────────────────────────┐
│       VS Code Extension     │  ← UI layer (TypeScript, VS Code API)
│  sidebar | drag-drop | menu │
└────────────┬────────────────┘
             │ HTTP (localhost)
┌────────────▼────────────────┐
│       Kindling Server       │  ← Core logic (language TBD)
│  upload | validate | batch  │
└────────────┬────────────────┘
             │ Firebase Admin SDK / Firebase Auth
┌────────────▼────────────────┐
│          Firebase           │
│     Firestore (documents)   │
└─────────────────────────────┘
```

## Key Architectural Decisions

### ADR-001 — Local Server Pattern

**Decision:** The CLI core runs as a local HTTP server. The VS Code extension spawns it on activation and kills it on deactivation.

**Rationale:**
- Clean separation of concerns — business logic lives in the server, not the extension
- Server can be used independently as a CLI tool (Phase 1)
- Allows the server to be written in any language
- Well-established pattern (mirrors how Language Servers work in VS Code)
- Academically interesting — clear layered architecture to analyse and write about

**Trade-offs:**
- Port management (need to pick/assign a port, handle conflicts)
- Process lifecycle management in the extension
- Slight latency overhead vs. direct function calls

---

### ADR-002 — Dual Authentication Strategy

**Decision:** Support both **service account** (Admin SDK) for CLI/automated use, and **Firebase Auth** (user login) for the VS Code extension.

**Rationale:**
- Service account: better for CI/CD pipelines and scripted seeding — no interactive login needed. Credentials stored as a JSON file per project.
- Firebase Auth: better for interactive use in the extension — a developer logs in once and the extension manages the session.

**Implementation:**
- CLI mode: reads `serviceAccountKey.json` from project root (or `GOOGLE_APPLICATION_CREDENTIALS` env var)
- Extension mode: handles OAuth login flow via Firebase Auth

---

### ADR-003 — Interactive Firestore Path Selection

**Decision:** Target Firestore collection paths are defined **interactively at upload time**, not via a static config file.

**Rationale:**
- The tool is general-purpose; it should not assume a fixed schema
- Developers have different Firestore structures per project
- Interactive prompts keep the tool flexible and zero-config
- Future enhancement: save/recall recent paths as favourites

**Trade-offs:**
- Less automation-friendly than a config file approach
- Could be extended later with a `kindling.config.json` for scripted/repeatable seeding

---

## Extension Lifecycle

```
VS Code loads → extension activates
  → finds/downloads kindling-server binary
  → spawns server on available local port
  → waits for /health endpoint
  → registers sidebar, commands, context menus

User closes VS Code → extension deactivates
  → sends shutdown signal to server
  → kills process if it doesn't exit cleanly
```

## VS Code Extension UX Features

| Feature | Description |
|---|---|
| **Sidebar panel** | Shows upload history, status, and connection state |
| **File selection** | Pick files from the explorer to add to a batch |
| **Drag & drop** | Drop files into the sidebar panel |
| **Right-click menu** | Right-click files/folders → "Upload to Firebase with Kindling" |
| **Command palette** | `Kindling: Upload batch` command |
