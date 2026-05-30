# Elevator React Client (Project 1 — React foundation)

Redesigned, maintainable React frontend for the elevator control system. This is
**Project 1** of the SaaS north-star plan
(`docs/superpowers/specs/2026-05-30-elevator-saas-north-star-design.md`): a
single-building, frontend-only rebuild that leaves the Go backend untouched.

The legacy SvelteKit client (`../client`) is kept as a reference and still works;
this app runs on a separate dev port so both can run side by side.

## Stack

- **Vite + React 19 + TypeScript** (SPA — keeps the static GitHub Pages + nginx deploy)
- **Tailwind** for styling (semantic status palette in `tailwind.config.js`)
- **TanStack Query** for REST commands (create / delete / request floor)
- **Zustand** for ephemeral live state (the subscribed building's elevators)

## Commands

From this directory (npm):

```bash
npm install
npm run dev        # dev server on http://localhost:5174
npm run build      # typecheck + production build → dist/
npm run test       # vitest unit tests
npm run lint       # eslint
npm run typecheck  # tsc project references
```

From the repo root (Makefile — preferred):

```bash
make react/install     # install deps
make react/dev         # dev server on :5174
make react/build       # production build
make react/test        # vitest
make react/typecheck   # tsc
make react/lint        # eslint
make react/check       # typecheck + lint + test (full gate)
make dev/react         # backend (:6660/:6661) + React client (:5174)
```

Requires the Go backend running (`make server-dev` from the repo root → REST on
6660, WebSocket on 6661, or `make dev/react` to start both). Dev API/WS URLs are
set in `.env.development`.

## Architecture (feature-sliced)

```
src/
  app/                     router-free shell: providers, Header, App
  features/
    building/
      domain/              pure logic: types, floors, validation (unit-tested)
      components/          ElevatorShaft, BuildingView, FloorPicker
      useBuildingLiveState.ts   ← THE SEAM (see below)
      mutations.ts         TanStack Query commands
    controls/              CreateElevatorModal
  shared/
    api/                   wire shapes, client, ws, mapper (single source of truth)
    store/                 liveStore (Zustand), toastStore, themeStore
    ui/                    Button, Modal, ToastContainer
    lib/                   cn, level-gated logger
```

### The key seam: `useBuildingLiveState(buildingId)`

A single hook owns the WebSocket lifecycle and exposes a building's live
elevators. In Project 1 `buildingId` is a constant (`CURRENT_BUILDING_ID`); in
Project 2 it becomes a route param and the socket subscribes per building — **the
component code does not change**. Connections are ref-counted so multiple
consumers share one socket.

## What this rebuild fixes (vs. the Svelte client)

- No WebSocket connect at import time — lifecycle tied to the viewed building.
- No full-store rewrite every tick — object identity is preserved for unchanged
  elevators, so memoized rows skip re-render.
- One backend→client mapper instead of duplicated transforms that had drifted.
- Dead "Not implemented" API stubs removed; only real endpoints are exposed.
- `console.log` noise replaced by a level-gated logger (off by default;
  `VITE_LOG=debug` to enable).

## Feature parity with the Svelte client

**Kept / redesigned:** header (connection · theme · create), elevator shaft with
animated car, per-floor trip requests, create-elevator modal with validation,
toasts, main/parking grouping. The Svelte left-sidebar control panel and the
separate floor-call-button view are folded into the simpler in-shaft picker.

**Restored after first pass:** delete confirmation (an accessible dialog, not a
blocking `window.confirm()`); a compact **fleet summary** (status counts +
utilization bars + pending requests) replacing the Svelte MonitoringDashboard,
driven by the live WS snapshot — no extra polling; a **status legend**; a
**connecting** state so the empty-state CTA no longer flashes before the first
snapshot; a **"Load sample fleet"** button on the empty state.

**Intentionally dropped:** doors-open / has-passenger indicators and the per-tick
30s REST sync (the backend never provided real values — the Svelte client
hard-coded them); the footer copyright + non-functional Doc/GitHub links; the
dead API stubs (`emergencyStop`, `scheduleMaintenance`, …).

## Backend contract

- REST `http://localhost:6660/v1`, responses wrapped `{success, data, meta, timestamp}`
  - `POST /elevators` `{name, min_floor, max_floor}`
  - `DELETE /elevators` `{name}` (graceful — removed when idle)
  - `POST /floors/request` `{from, to}`
  - `GET /health`
- WebSocket `ws://localhost:6661/ws/status` pushes a map keyed by elevator name
  (`current_floor`, `direction`, `requests`, `min_floor`, `max_floor`,
  `is_deleting`) roughly once per second. The CSS `top` transition animates the
  car between snapshots.
