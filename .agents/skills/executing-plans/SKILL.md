---
name: executing-plans
description: "Execute an approved implementation plan in a separate session with checkpoint reviews. Use after writing-plans when the user wants batched progress updates before more work continues."
---

# VibeMastery Executing Plans

Execute a written plan safely, in batches, with user review checkpoints.

## Hard Rules

1. Do not start if the plan is missing or unclear.
2. Follow the plan steps in order unless the user approves a change.
3. Stop immediately on blockers and ask for clarification.
4. Do not commit changes unless the user explicitly asks.

## Workflow

1. Read the selected plan file fully.
2. Restate the goal in simple terms.
3. Flag any gaps or contradictions before coding.
4. Create a task list from the plan.
5. Execute the first batch (default: 3 tasks).
6. Run all verification steps listed in the plan.
7. Report results and wait for user feedback.
8. Repeat with the next batch until done.

## Batch Execution Standard

For each task in the batch:

1. Mark task in progress.
2. Apply only the required changes.
3. Run the specified tests/checks.
4. Capture key output (pass/fail and reason).
5. Mark task complete only after verification passes.

## Report Format

Use this section structure after each batch:

- `Completed tasks`
- `What changed`
- `Verification results`
- `Open issues or blockers`
- `Ready for feedback`

## Beginner Explanation Standard

- Explain each completed task in plain English.
- Include what was done, why it was done, and how it was verified.
- Keep technical details concise but concrete (file paths, command names, test names).

## Completion Criteria

The execution session is complete only when:

1. All plan tasks are implemented and verified.
2. Final results are explained in beginner-friendly terms.
3. The user has reviewed changes.
4. No commit has been created unless explicitly requested.
