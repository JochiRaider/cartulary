import type { IncidentStreamMessage } from "@cartulary/protocol-ts/collaboration";

export type IncidentCollaborationSessionMessagePlan =
  | { readonly kind: "ignore"; readonly nextStreamSeq: number }
  | { readonly kind: "pong"; readonly nextStreamSeq: number }
  | {
      readonly connectionId: string | null;
      readonly kind: "established";
      readonly messageType: "hello_ack" | "resume_ack";
      readonly nextStreamSeq: number;
      readonly resetRequired: boolean;
      readonly resumeToken: string;
      readonly serverHighWaterStreamSeq: number | null;
      readonly status: string | null;
    }
  | {
      readonly kind: "terminate";
      readonly nextStreamSeq: number;
      readonly reason: "session_revoked" | "incident_closed";
    }
  | {
      readonly kind: "reset";
      readonly message: IncidentStreamMessage;
      readonly nextStreamSeq: number;
    }
  | {
      readonly kind: "message";
      readonly message: IncidentStreamMessage;
      readonly nextStreamSeq: number;
    };

function replaySequence(message: IncidentStreamMessage): number | null {
  return message.type === "record_changed" ||
    message.type === "extension_resource_changed" ||
    message.type === "job_progress"
    ? message.stream_seq
    : null;
}

export function planIncidentCollaborationSessionMessage(input: {
  readonly lastSeenStreamSeq: number;
  readonly message: IncidentStreamMessage;
  readonly resetting: boolean;
}): IncidentCollaborationSessionMessagePlan {
  const message = input.message;
  if (message.type === "ping") {
    return { kind: "pong", nextStreamSeq: input.lastSeenStreamSeq };
  }
  if (message.type === "hello_ack" || message.type === "resume_ack") {
    const highWater =
      message.type === "resume_ack"
        ? message.payload.server_high_water_stream_seq
        : undefined;
    const status =
      message.type === "resume_ack" ? message.payload.status : null;
    return {
      connectionId:
        message.type === "hello_ack" ? message.payload.connection_id : null,
      kind: "established",
      messageType: message.type,
      nextStreamSeq:
        highWater === undefined
          ? input.lastSeenStreamSeq
          : Math.max(input.lastSeenStreamSeq, highWater),
      resetRequired: status === "reset_required",
      resumeToken: message.payload.resume_token,
      serverHighWaterStreamSeq: highWater ?? null,
      status,
    };
  }
  if (message.type === "session_revoked") {
    return {
      kind: "terminate",
      nextStreamSeq: input.lastSeenStreamSeq,
      reason: "session_revoked",
    };
  }
  if (message.type === "error" && message.payload.code === "incident_closed") {
    return {
      kind: "terminate",
      nextStreamSeq: input.lastSeenStreamSeq,
      reason: "incident_closed",
    };
  }
  const sequence = replaySequence(message);
  if (sequence === null) {
    return { kind: "message", message, nextStreamSeq: input.lastSeenStreamSeq };
  }
  if (sequence <= input.lastSeenStreamSeq) {
    return { kind: "ignore", nextStreamSeq: input.lastSeenStreamSeq };
  }
  const gap =
    input.lastSeenStreamSeq > 0 && sequence > input.lastSeenStreamSeq + 1;
  if (gap) return { kind: "reset", message, nextStreamSeq: sequence };
  return input.resetting
    ? { kind: "ignore", nextStreamSeq: sequence }
    : { kind: "message", message, nextStreamSeq: sequence };
}
