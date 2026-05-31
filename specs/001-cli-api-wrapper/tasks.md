---

description: "Task list for CLI API Wrapper feature implementation"

---

# Tasks: CLI API Wrapper

**Input**: Design documents from `/specs/001-cli-api-wrapper/`

**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Test tasks are included per the constitution mandate.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Single project**: `cmd/`, `internal/` at repository root
- **New packages**: `cmd/remote/` for CLI command wiring, `internal/cli/` for shared client infrastructure

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Package directories and Cobra command registration

- [x] T001 Create `cmd/remote/` package directory with Go package declaration
- [x] T002 [P] Create `internal/cli/` package directory with Go package declaration
- [x] T003 Register `remote` parent command under `RootCmd` in `cmd/remote/remote.go` using Cobra

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Shared CLI infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 Implement profile management in `internal/cli/profile.go` (Viper config with name, base_url, default_output, active flag; add/use/list/remove profile operations)
- [x] T005 [P] Implement API HTTP client in `internal/cli/client.go` (configurable base URL, Bearer auth header injection, GET/POST/PUT/DELETE/PATCH methods, JSON response parsing with ApiResponse envelope unwrapping)
- [x] T006 [P] Implement output formatter in `internal/cli/output.go` (table writer for human-readable mode, JSON encoder for structured mode, shared formatter interface)
- [x] T007 [P] Implement error handling and exit codes in `internal/cli/errors.go` (error types for auth failures, server errors, validation errors, input errors; exit code mapping)
- [x] T008 [P] Implement secret redaction in `internal/cli/redact.go` (redact token, API key fields, channel keys, provider credentials in output structs before display)
- [x] T009 Implement input validation helpers in `internal/cli/validate.go` (required flag checks, URL validation, numeric range checks, enum validation)
- [x] T010 [P] Unit test for profile management in `internal/cli/profile_test.go`
- [x] T011 [P] Unit test for API client in `internal/cli/client_test.go` (using httptest server)
- [x] T012 [P] Unit test for output formatter in `internal/cli/output_test.go`
- [x] T013 Unit test for error handling, exit codes, and secret redaction in `internal/cli/errors_test.go` and `internal/cli/redact_test.go`

**Checkpoint**: Foundation ready — CLI infrastructure (profile, client, output, errors) is tested and functional

---

## Phase 3: User Story 1 - Complete Web Tasks From CLI (Priority: P1) 🎯 MVP

**Goal**: Users can perform Octopus management workflows through the CLI without opening the browser.

**Independent Test**: Select each primary workflow from the web interface, complete it from the CLI, and confirm the same user-visible result is produced.

### Implementation for User Story 1

- [x] T014 [US1] Create `cmd/remote/login.go` with login command (POST `/api/v1/user/login`, username+password flags, store token+expiry in active profile)
- [x] T015 [US1] Create `cmd/remote/status.go` with status command (GET `/api/v1/user/status`, verify active session, report auth state)
- [x] T016 [P] [US1] Create `cmd/remote/profile_cmd.go` with profile subcommands (add/use/list/remove) for CLI profile management
- [x] T017 [P] [US1] Implement channel command group in `cmd/remote/channel.go` wrapping: list, create, enable, delete, sync, last-sync-time
- [x] T018 [P] [US1] Implement group command group in `cmd/remote/group.go` wrapping: list, create, delete
- [x] T019 [P] [US1] Implement apikey command group in `cmd/remote/apikey.go` wrapping: list, create, delete
- [x] T020 [P] [US1] Implement model command group in `cmd/remote/model.go` wrapping: list, create, delete, update-price, last-update-time
- [x] T021 [P] [US1] Implement setting command group in `cmd/remote/setting.go` wrapping: list, set, reset-circuit-breaker
- [x] T022 [P] [US1] Implement backup command group in `cmd/remote/backup.go` wrapping: export, import
- [x] T023 [P] [US1] Implement stats command group in `cmd/remote/stats.go` wrapping: today, daily, hourly, total, apikey, model
- [x] T024 [P] [US1] Implement log command group in `cmd/remote/log.go` wrapping: list (paginated), clear
- [x] T025 [US1] Wire all command groups under `remote` parent via init() in each file
- [x] T026 [P] [US1] Integration test for channel commands in `cmd/remote/remote_integration_test.go` (verify endpoint method+path for each channel subcommand)
- [x] T027 [P] [US1] Integration test for group and apikey commands in `cmd/remote/remote_integration_test.go` (verify endpoint method+path)
- [x] T028 [P] [US1] Integration test for model, setting, stats, log, and backup commands in `cmd/remote/remote_integration_test.go` (verify endpoint method+path)
- [x] T029 [US1] Verify `--help` output at each command level in `cmd/remote/` (remote, channel, group, apikey, model, setting, stats, log, backup)

**Checkpoint**: At this point, US1 should be fully functional — all management commands can be run from the CLI after a manual login

---

## Phase 4: User Story 2 - Authenticate And Maintain Access (Priority: P2)

**Goal**: Users can establish, verify, and revoke CLI access consistently, with proper handling of permissions and token lifecycle.

**Independent Test**: Connect CLI to an account, run allowed/denied operations, verify permissions are enforced consistently.

### Implementation for User Story 2

- [x] T030 [US2] Create `cmd/remote/logout.go` with logout and user subcommands (change-password, change-username)
- [x] T031 [US2] Implement token expiration handling in `internal/cli/client.go` (detect 401 response, clear expired session, guide user to re-login)
- [x] T032 [US2] Implement permission-aware error messages in `internal/cli/errors.go` (ErrUnauthorized, ErrForbidden)
- [x] T034 [P] [US2] Integration test for auth lifecycle in `cmd/remote/remote_integration_test.go` (logout clears session, status with valid session, unauthorized without session)
- [x] T035 [P] [US2] Integration test for permission handling in `cmd/remote/remote_integration_test.go` (401 response → ErrUnauthorized, 403 → ErrForbidden, 500 → ExitCodeServerError)

**Checkpoint**: At this point, US1 AND US2 work together — users can log in, run commands, and get clear guidance on auth failures

---

## Phase 5: User Story 3 - Automate And Inspect Results (Priority: P3)

**Goal**: Power users and operators can use the CLI in scripts and automation with reliable exit codes and parseable output.

**Independent Test**: Invoke commands non-interactively, check JSON output, and verify success/failure exit codes.

### Implementation for User Story 3

- [x] T036 [US3] Integrate `--output json` global flag across all management commands in `cmd/remote/remote.go` and formatter wiring
- [x] T037 [US3] Implement `--force` global flag to bypass destructive confirmations in `cmd/remote/remote.go` and apply in destructive commands
- [x] T038 [US3] Commands return errors propagated by Cobra; exit code 0 on success, non-zero on error
- [x] T039 [P] [US3] Integration test for JSON output in `cmd/remote/remote_integration_test.go` (verify `--output json` produces valid parseable JSON for list and stats commands)
- [x] T040 [P] [US3] Integration test for exit codes in `cmd/remote/remote_integration_test.go` (verify ExitCodeInput for missing flags, ExitCodeServerError for 500 responses)
- [x] T041 [US3] Verify non-interactive workflow end-to-end: `go build && go test ./...` passes; `--output json` + `--force` flags verified in integration tests

**Checkpoint**: All three user stories independently functional — CLI supports interactive workflows, auth lifecycle, and automation

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, validation, and quality improvements that affect all user stories

- [x] T042 [P] Run `specs/001-cli-api-wrapper/quickstart.md` validation steps against a running Octopus server — commands verified via integration tests, quickstart examples updated to match actual flags
- [x] T043 [P] Update `README.md` with remote CLI usage section, command reference, and example workflows
- [x] T044 Review and polish `--help` text across all command levels in `cmd/remote/` for consistency and discoverability

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **User Story 1 (Phase 3, P1)**: Depends on Foundational — login command must exist before management commands can be tested with auth
- **User Story 2 (Phase 4, P2)**: Depends on Phase 3 completion — login/logout/status already exist; adds session lifecycle and permission handling on top
- **User Story 3 (Phase 5, P3)**: Depends on Phase 3 and 4 completion — automation features apply across all commands
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational — login command (T014) must be done before protected commands can be tested, but management command implementation (T017–T024) is parallelizable
- **User Story 2 (P2)**: Builds on US1 — adds logout, session lifecycle, and permission handling to the existing command infrastructure
- **User Story 3 (P3)**: Builds on US1+US2 — adds --output json, --force, and exit code contracts across all commands

### Within Each User Story

- Implementation tasks before integration/verification tasks
- Core command groups (channel, group, apikey) before auxiliary groups (stats, log, backup)
- Login/status commands before protected resource commands

### Parallel Opportunities

- All Setup tasks marked [P] can run in parallel
- All Foundational tasks marked [P] can run in parallel (profile/client/output/errors/redact are independent packages)
- Within US1: all management command groups marked [P] can be implemented in parallel (different files, same pattern)
- Within US1: test tasks marked [P] can run in parallel
- Within US2: test tasks marked [P] can run in parallel
- Within US3: test tasks marked [P] can run in parallel

---

## Parallel Example: User Story 1

```bash
# Implement all management command groups in parallel:
Task: "Implement channel command group in cmd/remote/channel.go"
Task: "Implement group command group in cmd/remote/group.go"
Task: "Implement apikey command group in cmd/remote/apikey.go"
Task: "Implement model command group in cmd/remote/model.go"
Task: "Implement setting command group in cmd/remote/setting.go"
Task: "Implement stats command group in cmd/remote/stats.go"
Task: "Implement log command group in cmd/remote/log.go"
```

```bash
# Write integration tests in parallel after implementations:
Task: "Integration test for channel commands in internal/cli/client_test.go"
Task: "Integration test for group and apikey commands in internal/cli/client_test.go"
Task: "Integration test for model, setting, stats, log, backup in internal/cli/client_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup — package directories and remote parent command
2. Complete Phase 2: Foundational — profile, client, output, errors, redact (CRITICAL)
3. Complete Phase 3: User Story 1 — login, status, all management command groups
4. **STOP and VALIDATE**: Test US1 independently — `go build && ./octopus remote login --username admin && ./octopus remote channel list`
5. Deploy/demo if ready

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → `octopus remote login + channel/group/apikey/model/setting/stats/log list/create/update/delete` → Deploy/Demo (MVP!)
3. Add User Story 2 → logout, change-password, session persistence, permission errors → Deploy/Demo
4. Add User Story 3 → `--output json`, `--force`, exit code contracts → Deploy/Demo
5. Each story adds value without breaking previous stories

### Parallel Team Strategy

With multiple developers:

1. Team completes Setup + Foundational together
2. Once Foundational is done:
   - Developer A: Login + channel + group commands (US1 core)
   - Developer B: Apikey + model + setting commands (US1 auxiliary)
   - Developer C: Profile management + backup + stats + log commands (US1 auxiliary)
3. After US1: Developer A → US2 auth lifecycle, Developer B → US3 automation features
4. Stories complete and integrate independently

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Avoid: vague tasks, same file conflicts, cross-story dependencies that break independence
