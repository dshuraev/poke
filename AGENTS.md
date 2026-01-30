# agents.md

This document defines how agents—human and automated—operate within this codebase.  
It follows **Extreme Programming (XP)** principles (communication, simplicity,
feedback, courage, respect) and adopts **Conventional Commits** for clear traceability.

## Code Generation Guidelines

### Core principals

- **Clarity over cleverness**: Use simple, explicit patterns. Avoid abstractions unless absolutely necessary.
- **Minimal changes**: Make one logical change at a time. Never bundle unrelated modifications.
- **Self-documenting code**: Choose descriptive names that explain intent without needing comments.

### Before any code change

1. Explain WHAT you're changing and WHY
2. Describe the design approach and alternatives you considered
3. List any assumptions or trade-offs
4. Generate code only when the user confirmed the proposal

### In the code itself

- Add docstrings for all functions/classes explaining purpose and usage
- Comment the "why" not the "what" - explain business logic, edge cases, non-obvious decisions
- Document any Magic numbers or configuration values

## Change Management

- Make ONE conceptual change per file modification
- If fixing a bug AND refactoring, do them separately
- Provide a brief summary after each change explaining what's different

## Code Style

- Prefer explicit over implicit (e.g., `getUserById(id)` not `get(id)`)
- Low cyclometric complexity: avoid nested ternaries, complex conditionals - break into named variables
- Maximum function length: ~50 lines. If longer, explain why or refactor

## Rationale Requirement

For any non-trivial design decision, document:

- Why this approach over alternatives
- What problems it solves
- What limitations it has
- Any future considerations

## Review Checklist

Before presenting code, verify:

- [ ] Can someone unfamiliar read this and understand the intent?
- [ ] Is the design rationale documented?
- [ ] Are changes isolated and minimal?
- [ ] Would I understand this code in 6 months?

## Testing

- All tests reside under `test/{unit,property,fuzz,integration,conformance}/`.
- Common generators, models, and case templates live in `test/support/`.

| Category        | Purpose                                                                      |
| --------------- | ---------------------------------------------------------------------------- |
| **unit**        | Fast, deterministic, pure logic.                                             |
| **property**    | Short property-based tests.                                                  |
| **fuzz**        | Long-running randomized or stateful tests, usually excluded from default CI. |
| **integration** | Cross-module, networking, or persistence behavior.                           |
| **conformance** | Spec and contract validation for CRDT behaviours and sync protocols.         |

## Definition of Done

- **Tests:** Appropriate coverage (unit/property/fuzz/conformance/integration as relevant).
- **Documentation:** README or guide reflects current design and usage.
- **Observability:** Metrics/logging hooks instrumented where applicable.
- **Debt:** All remaining TODOs are converted into tracked issues before merge.
- **CI:**: All code must clear the CI checks (run `fish -c go-task`).

## Review Policy

- Prefer **small, frequent merges** (<300 lines diff).  
- Every change must be demonstrable (test or demo).  
- Reject complexity not required by a failing test or explicit story.  
- Keep trunk green — integration tests must always pass.
