# Auth Spec

Authentication is configured under the top-level `auth` YAML node.

`auth` is required in the server config; configs that omit it fail to parse.

The `auth` node is a map keyed by auth method name. Each configured method is
validated at load time.

## Minimal Example

```yaml
commands:
  uptime: uptime

listeners:
  http: {}

auth:
  api_token:
    token: "my-secret-token"
```

## Methods

### `api_token`

The `api_token` method validates a caller-provided token against a configured
secret.

#### Example (token literal)

```yaml
auth:
  api_token:
    listeners: [http]
    token: "my-secret-token"
```

#### Example (token from environment variable)

```yaml
auth:
  api_token:
    env: "POKE_API_TOKEN"
```

#### Example (token from file)

```yaml
auth:
  api_token:
    file: "/run/secrets/poke_api_token"
```

#### Fields

- `listeners` (optional): List of listener types that can accept this auth kind.
  - Listener types match the keys under the top-level `listeners` node (e.g. `http`).
  - If omitted or empty, all listener types are allowed.
  - Values are normalized by trimming whitespace and lowercasing.
  - Empty values and duplicates are rejected.

Exactly one credential source must be configured:

- `token` (optional): Literal token value.
  - Must be non-empty after trimming whitespace.
- `env` (optional): Environment variable name to read the token from.
  - The variable must exist and be non-empty after trimming whitespace.
- `file` (optional): Path to a file containing the token.
  - The file must be readable by the server process.
  - File contents are trimmed with whitespace/newlines removed from both ends.

#### Notes

- Token values are trimmed with `strings.TrimSpace(...)` to avoid common mistakes
  (YAML indentation, trailing newlines in secret files, etc.).
- Prefer `env` or `file` over `token` so secrets do not live in plaintext configs.

