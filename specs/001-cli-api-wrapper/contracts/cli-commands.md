# CLI Command Contracts: CLI API Wrapper

## Global Behavior

All remote management commands use an active CLI profile unless `--profile` or `--server` overrides it.

Common options:
- `--profile <name>` selects a local connection profile.
- `--server <url>` overrides the server URL for one invocation.
- `--output table|json` selects human-readable or structured output.
- `--force` bypasses confirmation for destructive commands.

Common outcomes:
- Exit code `0` means the command completed successfully.
- Non-zero exit code means the command failed or was cancelled.
- Protected commands require authorization and must send an authorization token.
- Sensitive values are redacted in human-readable output by default.

## User And Session Commands

| Command | Method | Path | Purpose |
|---------|--------|------|---------|
| `octopus remote login` | POST | `/api/v1/user/login` | Authenticate and store local CLI session |
| `octopus remote logout` | Local | Local session file | Remove stored CLI session |
| `octopus remote status` | GET | `/api/v1/user/status` | Verify current session |
| `octopus remote user change-password` | POST | `/api/v1/user/change-password` | Change management password |
| `octopus remote user change-username` | POST | `/api/v1/user/change-username` | Change management username |

## Channel Commands

| Command | Method | Path | Purpose |
|---------|--------|------|---------|
| `octopus remote channel list` | GET | `/api/v1/channel/list` | List provider channels |
| `octopus remote channel create` | POST | `/api/v1/channel/create` | Create a provider channel |
| `octopus remote channel update` | POST | `/api/v1/channel/update` | Update a provider channel |
| `octopus remote channel enable` | POST | `/api/v1/channel/enable` | Enable or disable a channel |
| `octopus remote channel delete` | DELETE | `/api/v1/channel/delete/:id` | Delete a channel |
| `octopus remote channel fetch-model` | POST | `/api/v1/channel/fetch-model` | Fetch models for supplied channel connection data |
| `octopus remote channel sync` | POST | `/api/v1/channel/sync` | Sync channel model data |
| `octopus remote channel last-sync-time` | GET | `/api/v1/channel/last-sync-time` | Show last channel sync time |

## Group Commands

| Command | Method | Path | Purpose |
|---------|--------|------|---------|
| `octopus remote group list` | GET | `/api/v1/group/list` | List model groups |
| `octopus remote group create` | POST | `/api/v1/group/create` | Create a group |
| `octopus remote group update` | POST | `/api/v1/group/update` | Update a group and its items |
| `octopus remote group delete` | DELETE | `/api/v1/group/delete/:id` | Delete a group |

## API Key Commands

| Command | Method | Path | Purpose |
|---------|--------|------|---------|
| `octopus remote apikey list` | GET | `/api/v1/apikey/list` | List relay API keys |
| `octopus remote apikey create` | POST | `/api/v1/apikey/create` | Create a relay API key |
| `octopus remote apikey update` | POST | `/api/v1/apikey/update` | Update a relay API key |
| `octopus remote apikey delete` | DELETE | `/api/v1/apikey/delete/:id` | Delete a relay API key |
| `octopus remote apikey stats` | GET | `/api/v1/apikey/stats` | Show current API-key-authenticated stats when supported |

## Model And Price Commands

| Command | Method | Path | Purpose |
|---------|--------|------|---------|
| `octopus remote model list` | GET | `/api/v1/model/list` | List model prices |
| `octopus remote model channel` | GET | `/api/v1/model/channel` | List model-channel associations |
| `octopus remote model create` | POST | `/api/v1/model/create` | Create a model price entry |
| `octopus remote model update` | POST | `/api/v1/model/update` | Update a model price entry |
| `octopus remote model delete` | POST | `/api/v1/model/delete` | Delete a model price entry |
| `octopus remote model update-price` | POST | `/api/v1/model/update-price` | Sync model pricing |
| `octopus remote model last-update-time` | GET | `/api/v1/model/last-update-time` | Show last price update time |

## Settings, Stats, Logs, And Backup Commands

| Command | Method | Path | Purpose |
|---------|--------|------|---------|
| `octopus remote setting list` | GET | `/api/v1/setting/list` | List system settings |
| `octopus remote setting set` | POST | `/api/v1/setting/set` | Set a system setting |
| `octopus remote setting reset-circuit-breaker` | POST | `/api/v1/setting/circuit-breaker/reset` | Reset circuit breaker state |
| `octopus remote backup export` | GET | `/api/v1/setting/export` | Export database backup |
| `octopus remote backup import` | POST | `/api/v1/setting/import` | Import database backup |
| `octopus remote stats today` | GET | `/api/v1/stats/today` | Show today's stats |
| `octopus remote stats daily` | GET | `/api/v1/stats/daily` | Show daily stats |
| `octopus remote stats hourly` | GET | `/api/v1/stats/hourly` | Show hourly stats |
| `octopus remote stats total` | GET | `/api/v1/stats/total` | Show total stats |
| `octopus remote stats apikey` | GET | `/api/v1/stats/apikey` | Show API key stats |
| `octopus remote stats model` | GET | `/api/v1/stats/model` | Show model stats |
| `octopus remote log list` | GET | `/api/v1/log/list` | List relay logs |
| `octopus remote log clear` | DELETE | `/api/v1/log/clear` | Clear relay logs |

## Contract Test Expectations

- Each command validates required inputs before making an HTTP request.
- Each command sends the expected method, path, authorization header, query parameters, and payload shape.
- Read-only commands do not prompt for confirmation.
- Mutating commands report the changed resource or a success message.
- Destructive commands prompt unless `--force` is present.
- JSON output is valid JSON for success and failure cases.
- Human-readable output redacts API keys, provider keys, authorization tokens, and request/response bodies that may contain secrets unless explicitly requested.
