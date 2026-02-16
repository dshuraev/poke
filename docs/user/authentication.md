# Authentication Guide

Poke requires listener authentication.

For the HTTP listener, configure auth under `listeners.http.auth`.

## API Token Method

Supported method: `api_token`.

Configure exactly one token source:

- `token`: inline token in config.
- `env`: environment variable containing token.
- `file`: file path containing token.

Example:

```yaml
listeners:
  http:
    auth:
      api_token:
        env: POKE_API_TOKEN
```

## Request Headers

When auth is enabled, clients must send:

- `X-Poke-Auth-Method: api_token`
- `X-Poke-API-Token: <token>`

Example request:

```sh
curl -X PUT http://127.0.0.1:8008/ \
  -H "Content-Type: application/json" \
  -H "X-Poke-Auth-Method: api_token" \
  -H "X-Poke-API-Token: my-secret-token" \
  -d '{"command_id":"uptime"}'
```

## Security Notes

- Prefer `env` or `file` over inline `token`.
- Rotate token regularly.
- Use TLS to protect token headers on the network.

## See Also

- `docs/configuration/auth.md`
- `docs/configuration/listener.md`
- `docs/user/getting-started.md`
- `docs/user/troubleshooting.md`
