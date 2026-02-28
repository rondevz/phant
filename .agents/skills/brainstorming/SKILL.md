---
name: brainstorming
description: "Beginner-first idea-to-plan workflow for Laravel projects. Use before writing implementation code when the user wants to brainstorm features, shape requirements, compare options, validate architecture, and create a reviewed high-level plan for writing-plans."
---

# VibeMastery Brainstorming

Turn rough ideas into a clear, beginner-friendly high-level plan.

## Hard Rules

1. Do not write implementation code during this skill.
2. Do not commit changes. The user must review everything first.
3. Ask one question at a time when clarification is needed.
4. Never work on `main` (or the default branch). Check the current branch at the start — if on `main`, create a feature branch before any file changes.

## Agents (Optional)

Laravel validation agents enhance this skill with specialist reviews:

- `laravel-architect`
- `laravel-security-reviewer`
- `pest-tdd-expert`
- `pest-browser-testing`

If these agents are available, use them in Step 6. If they are not available, skip Step 6 and note in the plan that validation was skipped. Do not block the workflow — continue to produce and save the plan without agent feedback.

## Goals

1. Understand what the user wants to build and why.
2. Explain choices in simple terms for beginners.
3. Validate the design with Laravel specialist agents.
4. Produce a high-level plan doc in a user-approved location.
5. Handoff to `writing-plans` for detailed implementation planning.

## Workflow

1. Check the current git branch. If on `main` (or the default branch), create and switch to a feature branch named `feature/<short-topic>` before making any changes. Ask the user to confirm the branch name.
2. Explore project context (files, existing patterns, constraints).
3. Clarify the idea with short, plain-English questions.
4. Propose 2-3 implementation approaches with trade-offs.
5. Recommend one approach and explain why it is best.
6. Validate the design using these Laravel subagents:
   - `laravel-architect` for framework design quality
   - `laravel-security-reviewer` for security risks
   - `pest-tdd-expert` for test strategy
   - `pest-browser-testing` for browser/E2E coverage when UI is involved
7. Merge agent feedback, then present the refined design.
8. Ask where to save the plan before writing any plan file.
9. Write the approved high-level plan doc.
10. Invoke `interview` to conduct a technical interview on the plan. This stress-tests the design by probing edge cases, failure modes, and hidden assumptions before implementation begins.
11. Invoke `writing-plans` to generate the detailed implementation plan.

## Ask This Before Saving

Use this exact question:

"Where should I save your plan file? Press Enter for `docs/plans/`, or share another folder/path."

If the user presses Enter, default to `docs/plans/`.

Suggested filename format:

`YYYY-MM-DD-<topic>-plan.md`

## Output Structure

Always respond using these sections:

- `Idea in simple terms`
- `Options and trade-offs`
- `Recommended approach`
- `Workflow and why`
- `Validation feedback`
- `Plan file path`
- `Next step`

## Beginner Explanation Standard

- Use simple language, short sentences, and concrete examples.
- Avoid jargon when possible. If needed, define it in one line.
- Explain both what to do and why it matters.
- Keep each section concise and practical.

## Validation Standard

Before finalizing the plan:

1. Run design checks with relevant Laravel subagents.
2. Capture only actionable findings.
3. Resolve conflicts by prioritizing:
   1) security,
   2) correctness,
   3) maintainability,
   4) speed of delivery.
4. Re-explain updates in beginner terms.

## Completion Criteria

The brainstorming session is complete only when:

1. The user approves the recommended approach.
2. The plan has been validated with Laravel agents (if available) or the skip is noted.
3. The plan is saved to a user-approved location.
4. The `interview` skill has been invoked to stress-test the plan.
5. The `writing-plans` skill has been invoked.
6. No commit has been created.

## Mandatory Handoff

After saving the high-level plan:

1. Invoke `interview` to conduct a technical interview on the saved plan.
   - Announce: "I am now using `interview` to stress-test the plan before implementation."
   - The interview will probe edge cases, failure modes, and hidden assumptions.
   - After the interview, the spec will be updated with decisions and discoveries.
2. Then invoke `writing-plans` to generate the detailed implementation plan.
   - Announce: "I am now using `writing-plans` to produce the step-by-step implementation plan."
   - Pass the approved scope, constraints, validation feedback, and interview findings.
   - Keep the beginner-first tone.
   - Do not start coding.
