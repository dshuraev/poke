# Developer Onboarding

This guide helps contributors set up, run checks, and navigate the repository.

## Requirements

- Go 1.25
- `task` (go-task)

## Repository Layout

- Entrypoint: `cmd/server/main.go`
- Runtime wiring: `internal/server/main.go`
- Core packages:
  - `internal/server/listener`
  - `internal/server/dispatch`
  - `internal/server/executor`
  - `internal/server/auth`
  - `internal/server/logging`
- Tests: `test/{unit,property,fuzz,integration,conformance}`

## Setup

```sh
go mod download
```

Run server:

```sh
go run ./cmd/server -c /path/to/poke.yml
```

## Development Commands

```sh
task fmt
```

```sh
task lint
```

```sh
task test
```

```sh
task build
```

Run full checks:

```sh
task
```

For CI-equivalent workflow in project policy:

```sh
fish -c go-task
```

## Contribution Practices

- Follow XP values in `AGENTS.md`: communication, simplicity, feedback, courage, respect.
- Keep changes small and isolated.
- Prefer explicit code over abstraction.
- Use Conventional Commits.

## See Also

- `docs/developer/architecture.md`
- `docs/developer/testing.md`
- `docs/developer/review-checklist.md`
