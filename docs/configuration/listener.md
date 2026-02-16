# Listener Configuration Reference

Listeners receive command requests and forward them to dispatch.

Listeners are configured as a map under top-level `listeners`.

Supported listener types:

- `http`

## HTTP Listener Example

```yaml
listeners:
  http:
    host: 127.0.0.1
    port: 8080
    read_timeout: 5s
    write_timeout: 5s
    idle_timeout: 0s
    tls:
      cert_file: /etc/poke/server.crt
      key_file: /etc/poke/server.key
    auth:
      api_token:
        token: "my-secret-token"
```

## HTTP Defaults and Rules

- Default address: `127.0.0.1:8008`.
- If `tls` is configured, both `cert_file` and `key_file` are required.
- Environment variables are expanded in TLS file paths.
- `auth` is required and must define at least one method.

## HTTP Request Contract

- Method: `PUT`
- Path: `/`
- Body: `{"command_id":"<id>"}`

If accepted for execution, Poke returns `202 Accepted`.

## See Also

- `docs/configuration/auth.md`
- `docs/configuration/server.md`
- `docs/user/getting-started.md`
- `docs/user/authentication.md`
