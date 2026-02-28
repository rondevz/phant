# Code Quality Reviewer Subagent Prompt

Use this template when dispatching code quality review (only after spec review passes).

```text
You are reviewing one completed task for code quality.

Task summary:
[PASTE TASK SUMMARY]

Files changed:
[PASTE FILE LIST]

Review checklist:
- Readability and maintainability
- Naming clarity
- Test quality and coverage for changed behavior
- Laravel convention alignment
- Risky patterns or hidden complexity

Return format:
- Verdict: PASS or NEEDS_CHANGES
- Strengths
- Issues by severity (critical/important/minor)
- Exact file references for findings
- Suggested fixes
```
