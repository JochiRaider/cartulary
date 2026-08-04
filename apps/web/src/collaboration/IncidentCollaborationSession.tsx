import {
  type IncidentStreamMessage,
  incidentStreamMessageDecoder,
} from "@cartulary/protocol-ts/collaboration";
import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { SheetRef } from "../shared/sheetRef";

export type CollaborationPresence = {
  readonly sheet_ref: SheetRef;
  readonly mode: "viewing" | "editing" | "idle";
  readonly record_id?: string;
  readonly field_key?: string;
};

export type IncidentCollaborationEvent =
  | {
      readonly kind: "established";
      readonly messageType: "hello_ack" | "resume_ack";
      readonly payload: Readonly<{
        connection_id?: string;
        server_high_water_stream_seq?: number;
        status?: string;
      }>;
    }
  | { readonly kind: "message"; readonly message: IncidentStreamMessage }
  | {
      readonly kind: "reset_required";
      readonly generation: number;
      readonly reason: "resume_reset" | "sequence_gap";
    }
  | { readonly kind: "authorization_lost" }
  | { readonly kind: "session_revoked" }
  | { readonly kind: "incident_closed" };

export type IncidentCollaborationMessage = IncidentStreamMessage;

export type IncidentCollaborationStatus =
  | "connecting"
  | "connected"
  | "resetting"
  | "disconnected"
  | "authorization_lost"
  | "incident_closed";

type IncidentCollaborationListener = (
  event: IncidentCollaborationEvent,
) => void;

export type IncidentCollaborationSessionValue = {
  readonly clientInstanceId: string;
  readonly completeReset: (generation: number) => boolean;
  readonly connectionId: string | null;
  readonly disconnect: () => void;
  readonly publishPresence: (presence: CollaborationPresence) => void;
  readonly reconnect: () => void;
  readonly status: IncidentCollaborationStatus;
  readonly subscribe: (listener: IncidentCollaborationListener) => () => void;
};

const IncidentCollaborationContext =
  createContext<IncidentCollaborationSessionValue | null>(null);

function websocketPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    return `${protocol}//${window.location.host}${path}`;
  }
  const target = new URL(trimmedBase, window.location.origin);
  target.protocol = target.protocol === "https:" ? "wss:" : "ws:";
  target.pathname = path;
  target.search = "";
  target.hash = "";
  return target.toString();
}

function tabClientInstanceId(): string {
  const key = "cartulary.client_instance_id";
  try {
    const existing = window.sessionStorage.getItem(key);
    if (existing) {
      return existing;
    }
    const created =
      window.crypto?.randomUUID?.() ?? `${Date.now()}-${Math.random()}`;
    window.sessionStorage.setItem(key, created);
    return created;
  } catch {
    return `${Date.now()}-${Math.random()}`;
  }
}

function recordValue(value: unknown): Record<string, unknown> {
  if (value !== null && typeof value === "object" && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  return {};
}

function streamSequence(message: IncidentStreamMessage): number | null {
  const value = message.stream_seq;
  return typeof value === "number" ? value : null;
}

function replayable(message: IncidentStreamMessage): boolean {
  return (
    message.type === "record_changed" ||
    message.type === "extension_resource_changed" ||
    message.type === "job_progress"
  );
}

export function IncidentCollaborationSession({
  apiBase,
  children,
  incidentId,
  initialPresence,
}: {
  readonly apiBase?: string | undefined;
  readonly children: ReactNode;
  readonly incidentId: string;
  readonly initialPresence: CollaborationPresence;
}) {
  const clientInstanceIdRef = useRef<string | null>(null);
  if (clientInstanceIdRef.current === null) {
    clientInstanceIdRef.current = tabClientInstanceId();
  }
  const clientInstanceId = clientInstanceIdRef.current;
  const listenersRef = useRef(new Set<IncidentCollaborationListener>());
  const presenceRef = useRef(initialPresence);
  const socketRef = useRef<WebSocket | null>(null);
  const connectRef = useRef<() => void>(() => undefined);
  const resumeTokenRef = useRef<string | null>(null);
  const lastSeenStreamSeqRef = useRef(0);
  const resetGenerationRef = useRef(0);
  const reconnectSuppressedRef = useRef(false);
  const [connectionId, setConnectionId] = useState<string | null>(null);
  const [status, setStatus] =
    useState<IncidentCollaborationStatus>("disconnected");
  const statusRef = useRef<IncidentCollaborationStatus>("disconnected");

  const updateStatus = useCallback((next: IncidentCollaborationStatus) => {
    statusRef.current = next;
    setStatus(next);
  }, []);

  const emit = useCallback((event: IncidentCollaborationEvent) => {
    for (const listener of listenersRef.current) {
      listener(event);
    }
  }, []);

  const subscribe = useCallback((listener: IncidentCollaborationListener) => {
    listenersRef.current.add(listener);
    return () => {
      listenersRef.current.delete(listener);
    };
  }, []);

  const publishPresence = useCallback((presence: CollaborationPresence) => {
    presenceRef.current = presence;
    const socket = socketRef.current;
    if (
      statusRef.current !== "connected" ||
      socket === null ||
      socket.readyState !== WebSocket.OPEN
    ) {
      return;
    }
    socket.send(
      JSON.stringify({ type: "presence_update", payload: { presence } }),
    );
  }, []);

  const completeReset = useCallback(
    (generation: number) => {
      if (
        generation !== resetGenerationRef.current ||
        statusRef.current !== "resetting"
      ) {
        return false;
      }
      const socket = socketRef.current;
      if (socket !== null && socket.readyState === WebSocket.OPEN) {
        updateStatus("connected");
        return true;
      }
      return false;
    },
    [updateStatus],
  );

  const disconnect = useCallback(() => {
    reconnectSuppressedRef.current = true;
    socketRef.current?.close();
    socketRef.current = null;
    setConnectionId(null);
    updateStatus("disconnected");
  }, [updateStatus]);

  const reconnect = useCallback(() => {
    if (statusRef.current === "incident_closed") {
      return;
    }
    reconnectSuppressedRef.current = false;
    socketRef.current?.close();
    socketRef.current = null;
    setConnectionId(null);
    connectRef.current();
  }, []);

  useEffect(() => {
    if (incidentId.trim() === "" || typeof WebSocket === "undefined") {
      return;
    }
    reconnectSuppressedRef.current = false;
    resumeTokenRef.current = null;
    lastSeenStreamSeqRef.current = 0;
    resetGenerationRef.current = 0;
    setConnectionId(null);
    let disposed = false;
    let reconnectTimer: number | null = null;
    const url = websocketPath(apiBase, `/ws/v1/incidents/${incidentId}`);

    const scheduleReconnect = () => {
      if (
        disposed ||
        reconnectSuppressedRef.current ||
        reconnectTimer !== null
      ) {
        return;
      }
      reconnectTimer = window.setTimeout(() => {
        reconnectTimer = null;
        connectRef.current();
      }, 1000);
    };

    const terminate = (
      nextStatus: "authorization_lost" | "incident_closed",
      event: IncidentCollaborationEvent,
    ) => {
      reconnectSuppressedRef.current = true;
      resumeTokenRef.current = null;
      setConnectionId(null);
      updateStatus(nextStatus);
      emit(event);
      socketRef.current?.close();
    };

    const beginReset = (
      reason: Extract<
        IncidentCollaborationEvent,
        { kind: "reset_required" }
      >["reason"],
    ) => {
      resetGenerationRef.current += 1;
      updateStatus("resetting");
      emit({
        generation: resetGenerationRef.current,
        kind: "reset_required",
        reason,
      });
    };

    const handleMessage = (socket: WebSocket, raw: unknown) => {
      const decoded = incidentStreamMessageDecoder.decode(raw);
      if (!decoded.ok) {
        return;
      }
      const message = decoded.value;
      if (message.type === "ping") {
        socket.send(JSON.stringify({ type: "pong", payload: {} }));
        return;
      }
      if (message.type === "hello_ack" || message.type === "resume_ack") {
        const payload = recordValue(message.payload);
        const resumeToken = payload.resume_token;
        if (typeof resumeToken === "string") {
          resumeTokenRef.current = resumeToken;
        }
        const serverHighWater = payload.server_high_water_stream_seq;
        if (typeof serverHighWater === "number") {
          lastSeenStreamSeqRef.current = Math.max(
            lastSeenStreamSeqRef.current,
            serverHighWater,
          );
        }
        const resetRequired =
          message.type === "resume_ack" && payload.status === "reset_required";
        if (typeof payload.connection_id === "string") {
          setConnectionId(payload.connection_id);
        }
        updateStatus(resetRequired ? "resetting" : "connected");
        emit({
          kind: "established",
          messageType: message.type,
          payload: {
            ...(typeof payload.connection_id === "string"
              ? { connection_id: payload.connection_id }
              : {}),
            ...(typeof serverHighWater === "number"
              ? { server_high_water_stream_seq: serverHighWater }
              : {}),
            ...(typeof payload.status === "string"
              ? { status: payload.status }
              : {}),
          },
        });
        if (resetRequired) {
          beginReset("resume_reset");
        }
        return;
      }
      if (message.type === "session_revoked") {
        reconnectSuppressedRef.current = true;
        resumeTokenRef.current = null;
        setConnectionId(null);
        updateStatus("authorization_lost");
        emit({ kind: "session_revoked" });
        socket.close();
        return;
      }
      if (message.type === "error") {
        const payload = recordValue(message.payload);
        if (payload.code === "incident_closed") {
          terminate("incident_closed", { kind: "incident_closed" });
          return;
        }
      }
      if (replayable(message)) {
        const sequence = streamSequence(message);
        if (sequence !== null) {
          if (sequence <= lastSeenStreamSeqRef.current) {
            return;
          }
          const gap =
            lastSeenStreamSeqRef.current > 0 &&
            sequence > lastSeenStreamSeqRef.current + 1;
          lastSeenStreamSeqRef.current = sequence;
          if (gap) {
            beginReset("sequence_gap");
            return;
          }
        }
        if (statusRef.current === "resetting") {
          return;
        }
      }
      emit({ kind: "message", message });
    };

    const connect = () => {
      if (disposed || reconnectSuppressedRef.current) {
        return;
      }
      const socket = new WebSocket(url);
      socketRef.current = socket;
      updateStatus("connecting");
      socket.onopen = () => {
        if (socketRef.current !== socket) {
          return;
        }
        const resumeToken = resumeTokenRef.current;
        socket.send(
          JSON.stringify({
            type: resumeToken ? "resume" : "hello",
            payload: {
              client_instance_id: clientInstanceId,
              presence: presenceRef.current,
              ...(resumeToken
                ? {
                    resume_token: resumeToken,
                    last_seen_stream_seq: lastSeenStreamSeqRef.current,
                  }
                : null),
            },
          }),
        );
      };
      socket.onmessage = (event) => {
        if (socketRef.current !== socket || typeof event.data !== "string") {
          return;
        }
        try {
          handleMessage(socket, JSON.parse(event.data) as unknown);
        } catch {
          return;
        }
      };
      socket.onclose = (event) => {
        if (socketRef.current === socket) {
          socketRef.current = null;
        }
        if (
          event.code === 1008 &&
          (event.reason === "session_revoked" ||
            event.reason === "authorization_denied")
        ) {
          terminate("authorization_lost", { kind: "authorization_lost" });
          return;
        }
        if (!reconnectSuppressedRef.current) {
          updateStatus("disconnected");
          scheduleReconnect();
        }
      };
      socket.onerror = () => {
        socket.close();
      };
    };
    connectRef.current = connect;
    connect();

    return () => {
      disposed = true;
      reconnectSuppressedRef.current = true;
      if (reconnectTimer !== null) {
        window.clearTimeout(reconnectTimer);
      }
      socketRef.current?.close();
      socketRef.current = null;
      setConnectionId(null);
      connectRef.current = () => undefined;
    };
  }, [apiBase, clientInstanceId, emit, incidentId, updateStatus]);

  const value = useMemo<IncidentCollaborationSessionValue>(
    () => ({
      clientInstanceId,
      completeReset,
      connectionId,
      disconnect,
      publishPresence,
      reconnect,
      status,
      subscribe,
    }),
    [
      clientInstanceId,
      completeReset,
      connectionId,
      disconnect,
      publishPresence,
      reconnect,
      status,
      subscribe,
    ],
  );

  return (
    <IncidentCollaborationContext.Provider value={value}>
      {children}
    </IncidentCollaborationContext.Provider>
  );
}

export function IncidentCollaborationBoundary({
  apiBase,
  children,
  incidentId,
  initialPresence,
}: {
  readonly apiBase?: string | undefined;
  readonly children: ReactNode;
  readonly incidentId: string;
  readonly initialPresence: CollaborationPresence;
}) {
  const existing = useContext(IncidentCollaborationContext);
  if (existing !== null) {
    return children;
  }
  return (
    <IncidentCollaborationSession
      apiBase={apiBase}
      incidentId={incidentId}
      initialPresence={initialPresence}
    >
      {children}
    </IncidentCollaborationSession>
  );
}

export function useIncidentCollaborationSession() {
  const session = useContext(IncidentCollaborationContext);
  if (session === null) {
    throw new Error(
      "Incident collaboration consumers must be mounted inside IncidentCollaborationSession",
    );
  }
  return session;
}
