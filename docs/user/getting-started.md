# Getting Started

This guide gets Poke running with a minimal secure setup.

## Prerequisites

- Go 1.25
- A config file based on `docs/configuration/config.example.yaml`

## Minimal Config

Create `poke.yml`:

```yaml
commands:
  hello: ["echo", "hello"]

listeners:
  http:
    host: 127.0.0.1
    port: 8008
    auth:
      api_token:
        token: "my-secret-token"
```

## Start Server

```sh
go run ./cmd/server -c /path/to/poke.yml
```

If `-c`/`--config` is omitted, Poke searches default paths:

- `/etc/poke/poke.yml`
- `$XDG_CONFIG_HOME/poke/poke.yml`
- `$HOME/config/poke/poke.yml`
- `$HOME/.poke/poke.yml`

## Send Command Request

```sh
curl -X PUT http://127.0.0.1:8008/ \
  -H "Content-Type: application/json" \
  -H "X-Poke-Auth-Method: api_token" \
  -H "X-Poke-API-Token: my-secret-token" \
  -d '{"command_id":"hello"}'
```

Expected status code: `202 Accepted`.

## Important Behavior

- Request body only identifies the `command_id`.
- Poke executes only commands registered in `commands` config.
- Current API does not return execution output in the HTTP response.

## See Also

- `docs/user/configuration.md`
- `docs/user/authentication.md`
- `docs/user/troubleshooting.md`
- `docs/configuration/server.md`
