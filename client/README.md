# Elevator React Client

React frontend for the elevator control system — a single-building dashboard with real-time WebSocket updates.

## Prerequisites

- Node.js 18+
- Go backend running (provides REST on `:6660`, WebSocket on `:6661`)

## Get Started

From this directory:

```bash
npm install
npm run dev        # dev server on http://localhost:5173
```

From the repo root (recommended):

```bash
make client-install
make dev/local     # starts backend + client together
```

The app will be available at **http://localhost:5173**.

## Commands

| Command | Description |
|---------|-------------|
| `npm run dev` | Dev server on port 5173 |
| `npm run build` | Typecheck + production build → `dist/` |
| `npm run test` | Vitest unit tests |
| `npm run lint` | ESLint |
| `npm run typecheck` | TypeScript compiler check |

Or use the Makefile from the repo root:

```bash
make client-dev         # dev server
make client-build       # production build
make client-test        # tests
make lint/ts            # lint
make client-check       # full check (typecheck + lint + test)
```

## Architecture & Design

See **[docs/client.md](../docs/client.md)** for:
- Feature-sliced architecture details
- The WebSocket seam (`useBuildingLiveState`)
- Stack choices (React 19, TanStack Query, Zustand)
- Backend contract (REST + WebSocket API)

## History

This React client replaced the original Svelte client. The Svelte implementation
is recoverable from git history at commit `8da68df` (the last commit before the
React rebuild landed).
