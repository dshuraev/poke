# Troubleshooting

This page lists common runtime and configuration failures.

## Server Fails to Start

- Config path not found:
  - Pass explicit `-c /path/to/poke.yml`.
  - Confirm one of default config paths exists.
- Listener bind failure:
  - Check if host/port is already in use.
- TLS errors:
  - Ensure both `cert_file` and `key_file` are configured and readable.

## Request Rejected

- `405 Method Not Allowed`:
  - Use `PUT /`.
- `400 Bad Request`:
  - Send valid JSON and include non-empty `command_id`.
- `401 Unauthorized`:
  - Set `X-Poke-Auth-Method: api_token`.
  - Send valid `X-Poke-API-Token`.
- `503 Service Unavailable`:
  - Server context may be shutting down.

## Command Does Not Run

- `command_id` not defined in `commands` block.
- Command binary not available in `PATH`.
- Command timeout too low for expected runtime.

## How to Debug Quickly

1. Increase log level to `debug` in `logging.level`.
2. Verify request payload and auth headers with `curl -v`.
3. Validate command exists in config and is executable.

## See Also

- `docs/user/getting-started.md`
- `docs/user/configuration.md`
- `docs/configuration/logging.md`
