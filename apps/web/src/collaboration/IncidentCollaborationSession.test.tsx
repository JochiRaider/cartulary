import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { useEffect, useRef } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  type IncidentCollaborationEvent,
  IncidentCollaborationSession,
  useIncidentCollaborationSession,
} from "./IncidentCollaborationSession";

class FakeWebSocket {
  static readonly OPEN = 1;
  static instances: FakeWebSocket[] = [];
  readonly sent: string[] = [];
  readyState = 0;
  onopen: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onclose: ((event: { code: number; reason: string }) => void) | null = null;
  onerror: (() => void) | null = null;

  constructor(readonly url: string) {
    FakeWebSocket.instances.push(this);
  }

  send(payload: string) {
    this.sent.push(payload);
  }

  close() {
    this.readyState = 3;
  }
}

let serverEventOrdinal = 0;

function serverMessage(message: Record<string, unknown>) {
  serverEventOrdinal += 1;
  return {
    emitted_at: "2026-07-13T12:00:00Z",
    event_id: `event-${serverEventOrdinal}`,
    incident_id: "incident-1",
    ...message,
  };
}

function Consumer({
  onEvent,
  sheetId,
}: {
  readonly onEvent: (event: IncidentCollaborationEvent) => void;
  readonly sheetId: string;
}) {
  const session = useIncidentCollaborationSession();
  useEffect(() => session.subscribe(onEvent), [onEvent, session]);
  useEffect(() => {
    session.publishPresence({
      sheet_ref: { kind: "view_schema", id: sheetId },
      mode: "viewing",
    });
  }, [session, sheetId]);
  return null;
}

function ResetProbe({
  onEvent,
}: {
  readonly onEvent: (event: IncidentCollaborationEvent) => void;
}) {
  const session = useIncidentCollaborationSession();
  const generationRef = useRef<number | null>(null);
  useEffect(
    () =>
      session.subscribe((event) => {
        if (event.kind === "reset_required") {
          generationRef.current = event.generation;
        }
        onEvent(event);
      }),
    [onEvent, session],
  );
  return (
    <button
      onClick={() => {
        if (generationRef.current !== null) {
          session.completeReset(generationRef.current);
        }
      }}
      type="button"
    >
      {session.status}
    </button>
  );
}

describe("IncidentCollaborationSession", () => {
  afterEach(() => {
    cleanup();
    FakeWebSocket.instances = [];
    serverEventOrdinal = 0;
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it("ignores duplicate stream sequences and refreshes on sequence gaps", () => {
    vi.stubGlobal("WebSocket", FakeWebSocket);
    const onEvent = vi.fn();
    const view = render(
      <IncidentCollaborationSession
        incidentId="incident-1"
        initialPresence={{
          sheet_ref: { kind: "view_schema", id: "timeline" },
          mode: "viewing",
        }}
      >
        <Consumer onEvent={onEvent} sheetId="timeline" />
      </IncidentCollaborationSession>,
    );
    const socket = FakeWebSocket.instances[0];
    expect(socket).toBeDefined();
    act(() => {
      if (socket) {
        socket.readyState = FakeWebSocket.OPEN;
      }
      socket?.onopen?.();
    });
    expect(JSON.parse(socket?.sent[0] ?? "{}")).toMatchObject({
      type: "hello",
      payload: { presence: { sheet_ref: { id: "timeline" } } },
    });
    act(() => {
      socket?.onmessage?.({
        data: JSON.stringify(
          serverMessage({
            type: "hello_ack",
            payload: {
              connection_id: "connection-1",
              resume_token: "resume-1",
              server_time: "2026-07-13T12:00:00Z",
              heartbeat_interval_ms: 15_000,
              presence_ttl_ms: 45_000,
              resume_window_ms: 300_000,
            },
          }),
        ),
      });
    });

    view.rerender(
      <IncidentCollaborationSession
        incidentId="incident-1"
        initialPresence={{
          sheet_ref: { kind: "view_schema", id: "timeline" },
          mode: "viewing",
        }}
      >
        <Consumer onEvent={onEvent} sheetId="notes" />
      </IncidentCollaborationSession>,
    );
    expect(FakeWebSocket.instances).toHaveLength(1);
    expect(JSON.parse(socket?.sent.at(-1) ?? "{}")).toMatchObject({
      type: "presence_update",
      payload: { presence: { sheet_ref: { id: "notes" } } },
    });

    const change = serverMessage({
      ignored_envelope_route: "/must-not-follow",
      type: "extension_resource_changed",
      stream_seq: 1,
      payload: {
        extension_profile_id: "network_flow_activity",
        resource_kind: "network_flow_table",
        resource_id: "nft-1",
        change_kind: "invalidate",
        ignored_payload_action: "must-not-execute",
        reason_code: "changed",
      },
    });
    act(() => {
      socket?.onmessage?.({ data: JSON.stringify(change) });
      socket?.onmessage?.({ data: JSON.stringify(change) });
      socket?.onmessage?.({
        data: JSON.stringify({ ...change, stream_seq: 3 }),
      });
      socket?.onmessage?.({
        data: JSON.stringify(serverMessage({ type: "ping", payload: {} })),
      });
      socket?.onmessage?.({
        data: JSON.stringify(
          serverMessage({
            type: "future_extension",
            payload: {},
          }),
        ),
      });
    });
    expect(
      onEvent.mock.calls.filter(([event]) => event.kind === "message"),
    ).toHaveLength(1);
    expect(JSON.stringify(onEvent.mock.calls)).not.toContain("must-not");
    expect(onEvent).toHaveBeenCalledWith({
      generation: 1,
      kind: "reset_required",
      reason: "sequence_gap",
    });
    expect(JSON.parse(socket?.sent.at(-1) ?? "{}")).toEqual({
      type: "pong",
      payload: {},
    });

    view.rerender(
      <IncidentCollaborationSession
        incidentId="incident-2"
        initialPresence={{
          sheet_ref: { kind: "view_schema", id: "timeline" },
          mode: "viewing",
        }}
      >
        <Consumer onEvent={onEvent} sheetId="timeline" />
      </IncidentCollaborationSession>,
    );
    expect(socket?.readyState).toBe(3);
    expect(FakeWebSocket.instances).toHaveLength(2);
    const nextSocket = FakeWebSocket.instances[1];
    act(() => {
      if (nextSocket) {
        nextSocket.readyState = FakeWebSocket.OPEN;
      }
      nextSocket?.onopen?.();
    });
    expect(JSON.parse(nextSocket?.sent[0] ?? "{}")).toMatchObject({
      type: "hello",
      payload: { presence: { sheet_ref: { id: "timeline" } } },
    });
  });

  it("builds hello and resume session messages from socket resume state", () => {
    vi.useFakeTimers();
    vi.stubGlobal("WebSocket", FakeWebSocket);
    const onEvent = vi.fn();
    render(
      <IncidentCollaborationSession
        incidentId="incident-1"
        initialPresence={{
          sheet_ref: { kind: "view_schema", id: "timeline" },
          mode: "viewing",
        }}
      >
        <Consumer onEvent={onEvent} sheetId="timeline" />
      </IncidentCollaborationSession>,
    );
    const socket = FakeWebSocket.instances[0];
    act(() => {
      if (socket) {
        socket.readyState = FakeWebSocket.OPEN;
      }
      socket?.onopen?.();
      socket?.onmessage?.({
        data: JSON.stringify(
          serverMessage({
            type: "hello_ack",
            payload: {
              connection_id: "connection-1",
              resume_token: "resume-private",
              server_time: "2026-07-13T12:00:00Z",
              heartbeat_interval_ms: 15_000,
              presence_ttl_ms: 45_000,
              resume_window_ms: 300_000,
            },
          }),
        ),
      });
      socket?.onmessage?.({
        data: JSON.stringify(
          serverMessage({
            type: "extension_resource_changed",
            stream_seq: 7,
            payload: {
              extension_profile_id: "network_flow_activity",
              resource_kind: "network_flow_table",
              resource_id: "nft-1",
              change_kind: "invalidate",
              reason_code: "changed",
            },
          }),
        ),
      });
    });
    expect(JSON.parse(socket?.sent[0] ?? "{}")).toMatchObject({
      type: "hello",
      payload: { client_instance_id: expect.any(String) },
    });
    expect(onEvent).toHaveBeenCalledWith({
      kind: "established",
      messageType: "hello_ack",
      payload: { connection_id: "connection-1" },
    });
    expect(JSON.stringify(onEvent.mock.calls)).not.toContain("resume-private");

    act(() => {
      socket?.onclose?.({ code: 1006, reason: "" });
      vi.advanceTimersByTime(1000);
    });
    const resumedSocket = FakeWebSocket.instances[1];
    act(() => {
      if (resumedSocket) {
        resumedSocket.readyState = FakeWebSocket.OPEN;
      }
      resumedSocket?.onopen?.();
    });
    expect(JSON.parse(resumedSocket?.sent[0] ?? "{}")).toMatchObject({
      type: "resume",
      payload: {
        resume_token: "resume-private",
        last_seen_stream_seq: 7,
      },
    });
  });

  it("keeps replayable events unsynchronized until the owner completes reset", () => {
    vi.stubGlobal("WebSocket", FakeWebSocket);
    const onEvent = vi.fn();
    render(
      <IncidentCollaborationSession
        incidentId="incident-1"
        initialPresence={{
          sheet_ref: { kind: "view_schema", id: "timeline" },
          mode: "viewing",
        }}
      >
        <ResetProbe onEvent={onEvent} />
      </IncidentCollaborationSession>,
    );
    const socket = FakeWebSocket.instances[0];
    act(() => {
      if (socket) {
        socket.readyState = FakeWebSocket.OPEN;
      }
      socket?.onopen?.();
      socket?.onmessage?.({
        data: JSON.stringify(
          serverMessage({
            type: "resume_ack",
            payload: {
              resume_token: "resume-private",
              server_high_water_stream_seq: 1,
              status: "replayed",
            },
          }),
        ),
      });
      socket?.onmessage?.({
        data: JSON.stringify(
          serverMessage({
            type: "extension_resource_changed",
            stream_seq: 3,
            payload: {
              extension_profile_id: "network_flow_activity",
              resource_kind: "network_flow_table",
              resource_id: "nft-1",
              change_kind: "invalidate",
              reason_code: "changed",
            },
          }),
        ),
      });
      socket?.onmessage?.({
        data: JSON.stringify(
          serverMessage({
            type: "extension_resource_changed",
            stream_seq: 4,
            payload: {
              extension_profile_id: "network_flow_activity",
              resource_kind: "network_flow_table",
              resource_id: "nft-1",
              change_kind: "invalidate",
              reason_code: "changed",
            },
          }),
        ),
      });
    });

    expect(screen.getByRole("button").textContent).toBe("resetting");
    expect(onEvent).toHaveBeenCalledWith({
      generation: 1,
      kind: "reset_required",
      reason: "sequence_gap",
    });
    expect(
      onEvent.mock.calls.filter(
        ([event]) =>
          event.kind === "message" &&
          event.message.type === "extension_resource_changed",
      ),
    ).toHaveLength(0);

    fireEvent.click(screen.getByRole("button"));
    expect(screen.getByRole("button").textContent).toBe("connected");
  });

  it("suppresses reconnect and clears connection state after revocation", () => {
    vi.useFakeTimers();
    vi.stubGlobal("WebSocket", FakeWebSocket);
    const onEvent = vi.fn();
    render(
      <IncidentCollaborationSession
        incidentId="incident-1"
        initialPresence={{
          sheet_ref: { kind: "view_schema", id: "timeline" },
          mode: "viewing",
        }}
      >
        <ResetProbe onEvent={onEvent} />
      </IncidentCollaborationSession>,
    );
    const socket = FakeWebSocket.instances[0];
    act(() => {
      if (socket) {
        socket.readyState = FakeWebSocket.OPEN;
      }
      socket?.onopen?.();
      socket?.onmessage?.({
        data: JSON.stringify(
          serverMessage({
            type: "hello_ack",
            payload: {
              connection_id: "connection-1",
              resume_token: "resume-private",
              server_time: "2026-07-13T12:00:00Z",
              heartbeat_interval_ms: 15_000,
              presence_ttl_ms: 45_000,
              resume_window_ms: 300_000,
            },
          }),
        ),
      });
      socket?.onmessage?.({
        data: JSON.stringify(
          serverMessage({
            type: "session_revoked",
            payload: { reason_code: "session_revoked" },
          }),
        ),
      });
      vi.advanceTimersByTime(5_000);
    });

    expect(screen.getByRole("button").textContent).toBe("authorization_lost");
    expect(onEvent).toHaveBeenCalledWith({ kind: "session_revoked" });
    expect(FakeWebSocket.instances).toHaveLength(1);
  });
});
