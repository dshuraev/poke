# Listener Spec

Listeners are endpoints that accept command requests and forward them to executors.

All listeners are defined under top-level `listeners` node as a map. The key of
the map defines the type of listener, e.g. `http`.

Example configuration:

```yaml
listeners:
  http:
    host: 127.0.0.1
    port: 8080
    # optionally define read, write and idle timeouts
    read_timeout: 5s
    write_timeout: 5s
    idle_timeout: 0s
    # optionally enable TLS with server cert and key files
    tls:
      cert_file: /etc/poke/server.crt
      key_file: /etc/poke/server.key
```

If `host` or `port` are omitted, they default to `127.0.0.1:8008`.

If `tls` is set, both `cert_file` and `key_file` are required.
Environment variables in `cert_file` and `key_file` are expanded at load time.
The resolved paths must exist and be readable by the server.
