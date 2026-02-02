# Poke Server

Poke is a small server that listens for command requests and executes a
whitelisted set of commands defined in config.

## Quickstart

1. Create a config file (start from `docs/configuration/config.example.yaml`).
2. Run the server with an explicit config path:

   ```sh
   go run ./cmd/server -c /path/to/poke.yml
   ```

3. Send a command request (HTTP listener example):

   ```sh
   curl -X PUT http://127.0.0.1:8008/ \
     -H "Content-Type: application/json" \
     -d '{"command_id":"uptime"}'
   ```

## Configuration

Poke reads configuration from `poke.yml`/`poke.yaml`. By default it searches:

- `/etc/poke/poke.yml`
- `$XDG_CONFIG_HOME/poke/poke.yml`
- `$HOME/config/poke/poke.yml`
- `$HOME/.poke/poke.yml`

You can also pass a config path via `-c FILE` or `--config FILE`.

Documentation references:

- Server config: [server.md](./docs/configuration/server.md)
- Command definitions: [command.md](./docs/configuration/command.md)
- Listener definitions: [listener.md](./docs/configuration/listener.md)
- Example config: [config.example.yaml](./docs/configuration/config.example.yaml)
- Development guide: [development.md](./docs/development.md)

## Implemented Features

- HTTP listener that accepts `PUT /` with `{"command_id":"..."}` payload.
- Command registry with short and full command specs.
- Whitelisted command execution via the binary executor.
- Per-command environment strategies (`isolate`, `inherit`, `extend`, `override`).
- Per-command timeouts and basic structured logging.
- Graceful shutdown via context cancellation.

## Work In Progress

See `docs/roadmap.md` for the full list. Current WIP items:

- HTTP TLS support
- HTTP auth
- Async execution and flags
- Executor responses
- Better logging and telemetry
