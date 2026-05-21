---
tags: [project, overview]
created: 2026-05-21
status: planning
---

# 🔥 Kindling — Project Overview

Kindling is an open-source developer tool for batch-uploading seed data into Firebase (Firestore). It is designed to make it easy to populate a Firebase project with test/development data, eliminating the manual effort of seeding databases for testing environments.

## Purpose

> Drop batches of documents into Firebase — fast, repeatably, from your editor.

Developers working with Firebase often need to reset and re-seed Firestore with test data during development. Kindling solves this by providing a simple, interactive interface for uploading plain text and JSON files into any Firestore collection path, defined interactively at upload time.

## Key Characteristics

| Property | Value |
|---|---|
| **Type** | Developer / internal tooling |
| **Licence** | Open source |
| **Document types** | Plain text, JSON |
| **Batch size** | Small (tens of documents) |
| **Target users** | Developers using Firebase |
| **Primary interface** | VS Code extension (with CLI backing) |

## Phases

1. **Phase 1** — Console/CLI application (`kindling` CLI)
2. **Phase 2** — VS Code extension with sidebar UI (backed by Phase 1 server)

See [[Roadmap]] for detailed phase breakdown.

## Related Notes

- [[Architecture]] — System design and ADRs
- [[Tech Stack]] — Language and tooling decisions
- [[Roadmap]] — Development phases
- [[Academic Context]] — University project context and AI evaluation
