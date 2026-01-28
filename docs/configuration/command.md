# Command Spec

Commands can be defined under `command` yaml node.

## Env

Command can optionally contain a `env` node in the following format:

```yaml
env:
  strategy: isolate # (optional) must be one of "isolate", "inherit", "extend", "override"
  vals: # (optional) defines a map of ENV variables, defaults to empty map
    FOO: hello world
    BAR: 42
```

If `env` node is left out, default configuration is returned: `isolate` strategy
with empty environmental variables, i.e. command is executed in empty environment.
