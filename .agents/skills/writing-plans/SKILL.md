---
name: writing-plans
description: "Beginner-first implementation planning from an approved design. Use after brainstorming, or when the user already approved requirements and wants a step-by-step Laravel implementation plan before coding."
---

# VibeMastery Writing Plans

Turn an approved idea into a practical, beginner-friendly implementation plan.

## Hard Rules

1. Do not write implementation code in this skill.
2. Do not commit anything. The user must review first.
3. Ask one question at a time if details are missing.

## Goals

1. Explain what will be built in plain English.
2. Break work into small, testable steps.
3. Explain the workflow and why the order matters.
4. Save the plan to a user-approved path.

## Workflow

1. Restate the approved scope in simple terms.
2. Identify constraints (Laravel version, auth, admin, data model, testing scope).
3. Split implementation into phases and tasks.
4. For every task, include:
   - purpose (why this task exists)
   - exact files to create/modify
   - implementation steps
   - verification steps
   - rollback/fix hint if verification fails
5. Include testing tasks early (unit/feature/browser where relevant).
6. Explain why this order is safest for beginners.
7. Ask where to save the plan, then write it.

## Ask This Before Saving

Use this exact question:

"Where should I save your implementation plan? Press Enter for `docs/plans/`, or share another folder/path."

If the user presses Enter, default to `docs/plans/`.

Suggested filename format:

`YYYY-MM-DD-<topic>-implementation-plan.md`

## Output Structure

Always respond using these sections:

- `Goal in simple terms`
- `Plan strategy`
- `Step-by-step tasks`
- `Workflow and why`
- `Verification checkpoints`
- `Plan file path`
- `Review before coding`

## Beginner Explanation Standard

- Use short sentences and plain words.
- Define Laravel jargon in one line when needed.
- Give concrete examples (file names, route names, command examples).
- Prefer "what + why + how to verify" for each task.

## Completion Criteria

This planning session is complete only when:

1. The step-by-step plan is complete and clear.
2. The plan is saved to a user-approved location.
3. The user is asked to review the plan before coding starts.
4. No commit has been created.

## Execution Handoff

After saving the plan, offer both execution modes:

"Plan complete and saved. Choose your execution style:\n1) `subagent-driven-development` (same session, task-by-task)\n2) `executing-plans` (separate session, batch checkpoints)"

If the user chooses option 1:

- Announce: "I am now using `subagent-driven-development` to execute this plan in this session."
- Stay in the current session.

If the user chooses option 2:

- Announce: "Please open a separate session and run `executing-plans` with this plan file."
- Share the exact plan file path.
