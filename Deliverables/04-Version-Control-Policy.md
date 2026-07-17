# Version Control Policy — Kindling

> Research-driven policy to govern how version control is used throughout the project.
> Every decision should be backed by **referenced justification** — avoid "we do this because that's how it's usually done."

---

## 1. Git Strategy Selection

### Options Considered

| Strategy | Description | Pros | Cons | Score |
|---|---|---|---|---|
| **GitHub Flow** | Single main branch, feature branches, PRs | Simple, fits small teams, CI-friendly | No release branches; requires discipline | *(Rate)* |
| **Trunk-Based Development** | Short-lived branches, frequent merges to main | Fast iteration, good for CI/CD | Requires feature flags; high trust needed | *(Rate)* |
| **Git Flow** | `develop`, `main`, `feature`, `release`, `hotfix` branches | Structured releases, clear history | Complex; the author now advises against it for most projects | *(Rate)* |
| **Forking Flow** | Contributors fork, then PR back | Low-trust, good for open-source | Overhead for single-dev project | *(Rate)* |

### Decision: GitHub Flow (with small alterations)

**Justification:**
- *Team size:* Solo developer: As a solo developer, working with a single main branch and then  having different feature branches off the central branch makes sense. This allows me to mostly disregard most of the busy work that comes about with large complex branching structures
- *Release cadence:* Single deadline (Aug 6th): This strategy favours the shorter development timeline I'll be working under. By removing the bloat in the strategy, it allows me to quickly iterate and merge.
- *CI/CD:* For such a small project that will have a relatively small lifetime and only one developer, there is a lack of benefits in using CI/CD workflows.
- *Failed releases:* Breaking slightly from convention, I will keep snapshots at the end of each day. This is where I deviate from the GitHub Flow slightly, but for a project that is intended to be used by a large range of people, having a snapshot to roll back to can be useful.
- *Regulatory compliance:* Kindling is an open-source developer tool with no regulated data processing requirements. No external regulatory framework (e.g. PCI-DSS, HIPAA, ISO 27001) applies. However, the policy still enforces audit trail discipline (issue-linked commits, reviewed PRs) as a professional best practice, ensuring traceability if the project were adopted in a regulated environment.

**Referenced justification:**

> **Team size:** Chacon (2011) introduced GitHub Flow as a simplified alternative to Git Flow, arguing that most teams do not require the complexity of Git Flow's multi-branch structure and that simplicity leads to fewer mistakes. Chen (2026) directly compares Git Flow — "17 branches, 5 merge targets, needs a diagram to explain" — with GitHub Flow and concludes that for solo developers, "simple, safe, and works alone" is the right fit. Antolín (2026) reinforces this: *"A good git workflow for solo developers doesn't need branching strategies designed for teams of twenty."* This simplicity is especially valuable in an AI-assisted workflow — AI accelerates code generation significantly, so the branching model must minimise administrative overhead to avoid becoming a bottleneck. Short-lived feature branches and a single main branch keep the friction between AI generation and human review low.
>
> **Release cadence:** Driessen (2010, updated 2020), the original author of Git Flow, now advises: *"If your team is doing continuous delivery of software, I would suggest adopting a much simpler workflow (like GitHub Flow)."* With a single deadline (Aug 6th) and no requirement for versioned releases, the fast iteration cycle of GitHub Flow is preferable to the release-branch overhead of Git Flow.
>
> **CI/CD:** Chen (2026) notes that trunk-based development is "great for teams with CI/CD" but "overkill for solo." For a short-lifecycle solo project, the overhead of maintaining a CI/CD pipeline delivers marginal benefit when manual testing and self-review can achieve comparable quality assurance (Forsgren et al., 2019 — DORA research shows CI/CD is correlated with high performance at scale, but the effect is less pronounced for single-developer projects with a single deployment).
>
> **Failed releases / snapshots:** The use of daily tags as rollback points is supported by W3Schools (n.d.): *"Use tags to mark release points... to roll back to previous versions if needed."* DevToolHub (2025) recommends annotated tags with semantic versioning for creating recoverable release points, and Corrux (2025) describes tags as *"vital communication tool[s]"* marking significant milestones. Daily snapshots also provide clear rollback targets if AI-generated code introduces unexpected behaviour, giving a safety net for rapid AI iterations.
>
> **Regulatory compliance:** Kindling processes no regulated data; therefore, frameworks such as PCI-DSS, HIPAA, and ISO 27001 are not applicable. However, best-practice audit trail discipline (issue-linked commits and reviewed pull requests) is retained to ensure traceability if the project is adopted in a regulated environment.

---

## 2. Branch Naming Convention

```
<type>/<short-description>
```

| Type | Example |
|---|---|
| `feature/` | `feature/cli-csv-parser` |
| `fix/` | `fix/firebase-auth-error` |
| `docs/` | `docs/readme-update` |
| `chore/` | `chore/update-dependencies` |

---

## 3. Commit Message Convention

**Format:** *Imperative mood, capitalised, no period at end*

```
<type>(<scope>): <imperative summary>

[optional body explaining why, not what]
```

| Type | Example |
|---|---|
| `feat` | `feat(cli): add CSV batch upload command` |
| `fix` | `fix(api): handle Firebase timeout gracefully` |
| `docs` | `docs(readme): add installation instructions` |
| `refactor` | `refactor(parser): extract validation to separate module` |
| `chore` | `chore(deps): update firebase-admin to 12.1` |

**Justification:**
> *Conventional Commits (ConventionalCommits.org, 2023) provides a standardised format that enables automated changelog generation and semantic versioning. Imperative mood follows Git's own convention (Torvalds, 2007) — "merge" not "merging" or "merged".*
>
> This structured format is also well-suited to AI-assisted development. AI models can reliably generate formatted commit messages from staged diffs, reducing friction while maintaining consistency across the commit history.

---

## 4. Pull Request Workflow

1. Create feature branch from `main`
2. Make changes, commit following convention
3. Open PR against `main`
4. Automated AI Review: Catch obvious bugs and errors
5. Self-review: verify code quality, readability, tests and other issues the AI didn't find
6. Merge via **squash merge** (keeps history clean)
7. Delete feature branch

**Justification:**
> *"Squash merging ensures the main branch history remains linear and readable, while preserving the full detail in the squashed commit body."*
>
> Given that AI generates a significant portion of this codebase, every PR serves as a critical human review gate. Step 4 (self-review) is where AI output is checked for correctness, readability, and appropriateness — directly addressing the quality concerns identified in the project proposal. Without this gate, AI-generated code could introduce bugs, inconsistent style, or unnecessary complexity unchecked.

---

## 5. Enforcement Mechanisms

| Mechanism | Tool | Purpose |
|---|---|---|
| Pre-commit hooks | *(e.g. husky, pre-commit)* | Enforce commit message format, run linters, auto-format AI-generated code |
| AI-assisted review | *(e.g. Copilot Code Review, ChatGPT)* | Catch issues in AI-generated code before human review; acts as a first-pass quality filter |
| PR template | GitHub PR template | Ensure every PR has context, checklist |
| Code review | Self-review checklist + AI-assisted review | Maintain code quality standards — verify AI output for correctness, readability, and appropriateness |
| CI checks | *(e.g. GitHub Actions)* | Run tests + lint on every PR |
| Commit message generation | AI tooling | Auto-generate conventional commit messages from the diff, reducing friction while maintaining consistency |
| Issue linking | GitHub — "Closes #X" | Trace commits back to project management |

---

## 6. References

- *Antolín, M. (2026). Git for solo developers: the workflow that actually works. miguelangelantolin.com*
- *Chacon, S. (2011). GitHub Flow. scottchacon.com*
- *Chen, A. (2026). The Git Workflow That Actually Works for Solo Developers. dev.to*
- *Conventional Commits. (2023). conventionalcommits.org*
- *Corrux (2025). How I utilize Git tags effectively. corrux.io*
- *DevToolHub (2025). Git Tags and Releases Best Practices. devtoolhub.com*
- *Driessen, V. (2010, updated 2020). A successful Git branching model. nvie.com*
- *Forsgren, N., Humble, J., & Kim, G. (2019). Accelerate: The Science of Lean Software and DevOps. IT Revolution Press.*
- *GitHub. (n.d.). GitHub Flow. docs.github.com*
- *Hammant, P. (n.d.). Trunk-Based Development. trunkbaseddevelopment.com*
- *Koalr (2026). GitHub Flow vs Gitflow: Which Branching Strategy Improves DORA Metrics? koalr.com*
- *Torvalds, L. (2007). Git commit message comments. git mailing list*
- *W3Schools (n.d.). Git Best Practices. w3schools.com*
