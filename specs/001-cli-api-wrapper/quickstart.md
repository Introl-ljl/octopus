# Quickstart: CLI API Wrapper

## Prerequisites

- Build or download the Octopus binary.
- Start an Octopus server with `octopus start`.
- Have management credentials for the server.

## Configure A Remote Profile

```bash
octopus remote profile add --name local --url http://127.0.0.1:8080
octopus remote profile use local
```

## Login And Check Status

```bash
octopus remote login --username admin
octopus remote status
```

Expected result: the status command reports an authorized session without printing the stored token.

## Discover Commands

```bash
octopus remote --help
octopus remote channel --help
octopus remote group --help
```

Expected result: users can find commands, required inputs, and examples without opening the web UI.

## Run Core Management Workflows

```bash
octopus remote channel list
octopus remote group list --output json
octopus remote apikey list
octopus remote model list
octopus remote setting list
octopus remote stats total
octopus remote log list --page-size 20
```

Expected result: commands return the same resource data visible in the web management panel, with secrets redacted by default.

## Run A Mutating Workflow

```bash
octopus remote model create --name example-model --input 1 --output 2
octopus remote model update --name example-model --input 1.5 --output 2.5
octopus remote model delete --name example-model --force
```

Expected result: create, update, and delete operations preserve the same validation behavior as the web UI and return non-zero exit codes on failure.

## Verify Destructive Safeguards

```bash
octopus remote log clear
octopus remote log clear --force
```

Expected result: the first command asks for confirmation; the second command runs non-interactively.

## Test Requirements

```bash
go test ./...
```

Required coverage before delivery:
- Unit tests for command validation, output formatting, config/session handling, and secret redaction.
- Integration tests for authenticated and unauthenticated HTTP client behavior.
- Contract tests for command-to-endpoint method, path, authorization, payload, and exit behavior.
