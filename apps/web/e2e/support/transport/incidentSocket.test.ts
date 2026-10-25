// @vitest-environment node

import type { BrowserContext } from "@playwright/test";
import { describe, expect, it } from "vitest";
import { OwnedBrowserScope } from "../runtime/cleanup";

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
      context: () => context as unknown as BrowserContext,
      unrouteAll: async () => {},
      on: (_event: string, listener: (socket: FakeSocket) => void) => {
        onSocket = listener;
      },
      off: () => {
        onSocket = undefined;
      },
    };
    const context = {
      pages: () => [page],
      request: {},
      unrouteAll: async () => {},
      close: async () => {},
    };
    const scope = new OwnedBrowserScope({
      browser: { newContext: async () => context as unknown as BrowserContext },
      revokeAllSessions: async () => {},
      verifyRevokedSession: async () => {},
    });
    await scope.newContext();
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
    const pending = expect(monitor.waitForMessage("never")).rejects.toThrow(
      "Owned socket observation cancelled",
    );
    const closing = expect(monitor.waitForClose(0)).rejects.toThrow(
      "Owned socket observation cancelled",
    );
    await scope.dispose();
    await Promise.all([pending, closing]);
    expect(onSocket).toBeUndefined();
    socket.emit("framereceived", {
      payload: JSON.stringify({ type: "after-disposal" }),
    });
    expect(monitor.messageCount()).toBe(2);
    expect(() => monitor.waitForMessage("late")).toThrow(
      "observation has stopped",
    );
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

  off(event: string, listener: (event: never) => void) {
    this.#listeners.set(
      event,
      (this.#listeners.get(event) ?? []).filter(
        (candidate) => candidate !== listener,
      ),
    );
  }

  emit(event: string, value: unknown) {
    for (const listener of this.#listeners.get(event) ?? []) {
      listener(value as never);
    }
  }
}
