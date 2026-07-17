---
tags: [tech-stack, tooling]
created: 2026-05-21
status: in-progress
---

# 🛠️ Tech Stack

## Decided

| Layer | Technology | Reason |
|---|---|---|
| **VS Code Extension** | TypeScript + VS Code Extension API | Required — VS Code extensions must be TypeScript/JavaScript |
| **Firebase** | Firestore | Document storage for seed data |
| **Auth (CLI)** | Firebase Admin SDK + Service Account | No interactive login for scripted use |
| **Auth (Extension)** | Firebase Auth (Google / email) | User-facing login in the extension UI |
| **Document types** | Plain text (`.txt`), JSON (`.json`) | Initial scope; more types can be added |

## Undecided

### Core Server Language

The server (core logic) can be written in any language since it communicates with the extension over HTTP. Options:

| Language | Pros | Cons |
|---|---|---|
| **TypeScript / Node.js** | Same language as extension, share types/utils, Firebase JS SDK is excellent | Runtime required (Node), not a single binary |
| **Python** | Familiar for many devs, rich ecosystem, easy scripting | Runtime required, not a single binary |
| **Go** | Compiles to a single binary (easy to distribute), fast startup, great HTTP stdlib | Less familiar to most JS/TS devs, separate Firebase SDK |

> **Note:** If distributing as a VS Code extension via the Marketplace, bundling a Go binary or requiring Node.js are both viable but have different packaging implications. TypeScript compiled to a Node bundle is the simplest distribution story.

## To Decide

- [ ] Core server language
- [ ] HTTP framework for the server (Express, Fastify, Flask, Gin, etc.)
- [ ] How the binary/server is distributed alongside the extension (bundled vs downloaded)
- [ ] Local port strategy (fixed vs dynamic)
- [ ] Testing framework

## Notes

- The VS Code extension is always TypeScript — this is non-negotiable (VS Code API requirement)
- The extension communicates with the server over `localhost` HTTP — the server language is an implementation detail
- Firebase JS/TS SDK (`firebase-admin`, `firebase`) has the best documentation and community support
