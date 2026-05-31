# Data Model: CLI API Wrapper

## CLI Profile

Represents a named connection from the CLI to an Octopus server.

**Fields**:
- `name`: Stable local profile name.
- `base_url`: Server address used by remote commands.
- `default_output`: Preferred output mode, either human-readable or structured.
- `active`: Whether this profile is the default for commands.

**Validation Rules**:
- `name` must be non-empty and unique locally.
- `base_url` must be a valid HTTP or HTTPS URL.
- Only one profile may be active at a time.

**Relationships**:
- Has zero or one active CLI User Session.

## CLI User Session

Represents the authenticated state for a CLI profile.

**Fields**:
- `profile_name`: Associated CLI profile.
- `username`: Authenticated management user name when known.
- `token`: Authentication token stored locally and redacted in output.
- `expire_at`: Token expiration time when provided by the server.
- `last_verified_at`: Last successful status check time.

**Validation Rules**:
- `token` must never be printed by default.
- Expired sessions must be treated as unauthenticated.
- Logout must remove the locally stored token for the selected profile.

**State Transitions**:
- `unauthenticated` to `authenticated`: Successful login.
- `authenticated` to `expired`: Token expiration or failed status check caused by authorization.
- `authenticated` to `revoked`: Logout or password/username change invalidates access.

## Command

Represents a CLI operation exposed to users.

**Fields**:
- `name`: Command path, such as `channel list`.
- `category`: Functional area such as user, channel, group, apikey, model, setting, stats, log, backup, or sync.
- `required_inputs`: Inputs that must be present before execution.
- `optional_inputs`: Inputs that alter filtering, formatting, or behavior.
- `safety_level`: Read-only, mutating, or destructive.
- `supports_structured_output`: Whether command results can be emitted in structured form.

**Validation Rules**:
- Required inputs must be validated before sending a request.
- Destructive commands require confirmation unless an explicit force option is provided.
- Unsupported web functions must be discoverable as unsupported or omitted with documented scope.

## Operation Result

Represents the output and exit behavior of a CLI command.

**Fields**:
- `status`: Success, failure, or partial completion.
- `exit_code`: Process outcome for automation.
- `message`: Human-readable summary.
- `data`: Returned records or details when applicable.
- `next_step`: Corrective action for failures when known.

**Validation Rules**:
- Success and failure must be distinguishable by exit code.
- Failure messages must not include credentials or sensitive request payloads.
- Structured output must remain parseable for both empty and non-empty results.

## Application Resource

Represents existing Octopus resources manipulated through web and CLI workflows.

**Resource Types**:
- `Channel`: Provider connection, base URLs, model mapping, keys, headers, proxy settings, and status.
- `Group`: Aggregated model name, routing mode, match rules, timeout settings, and channel items.
- `API Key`: Relay key, enablement state, supported models, expiration, budget limit, and statistics.
- `Model Price`: Model name and pricing values.
- `Setting`: System setting key and value.
- `Relay Log`: Recorded request, response, attempts, timing, cost, and error details.
- `Stats`: Daily, hourly, total, model, channel, and API-key usage metrics.
- `Backup`: Exported or imported database data according to selected options.

**Validation Rules**:
- CLI operations must preserve server-side validation and state transition rules.
- Secrets inside resources must be redacted unless the command explicitly supports revealing them.
- Pagination, filtering, and large output handling must prevent unreadable terminal dumps by default.
