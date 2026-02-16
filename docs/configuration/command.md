# Command Configuration Reference

Commands are defined under top-level `commands`.

Each key under `commands` is a command identifier (`command_id`) callable via
API requests.

## Minimal Forms

Single argument shorthand:

```yaml
commands:
  uptime: uptime
```

Argument list shorthand:

```yaml
commands:
  query-fs: ["df", "-h"]
```

## Full Command Form

```yaml
commands:
  hello:
    name: "Hello"
    description: "Say hello"
    args: ["echo", "hello world"]
    executor: "bin"
    timeout: 5s
    env:
      strategy: isolate
      vals:
        FOO: bar
```

## Fields

- `args` (required in object form): command and arguments.
- `name` (optional): human-readable name.
- `description` (optional): human-readable description.
- `executor` (optional): executor name (`bin` is currently supported).
- `timeout` (optional): Go duration string, for example `5s`, `1500ms`.
- `env` (optional): environment config.

## Environment Strategy

```yaml
env:
  strategy: isolate
  vals:
    FOO: hello
```

Supported `strategy` values:

- `isolate`: only `vals` are provided.
- `inherit`: parent process environment only.
- `extend`: parent env plus missing keys from `vals`.
- `override`: parent env with `vals` overriding existing keys.

Default when `env` is omitted: `isolate` with empty `vals`.

## See Also

- `docs/configuration/server.md`
- `docs/configuration/listener.md`
- `docs/user/configuration.md`
