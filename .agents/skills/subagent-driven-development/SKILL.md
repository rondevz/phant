---
name: subagent-driven-development
description: "Execute an approved implementation plan in the current session by dispatching one subagent per task, with spec and quality reviews after each task. Use after writing-plans when tasks are mostly independent."
---

# VibeMastery Subagent-Driven Development

Implement plan tasks using focused subagents and strict review gates.

## Hard Rules

1. Run only one implementation subagent at a time.
2. Require two review gates per task:
   - spec compliance review
   - code quality review
3. Fix all review issues before moving to the next task.
4. Do not commit changes unless the user explicitly asks.

## Workflow

1. Read the plan once and extract all tasks.
2. For each task:
   - dispatch implementer subagent with full task text + context
   - answer clarifying questions before implementation continues
   - run task verification checks
   - dispatch spec compliance reviewer
   - if issues: implementer fixes and spec reviewer re-checks
   - dispatch code quality reviewer
   - if issues: implementer fixes and reviewer re-checks
   - mark task complete
3. After all tasks, run one final end-to-end review.
4. Present results and request user review.

## Prompt Templates

Use these helper templates when dispatching subagents. They are located in the same directory as this skill file (e.g. `skills/subagent-driven-development/` or `.claude/skills/vibemastery-toolkit/subagent-driven-development/` depending on your install):

- `implementer-prompt.md`
- `spec-reviewer-prompt.md`
- `code-quality-reviewer-prompt.md`

Read the relevant file and use its contents as the subagent prompt, filling in the task text and context placeholders.

## Per-Task Output Format

After each task, report:

- `Task completed`
- `Files changed`
- `Spec review result`
- `Quality review result`
- `Verification result`

## Beginner Explanation Standard

- Keep wording simple and direct.
- Explain what changed and why it matters.
- Translate reviewer findings into plain English.

## Completion Criteria

This workflow is complete only when:

1. Every task passes both review gates.
2. Plan verification checks pass.
3. Results are explained clearly for beginners.
4. No commit has been created unless explicitly requested.
