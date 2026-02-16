# Review Checklist

Use this checklist for pull requests and self-review.

## Product and Behavior

- Change is scoped to one clear concept.
- Behavior is documented and testable.
- New behavior does not bypass command allowlisting/auth constraints.

## Code Quality

- Naming is explicit and intention-revealing.
- Complexity remains low; conditionals are readable.
- Non-obvious decisions include rationale.

## Tests

- Appropriate tests added or updated in `test/`.
- Existing suite remains green.
- Edge cases and failure modes are covered.

## Documentation

- User docs updated when user-facing behavior changes.
- Developer docs updated when architecture/workflow changes.
- Cross-links remain valid from `README.md` and `docs/index.md`.

## Operational Readiness

- Logging includes enough context for debugging.
- Config validation rejects ambiguous or unsafe inputs.
- Roadmap updated when scope changes materially.

## Done Criteria

- Run project checks: `fish -c go-task`.
- Ensure no unrelated changes are bundled.
- Ensure remaining TODOs are tracked as issues.

## See Also

- `docs/developer/onboarding.md`
- `docs/developer/testing.md`
- `AGENTS.md`
