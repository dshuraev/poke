# Architecture

Poke is a small request-to-command runtime with explicit command allowlisting.

## Runtime Flow

1. `cmd/server/main.go` resolves config path and parses YAML config.
2. `internal/server.Start(...)` creates request channel and starts listeners.
3. Listeners enqueue `request.CommandRequest{CommandID: ...}`.
4. `dispatch.SyncDispatcher` consumes requests and resolves command config.
5. Dispatcher calls configured executor (`bin` today).
6. `executor.ExecuteBinary` runs OS command with timeout/env strategy.
7. Structured logs report request, execution start, and execution outcome.

## Core Components

- Listener (`internal/server/listener`)
  - HTTP listener supports `PUT /` with JSON `{ "command_id": "..." }`.
  - Validates auth before enqueue.
- Dispatch (`internal/server/dispatch`)
  - Synchronous processing loop.
  - One request handled at a time.
- Executor (`internal/server/executor`)
  - Binary executor (`os/exec`) with command timeout and env merging.
- Auth (`internal/server/auth`)
  - `api_token` validator with `token`/`env`/`file` sources.
- Logging (`internal/server/logging`)
  - Text/JSON output.
  - stdout or journald sink.

## Design Constraints

- Commands must be pre-registered in config.
- Listener auth is required for HTTP listener config.
- Request response currently indicates acceptance (`202`) only.
- Execution output is logged/executor-internal, not returned to API clients.

## Why This Shape

- Keeps MVP surface small and auditable.
- Separates concerns cleanly (listener, dispatch, executor, auth).
- Supports future extension points (async dispatch, richer responses, telemetry).

## Current Limitations

- No async queue/job status API yet.
- No executor response payload contract for clients.
- Telemetry not implemented yet.

## See Also

- `docs/developer/onboarding.md`
- `docs/developer/testing.md`
- `docs/roadmap.md`
