---
tags: [roadmap, planning]
created: 2026-05-21
status: planning
---

# 🗺️ Roadmap

## Phase 1 — CLI / Console App

**Goal:** A working command-line tool that can batch-upload plain text and JSON files to Firestore.

### Milestones

- [ ] Firebase project setup (Firestore, Auth, service account)
- [ ] Core upload logic — read files, parse JSON/text, write to Firestore
- [ ] Interactive Firestore path selection (prompt user for collection path at runtime)
- [ ] Batch handling — upload multiple files in a single operation
- [ ] Basic error handling and upload status reporting
- [ ] Service account authentication (via credentials JSON or env var)
- [ ] HTTP server wrapper around core logic (in preparation for Phase 2)
- [ ] `/health` endpoint for extension to poll on startup
- [ ] `/upload` endpoint accepting batch payloads
- [ ] CLI entrypoint (`kindling start`) to launch the server

---

## Phase 2 — VS Code Extension

**Goal:** A VS Code extension that provides a GUI for Kindling, backed by the Phase 1 server.

### Milestones

- [ ] Extension scaffold (`yo code` or manual)
- [ ] Extension spawns Kindling server on activation
- [ ] Sidebar panel — shows connection status, upload history
- [ ] File selection — pick files from VS Code explorer
- [ ] Drag & drop — drop files into the sidebar
- [ ] Right-click context menu — "Upload to Firebase with Kindling" on files/folders
- [ ] Command palette integration — `Kindling: Upload batch`
- [ ] Firebase Auth login flow in extension
- [ ] Interactive collection path picker (dropdown or input in sidebar)
- [ ] Upload progress and result feedback in sidebar
- [ ] Extension deactivation cleans up server process

---

## Stretch Goals

- [ ] Save/recall recently used Firestore collection paths
- [ ] Support for additional file types (CSV → JSON auto-conversion, Markdown)
- [ ] `kindling.config.json` for fully automated/scripted seeding
- [ ] VS Code Marketplace publication
- [ ] CI/CD integration — `kindling` CLI usable in GitHub Actions

---

## Timeline (Summer)

| Month | Focus |
|---|---|
| **Month 1** | Phase 1 — CLI and core upload logic |
| **Month 2** | Phase 1 complete + Phase 2 scaffold and server integration |
| **Month 3** | Phase 2 UI features + polish + academic write-up |
