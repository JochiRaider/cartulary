import { performance } from "node:perf_hooks";
import type { Page, WebSocket } from "@playwright/test";

export type SocketMessage = {
  payload: Record<string, unknown>;
  receivedAtMs: number;
  socketIndex: number;
  type: string;
};

export type AcceptedSocket = {
  message: SocketMessage;
  socketIndex: number;
};

export function installIncidentSocketMonitor(page: Page, incidentId: string) {
  const messages: SocketMessage[] = [];
  const sentMessages: SocketMessage[] = [];
  const closes: Array<{ closedAtMs: number; socketIndex: number }> = [];
  const sockets: WebSocket[] = [];
  const messageWaiters: Array<{
    matches: (message: SocketMessage) => boolean;
    reject: (error: Error) => void;
    resolve: (message: SocketMessage) => void;
    timeout: ReturnType<typeof setTimeout>;
  }> = [];
  const closeWaiters: Array<{
    matches: (socketIndex: number) => boolean;
    reject: (error: Error) => void;
    resolve: (socketIndex: number) => void;
    timeout: ReturnType<typeof setTimeout>;
  }> = [];

  page.on("websocket", (socket) => {
    if (!socket.url().includes(`/ws/v1/incidents/${incidentId}`)) {
      return;
    }
    const socketIndex = sockets.length;
    sockets.push(socket);
    socket.on("framesent", ({ payload }) => {
      const message = decodeIncidentSocketFrame(payload, socketIndex);
      if (message) sentMessages.push(message);
    });
    socket.on("framereceived", ({ payload }) => {
      const message = decodeIncidentSocketFrame(payload, socketIndex);
      if (!message) return;
      messages.push(message);
      for (const waiter of [...messageWaiters]) {
        if (!waiter.matches(message)) continue;
        clearTimeout(waiter.timeout);
        messageWaiters.splice(messageWaiters.indexOf(waiter), 1);
        waiter.resolve(message);
      }
    });
    socket.on("close", () => {
      closes.push({ closedAtMs: performance.now(), socketIndex });
      for (const waiter of [...closeWaiters]) {
        if (!waiter.matches(socketIndex)) continue;
        clearTimeout(waiter.timeout);
        closeWaiters.splice(closeWaiters.indexOf(waiter), 1);
        waiter.resolve(socketIndex);
      }
    });
  });

  return {
    messageCount: () => messages.length,
    receivedMessages: () => [...messages],
    sentMessages: () => [...sentMessages],
    socketCount: () => sockets.length,
    latestEstablishedSocket: () => latestEstablishedSocket(),
    waitForAcceptedSocket: (
      options: { startAt?: number; timeoutMs?: number } = {},
    ) =>
      waitForMessage("hello_ack", options).then((message) => ({
        message,
        socketIndex: message.socketIndex,
      })),
    waitForClose: (socketIndex: number, timeoutMs = 10_000) => {
      if (closes.some((closed) => closed.socketIndex === socketIndex)) {
        return Promise.resolve(socketIndex);
      }
      return new Promise<number>((resolve, reject) => {
        const waiter = {
          matches: (candidate: number) => candidate === socketIndex,
          reject,
          resolve,
          timeout: setTimeout(() => {
            closeWaiters.splice(closeWaiters.indexOf(waiter), 1);
            reject(
              new Error(
                `timed out waiting for socket ${socketIndex} close; ${describeSocketMonitorState(
                  { closes, messages, sentMessages, sockets },
                )}`,
              ),
            );
          }, timeoutMs),
        };
        closeWaiters.push(waiter);
      });
    },
    waitForMessageOnSocket: (
      type: string,
      socketIndex: number,
      options: {
        matches?: (message: SocketMessage) => boolean;
        startAt?: number;
        timeoutMs?: number;
      } = {},
    ) =>
      waitForMessage(type, {
        ...options,
        matches: (message) =>
          message.socketIndex === socketIndex &&
          (options.matches?.(message) ?? true),
      }),
    waitForMessage: (
      type: string,
      options: {
        matches?: (message: SocketMessage) => boolean;
        startAt?: number;
        timeoutMs?: number;
      } = {},
    ) => waitForMessage(type, options),
    waitForMessageWhere: (
      label: string,
      options: {
        matches: (message: SocketMessage) => boolean;
        startAt?: number;
        timeoutMs?: number;
      },
    ) => waitForMatchingMessage(label, options),
  };

  function waitForMessage(
    type: string,
    options: {
      matches?: (message: SocketMessage) => boolean;
      startAt?: number;
      timeoutMs?: number;
    } = {},
  ) {
    return waitForMatchingMessage(`socket message ${type}`, {
      ...options,
      matches: (message) =>
        message.type === type && (options.matches?.(message) ?? true),
    });
  }

  function waitForMatchingMessage(
    label: string,
    options: {
      matches: (message: SocketMessage) => boolean;
      startAt?: number;
      timeoutMs?: number;
    },
  ) {
    const startAt = options.startAt ?? 0;
    const existing = messages.slice(startAt).find(options.matches);
    if (existing) return Promise.resolve(existing);
    return new Promise<SocketMessage>((resolve, reject) => {
      const waiter = {
        matches: (message: SocketMessage) =>
          messages.indexOf(message) >= startAt && options.matches(message),
        reject,
        resolve,
        timeout: setTimeout(() => {
          messageWaiters.splice(messageWaiters.indexOf(waiter), 1);
          reject(
            new Error(
              `timed out waiting for ${label}; ${describeSocketMonitorState({
                closes,
                messages,
                sentMessages,
                sockets,
              })}`,
            ),
          );
        }, options.timeoutMs ?? 10_000),
      };
      messageWaiters.push(waiter);
    });
  }

  function latestEstablishedSocket(): AcceptedSocket | null {
    for (let index = messages.length - 1; index >= 0; index -= 1) {
      const message = messages[index];
      if (
        message &&
        (message.type === "hello_ack" || message.type === "resume_ack")
      ) {
        return { message, socketIndex: message.socketIndex };
      }
    }
    return null;
  }
}

export function decodeIncidentSocketFrame(
  payload: string | Buffer,
  socketIndex: number,
  now: () => number = performance.now.bind(performance),
): SocketMessage | null {
  const text = Buffer.isBuffer(payload) ? payload.toString("utf8") : payload;
  try {
    const parsed = JSON.parse(text) as {
      payload?: Record<string, unknown>;
      type?: unknown;
    };
    if (typeof parsed.type !== "string") return null;
    return {
      payload: parsed.payload ?? {},
      receivedAtMs: now(),
      socketIndex,
      type: parsed.type,
    };
  } catch {
    return null;
  }
}

function describeSocketMonitorState({
  closes,
  messages,
  sentMessages,
  sockets,
}: {
  closes: Array<{ closedAtMs: number; socketIndex: number }>;
  messages: readonly SocketMessage[];
  sentMessages: readonly SocketMessage[];
  sockets: readonly WebSocket[];
}) {
  const received = messages
    .map((message) => `${message.socketIndex}:${message.type}`)
    .join(", ");
  const sent = sentMessages
    .map((message) => `${message.socketIndex}:${message.type}`)
    .join(", ");
  const closed = closes
    .map((close) => `${close.socketIndex}@${Math.round(close.closedAtMs)}ms`)
    .join(", ");
  const presence = messages
    .filter(
      (message) =>
        message.type === "presence_delta" ||
        message.type === "presence_snapshot",
    )
    .slice(-8)
    .map((message) => `${message.socketIndex}:${message.type}`)
    .join("; ");
  return `sockets=${sockets.length}; received=[${received}]; sent=[${sent}]; closes=[${closed}]; presence=[${presence}]`;
}
