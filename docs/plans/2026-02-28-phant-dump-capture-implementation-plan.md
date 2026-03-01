# Phant Laravel Dump Capture — Implementation Plan

Date: 2026-02-28
Project: `phant` (Wails v2 + React + Go)
Status: Approved for planning (no coding yet)

## Goal in simple terms

Build a Linux desktop experience similar to Laravel Herd, starting with dump capture.  
Phant should catch and show `dump()` and `dd()` from Laravel apps **without adding any app-side interface or package**.

Primary architecture:
- Global PHP hook as the default mode
- Symfony Dump Server ingestion as optional fallback mode

MVP coverage:
- HTTP requests (PHP-FPM/Nginx)
- CLI commands (`artisan`, `php ...`)
- Queue workers (`queue:work`, Horizon)
- Scheduled tasks (cron)

## Plan strategy

1. Build a local collector daemon in Go (`phantd`) that accepts structured dump events.
2. Add a one-time onboarding flow in Phant that safely configures global PHP hook files with backup + rollback.
3. Add metadata-aware capture bootstrap in PHP (machine-level prepend script) to forward dumps to `phantd`.
4. Expose a Wails backend stream for live updates to the React frontend.
5. Start with a minimal developer UI: raw JSON event log + health/errors panel.
6. Add resilience: ring buffer, dropped-count warnings, diagnostics mode, and auto-revert on config failure.

## Step-by-step tasks

### Phase 1 — Collector and protocol foundation

### Task 1: Define event schema and transport contract
- Purpose: Make all components speak one stable format.
- Files to create/modify:
  - `docs/specs/dump-event-schema.md` (new)
  - `app.go` (new methods later consume schema)
- Implementation steps:
  1. Define event JSON fields: `id`, `timestamp`, `sourceType`, `projectRoot`, `phpSapi`, `requestId`, `http` (optional), `command` (optional), `isDd`, `payloadFormat`, `payload`, `trace`, `host`.
  2. Define transport framing (newline-delimited JSON over Unix socket).
  3. Define versioning field (`schemaVersion`) and compatibility behavior.
- Verification:
  - Run schema examples through a JSON validator.
  - Confirm required fields cover HTTP + CLI + worker use cases.
- Rollback/fix hint:
  - If schema is too large/complex, trim non-essential optional fields and keep only capture-critical data.

### Task 2: Implement local collector daemon core (`phantd`)
- Purpose: Receive dumps from PHP and provide live stream to desktop app.
- Files to create/modify:
  - `internal/collector/server.go` (new)
  - `internal/collector/buffer.go` (new)
  - `internal/collector/types.go` (new)
  - `main.go` (wire collector lifecycle)
- Implementation steps:
  1. Create Unix domain socket listener (e.g. `${XDG_RUNTIME_DIR}/phant/collector.sock`).
  2. Parse NDJSON events and validate required fields.
  3. Store recent events in bounded ring buffer.
  4. Track dropped event count when buffer limit is exceeded.
  5. Publish events to in-process subscribers (Wails bridge).
- Verification:
  - Send sample events with `socat`/test client; confirm parse + fanout.
  - Flood test to ensure old events are dropped and dropped counter increments.
- Rollback/fix hint:
  - If Unix socket permissions fail, fall back to user-home socket path with strict file mode.

### Phase 2 — Safe global hook onboarding (best DX)

### Task 3: Add PHP environment discovery and diagnostics
- Purpose: Detect supported PHP runtimes before writing config.
- Files to create/modify:
  - `internal/setup/php_discovery.go` (new)
  - `app.go` (expose discovery method to frontend)
- Implementation steps:
  1. Detect PHP binaries and versions (`php -v`, common paths).
  2. Detect loaded ini files (`php --ini`) and FPM pool config locations.
  3. Detect service manager commands available (`systemctl`, `service`).
  4. Generate a diagnostic report with pass/fail per check.
- Verification:
  - Run discovery on machine with at least one PHP version.
  - Confirm report includes clear actionable failures.
- Rollback/fix hint:
  - If some paths cannot be auto-detected, mark them as manual-input required, not hard failure.

### Task 4: Implement config patcher with backup and rollback
- Purpose: Apply hook changes safely and reversibly.
- Files to create/modify:
  - `internal/setup/config_patcher.go` (new)
  - `internal/setup/rollback.go` (new)
  - `internal/setup/history.go` (new)
- Implementation steps:
  1. Create timestamped backups before any write.
  2. Add/update `auto_prepend_file` to point to Phant bootstrap script.
  3. Apply changes separately for CLI and FPM contexts.
  4. Reload/restart relevant services.
  5. On failure, restore backups automatically and emit detailed error.
- Verification:
  - Simulate syntax error in target config and validate auto-rollback.
  - Confirm post-check: `php -i` reflects prepend path.
- Rollback/fix hint:
  - If patching parser fails, switch to line-based insertion with idempotent markers.

### Task 5: Build onboarding UX for transparent changes
- Purpose: Keep trust and reduce setup fear/noise.
- Files to create/modify:
  - `frontend/src/App.tsx` (replace template with setup + log UI)
  - `frontend/src/components/SetupSummary.tsx` (new)
  - `frontend/src/components/HealthPanel.tsx` (new)
- Implementation steps:
  1. Show “before vs after” file preview for each config change.
  2. Show exact commands Phant will run (reload services, checks).
  3. Require explicit user confirmation before write.
  4. Show rollback status and docs links when failure occurs.
- Verification:
  - Dry-run mode renders all planned edits without writing files.
  - Failed setup produces clear error + “reverted successfully” message.
- Rollback/fix hint:
  - If visual diff is complex, render plain text summary first, then add unified diff later.

### Phase 3 — PHP capture bootstrap and fallback transport

### Task 6: Implement machine-level PHP prepend bootstrap script
- Purpose: Capture `dump()` and `dd()` with no app integration.
- Files to create/modify:
  - `assets/php/phant_prepend.php` (new)
  - `internal/setup/bootstrap_installer.go` (new)
- Implementation steps:
  1. In prepend file, register Symfony VarDumper handler early.
  2. Normalize payload into schema format.
  3. Detect context (HTTP vs CLI vs queue/cron heuristics).
  4. Send event non-blocking to collector socket with short timeout.
  5. Ensure `dd()` event is sent before process terminates.
- Verification:
  - Test simple Laravel route with `dump()` and `dd()`.
  - Test `php artisan tinker`/custom command with `dump()`.
- Rollback/fix hint:
  - If handler collides with existing dumper behavior, chain to previous handler after sending event.

### Task 7: Add Symfony Dump Server compatibility mode
- Purpose: Provide fallback where hook mode cannot be applied.
- Files to create/modify:
  - `internal/collector/dumpserver_adapter.go` (new)
  - `docs/specs/dump-server-mode.md` (new)
- Implementation steps:
  1. Add optional listener/adapter for Dump Server payloads.
  2. Transform incoming messages into common event schema.
  3. Flag source mode in event metadata.
- Verification:
  - Trigger dumps with dump-server enabled env and confirm ingestion.
- Rollback/fix hint:
  - If payload fidelity is reduced, mark missing fields as null and annotate source limitations in UI.

### Phase 4 — Wails bridge and frontend MVP

### Task 8: Expose live event stream from Go to React
- Purpose: Feed realtime dump events to UI.
- Files to create/modify:
  - `app.go` (stream subscription / event APIs)
  - `frontend/wailsjs/go/main/App.d.ts` (generated)
  - `frontend/wailsjs/go/main/App.js` (generated)
- Implementation steps:
  1. Add backend methods: `GetRecentEvents`, `GetHealth`, `StartSetup`, `RollbackLastSetup`.
  2. Emit Wails runtime events when new dump arrives.
  3. Include dropped-count and transport-mode state in health responses.
- Verification:
  - Frontend receives events in dev mode and after app restart.
- Rollback/fix hint:
  - If event flooding causes UI lag, batch event emissions at short interval.

### Task 9: Build raw JSON event log UI (MVP)
- Purpose: Ship fast with maximum debugging value.
- Files to create/modify:
  - `frontend/src/App.tsx`
  - `frontend/src/components/EventLog.tsx` (new)
  - `frontend/src/components/ErrorBanner.tsx` (new)
- Implementation steps:
  1. Render append-only JSON log with timestamp and source tags.
  2. Show health state, dropped events, and current mode (Hook / Dump Server / Diagnostics).
  3. Add clear-log and delete-stored-events actions.
- Verification:
  - High-frequency dumps stay responsive with virtualized or capped render list.
  - Error panel shows actionable diagnostics and docs links.
- Rollback/fix hint:
  - If render cost is high, reduce visible entries and keep full data in backend store.

### Phase 5 — Persistence, safety, and documentation

### Task 10: Add local persistence with user-controlled deletion
- Purpose: Support long sessions while respecting privacy control.
- Files to create/modify:
  - `internal/storage/events_store.go` (new)
  - `internal/storage/migrations.go` (new)
  - `app.go` (delete/export methods)
- Implementation steps:
  1. Persist events in local app data directory.
  2. Add APIs to delete all events and optionally delete by project/session.
  3. Add startup load of recent events with cap.
- Verification:
  - Restart app and confirm previous events are restored.
  - Delete action removes entries from storage and UI immediately.
- Rollback/fix hint:
  - If DB introduces complexity, start with newline JSON file rotation and migrate later.

### Task 11: Document failure modes and recovery paths
- Purpose: Make operations predictable for users.
- Files to create/modify:
  - `README.md`
  - `docs/setup/linux-php-hook.md` (new)
  - `docs/troubleshooting/dump-capture.md` (new)
- Implementation steps:
  1. Document onboarding flow, required permissions, and what gets changed.
  2. Document rollback process and diagnostics mode behavior.
  3. Document Dump Server fallback setup.
- Verification:
  - Follow docs on a clean environment and complete setup end-to-end.
- Rollback/fix hint:
  - If docs are too long, keep quickstart + troubleshooting table first.

## Workflow and why

- Start with schema + collector first so transport is stable before touching system config.
- Add discovery and safe patching before PHP hook install to avoid risky writes.
- Build prepend capture only after rollback mechanics exist.
- Add frontend after backend stream is stable to reduce UI thrash.
- Add persistence and documentation last because they depend on final event semantics.

This order is safest for beginners because each phase has a visible checkpoint and easy rollback path.

## Verification checkpoints

1. Collector accepts and broadcasts test events.
2. Discovery reports complete PHP/FPM status.
3. Config patch applies and rolls back correctly in forced-failure test.
4. `dump()` and `dd()` appear in Phant for HTTP + CLI + queue + cron contexts.
5. Ring buffer drops oldest events and shows dropped count in UI.
6. Dump Server fallback mode ingests events when hook mode is unavailable.
7. Persisted events survive restart and can be deleted by the user.

## Interview decisions incorporated

- MVP includes HTTP, CLI, queue workers, and cron.
- Setup should be fully automatic with sudo prompt, but transparent about exact changes.
- Primary mode: global hook; secondary mode: Dump Server fallback.
- If setup fails: auto-rollback to prior configuration, clear error explanation, and docs guidance.
- Persistence default: keep events locally; user can delete stored dumps.
- Frontend MVP: raw JSON event log (developer mode) is acceptable first release.
- Performance guardrail: ring buffer + dropped-count warning.

## Open questions for implementation kickoff

1. Preferred local persistence format for v1: SQLite vs JSONL file rotation?
2. Maximum default in-memory buffer size (e.g. 2,000 / 10,000 events)?
3. Do we need per-project retention limits in v1 or only global delete?

## Plan file path

`docs/plans/2026-02-28-phant-dump-capture-implementation-plan.md`

## Review before coding

Please review this plan and confirm the three open questions above.  
After confirmation, implementation can start in small phases with validation after each checkpoint.
