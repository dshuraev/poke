# Development

## Requirements

- Go 1.25 (see `taskfile.yaml`)
- [go-task](https://taskfile.dev/) (`task`) for convenience targets

## Quickstart

```sh
go mod download
```

```sh
# run the server

go run ./cmd/server -c /path/to/poke.yml
```

## Tasks

Common targets defined in `taskfile.yaml`:

```sh
# format

task fmt
```

```sh
# lint (installs tools first)

task lint
```

```sh
# test (race + coverage)

task test
```

```sh
# build

task build
```

```sh
# full check (lint + test + build)

task
```

## Tests

Tests live under `test/` and follow these categories:

- `test/unit/`
- `test/property/`
- `test/fuzz/`
- `test/integration/`
- `test/conformance/`

Common support helpers live in `test/support/`.

## Configuration

Use `docs/configuration/config.example.yaml` as a starting point. For details:

- `docs/configuration/server.md`
- `docs/configuration/command.md`
- `docs/configuration/listener.md`
