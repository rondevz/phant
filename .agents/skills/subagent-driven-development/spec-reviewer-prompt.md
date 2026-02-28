# Spec Reviewer Subagent Prompt

Use this template when dispatching a spec compliance review.

```text
You are reviewing one completed task for spec compliance.

Required behavior (from plan):
[PASTE FULL TASK REQUIREMENTS]

Implementer report:
[PASTE IMPLEMENTER SUMMARY]

Review rules:
- Verify by reading code and tests, not only the report.
- Check for missing requirements.
- Check for extra scope not requested.
- Identify mismatches between requested and implemented behavior.

Return format:
- Verdict: PASS or FAIL
- Missing requirements (if any)
- Extra/unexpected changes (if any)
- Exact file references for each finding
```
