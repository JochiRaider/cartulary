import {
  type IncidentStreamMessage,
  incidentStreamMessageDecoder,
} from "@cartulary/protocol-ts/collaboration";
import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { SheetRef } from "../shared/sheetRef";
import {
  type IncidentAuthorizationRevocation,
  type IncidentCollaborationSessionMessagePlan,
  planIncidentCollaborationSessionMessage,
} from "./incidentCollaborationSessionPlan";

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
  | IncidentAuthorizationRevocation
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

  useLayoutEffect(() => {
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
    const terminalOutcomes = new WeakMap<
      WebSocket,
      IncidentCollaborationEvent
    >();
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
      socket: WebSocket,
      nextStatus: "authorization_lost" | "incident_closed",
      event: IncidentCollaborationEvent,
    ) => {
      if (
        disposed ||
        socketRef.current !== socket ||
        terminalOutcomes.has(socket)
      )
        return;
      terminalOutcomes.set(socket, event);
      reconnectSuppressedRef.current = true;
      resumeTokenRef.current = null;
      lastSeenStreamSeqRef.current = 0;
      resetGenerationRef.current += 1;
      if (reconnectTimer !== null) {
        window.clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
      setConnectionId(null);
      updateStatus(nextStatus);
      emit(event);
      socket.close();
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

    type EstablishmentPlan = Extract<
      IncidentCollaborationSessionMessagePlan,
      { readonly kind: "established" }
    >;
    const applyEstablishmentPlan = (plan: EstablishmentPlan) => {
      resumeTokenRef.current = plan.resumeToken;
      if (plan.connectionId !== null) setConnectionId(plan.connectionId);
      updateStatus(plan.resetRequired ? "resetting" : "connected");
      emit({
        kind: "established",
        messageType: plan.messageType,
        payload: {
          ...(plan.connectionId === null
            ? {}
            : { connection_id: plan.connectionId }),
          ...(plan.serverHighWaterStreamSeq === null
            ? {}
            : {
                server_high_water_stream_seq: plan.serverHighWaterStreamSeq,
              }),
          ...(plan.status === null ? {} : { status: plan.status }),
        },
      });
      if (plan.resetRequired) beginReset("resume_reset");
    };

    const applyMessagePlan = (
      socket: WebSocket,
      plan: IncidentCollaborationSessionMessagePlan,
    ) => {
      lastSeenStreamSeqRef.current = plan.nextStreamSeq;
      if (plan.kind === "ignore") return;
      if (plan.kind === "pong") {
        socket.send(JSON.stringify({ type: "pong", payload: {} }));
        return;
      }
      if (plan.kind === "established") {
        applyEstablishmentPlan(plan);
        return;
      }
      if (plan.kind === "terminate") {
        terminate(
          socket,
          plan.event.kind === "incident_closed"
            ? "incident_closed"
            : "authorization_lost",
          plan.event,
        );
        return;
      }
      if (plan.kind === "reset") {
        beginReset("sequence_gap");
        return;
      }
      emit({ kind: "message", message: plan.message });
    };

    const handleMessage = (socket: WebSocket, raw: unknown) => {
      const decoded = incidentStreamMessageDecoder.decode(raw);
      if (!decoded.ok) {
        if (
          raw !== null &&
          typeof raw === "object" &&
          "type" in raw &&
          raw.type === "session_revoked" &&
          "incident_id" in raw &&
          raw.incident_id === incidentId
        ) {
          terminate(socket, "authorization_lost", {
            kind: "authorization_lost",
          });
        }
        return;
      }
      if (decoded.value.incident_id !== incidentId) return;
      applyMessagePlan(
        socket,
        planIncidentCollaborationSessionMessage({
          lastSeenStreamSeq: lastSeenStreamSeqRef.current,
          message: decoded.value,
          resetting: statusRef.current === "resetting",
        }),
      );
    };

    const connect = () => {
      if (disposed || reconnectSuppressedRef.current) {
        return;
      }
      const socket = new WebSocket(url);
      socketRef.current = socket;
      updateStatus("connecting");
      socket.onopen = () => {
        if (
          disposed ||
          socketRef.current !== socket ||
          terminalOutcomes.has(socket)
        ) {
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
        if (
          disposed ||
          socketRef.current !== socket ||
          terminalOutcomes.has(socket) ||
          typeof event.data !== "string"
        ) {
          return;
        }
        try {
          handleMessage(socket, JSON.parse(event.data) as unknown);
        } catch {
          return;
        }
      };
      socket.onclose = (event) => {
        if (disposed || socketRef.current !== socket) return;
        if (terminalOutcomes.has(socket)) {
          socketRef.current = null;
          return;
        }
        if (event.code === 1008 && event.reason !== "heartbeat_timeout") {
          terminate(socket, "authorization_lost", {
            kind: "authorization_lost",
          });
          socketRef.current = null;
          return;
        }
        socketRef.current = null;
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
