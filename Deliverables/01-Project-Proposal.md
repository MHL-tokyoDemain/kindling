# Project Proposal — Kindling

> **Spec requirement (Stage 1a, 15%):** Project definition (Aims), Outline business case, Brief project product descriptions, Outline project management structure and roles, Project approach including dependencies.

---

## 1. Project Definition / Aims

**Problem statement:**
*What unfriendly ways of batching data in Firebase currently exist? Why is manual insertion slow and inconvenient?*
Currently in firebase there is an issue for developers regarding the way you insert data into the documents. You have to open the website for firebase and write out each key, value and data type into the ui, or you can manually right it into the firebase console. both are slow.

**Project aim:**
Create an open-source tool (CLI/VS Code extension) for batch uploading data into Firebase.

**Specific objectives:**
1. Develop a CLI that accepts CSV and JSON files and inserts them into Firebase via the Admin SDK, processing at least 1,000 records per batch
2. Build a VS Code extension wrapping the CLI, providing a graphical interface for initiating uploads and viewing results
3. Support at least CSV and JSON input formats with automatic schema detection
4. Display per-record success/failure feedback with error details for failed inserts
5. Achieve test coverage of at least 70% across CLI and extension codebases

---

## 2. Outline Business Case

**Why this project?**
- *Target audience:* Developers using Firebase regularly
- *Value proposition:* Saves time on manual data entry for testing and performance benchmarking
- *Differentiation:* Open-source, free, developer-first tooling
- *Success criteria:* CLI functional at prototype stage; extension usable for common batch tasks

**Cost-benefit summary:**

| Cost Driver | Estimate | Benefit | Estimate |
| --- | --- | --- | --- |
| Developer hours | ~65 hours | Time saved per developer | ~5 hours/month |
| AI tooling costs | £100 | Community adoption | Open-source growth |

---

## 3. Brief Project Product Descriptions

| Product                    | Description                                                              | Priority         |
| -------------------------- | ------------------------------------------------------------------------ | ---------------- |
| Kindling CLI               | Command-line interface for batch Firebase insertion via the Firebase API | P0 — Must have   |
| Kindling VS Code Extension | GUI wrapper around the CLI, integrated into VS Code                      | P1 — Should have |
| Documentation              | README, usage guide, API reference                                       | P0 — Must have   |
| Test Suite                 | Unit + integration tests for core functionality                          | P1 — Should have |

---

## 4. Outline Project Management Structure and Roles

**Team:**

| Role                        | Name          | Responsibilities                                              |
| --------------------------- | ------------- | ------------------------------------------------------------- |
| Developer / Project Manager | You           | All development, planning, documentation, testing, deployment |
| Module Leader / Supervisor  | Barry Hebbron | Formative feedback, assessment guidance                       |
| Employer / Client           | Hart ltd      | Stakeholder input, real-world use case validation             |

**Communication:**
- Weekly check-ins with supervisor
- Issue tracking via GitHub Projects
- Version control via Git (see Version Control Policy, deliverable 04)

---

## 5. Project Approach Including Dependencies

**Methodology:** Agile (Scrum-inspired) (Daraojimba et al., 2024)
- 1-week sprints (suit part-time, solo capacity)
- Sprint planning + retrospective
- Backlog managed via issues
- Story points for effort estimation

**Timeline overview (6 working weeks, weekends included — primary work windows):**

| Phase | Weeks | Hours |
|---|---|---|
| Documentation | 1–2 (Jun 22–Jul 5) | ~20 |
| Development | 3–4 (Jul 6–19) | ~20 |
| Reflective report + presentation | 5–6 (Jul 20–Aug 2) | ~20 |
| Buffer | Aug 3–6 | ~5 |

**Dependencies:**

| Dependency | Type | Notes |
|---|---|---|
| Firebase API | External | Requires valid Firebase project + credentials |
| VS Code Extension API | External | Needs research — no prior experience |
| AI tooling (Claude, GPT) | Tooling | Budget-dependent model selection |
| Go runtime + Node.js | Technical | CLI built in Go; VS Code extension in TypeScript/Node.js |

**Risk mitigation:**
- Build the CLI first (known territory) before tackling the VS Code extension (unknown)
- AI handles research-heavy / unfamiliar tasks
- Code reviews on every PR to maintain quality

---

## References

Daraojimba, E. et al. (2024) 'Comprehensive review of Agile methodologies in project management', *Computer Science & IT Research Journal*, 5(1), pp. 190–218. Available at: https://doi.org/10.51594/csitrj.v5i.717 (Accessed: 20 June 2026).
