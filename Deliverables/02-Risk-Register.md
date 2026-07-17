# Risk Register — Kindling

> Part of the documentation pack required for next-week deadline.

Each risk is rated on a 1–5 scale for Probability (P) and Impact (I). Risk Score = P × I.

| ID | Risk Description | Cause | Impact Description | P | I | Score | Mitigation Strategy | Contingency Plan | Owner | Status |
|---|---|---|---|---|---|---|---|---|---|---|
| R01 | VS Code extension too complex to complete on time | No prior experience with VS Code extension API | CLI-only delivery, missing a planned feature | 4 | 3 | 12 | Research first; build CLI first as fallback; use AI to accelerate learning | Deliver CLI-only with strong documentation | You | Open |
| R02 | AI model costs exceed budget | Heavy reliance on frontier models for development | Reduced AI usage slows development | 3 | 2 | 6 | Tiered model strategy: cheap models for simple tasks, expensive models for complex ones | Set hard budget cap per sprint; switch entirely to cheaper models if needed | You | Open |
| R03 | Firebase API changes or deprecation | Google updates Firebase API | Core functionality breaks during development | 2 | 4 | 8 | Pin SDK version; monitor Firebase changelog | Abstract Firebase layer to minimise impact of changes | You | Open |
| R04 | AI-generated code introduces bugs or poor quality | AI lacks context awareness and produces inconsistent code | Increased review time; potential production issues | 4 | 3 | 12 | In-depth PR code reviews; enforce linting; maintain coding standards doc | Roll back problematic PRs; rewrite flagged sections manually | You | Open |
| R05 | Scope creep / feature bloat | Loosely-defined problem and Agile approach | Delayed delivery; missed deadline | 3 | 4 | 12 | Strict backlog prioritisation; definition of done for each sprint | Hard deadline gate — freeze features 2 weeks before hand-in | You | Open |
| R06 | Time management failure | Competing university/work commitments | Incomplete project or rushed final submission | 3 | 5 | 15 | Sprint planning with realistic velocity; buffer in schedule | Prioritise MVP features; drop P2 items | You | Open |
| R07 | Data validation failure | Malformed CSV/JSON input not properly validated | Data corruption, Firebase writes invalid data, tool crashes | 3 | 4 | 12 | Implement strict schema validation; test with edge cases and malformed files | Input sanitisation layer; fail-safe mode that rejects suspicious data | You | Open |
| R08 | Testing complexity and gaps | Difficult to mock Firebase; integration tests require real Firebase project | Undetected bugs ship to production; demo failures | 4 | 3 | 12 | Test-driven development; dedicated test Firebase project; automated test suite in CI | Manual smoke testing before demo; rollback to last known-good version | You | Open |
| R09 | Hardware or infrastructure failure | Laptop failure, corrupted repository, lost work | Days/weeks of lost development time | 2 | 5 | 10 | Daily Git commits; remote backup (GitHub); Time Machine/cloud backup for local files | Borrow replacement hardware; restore from backup; replicate environment quickly | You | Open |
| R10 | Dependency breaking changes | Firebase SDK, Go packages, or Node modules update with breaking changes mid-project | Core functionality stops working; time lost debugging/refactoring | 3 | 3 | 9 | Lock dependency versions in go.mod and package-lock.json; test updates in isolation | Pin to last working versions; defer updates until post-submission | You | Open |
| R11 | Assessment criteria misalignment | Misunderstanding module rubric; missing key deliverables or evidence | Lower grade despite functional software | 3 | 4 | 12 | Review ICA spec weekly; map deliverables to rubric; get supervisor feedback early | Retrospective documentation; add missing evidence in final week buffer | You | Open |
| R12 | Performance issues with large datasets | Tool too slow with 10K+ record files; memory issues; Firebase rate limits hit | Unusable for real-world use cases; poor demo impression | 3 | 3 | 9 | Batch writes in chunks; stream large files; benchmark with realistic data sizes | Document known limitations; demonstrate with smaller datasets; add "future work" note | You | Open |
| R13 | Open-source licensing conflict | Using incompatible licenses across dependencies | Legal issues; forced code rewrite; withdrawal from marketplace | 1 | 4 | 4 | Review all dependency licenses upfront; use permissive licenses (MIT, Apache 2.0) | Relicense project; replace problematic dependencies | You | Open |
| R14 | Documentation quality insufficient | Poor README, missing usage examples, unclear error messages | Low adoption; support burden; poor assessment grade on documentation criteria | 2 | 2 | 4 | Write docs alongside code; use AI for drafting; peer review documentation | Write comprehensive guide in final week; add FAQ section | You | Open |
| R15 | Naming/branding conflict | "Kindling" already used by another Firebase tool | Confusion; potential trademark issues; need to rebrand | 1 | 2 | 2 | Quick trademark/GitHub search before finalising name | Rename to "Firelog", "BatchFire", or similar; update all references | You | Open |
| R16 | Minor security vulnerability in dependencies | Low-severity CVE reported in transitive dependency | GitHub security alerts; potential exploit (low likelihood) | 2 | 2 | 4 | Dependabot alerts enabled; regular `npm audit` / `go mod tidy`; patch promptly | Accept risk if non-exploitable; document in security.md; defer non-critical patches | You | Open |

---

## Risk Scoring Matrix Reference

| Score | Rating | Action |
|---|---|---|
| 1–4 | Low | Accept — monitor |
| 5–9 | Medium | Active mitigation |
| 10–15 | High | Aggressive mitigation required |
| 16–25 | Critical | Must address before proceeding |

*Mark risks as "Open", "Mitigated", or "Closed" as they evolve.*
