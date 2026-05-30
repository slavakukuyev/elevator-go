# Elevator React Client

React frontend for the elevator control system — a single-building dashboard with real-time WebSocket updates.

## Prerequisites

- Node.js 18+
- Go backend running (provides REST on `:6660`, WebSocket on `:6661`)

## Get Started

From this directory:

```bash
npm install
npm run dev        # dev server on http://localhost:5174
```

From the repo root (recommended):

```bash
make react/install
make dev/react     # starts backend + React client together
```

The app will be available at **http://localhost:5174**.

## Commands

| Command | Description |
|---------|-------------|
| `npm run dev` | Dev server on port 5174 |
| `npm run build` | Typecheck + production build → `dist/` |
| `npm run test` | Vitest unit tests |
| `npm run lint` | ESLint |
| `npm run typecheck` | TypeScript compiler check |

Or use the Makefile from the repo root:

```bash
make react/dev          # dev server
make react/build        # production build
make react/test         # tests
make react/lint         # lint
make react/check        # full check (typecheck + lint + test)
```

## Architecture & Design

See **[docs/client-react.md](../docs/client-react.md)** for:
- Feature-sliced architecture details
- The WebSocket seam (`useBuildingLiveState`)
- Stack choices (React 19, TanStack Query, Zustand)
- Backend contract (REST + WebSocket API)
- Feature parity vs. the legacy Svelte client

## Legacy Svelte Client

The original Svelte client (`../client`) is untouched and runs on port 5173. Both clients work side by side.
