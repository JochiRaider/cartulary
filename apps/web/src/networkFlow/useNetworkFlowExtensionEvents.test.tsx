import { cleanup, render } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useNetworkFlowExtensionEvents } from "./useNetworkFlowExtensionEvents";

class FakeWebSocket {
  static instances: FakeWebSocket[] = [];
  readonly sent: string[] = [];
  onopen: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onclose: (() => void) | null = null;
  onerror: (() => void) | null = null;

  constructor(readonly url: string) {
    FakeWebSocket.instances.push(this);
  }

  send(payload: string) {
    this.sent.push(payload);
  }

  close() {}
}

function Harness({ onChange }: { readonly onChange: () => void }) {
  useNetworkFlowExtensionEvents({
    enabled: true,
    incidentId: "incident-1",
    onResourceChange: onChange,
  });
  return null;
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
    first?.onopen?.();
    expect(JSON.parse(first?.sent[0] ?? "{}").type).toBe("hello");
    first?.onmessage?.({
      data: JSON.stringify({
        type: "hello_ack",
        payload: { resume_token: "resume-1", server_high_water_stream_seq: 7 },
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

    first?.onclose?.();
    vi.advanceTimersByTime(1000);
    const second = FakeWebSocket.instances[1];
    second?.onopen?.();
    expect(JSON.parse(second?.sent[0] ?? "{}")).toMatchObject({
      type: "resume",
      payload: { resume_token: "resume-1", last_seen_stream_seq: 8 },
    });
  });
});
