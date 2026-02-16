# Authentication Configuration Reference

Authentication is configured per listener under `listeners.<type>.auth`.

For HTTP, auth configuration lives in `listeners.http.auth`.

## Supported Method

- `api_token`

## API Token Config

Exactly one token source is required:

- `token`: inline token value.
- `env`: environment variable name.
- `file`: path to file containing token.

### Examples

Literal token:

```yaml
listeners:
  http:
    auth:
      api_token:
        token: "my-secret-token"
```

Environment variable:

```yaml
listeners:
  http:
    auth:
      api_token:
        env: "POKE_API_TOKEN"
```

File:

```yaml
listeners:
  http:
    auth:
      api_token:
        file: "/run/secrets/poke_api_token"
```

## HTTP Headers

When using `api_token`, clients send:

- `X-Poke-Auth-Method: api_token`
- `X-Poke-API-Token: <token>`

## Notes

- Token inputs are trimmed for surrounding whitespace.
- Prefer `env` or `file` to avoid plain-text secrets in config files.

## See Also

- `docs/configuration/listener.md`
- `docs/user/authentication.md`
- `docs/user/troubleshooting.md`
