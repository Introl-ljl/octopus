# Research: CLI API Wrapper

## Decision: Implement CLI as additional Cobra commands in the existing Go binary

**Rationale**: The project already uses Cobra for `start` and `version`, so adding management commands to the same binary avoids another distribution artifact and keeps installation simple for existing users.

**Alternatives considered**: A separate Node/TypeScript CLI would reuse web endpoint type definitions but would require a second runtime and package pipeline. A shell-script wrapper would be harder to test and unsafe for structured operations.

## Decision: Treat the CLI as an external API client

**Rationale**: The feature explicitly asks for a plug-in style CLI that accesses APIs. Calling existing management endpoints preserves server-side validation, permissions, logs, and business rules without duplicating behavior in the CLI.

**Alternatives considered**: Direct database access from the CLI was rejected because it bypasses permission checks and business rules. Calling internal `op` packages directly was rejected because it would only work locally and would duplicate server execution paths.

## Decision: Support both interactive and non-interactive modes

**Rationale**: The spec requires terminal workflows and automation. Interactive prompts improve safety for destructive operations, while flags and structured output enable scripts and CI jobs.

**Alternatives considered**: Interactive-only commands would block automation. Non-interactive-only commands would make destructive actions too easy to run accidentally.

## Decision: Store CLI connection profile locally, not in server storage

**Rationale**: Server-side storage already owns users, permissions, and resources. The CLI only needs base URL, active profile, output preference, and token metadata to connect to a server.

**Alternatives considered**: Environment-only configuration is simple but inconvenient for routine use. Server-side CLI profile storage would add unnecessary scope and would not help before login.

## Decision: Use existing admin JWT login for management commands in v1

**Rationale**: The web management panel already authenticates with `/api/v1/user/login` and protected management endpoints use the same authorization model. This satisfies parity with web operations while avoiding a new credential system.

**Alternatives considered**: API-key authentication is already used for relay/client access but does not represent full management-panel permissions. A new device-code flow would improve UX but adds server-side scope not required for v1.

## Decision: Redact sensitive values by default

**Rationale**: The project manages LLM provider keys and Octopus API keys. CLI output and error handling must prevent accidental credential disclosure in terminal scrollback, logs, and automation output.

**Alternatives considered**: Always printing full records would be simpler but violates the constitution's security guidance. A per-command reveal flag keeps advanced workflows possible while making safe behavior the default.

## Decision: Contract-test command mapping against management endpoints

**Rationale**: CLI correctness depends on preserving parity with web operations. Contract tests can verify that each command sends the expected method, path, authorization, and payload shape without requiring a full production server.

**Alternatives considered**: Only unit-testing command parsing would miss API integration regressions. Only end-to-end tests would be slower and less precise for command coverage.
