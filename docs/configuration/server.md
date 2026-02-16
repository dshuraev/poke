# Poke Server Configuration

Poke server is configured with `poke.yml`. The file defines what commands the
server can execute and how clients connect to request those commands.

The main configuration blocks are `commands` and `listeners`.
The optional `logging` block configures log level, format, and sink.

## Full Example

```yaml
commands:
  hello:
    name: "Hello"
    description: "say hello, via /usr/bin/echo binary"
    args: ["echo", "hello world"]
  uptime: uptime

listeners:
  http:
    host: 127.0.0.1
    port: 8080
    auth:
      api_token:
        token: "my-secret-token"

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

## Commands

Poke server cannot execute arbitrary commands. All commands must be explicitly
defined inside of `commands` block, e.g.

```yaml
commands:
  hello: ["echo", "hello"]
```

for more information, see [commands configuration](./command.md).

## Listeners

`listeners` block defines endpoints which `poke` uses to listen to commands from
clients requesting command execution, e.g.

```yaml
listeners:
  http:
    host: 127.0.0.1
    port: 8080
    auth:
      api_token:
        token: "my-secret-token"
```

for more information, see [listeners configuration](./listener.md).

## Logging

`logging` controls server log output shape and sink:

```yaml
logging:
  level: info # debug, info, warn, error
  format: text # text or json
  add_source: false
  static_fields:
    service: poke
  sink:
    type: stdout # stdout or journald
```

For full options, see [logging configuration](./logging.md).

## Defaults

`logging` defaults to:

```yaml
logging:
  level: info
  format: text
  add_source: false
  sink:
    type: stdout
```

Defaults for command execution and listener settings, including listener auth,
are documented in their respective configuration guides.
