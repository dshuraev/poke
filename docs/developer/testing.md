# Testing Strategy

Poke tests are organized by test intent.

## Test Categories

- `test/unit/`: Fast deterministic logic tests.
- `test/property/`: Short property-based tests.
- `test/fuzz/`: Long randomized/stateful tests.
- `test/integration/`: Cross-module behavior.
- `test/conformance/`: Spec/contract behavior.
- `test/support/`: Shared generators/helpers/models.

## Typical Commands

Run full suite with race and coverage:

```sh
task test
```

Run quicker test pass:

```sh
task test-short
```

Run package-level checks:

```sh
go test ./...
```

## What To Test

- Parser and config validation for new config behavior.
- Listener request validation and auth behavior for API changes.
- Dispatcher behavior for command routing/execution control flow.
- Executor timeout/env semantics for command execution changes.
- Logging config validation when adding sink/format options.

## Review Expectations

- Every behavioral change should be demonstrated by tests.
- Keep tests close to behavior contracts.
- Prefer focused unit tests before adding broader integration tests.

## See Also

- `docs/developer/review-checklist.md`
- `docs/developer/architecture.md`
- `AGENTS.md`
