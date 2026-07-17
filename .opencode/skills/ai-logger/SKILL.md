---
name: ai-logger
description: Log every user prompt and AI response summary to a plain-text log file. Use when the user wants to record, save, log, or track their conversation with the AI.
---

# AI Conversation Logger

Log directory: `/Users/tokyodemain/Desktop/ai-logs/`

## Per-session file

At the start of each conversation (first prompt), create a new file named:

```
YYYY-MM-DD_HHmm.md
```

For example: `2026-07-04_1430.md`

The HHmm uses 24-hour format based on when the session started.

## Format within each file

```
# Session — 2026-07-04 14:30  |  Branch: $(git branch --show-current)

============================================================
Prompt #1
============================================================
<prompt text>

------------------------------------------------------------
Response
------------------------------------------------------------
<one paragraph summary of the response>

---
============================================================
Prompt #2
============================================================
<prompt text>

------------------------------------------------------------
Response
------------------------------------------------------------
<one paragraph summary of the response>

---
```

- Use a session heading at the top with the date and time
- Number prompts sequentially
- Keep response summaries short (1-4 sentences)
- Separate conversation turns with `---`
