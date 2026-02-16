# Server Configuration Reference

Poke loads configuration from `poke.yml` (or `poke.yaml`).

Top-level blocks:

- `commands`: command allowlist and execution settings.
- `listeners`: inbound request endpoints.
- `logging`: structured logging settings.

## Example

```yaml
commands:
  hello:
    name: "Hello"
    description: "Say hello"
    args: ["echo", "hello world"]
  uptime: uptime

listeners:
  http:
    host: 127.0.0.1
    port: 8080
    auth:
      api_token:
        env: POKE_API_TOKEN

logging:
  level: info
  format: text
  add_source: false
  static_fields:
    service: poke
    env: prod
  sink:
    type: stdout
```

## Notes

- Commands must be explicitly defined in `commands`.
- Listener auth is configured per listener under `listeners.<type>.auth`.
- Logging defaults are applied when `logging` is omitted.

## Defaults

When omitted, `logging` defaults to:

```yaml
logging:
  level: info
  format: text
  add_source: false
  sink:
    type: stdout
```

## See Also

- `docs/configuration/command.md`
- `docs/configuration/listener.md`
- `docs/configuration/logging.md`
- `docs/user/configuration.md`
