# Poke Server Configuration

Poke server is configured with `poke.yml`. The file defines what commands the
server can execute and how clients connect to request those commands.

Two main configuration blocks are `commands` and `listeners`.

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
```

for more information, see [listeners configuration](./listener.md).

## Defaults

There are no server-level defaults beyond the presence of `commands` and
`listeners`. Defaults for command execution and listener settings are documented
in their respective configuration guides.
