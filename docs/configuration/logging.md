# Logging Configuration

Poke logging is configured under the top-level `logging` node.

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

- `level` (optional): `debug`, `info`, `warn`, or `error`.
- `format` (optional): `json` or `text`.
- `add_source` (optional): Include source file and line metadata in log records.
- `static_fields` (optional): Static key/value metadata added to every log record.
- `sink` (optional): Sink configuration.

## Sink

```yaml
sink:
  type: stdout # stdout or journald
```

- `sink.type`: Selects the active sink.
- `stdout`: Writes logs to standard output.
- `journald`: Writes logs to the native journald socket. When journald is
  unavailable, logging falls back to the configured fallback sink.

### Journald Sink

```yaml
sink:
  type: journald
  journald:
    identifier: poke-server
    fallback: stdout
```

- `identifier` (required when `type: journald`): journald identifier used by the service.
- `fallback` (optional): sink used when journald is unavailable. Current valid value: `stdout`.

## Defaults

```yaml
logging:
  level: info
  format: text
  add_source: false
  sink:
    type: stdout
```
