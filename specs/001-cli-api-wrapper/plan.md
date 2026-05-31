# Implementation Plan: CLI API Wrapper

**Branch**: `dev` | **Date**: 2026-05-29 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `/specs/001-cli-api-wrapper/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Add external CLI operations for Octopus management workflows by extending the existing Cobra-based binary with remote API client commands. The CLI will authenticate against a running Octopus server, store local session configuration, call the same management endpoints used by the web UI, and return both human-readable and script-friendly output.

## Technical Context

**Language/Version**: Go 1.24.4 for the server/CLI binary; TypeScript 5.9.3 for existing web API references.

**Primary Dependencies**: Cobra for commands, Viper for configuration, Gin for existing HTTP API, standard Go HTTP client for CLI-to-server calls.

**Storage**: Existing SQLite/MySQL/PostgreSQL server storage remains unchanged; CLI stores only local client configuration and authentication token metadata.

**Testing**: Go unit tests, integration tests using HTTP test servers, and contract tests for CLI command-to-endpoint behavior.

**Target Platform**: Cross-platform terminal use wherever the Octopus Go binary is distributed, with primary validation on Linux.

**Project Type**: Existing Go web service with embedded web app, extended with external API client CLI commands.

**Performance Goals**: CLI help should render within 1 second; typical list/detail commands should complete within 3 seconds on a healthy local server; CLI overhead should be negligible compared with server response time.

**Constraints**: Must not duplicate server-side business logic; must not log credentials or API keys; must preserve web permission checks; destructive commands require explicit confirmation or a force flag.

**Scale/Scope**: Covers basic web management workflows: login/status, channels, groups, API keys, models/prices, settings, stats, logs, backup export/import, circuit breaker reset, and sync/update actions.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Functionality Completeness**: PASS. Plan covers each basic web management area exposed by existing `/api/v1/*` endpoints and explicitly requires unsupported operations to be visible.

**Test Coverage Mandate**: PASS. Unit, integration, and contract tests are planned for command parsing, API client behavior, and command-to-endpoint contracts.

**Modular Design & Interface Contracts**: PASS. CLI command layer, API client layer, output formatting, and local session/config handling will be separately testable.

**Observability & Monitoring**: PASS. CLI failures will provide clear user-facing errors while server-side structured logging remains the source for request traces.

**Simplicity First**: PASS. Reuses existing Cobra binary and existing HTTP endpoints instead of adding a second CLI application or duplicating business logic.

**Security & Data Protection**: PASS. Tokens and secrets must not be printed by default, stored only in local config, redacted in errors, and revocable by logout or password change.

## Project Structure

### Documentation (this feature)

```text
specs/001-cli-api-wrapper/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
cmd/
├── root.go               # Existing Cobra root command
├── start.go              # Existing server command
├── version.go            # Existing version command
└── remote/               # New CLI command group for remote API operations

internal/
├── cli/                  # New API client, output, config, and session helpers
├── server/handlers/      # Existing management API endpoints reused by CLI
├── model/                # Existing request/response models reused where suitable
└── op/                   # Existing business operations remain server-side only

web/src/api/endpoints/    # Existing web endpoint references used for coverage mapping

tests are colocated as Go *_test.go files under cmd/ and internal/ packages.
```

**Structure Decision**: Keep a single Go binary with additional Cobra subcommands. Place reusable CLI client logic under `internal/cli` and thin command wiring under `cmd/remote` to keep API access testable without invoking the full process.

## Complexity Tracking

No constitution violations identified.

## Post-Design Constitution Check

**Functionality Completeness**: PASS. Design artifacts define command coverage, API contracts, configuration, and verification paths for all scoped workflows.

**Test Coverage Mandate**: PASS. Contract artifact identifies command-to-endpoint assertions; quickstart includes verification scenarios; implementation tasks must include unit, integration, and contract tests.

**Modular Design & Interface Contracts**: PASS. Data model and contracts separate CLI session, command metadata, operation results, and application resources.

**Observability & Monitoring**: PASS. Error handling and exit behavior are specified; server logs remain authoritative for API request diagnostics.

**Simplicity First**: PASS. No new service or separate distribution system is introduced for v1.

**Security & Data Protection**: PASS. Contracts and data model require token redaction, explicit destructive confirmations, and existing permission enforcement.
