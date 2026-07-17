# Methodology Outline — Kindling

> Defines how the project will be developed, managed, and delivered.

---

## 1. Chosen Methodology: Agile (Scrum-Inspired)

**Why Agile?**
- Problem is loosely defined — requirements will emerge during development (Daraojimba et al., 2024)
- No prior experience with VS Code extensions → need flexibility to adapt (Lindvall et al., 2002)
- AI-assisted development means iteration speed is high
- Solo developer → ceremonies are lightweight but still provide structure

**Why not Waterfall?**
- Waterfall assumes all requirements are known upfront — not the case here
- Risk of discovering late that a feature is infeasible (especially VS Code extension)
- The Standish Group (2020) found that only 31% of projects succeed under traditional plan-driven approaches, with 66% ending in partial or total failure — changing requirements late in the lifecycle is a primary driver of this

---

## 2. Sprint Structure

| Element              | Detail                                                 |
|--------------------|------------------------------------------------------|
| Sprint length        | 1 week — shorter sprints suit part-time, solo capacity (Gant, 2019; Metzger, 2018) |
| Sprint planning      | Start of sprint — define goals, select backlog items   |
| Daily standup        | Solo: self-check-in / brief journal entry              |
| Sprint review        | End of sprint — what was delivered                     |
| Sprint retrospective | End of sprint — what went well, what to improve        |
| Backlog refinement   | Ongoing — triage and estimate new items                |

**Timeline:**

*Start date: June 22 — Deadline: August 6 (6 working weeks)*  
*Capacity: ~10 hours/week (evenings + weekends alongside full-time work)*

| Week | Dates | Phase | Hours | Goals |
|---|---|---|---|---|
| 1 | Jun 22–28 | Documentation Sprint 1 | ~10 | Finalise project proposal, risk register, risk matrix, version control policy |
| 2 | Jun 29–Jul 5 | Documentation Sprint 2 | ~10 | Finalise methodology outline, cost analysis, presentation draft |
| 3 | Jul 6–12 | Development Sprint 1 | ~10 | Project setup, CLI prototype, Firebase API integration |
| 4 | Jul 13–19 | Development Sprint 2 | ~10 | CLI polish + testing; VS Code extension research + prototype |
| 5 | Jul 20–26 | Reflective Sprint 1 | ~10 | Write first draft of reflective report (~2,500 words); send to professors for feedback |
| 6 | Jul 27–Aug 2 | Reflective Sprint 2 | ~10 | Incorporate feedback, finalise report; polish presentation; any last bug fixes |
| *Buffer* | *Aug 3–6* | *Slack* | *~5* | *Final review of all submissions, submission logistics, contingency for overruns* |

**Total estimated: ~65 hours**

---

## 3. Tools

| Tool              | Purpose                                |
|-----------------|--------------------------------------|
| GitHub            | Version control, issue tracking, CI/CD |
| GitHub Projects   | Sprint board, backlog management (Schwaber and Sutherland, 2020)       |
| Opencode          | AI-assisted development                |
| Go Runtime + node | Development runtime                    |
| Testify + vitest  | Testing framework                      |

---

## 4. Quality Assurance

- Every PR must pass linting + tests before merge
- Code review checklist for each PR:
  - Does the code work?
  - Is it readable? (sensible names, no dead code, consistent style)
  - Are there tests?
  - Is AI-generated code reviewed critically?
- Definition of Done (Beck et al., 2001):
  - Feature implemented
  - Tests written + passing
  - Code reviewed (self-review + AI-assisted review)
  - Documentation updated if needed

---

## 5. References

Beck, K. et al. (2001) *Manifesto for Agile Software Development*. Available at: https://agilemanifesto.org (Accessed: 20 June 2026).

Daraojimba, E. et al. (2024) 'Comprehensive review of Agile methodologies in project management', *Computer Science & IT Research Journal*, 5(1), pp. 190–218. Available at: https://doi.org/10.51594/csitrj.v5i.717 (Accessed: 20 June 2026).

Digital.ai (n.d.) *The benefits of Agile software development*. Available at: https://digital.ai/glossary/agile-software-development-benefits (Accessed: 20 June 2026).

Gant, M. (2019) *Scrum and the solo dev*. Available at: https://medium.com/swlh/scrum-and-the-solo-dev-fb8e810ed42b (Accessed: 20 June 2026).

Lindvall, M. et al. (2002) 'Empirical findings in Agile methods', *Lecture Notes in Computer Science*. Available at: https://www.cs.umd.edu/~mvz/pub/agile.pdf (Accessed: 20 June 2026).

Metzger, E. (2018) *How to Scrum as a one person operation*. Available at: https://medium.com/hackernoon/how-to-scrum-for-one-man-operations-e8fc0dc5a58c (Accessed: 20 June 2026).

Schwaber, K. and Sutherland, J. (2020) *The Scrum Guide*. Available at: https://scrumguides.org (Accessed: 20 June 2026).

Standish Group (2020) *CHAOS 2020: Beyond Infinity*. The Standish Group.
