import { useEffect } from "react";
import { clientTxnID } from "../services/browserApi";
import { networkAnalysisSheetRef } from "./networkFlowClient";
import {
  type ExtensionResourceMessage,
  interpretNetworkFlowCollaborationMessage,
  type NetworkFlowExtensionResourceChange,
} from "./networkFlowCollaborationInterpreter";

export type { NetworkFlowExtensionResourceChange } from "./networkFlowCollaborationInterpreter";

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

export function useNetworkFlowExtensionEvents({
  apiBase,
  enabled,
  incidentId,
  onResourceChange,
}: {
  readonly apiBase?: string | undefined;
  readonly enabled: boolean;
  readonly incidentId: string;
  readonly onResourceChange: (
    change: NetworkFlowExtensionResourceChange,
  ) => void;
}) {
  useEffect(() => {
    if (
      !enabled ||
      incidentId.trim() === "" ||
      typeof WebSocket === "undefined"
    ) {
      return;
    }

    let closed = false;
    let socket: WebSocket | null = null;
    let reconnectTimer: number | null = null;
    const appliedStreamSeqs = new Set<number>();
    let lastSeenStreamSeq = 0;
    let resumeToken: string | null = null;
    const clientInstanceId = clientTxnID("nf-ws-client");
    const url = websocketPath(apiBase, `/ws/v1/incidents/${incidentId}`);

    const sendHello = (target: WebSocket) => {
      target.send(
        JSON.stringify({
          type: resumeToken ? "resume" : "hello",
          payload: {
            client_instance_id: clientInstanceId,
            presence: {
              sheet_ref: networkAnalysisSheetRef(),
              mode: "viewing",
            },
            ...(resumeToken
              ? {
                  resume_token: resumeToken,
                  last_seen_stream_seq: lastSeenStreamSeq,
                }
              : null),
          },
        }),
      );
    };

    const scheduleReconnect = () => {
      if (closed || reconnectTimer !== null) {
        return;
      }
      reconnectTimer = window.setTimeout(() => {
        reconnectTimer = null;
        connect();
      }, 1000);
    };

    const handleMessage = (target: WebSocket, raw: unknown) => {
      if (!raw || typeof raw !== "object") {
        return;
      }
      const message = raw as ExtensionResourceMessage & {
        readonly payload?: Record<string, unknown>;
      };
      if (message.type === "ping") {
        target.send(JSON.stringify({ type: "pong", payload: {} }));
        return;
      }
      if (message.type === "hello_ack" || message.type === "resume_ack") {
        const nextResumeToken = message.payload?.resume_token;
        if (typeof nextResumeToken === "string") {
          resumeToken = nextResumeToken;
        }
        const serverHighWater = message.payload?.server_high_water_stream_seq;
        if (typeof serverHighWater === "number") {
          lastSeenStreamSeq = Math.max(lastSeenStreamSeq, serverHighWater);
        }
        return;
      }
      const change = interpretNetworkFlowCollaborationMessage(message);
      if (change === null) {
        return;
      }
      const streamSeq = message.stream_seq;
      if (typeof streamSeq === "number") {
        if (appliedStreamSeqs.has(streamSeq)) {
          return;
        }
        appliedStreamSeqs.add(streamSeq);
        lastSeenStreamSeq = Math.max(lastSeenStreamSeq, streamSeq);
      }
      onResourceChange(change);
    };

    const connect = () => {
      if (closed) {
        return;
      }
      socket = new WebSocket(url);
      socket.onopen = () => {
        if (socket) {
          sendHello(socket);
        }
      };
      socket.onmessage = (event) => {
        if (!socket) {
          return;
        }
        if (typeof event.data !== "string") {
          return;
        }
        try {
          handleMessage(socket, JSON.parse(event.data) as unknown);
        } catch {
          return;
        }
      };
      socket.onclose = () => {
        scheduleReconnect();
      };
      socket.onerror = () => {
        socket?.close();
      };
    };

    connect();

    return () => {
      closed = true;
      if (reconnectTimer !== null) {
        window.clearTimeout(reconnectTimer);
      }
      socket?.close();
    };
  }, [apiBase, enabled, incidentId, onResourceChange]);
}
