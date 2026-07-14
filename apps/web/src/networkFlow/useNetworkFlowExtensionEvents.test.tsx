import { cleanup, render } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { IncidentCollaborationSession } from "../collaboration/IncidentCollaborationSession";
import { useNetworkFlowExtensionEvents } from "./useNetworkFlowExtensionEvents";

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

function NetworkFlowConsumer({ onChange }: { readonly onChange: () => void }) {
  useNetworkFlowExtensionEvents({
    enabled: true,
    incidentId: "incident-1",
    onResourceChange: onChange,
  });
  return null;
}

function Harness({ onChange }: { readonly onChange: () => void }) {
  return (
    <IncidentCollaborationSession
      incidentId="incident-1"
      initialPresence={{
        sheet_ref: {
          kind: "extension_workspace",
          extension_profile_id: "network_flow_activity",
          workspace_key: "network_analysis",
        },
        mode: "viewing",
      }}
    >
      <NetworkFlowConsumer onChange={onChange} />
    </IncidentCollaborationSession>
  );
}

describe("useNetworkFlowExtensionEvents", () => {
  afterEach(() => {
    cleanup();
    vi.useRealTimers();
    vi.unstubAllGlobals();
    FakeWebSocket.instances = [];
  });

  it("deduplicates changes and resumes after reconnect", () => {
    vi.useFakeTimers();
    vi.stubGlobal("WebSocket", FakeWebSocket);
    const onChange = vi.fn();
    render(<Harness onChange={onChange} />);

    const first = FakeWebSocket.instances[0];
    expect(first).toBeDefined();
    if (first) {
      first.readyState = FakeWebSocket.OPEN;
    }
    first?.onopen?.();
    expect(JSON.parse(first?.sent[0] ?? "{}").type).toBe("hello");
    first?.onmessage?.({
      data: JSON.stringify({
        type: "hello_ack",
        payload: {
          connection_id: "connection-1",
          resume_token: "resume-1",
          server_time: "2026-07-13T12:00:00Z",
          heartbeat_interval_ms: 15_000,
          presence_ttl_ms: 45_000,
          resume_window_ms: 60_000,
          server_high_water_stream_seq: 7,
        },
      }),
    });
    const change = JSON.stringify({
      type: "extension_resource_changed",
      stream_seq: 8,
      payload: {
        extension_profile_id: "network_flow_activity",
        resource_kind: "network_flow_table",
        resource_id: "nft_a",
        change_kind: "remove",
        reason_code: "soft_deleted",
      },
    });
    first?.onmessage?.({ data: change });
    first?.onmessage?.({ data: change });
    expect(onChange).toHaveBeenCalledTimes(1);

    first?.onclose?.({ code: 1006, reason: "connection_lost" });
    vi.advanceTimersByTime(1000);
    const second = FakeWebSocket.instances[1];
    if (second) {
      second.readyState = FakeWebSocket.OPEN;
    }
    second?.onopen?.();
    expect(JSON.parse(second?.sent[0] ?? "{}")).toMatchObject({
      type: "resume",
      payload: { resume_token: "resume-1", last_seen_stream_seq: 8 },
    });
  });
});
