# Elevator Identifier Design

## Current: Name as Primary Key

Elevators are identified solely by `Name string`. The manager stores elevators in `[]*elevator.Elevator` (a slice) and looks them up by iterating and matching `e.Name()`.

Name is set at creation (`POST /v1/elevators {"name": "..."}`) and flows through:
- Manager slice lookup
- HTTP API request/response bodies
- WebSocket status broadcasts (`"name"` field)
- Frontend store keys

## Limitations

- **Not guaranteed unique** — manager does not enforce uniqueness on creation; two elevators named `"A"` cause undefined delete/lookup behavior
- **Mutable** — `SetName()` exists; a rename between request and delete could miss
- **Poor primary key** — user-visible labels shouldn't be system identifiers

## Why No UID Yet

The system predates this concern. Names work in practice because:
1. UI enforces unique names at creation (via modal validation)
2. `SetName()` is never called after construction in current code
3. Fleet size is small (≤100), linear scan is negligible

## Future: UID-based Identification

If added, the change would require:
1. Add `id string` (UUID v4) to `Elevator` struct, assigned in constructor
2. Add `ID` to `ElevatorStatus` and WebSocket broadcast
3. Add `map[string]*elevator.Elevator` index in Manager for O(1) lookup by ID
4. Update `DELETE /v1/elevators` to accept `id` instead of (or alongside) `name`
5. Update frontend store to key by `id`, use `id` for delete calls

Until then, treat `name` as effectively immutable and unique — enforced by the UI.
