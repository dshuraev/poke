# Poke

Poke is a small command-execution server.

It accepts remote command requests, validates them, and executes only commands
that are explicitly whitelisted in configuration.

## Who This Is For

- Operators who need a small remote command runner with explicit allowlists.
- Developers integrating infrastructure commands into automation workflows.

## Quickstart

1. Copy `docs/configuration/config.example.yaml` and adjust values.
2. Start the server:

   ```sh
   go run ./cmd/server -c /path/to/poke.yml
   ```

3. Send a request:

```sh
curl -X PUT http://127.0.0.1:8008/ \
  -H "Content-Type: application/json" \
  -H "X-Poke-Auth-Method: api_token" \
  -H "X-Poke-API-Token: my-secret-token" \
  -d '{"command_id":"hello"}'
```

The server returns `202 Accepted` when a command request is accepted for
execution.

## Documentation

Start here for full documentation maps:

- User documentation: `docs/user/README.md`
- Developer documentation: `docs/developer/README.md`
- Full docs index: `docs/index.md`

## Current Scope

Implemented:

- HTTP listener (`PUT /`) for `{"command_id":"..."}` requests.
- API token auth per listener.
- Optional TLS for HTTP listener.
- Binary command executor with command allowlist.
- Per-command timeout and environment strategy.
- Structured logging (stdout and journald sink options).

In progress (see `docs/roadmap.md`):

- Async execution and flags.
- Executor responses.
- Telemetry.
