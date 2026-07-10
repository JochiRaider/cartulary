import { useEffect } from "react";
import { clientTxnID } from "../services/browserApi";
import {
  networkAnalysisSheetRef,
  networkFlowActivityProfileId,
} from "./networkFlowClient";

export type NetworkFlowExtensionResourceChange = {
  readonly changeKind: "invalidate" | "remove";
  readonly reasonCode: string;
  readonly resourceId: string;
};

type ExtensionResourceMessage = {
  readonly type?: string;
  readonly stream_seq?: number;
  readonly payload?: {
    readonly extension_profile_id?: string;
    readonly resource_kind?: string;
    readonly resource_id?: string;
    readonly change_kind?: string;
    readonly reason_code?: string;
  };
};

type NetworkFlowTableChangeMessage = {
  readonly type: "extension_resource_changed";
  readonly stream_seq?: number;
  readonly payload: {
    readonly extension_profile_id: typeof networkFlowActivityProfileId;
    readonly resource_kind: "network_flow_table";
    readonly resource_id: string;
    readonly change_kind: "invalidate" | "remove";
    readonly reason_code: string;
  };
};

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

function isNetworkFlowTableChange(
  message: ExtensionResourceMessage,
): message is NetworkFlowTableChangeMessage {
  const payload = message.payload;
  return (
    message.type === "extension_resource_changed" &&
    payload?.extension_profile_id === networkFlowActivityProfileId &&
    payload.resource_kind === "network_flow_table" &&
    typeof payload.resource_id === "string" &&
    (payload.change_kind === "invalidate" ||
      payload.change_kind === "remove") &&
    typeof payload.reason_code === "string"
  );
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
      if (!isNetworkFlowTableChange(message)) {
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
      onResourceChange({
        changeKind: message.payload.change_kind,
        reasonCode: message.payload.reason_code,
        resourceId: message.payload.resource_id,
      });
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
