// THE KEY SEAM (north-star design §4.3).
//
// `useBuildingLiveState(buildingId)` is the single entry point for a building's
// live elevator state. In Project 1, buildingId is a constant ("the one
// building"); in Project 2 it becomes a route param and the WS subscribes
// per-building. Component code does NOT change — only what feeds the id.
//
// Today the backend has one global fleet and one global /ws/status stream, so
// buildingId is accepted but not yet sent on the wire. The hook owns the socket
// lifecycle so the connection is tied to the viewed building, not to import.
//
// The socket is ref-counted per buildingId: multiple components calling this
// hook for the same building share ONE connection. The last consumer to unmount
// tears it down.

import { useEffect, useMemo } from 'react';
import { ElevatorSocket } from '@/shared/api/ws';
import { mapBroadcast } from '@/shared/api/mapper';
import { useLiveStore } from '@/shared/store/liveStore';
import type { Elevator } from './domain/types';

export interface BuildingLiveState {
  elevators: Elevator[];
  connected: boolean;
  retryCount: number;
}

interface SharedConnection {
  socket: ElevatorSocket;
  refs: number;
  cleanup: () => void;
}

const connections = new Map<string, SharedConnection>();

function acquire(buildingId: string): SharedConnection {
  let conn = connections.get(buildingId);
  if (conn) {
    conn.refs += 1;
    return conn;
  }

  const socket = new ElevatorSocket(); // P2: new ElevatorSocket({ buildingId })
  const { setSnapshot, setConnection } = useLiveStore.getState();

  const offStatus = socket.onStatus((raw) => setSnapshot(mapBroadcast(raw)));
  const offState = socket.onStateChange((connected, retryCount) =>
    setConnection(connected, retryCount),
  );
  socket.connect();

  conn = {
    socket,
    refs: 1,
    cleanup: () => {
      offStatus();
      offState();
      socket.disconnect();
    },
  };
  connections.set(buildingId, conn);
  return conn;
}

function release(buildingId: string): void {
  const conn = connections.get(buildingId);
  if (!conn) return;
  conn.refs -= 1;
  if (conn.refs <= 0) {
    conn.cleanup();
    connections.delete(buildingId);
    useLiveStore.getState().reset();
  }
}

export function useBuildingLiveState(buildingId: string): BuildingLiveState {
  useEffect(() => {
    acquire(buildingId);
    return () => release(buildingId);
  }, [buildingId]);

  const byName = useLiveStore((s) => s.byName);
  const order = useLiveStore((s) => s.order);
  const connected = useLiveStore((s) => s.connected);
  const retryCount = useLiveStore((s) => s.retryCount);

  const elevators = useMemo(() => order.map((name) => byName[name]), [order, byName]);

  return { elevators, connected, retryCount };
}
