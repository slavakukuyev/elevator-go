# Elevator Graceful Deletion

## Flow

```
DELETE /v1/elevators {"name": "X"}
    ↓
Manager acquires write lock
  → lookup + IsMarkedForDeletion() check + MarkForDeletion() — atomic under same lock
  → releases lock
    ↓
Elevator continues Run() loop (direction unchanged)
  → isDeleting=true blocks CanAcceptRequests()
  → GetStatus() sets IsDeleting=true in broadcasts
    ↓
WebSocket: broadcasts is_deleting:true with live floor/direction
Frontend: shows "Deleting..." badge, filters elevator from floor routing
    ↓
waitForElevatorToFinish() polls until PendingRequestCount()==0
  → timeout → force-remove anyway (no zombie)
    ↓
removeElevatorFromList() + elevator.Shutdown()
    ↓
WebSocket stops broadcasting this elevator
Frontend detects absence → removes from store → UI clears
```

## Key Design Decisions

### atomic.Bool instead of DirectionDeleting

`MarkForDeletion()` sets `isDeleting atomic.Bool = true` — it does NOT change the elevator's direction.

**Why**: The `Run()` loop only handles `DirectionUp`, `DirectionDown`, `DirectionIdle`. Setting `DirectionDeleting` caused the SCAN algorithm to ignore pending requests — elevator froze mid-flight.

**Result**: Elevator keeps moving in its current direction, serves all queued requests, then becomes idle naturally.

### Race condition protection

The get → check → mark sequence is atomic under `m.mu.Lock()`:

```go
m.mu.Lock()
// inline lookup (not GetElevator which re-acquires lock)
elevator := findByName(name)
if elevator.IsMarkedForDeletion() { m.mu.Unlock(); return conflict error }
elevator.MarkForDeletion()
m.mu.Unlock()
// wait + remove outside lock
```

Prevents two concurrent DELETE requests both passing the `IsMarkedForDeletion()` check.

### Force-remove on timeout

`waitForElevatorToFinish()` uses a context with timeout. On timeout, the elevator is **still removed** — the error is logged as a warning but does not abort deletion:

```go
waitErr := m.waitForElevatorToFinish(deleteCtx, elevator)
// waitErr != nil → log warn, continue
removeElevatorFromList(name)
elevator.Shutdown()
```

Prevents zombie elevators stuck in "Deleting" forever.

## Files

| File | Role |
|------|------|
| `internal/elevator/elevator.go` | `isDeleting atomic.Bool`, `MarkForDeletion()`, `CanAcceptRequests()`, `GetStatus()` |
| `internal/manager/manager.go` | `DeleteElevator()` — atomic check+mark, `waitForElevatorToFinish()`, force-remove |
| `internal/domain/elevator_status.go` | `ElevatorStatus.IsDeleting` field, JSON `is_deleting` |
| `client/src/utils/floorSelection.ts` | Filters `isDeleting` elevators from routing |
| `client/src/services/websocket.ts` | Maps `is_deleting` → `status: 'deleting'` |
| `client/src/components/elevator/ElevatorBuilding.svelte` | "Deleting..." badge, confirm dialog |

## API

```http
DELETE /v1/elevators
Content-Type: application/json

{"name": "Elevator-1"}
```

Responses:
- `200` — deletion initiated (elevator will disappear from WS when complete)
- `404` — elevator not found
- `409` — elevator already being deleted
- `500` — internal error (elevator was still force-removed)
