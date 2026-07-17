// @vitest-environment node

import { describe, expect, it } from "vitest";

import {
  decodeIncidentSocketFrame,
  installIncidentSocketMonitor,
} from "./incidentSocket";

describe("incident socket transport", () => {
  it("decodes typed protocol envelopes without accepting malformed frames", () => {
    expect(
      decodeIncidentSocketFrame(
        JSON.stringify({ type: "presence_delta", payload: { mode: "edit" } }),
        2,
        () => 17,
      ),
    ).toEqual({
      payload: { mode: "edit" },
      receivedAtMs: 17,
      socketIndex: 2,
      type: "presence_delta",
    });
    expect(decodeIncidentSocketFrame("not-json", 0)).toBeNull();
    expect(
      decodeIncidentSocketFrame(JSON.stringify({ payload: {} }), 0),
    ).toBeNull();
  });

  it("retains unrelated ordered events while satisfying a semantic wait", async () => {
    let onSocket: ((socket: FakeSocket) => void) | undefined;
    const page = {
      on: (_event: string, listener: (socket: FakeSocket) => void) => {
        onSocket = listener;
      },
    };
    const monitor = installIncidentSocketMonitor(page as never, "incident-1");
    const socket = new FakeSocket();
    onSocket?.(socket);
    socket.emit("framereceived", {
      payload: JSON.stringify({ type: "unrelated", payload: { sequence: 1 } }),
    });
    socket.emit("framereceived", {
      payload: JSON.stringify({ type: "hello_ack", payload: { sequence: 2 } }),
    });

    await expect(monitor.waitForMessage("hello_ack")).resolves.toMatchObject({
      payload: { sequence: 2 },
    });
    expect(monitor.receivedMessages().map((message) => message.type)).toEqual([
      "unrelated",
      "hello_ack",
    ]);
  });
});

class FakeSocket {
  readonly #listeners = new Map<string, Array<(event: never) => void>>();

  url() {
    return "ws://127.0.0.1/ws/v1/incidents/incident-1";
  }

  on(event: string, listener: (event: never) => void) {
    const listeners = this.#listeners.get(event) ?? [];
    listeners.push(listener);
    this.#listeners.set(event, listeners);
  }

  emit(event: string, value: unknown) {
    for (const listener of this.#listeners.get(event) ?? []) {
      listener(value as never);
    }
  }
}
