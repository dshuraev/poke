# Logging Configuration Reference

Logging is configured under top-level `logging`.

## Example

```yaml
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

## Fields

- `level`: `debug`, `info`, `warn`, `error`.
- `format`: `text` or `json`.
- `add_source`: include source file and line metadata.
- `static_fields`: key/value attributes added to all log records.
- `sink`: sink settings.

## Sink

```yaml
sink:
  type: stdout
```

Supported sink types:

- `stdout`
- `journald`

### Journald

```yaml
sink:
  type: journald
  journald:
    identifier: poke-server
    fallback: stdout
```

Rules:

- `identifier` is required for `journald`.
- `fallback` currently supports `stdout`.

## Defaults

```yaml
logging:
  level: info
  format: text
  add_source: false
  sink:
    type: stdout
```

## See Also

- `docs/configuration/server.md`
- `docs/user/troubleshooting.md`
