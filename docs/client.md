# Elevator React Client — Architecture & Design

## Overview

This is **Project 1** of the SaaS north-star plan ([specs/2026-05-30-elevator-saas-north-star-design.md](superpowers/specs/2026-05-30-elevator-saas-north-star-design.md)): a single-building, frontend-only rebuild that leaves the Go backend untouched. The React client (`client/`) is a redesigned, maintainable replacement for the original Svelte client, which has been removed (recoverable from git history at commit `8da68df`).

## Stack

| Layer | Choice | Rationale |
|-------|--------|-----------|
| **Build** | Vite + React 19 + TypeScript | SPA keeps the static GitHub Pages + nginx deploy; React 19 for modern concurrent rendering; Vite for fast HMR |
| **Styling** | Tailwind CSS | Semantic status palette in `tailwind.config.js`; utility-first consistency |
| **State** | Zustand | Ephemeral live state (the subscribed building's elevators) — simple, no boilerplate |
| **Server state** | TanStack Query | REST commands (create / delete / request floor) — caching, retries, optimistic updates |

## Architecture (Feature-Sliced Design)

```
client/src/
  app/
    App.tsx             router-free shell, providers, layout
    Header.tsx          connection status, theme toggle, create button
    main.tsx            entry point
    styles.css          Tailwind imports + global CSS

  features/
    building/
      domain/           pure logic (types, floors, validation, stats) — unit-tested
        types.ts
        floors.ts       floor range utilities, parking vs main grouping
        validation.ts   elevator config + floor request validation
        stats.ts        fleet summary: status counts, utilization, pending
      components/
        BuildingView.tsx         main dashboard container
        ElevatorShaft.tsx        single shaft visualization (animated car)
        FloorPicker.tsx          in-shaft floor-request form
        FleetSummary.tsx         status counts + utilization bars
        StatusLegend.tsx         color key
      useBuildingLiveState.ts   ← THE SEAM (see below)
      mutations.ts              TanStack Query: createElevator, deleteElevator, requestFloor
      constants.ts              CURRENT_BUILDING_ID (hardcoded "default" for Project 1)

    controls/
      CreateElevatorModal.tsx   modal with form, validation, TanStack mutation

  shared/
    api/
      client.ts         REST client (fetch wrapper, response unwrap)
      ws.ts             WebSocket manager (reconnect, ref-count, lifecycle)
      mapper.ts         wire → domain mapper (single source of truth for transforms)
      wire.ts           backend wire types (ElevatorStatusWire, etc.)
    store/
      liveStore.ts      Zustand: live elevators for the subscribed building
      toastStore.ts     Zustand: toast queue
      themeStore.ts     Zustand: dark mode toggle + localStorage
    ui/
      Button.tsx        reusable button (variants: primary, danger, ghost)
      Modal.tsx         accessible dialog with focus trap
      ConfirmDialog.tsx delete confirmation (not window.confirm)
      ToastContainer.tsx notification display
    lib/
      cn.ts             tailwind-merge + clsx wrapper
      log.ts            level-gated console (off by default; VITE_LOG=debug to enable)
```

## The Key Seam: `useBuildingLiveState(buildingId)`

A single hook that:
1. **Owns the WebSocket lifecycle**: connects on mount, ref-counts, disconnects when no subscribers
2. **Subscribes to a building**: in Project 1 `buildingId` is `CURRENT_BUILDING_ID` (constant); in Project 2 it becomes a route param
3. **Exposes live elevators**: `{ elevators, isConnecting, isConnected, error }`

**Why this matters**: when we go multi-building (Project 2), the socket will subscribe per building, and **component code does not change** — the same `useBuildingLiveState(buildingId)` call just uses the route param instead of the constant.

```typescript
// Project 1 (now)
const { elevators, isConnecting } = useBuildingLiveState(CURRENT_BUILDING_ID);

// Project 2 (future)
const { elevators, isConnecting } = useBuildingLiveState(buildingIdFromRoute);
// ← components unchanged; only the constant → param swap
```

### Connection Lifecycle

- **Ref-counted**: multiple `useBuildingLiveState(id)` calls share one socket per `id`
- **Auto-reconnect**: exponential backoff on disconnect
- **Object identity**: unchanged elevators keep the same object reference → memoized rows skip re-render

## What This Rebuild Fixes (vs. Svelte Client)

| Issue in Svelte | Fix in React |
|-----------------|--------------|
| WebSocket connects at import time (no lifecycle control) | Lifecycle tied to the viewed building (mount/unmount) |
| Full store rewrite every tick | Object identity preserved → memoized rows skip re-render |
| Duplicated backend→client transforms (drift) | One mapper (`shared/api/mapper.ts`) — single source of truth |
| Dead "Not implemented" API stubs | Only real endpoints exposed; stubs removed |
| `console.log` noise | Level-gated logger (off by default; `VITE_LOG=debug` to enable) |
| Blocking `window.confirm()` for delete | Accessible `ConfirmDialog` (focus trap, ESC to cancel) |

## Feature Parity with Svelte Client

### Kept / Redesigned

- **Header**: connection status indicator, theme toggle, create elevator button
- **Elevator shaft**: animated car, per-floor display, status color coding
- **Per-floor trip requests**: in-shaft `FloorPicker` (folded the Svelte left-sidebar control panel into the shaft)
- **Create elevator modal**: form with validation (name, min/max floor)
- **Toasts**: success/error feedback
- **Main/parking grouping**: floors ≥0 (main), <0 (parking)

### Restored (after first pass)

- **Delete confirmation**: accessible dialog instead of `window.confirm()`
- **Fleet summary**: replaces Svelte's MonitoringDashboard — status counts + utilization bars + pending requests, driven by the live WS snapshot (no extra polling)
- **Status legend**: color key for idle/moving-up/moving-down/deleting
- **Connecting state**: empty-state CTA no longer flashes before first snapshot
- **"Load sample fleet" button**: on empty state (calls backend sample data endpoint)

### Intentionally Dropped

- **Doors-open / has-passenger indicators**: backend never provided real values; Svelte hard-coded them
- **Per-tick 30s REST sync**: WebSocket already pushes updates; REST sync was redundant
- **Footer copyright + non-functional Doc/GitHub links**: clutter
- **Dead API stubs**: `emergencyStop`, `scheduleMaintenance`, etc. — backend doesn't implement, so removed from client

## Backend Contract

### REST API

Base URL: `http://localhost:6660/v1`

All responses wrapped in:

```json
{
  "success": true,
  "data": { ... },
  "meta": { "timestamp": "..." },
  "timestamp": "2026-05-30T10:30:00Z"
}
```

| Method | Path | Body | Description |
|--------|------|------|-------------|
| POST | `/elevators` | `{name, min_floor, max_floor}` | Create elevator |
| DELETE | `/elevators` | `{name}` | Graceful delete (removed when idle) — see [elevator_deletion.md](elevator_deletion.md) |
| POST | `/floors/request` | `{from, to}` | Request trip |
| GET | `/health` | — | Health check |

### WebSocket

URL: `ws://localhost:6661/ws/status`

Pushes a map keyed by elevator name roughly once per second:

```json
{
  "Elevator-1": {
    "current_floor": 5,
    "direction": "up",
    "requests": [3, 7, 9],
    "min_floor": -2,
    "max_floor": 20,
    "is_deleting": false
  },
  "Elevator-2": { ... }
}
```

The CSS `top` transition animates the elevator car between snapshots.

## Testing

- **Unit tests**: `features/building/domain/*.test.ts` (pure logic: floors, validation, stats)
- **Vitest**: test runner
- **@testing-library/react**: component test utilities (future)

Run: `npm run test` or `make client-test`

## Configuration

`.env.development`:

```bash
VITE_API_URL=http://localhost:6660
VITE_WS_URL=ws://localhost:6661
VITE_LOG=info  # off | error | warn | info | debug
```

## Project 2 Preview (Multi-Building SaaS)

When we go multi-building:

1. **Add React Router**: `/:buildingId` route param
2. **Swap the constant**: `useBuildingLiveState(buildingIdFromRoute)` instead of `CURRENT_BUILDING_ID`
3. **Backend adds building filtering**: WebSocket subscribes per building, REST endpoints scoped by building

**Component code unchanged** — the seam was designed for this.

## Deployment

- **Dev**: `make dev/local` (backend + client; dev server on `:5173`)
- **Production**: `make client-build` → `dist/` (static SPA, nginx or GitHub Pages); the `frontend` service in `docker-compose.full.yml` builds `client/Dockerfile` and serves the SPA via nginx

## Related Docs

- [elevator.md](elevator.md) — SCAN/LOOK algorithm, elevator movement
- [manager.md](manager.md) — fleet coordination, 3-phase selection
- [elevator_deletion.md](elevator_deletion.md) — graceful delete behavior
- [superpowers/specs/2026-05-30-elevator-saas-north-star-design.md](superpowers/specs/2026-05-30-elevator-saas-north-star-design.md) — full SaaS roadmap
