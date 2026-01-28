# Command Spec

Commands are defined under the top-level `command` YAML node. Each key under
`command` becomes a command ID that you can call through the remote API. Command
IDs must be unique within a running server.

The minimal command definition is a list of arguments (or a single argument
string). The first argument must be a valid binary name that can be resolved via
`PATH`.

## Minimal Examples

```yaml
command:
  uptime: uptime # single argument; executed as "uptime"
  query-fs: ["df", "-h"] # list form; executed as "df -h"
```

Calling the `query-fs` command by ID:

```sh
curl -X PUT http://127.0.0.1:8080/ \
  -H "Content-Type: application/json" \
  -d '{"command_id":"query-fs"}'
```

## Full Command Object

Use an object when you need metadata or configuration beyond raw arguments.

```yaml
command:
  hello:
    name: "Hello"
    description: "say hello, via /usr/bin/echo binary"
    args: ["echo", "hello world"]
    executor: "bin"
  list-home:
    name: "ls"
    description: "Commands can have ENV configuration"
    env: # read more in the Env section
      HOME: /srv
    timeout: "5s" # duration with suffix, e.g. "5s", "1500ms"
```

### Fields

- `args` (required if object form): List of command arguments. The first entry
  is the binary to execute.
- `name` (optional): Human-readable command name.
- `description` (optional): Human-readable command description.
- `executor` (optional): Executor name used to run the command (e.g. `bin`).
- `env` (optional): Environment configuration for this command.
- `timeout` (optional): Timeout duration string with suffix (e.g. `5s`, `1500ms`).

## Env

Commands can optionally define an `env` node. Keys and values are always treated
as strings.

```yaml
env:
  strategy: isolate # optional; "isolate", "inherit", "extend", or "override"
  vals: # optional; map of environment variables
    FOO: hello world
    BAR: 42
```

### Strategies

- `isolate`: Use only the variables provided in `vals` (clean environment).
- `inherit`: Use the parent process environment as-is.
- `extend`: Start from the parent environment and add only missing keys from `vals`.
- `override`: Start from the parent environment and overwrite with `vals`.

If the `env` node is omitted, the default is `isolate` with an empty `vals` map,
meaning the command runs in an empty environment.
