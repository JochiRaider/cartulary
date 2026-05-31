import { performance } from "node:perf_hooks";

import {
  conflictMarkerTestId,
  currentIncidentRoleTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  rowCellTestId,
  saveStateTestId,
} from "@cartulary/ui-contracts";
import type { Browser, Page, Route, WebSocket } from "@playwright/test";
import { expect } from "./fixtures";
import {
  apiBase,
  createIncident,
  createIncidentMemberUser,
  createViewRow,
  csrfHeaders,
  queryViewRows,
  safeUnroute,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";

export const timelineViewSchemaId = "cartulary.view.timeline.v1";

export type PatchCall = {
  body: Record<string, unknown>;
  recordId: string;
  status: number;
};

type PatchBehavior =
  | {
      hold: Promise<void>;
      recordId?: string;
      release: () => void;
      resolveCompletion: (call: PatchCall) => void;
      resolveHit: (call: PatchCall) => void;
      type: "hold";
      waitForCompletion: Promise<PatchCall>;
      waitForHit: Promise<PatchCall>;
    }
  | {
      code: string;
      recordId?: string;
      status: number;
      type: "error";
    };

type PatchBehaviorOptions = {
  recordId?: string;
};

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

export type SessionTracker = {
  loginTrackedUser: (
    page: Page,
    details: {
      createdBy: string;
      email: string;
      password: string;
      purpose: string;
      userId: string;
    },
  ) => Promise<void>;
};

export async function openIncidentAsTrackedUserReady(
  browser: Browser,
  sessionTracker: SessionTracker,
  options: {
    createdBy: string;
    email: string;
    incidentId: string;
    password: string;
    purpose: string;
    readyRecordId: string;
    userId: string;
  },
) {
  const context = await browser.newContext();
  const page = await context.newPage();
  await sessionTracker.loginTrackedUser(page, {
    createdBy: options.createdBy,
    email: options.email,
    password: options.password,
    purpose: options.purpose,
    userId: options.userId,
  });
  const socketMonitor = installIncidentSocketMonitor(page, options.incidentId);
  await page.goto(`/?incident_id=${options.incidentId}`);
  const acceptedSocket = await socketMonitor.waitForAcceptedSocket();
  await expect(
    page.getByTestId(rowCellTestId(options.readyRecordId, "timeline.summary")),
  ).toBeVisible();
  return { acceptedSocket, page, socketMonitor };
}

export type Phase6RecoveryScenario = {
  createdBy: string;
  incidentKeyPrefix: string;
  localValues?: readonly string[];
  page: Page;
  scenario: string;
  sessionTracker: SessionTracker;
  triggerRevocation: (context: {
    incidentId: string;
    member: {
      email: string;
      initial_password: string;
      user_id: string;
    };
  }) => Promise<void>;
};

export async function createTimelineRow(
  page: Page,
  incidentId: string,
  summary: string,
) {
  return createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("phase6-timeline-row"),
    "timeline.summary": summary,
  });
}

export async function editTimelineSummary(
  page: Page,
  recordId: string,
  value: string,
) {
  const input = page.getByTestId(rowCellTestId(recordId, "timeline.summary"));
  await input.fill(value);
  await input.press("Enter");
  await expect(input).toHaveValue(value);
}

export async function patchTimelineField(
  page: Page,
  recordId: string,
  baseRowVersion: number,
  fieldKey: string,
  value: string,
  txnPrefix: string,
) {
  const response = await page.request.patch(
    `${apiBase}/api/v1/records/${recordId}`,
    {
      headers: await csrfHeaders(page),
      data: {
        view_schema_id: timelineViewSchemaId,
        base_row_version: baseRowVersion,
        client_txn_id: uniqueTxn(txnPrefix),
        changes: [
          {
            field_key: fieldKey,
            value,
          },
        ],
      },
    },
  );
  expect(response.ok()).toBeTruthy();
}

export async function driveRealTimelineSummaryConflict({
  afterLocalPatchHeld,
  baseRowVersion,
  expectConflictMarker = true,
  expectEditedCellMounted = true,
  localValue,
  page,
  patchController,
  recordId,
  remotePatchPage,
  remoteValue,
  txnPrefix,
}: {
  afterLocalPatchHeld?: () => Promise<void>;
  baseRowVersion: number;
  expectConflictMarker?: boolean;
  expectEditedCellMounted?: boolean;
  localValue: string;
  page: Page;
  patchController: Awaited<ReturnType<typeof installPatchController>>;
  recordId: string;
  remotePatchPage?: Page;
  remoteValue: string;
  txnPrefix: string;
}) {
  const heldPrimaryPatch = patchController.holdNextPatch({ recordId });
  await editTimelineSummary(page, recordId, localValue);
  await heldPrimaryPatch.waitForHit;
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");

  await afterLocalPatchHeld?.();

  await patchTimelineField(
    remotePatchPage ?? page,
    recordId,
    baseRowVersion,
    "timeline.summary",
    remoteValue,
    txnPrefix,
  );

  heldPrimaryPatch.release();
  const primaryPatch = await heldPrimaryPatch.waitForCompletion;
  expect(primaryPatch.status).toBe(409);
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
  if (expectConflictMarker) {
    await expect(
      page.getByTestId(conflictMarkerTestId(recordId, "timeline.summary")),
    ).toBeVisible();
  }
  await expect(page.getByTestId("conflict-resolver")).toBeVisible();
  if (expectEditedCellMounted) {
    await expect(
      page.getByTestId(rowCellTestId(recordId, "timeline.summary")),
    ).toHaveValue(localValue);
  }
  await expect(page.getByTestId("conflict-server-value")).toHaveValue(
    remoteValue,
  );
  await expect(page.getByTestId("conflict-local-value")).toHaveValue(
    localValue,
  );
}

export async function exerciseRevokedPendingReplay({
  createdBy,
  incidentKeyPrefix,
  localValues,
  page,
  scenario,
  sessionTracker,
  triggerRevocation,
}: Phase6RecoveryScenario) {
  const replayValues = localValues ?? [
    `Phase 6 ${createdBy} ${scenario} first local`,
    `Phase 6 ${createdBy} ${scenario} second local`,
  ];
  if (replayValues.length < 2) {
    throw new Error(
      "revoked pending replay requires at least two local values",
    );
  }
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey(incidentKeyPrefix),
    `Phase 6 ${createdBy} ${scenario} revocation recovery`,
  );
  const member = await createIncidentMemberUser(page, incidentId, {
    display_name: `Phase 6 ${createdBy} ${scenario} Analyst`,
    email: uniqueEmail(`phase6-${createdBy.toLowerCase()}-${scenario}`),
    initial_password: `Phase6${createdBy.replaceAll("-", "")}${scenario}Pass!`,
    role: "editor",
  });
  const recordIds: string[] = [];
  for (const [index] of replayValues.entries()) {
    recordIds.push(
      requireRecordId(
        await createTimelineRow(
          page,
          incidentId,
          `Phase 6 ${createdBy} ${scenario} ${index + 1} base`,
        ),
      ),
    );
  }
  const replayItems = requireReplayItems(recordIds, replayValues);
  const firstReplayItem = replayItems[0];
  if (!firstReplayItem) {
    throw new Error("revoked pending replay did not create a first row");
  }
  const patchController = await installPatchController(page);

  try {
    await sessionTracker.loginTrackedUser(page, {
      createdBy,
      email: member.email,
      password: member.initial_password,
      purpose: `Phase 6 ${createdBy} ${scenario} analyst runtime`,
      userId: member.user_id,
    });

    const socketMonitor = installIncidentSocketMonitor(page, incidentId);
    await page.goto(`/?incident_id=${incidentId}`);
    await socketMonitor.waitForAcceptedSocket();
    await expect(
      page.getByTestId(
        rowCellTestId(firstReplayItem.recordId, "timeline.summary"),
      ),
    ).toHaveValue(`Phase 6 ${createdBy} ${scenario} 1 base`);
    await expect(page.getByTestId(currentIncidentRoleTestId())).toHaveText(
      "Current incident role: editor",
    );

    const heldPatch = patchController.holdNextPatch();
    await editTimelineSummary(
      page,
      firstReplayItem.recordId,
      firstReplayItem.value,
    );
    await heldPatch.waitForHit;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");

    const establishedSocket = socketMonitor.latestEstablishedSocket();
    if (!establishedSocket) {
      throw new Error("revoked pending replay had no established socket");
    }
    await triggerRevocation({ incidentId, member });
    const revocation = await socketMonitor.waitForMessage("session_revoked", {
      timeoutMs: 25_000,
    });
    await socketMonitor.waitForClose(revocation.socketIndex, 25_000);

    heldPatch.release();
    await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();

    for (const item of replayItems.slice(1)) {
      await editTimelineSummary(page, item.recordId, item.value);
    }
    for (const item of replayItems) {
      await expect(
        page.getByTestId(rowCellTestId(item.recordId, "timeline.summary")),
      ).toHaveValue(item.value);
    }
    await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
      String(replayValues.length),
    );

    const messageStart = socketMonitor.messageCount();
    await sessionTracker.loginTrackedUser(page, {
      createdBy,
      email: member.email,
      password: member.initial_password,
      purpose: `Phase 6 ${createdBy} ${scenario} analyst re-authentication`,
      userId: member.user_id,
    });
    await socketMonitor.waitForAcceptedSocket({
      startAt: messageStart,
    });

    await expect
      .poll(() => successfulPatchCalls(patchController.calls).length)
      .toBeGreaterThanOrEqual(replayValues.length);
    const replayed = successfulPatchCalls(patchController.calls).slice(
      -replayValues.length,
    );
    expect(replayed.map((call) => call.recordId)).toEqual(recordIds);
    expect(replayed.map((call) => summaryPatchValue(call.body))).toEqual(
      replayValues,
    );

    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
    await expectServerSummaries(
      page,
      incidentId,
      Object.fromEntries(
        replayItems.map((item) => [item.recordId, item.value]),
      ),
    );
  } finally {
    await patchController.dispose();
  }
}

function requireReplayItems(
  recordIds: readonly string[],
  values: readonly string[],
) {
  return recordIds.map((recordId, index) => {
    const value = values[index];
    if (value === undefined) {
      throw new Error(`missing replay value for record ${recordId}`);
    }
    return { recordId, value };
  });
}

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
      const message = parseSocketPayload(payload, socketIndex);
      if (message) {
        sentMessages.push(message);
      }
    });
    socket.on("framereceived", ({ payload }) => {
      const message = parseSocketPayload(payload, socketIndex);
      if (!message) {
        return;
      }
      messages.push(message);
      for (const waiter of [...messageWaiters]) {
        if (!waiter.matches(message)) {
          continue;
        }
        clearTimeout(waiter.timeout);
        messageWaiters.splice(messageWaiters.indexOf(waiter), 1);
        waiter.resolve(message);
      }
    });
    socket.on("close", () => {
      closes.push({ closedAtMs: performance.now(), socketIndex });
      for (const waiter of [...closeWaiters]) {
        if (!waiter.matches(socketIndex)) {
          continue;
        }
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
      waitForMessage("hello_ack", {
        ...(options.startAt === undefined ? {} : { startAt: options.startAt }),
        ...(options.timeoutMs === undefined
          ? {}
          : { timeoutMs: options.timeoutMs }),
      }).then((message) => ({
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
                  {
                    closes,
                    messages,
                    sentMessages,
                    sockets,
                  },
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
  };

  function waitForMessage(
    type: string,
    options: {
      matches?: (message: SocketMessage) => boolean;
      startAt?: number;
      timeoutMs?: number;
    } = {},
  ) {
    const startAt = options.startAt ?? 0;
    const existing = messages
      .slice(startAt)
      .find(
        (message) =>
          message.type === type && (options.matches?.(message) ?? true),
      );
    if (existing) {
      return Promise.resolve(existing);
    }
    return new Promise<SocketMessage>((resolve, reject) => {
      const waiter = {
        matches: (message: SocketMessage) =>
          messages.indexOf(message) >= startAt &&
          message.type === type &&
          (options.matches?.(message) ?? true),
        reject,
        resolve,
        timeout: setTimeout(() => {
          messageWaiters.splice(messageWaiters.indexOf(waiter), 1);
          reject(
            new Error(
              `timed out waiting for socket message ${type}; ${describeSocketMonitorState(
                {
                  closes,
                  messages,
                  sentMessages,
                  sockets,
                },
              )}`,
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
        return {
          message,
          socketIndex: message.socketIndex,
        };
      }
    }
    return null;
  }
}

export async function installPatchController(page: Page) {
  const calls: PatchCall[] = [];
  const behaviors: PatchBehavior[] = [];
  const routePattern = "**/api/v1/records/*";
  const takeBehavior = (call: PatchCall) => {
    const behaviorIndex = behaviors.findIndex(
      (behavior) =>
        behavior.recordId === undefined || behavior.recordId === call.recordId,
    );
    if (behaviorIndex === -1) {
      return null;
    }
    const [behavior] = behaviors.splice(behaviorIndex, 1);
    return behavior ?? null;
  };
  const handler = async (route: Route) => {
    const request = route.request();
    if (request.method().toUpperCase() !== "PATCH") {
      await route.fallback();
      return;
    }

    const call: PatchCall = {
      body: parseRequestBody(request.postData()),
      recordId: recordIdFromURL(request.url()),
      status: 0,
    };
    const behavior = takeBehavior(call);
    if (behavior?.type === "hold") {
      behavior.resolveHit(call);
      await behavior.hold;
    }
    if (behavior?.type === "error") {
      call.status = behavior.status;
      calls.push(call);
      await fulfillJSONError(route, behavior.status, behavior.code);
      return;
    }

    const response = await route.fetch();
    call.status = response.status();
    calls.push(call);
    await route.fulfill({ response });
    if (behavior?.type === "hold") {
      behavior.resolveCompletion(call);
    }
  };

  await page.route(routePattern, handler);

  return {
    calls,
    dispose: async () => {
      await safeUnroute(page, routePattern, handler);
    },
    failNextPatch: (
      status: number,
      code: string,
      options: PatchBehaviorOptions = {},
    ) => {
      behaviors.push({
        code,
        status,
        type: "error",
        ...(options.recordId === undefined
          ? {}
          : { recordId: options.recordId }),
      });
    },
    holdNextPatch: (options: PatchBehaviorOptions = {}) => {
      let releaseHold!: () => void;
      let resolveCompletion!: (call: PatchCall) => void;
      let resolveHit!: (call: PatchCall) => void;
      const waitForHit = new Promise<PatchCall>((resolve) => {
        resolveHit = resolve;
      });
      const waitForCompletion = new Promise<PatchCall>((resolve) => {
        resolveCompletion = resolve;
      });
      const hold = new Promise<void>((resolve) => {
        releaseHold = resolve;
      });
      const behavior = {
        hold,
        release: releaseHold,
        resolveCompletion,
        resolveHit,
        type: "hold" as const,
        waitForCompletion,
        waitForHit,
        ...(options.recordId === undefined
          ? {}
          : { recordId: options.recordId }),
      };
      behaviors.push(behavior);
      return {
        release: releaseHold,
        waitForCompletion,
        waitForHit,
      };
    },
  };
}

export async function installPatchTransportFailureController(page: Page) {
  const calls: PatchCall[] = [];
  let connected = true;
  const routePattern = "**/api/v1/records/*";
  const handler = async (route: Route) => {
    const request = route.request();
    if (request.method().toUpperCase() !== "PATCH") {
      await route.fallback();
      return;
    }

    const call: PatchCall = {
      body: parseRequestBody(request.postData()),
      recordId: recordIdFromURL(request.url()),
      status: 0,
    };
    if (!connected) {
      calls.push(call);
      await route.abort("internetdisconnected");
      return;
    }

    const response = await route.fetch();
    call.status = response.status();
    calls.push(call);
    await route.fulfill({ response });
  };

  await page.route(routePattern, handler);

  return {
    calls,
    connect: () => {
      connected = true;
    },
    disconnect: () => {
      connected = false;
    },
    dispose: async () => {
      await safeUnroute(page, routePattern, handler);
    },
  };
}

export function successfulPatchCalls(calls: PatchCall[]) {
  return calls.filter((call) => call.status >= 200 && call.status < 300);
}

export function summaryPatchValue(body: Record<string, unknown>) {
  const changes = Array.isArray(body.changes) ? body.changes : [];
  const summaryChange = changes.find(
    (change): change is { field_key: string; value: unknown } =>
      typeof change === "object" &&
      change !== null &&
      "field_key" in change &&
      change.field_key === "timeline.summary",
  );
  return summaryChange?.value;
}

export function requireRecordId(row: Record<string, unknown>) {
  if (typeof row.record_id !== "string") {
    throw new Error(`missing record_id in row ${JSON.stringify(row)}`);
  }
  return row.record_id;
}

export async function expectServerSummaries(
  page: Page,
  incidentId: string,
  expected: Record<string, string>,
) {
  await expect
    .poll(async () => {
      const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
      const summaries: Record<string, string> = {};
      for (const row of rows) {
        const recordId =
          typeof row.record_id === "string" ? row.record_id : undefined;
        if (!recordId || !(recordId in expected)) {
          continue;
        }
        summaries[recordId] = readTimelineSummary(row);
      }
      return summaries;
    })
    .toEqual(expected);
}

export async function expectServerTimelineCells(
  page: Page,
  incidentId: string,
  recordId: string,
  expected: Record<string, string>,
) {
  await expect
    .poll(async () => {
      const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
      const row = rows.find((candidate) => candidate.record_id === recordId);
      if (!row) {
        return {};
      }
      const values: Record<string, string> = {};
      for (const fieldKey of Object.keys(expected)) {
        values[fieldKey] = readTimelineCell(row, fieldKey);
      }
      return values;
    })
    .toEqual(expected);
}

function parseSocketPayload(
  payload: string | Buffer,
  socketIndex: number,
): SocketMessage | null {
  const text = Buffer.isBuffer(payload) ? payload.toString("utf8") : payload;
  try {
    const parsed = JSON.parse(text) as {
      payload?: Record<string, unknown>;
      type?: unknown;
    };
    if (typeof parsed.type !== "string") {
      return null;
    }
    return {
      payload: parsed.payload ?? {},
      receivedAtMs: performance.now(),
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
  return `sockets=${sockets.length}; received=[${received}]; sent=[${sent}]; closes=[${closed}]`;
}

function parseRequestBody(postData: string | null): Record<string, unknown> {
  if (!postData) {
    return {};
  }
  return JSON.parse(postData) as Record<string, unknown>;
}

function recordIdFromURL(url: string) {
  const parsed = new URL(url);
  const prefix = "/api/v1/records/";
  if (!parsed.pathname.startsWith(prefix)) {
    throw new Error(`unexpected record patch URL: ${url}`);
  }
  return parsed.pathname.slice(prefix.length);
}

async function fulfillJSONError(route: Route, status: number, code: string) {
  await route.fulfill({
    body: JSON.stringify({
      error: {
        code,
        details: {},
        message: code,
        request_id: `phase6-e2e-${code}`,
        retryable: false,
        status,
      },
    }),
    contentType: "application/json",
    status,
  });
}

function readTimelineSummary(row: Record<string, unknown>) {
  return readTimelineCell(row, "timeline.summary");
}

function readTimelineCell(row: Record<string, unknown>, fieldKey: string) {
  const cells = row.cells;
  if (!cells || typeof cells !== "object" || Array.isArray(cells)) {
    return "";
  }
  const cell = (cells as Record<string, { value?: unknown }>)[fieldKey];
  return typeof cell?.value === "string" ? cell.value : "";
}
