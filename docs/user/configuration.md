# Configuration Guide

Poke is configured with a YAML file (`poke.yml` or `poke.yaml`).

## Top-Level Blocks

- `commands`: Whitelisted commands callable by `command_id`.
- `listeners`: Request entrypoints (currently `http`).
- `logging`: Log level, format, and sink options.

## Minimal Secure Example

```yaml
commands:
  hello: ["echo", "hello"]

listeners:
  http:
    host: 127.0.0.1
    port: 8008
    auth:
      api_token:
        env: POKE_API_TOKEN

logging:
  level: info
  format: text
  sink:
    type: stdout
```

## Configuration Workflow

1. Start from `docs/configuration/config.example.yaml`.
2. Define only commands you intend to expose.
3. Configure listener auth (required).
4. Add TLS when requests cross untrusted networks.
5. Start with default logging, then tune as needed.

## Full Reference Specs

- Server layout: `docs/configuration/server.md`
- Command spec: `docs/configuration/command.md`
- Listener spec: `docs/configuration/listener.md`
- Auth spec: `docs/configuration/auth.md`
- Logging spec: `docs/configuration/logging.md`

## See Also

- `docs/user/getting-started.md`
- `docs/user/authentication.md`
- `docs/user/troubleshooting.md`
