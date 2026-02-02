# Poke Server

Poke server is responsible for listening to command requests from clients and
executing whitelisted commands via runners.

## Quickstart

Poke server is configured via `poke.yml`/`poke.yaml` config file. By default,
it searches the following paths

- `/etc/poke/poke.yml`
- `$XDG_CONFIG_HOME/poke/poke.yml`
- `$HOME/config/poke/poke.yml`
- `$HOME/.poke/poke.yml`

Config can be also profided via `-c FILE`/`--config FILE` flag.

See [documentation for config](./configuration/server.md) for more details about
config's structure.
