import { scrollGridCellIntoView } from "@cartulary/test-utils/grid";
import {
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  currentIncidentRoleTestId,
  gridShellTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  rowCellTestId,
  rowPresenceMarkerTestId,
  saveStateTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import type { Browser, Page, Route } from "@playwright/test";
import { expect } from "@playwright/test";
import { csrfHeaders } from "../auth/browserSession";
import { timelineViewSchemaId } from "../contracts/workbookSurfaces";
import { createIncident } from "../incidents/fixtures";
import { createIncidentMemberUser } from "../incidents/memberships";
import { apiBase } from "../runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "../runtime/fixtureIdentity";
import {
  installIncidentSocketMonitor,
  type SocketMessage,
} from "../transport/incidentSocket";
import { safelyRemoveRoute as safeUnroute } from "../transport/requestInterception";
import { createViewRow, queryViewRows } from "../workbook/query";

async function expectCurrentIncidentRole(page: Page, roleText: string) {
  const accountMenuTrigger = page.getByLabel(
    "Account and application navigation",
  );
  await accountMenuTrigger.click();
  await expect(page.getByTestId(currentIncidentRoleTestId())).toHaveText(
    roleText,
  );
  await accountMenuTrigger.click();
}

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
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId: options.readyRecordId,
    surface: timelineViewSchemaId,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(options.readyRecordId, "timeline.activity_synopsis_text"),
    ),
  ).toBeVisible();
  return { acceptedSocket, page, socketMonitor };
}

export type SessionRecoveryScenario = {
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
    client_txn_id: uniqueTxn("collaboration-timeline-row"),
    "timeline.activity_synopsis_text": summary,
  });
}

export async function editTimelineSummary(
  page: Page,
  recordId: string,
  value: string,
) {
  await scrollGridCellIntoView({
    cellKey: "timeline.activity_synopsis_text",
    page,
    recordId,
    surface: timelineViewSchemaId,
  });
  const display = page.getByTestId(
    rowCellTestId(recordId, "timeline.activity_synopsis_text"),
  );
  await display.click();
  const input = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId,
      surface: "grid",
    }),
  );
  await expect(input).toBeFocused();
  await input.fill(value);
  await input.press("Enter");
  const currentValue = input.or(display).first();
  await expect
    .poll(() =>
      currentValue.evaluate((element) =>
        element instanceof HTMLInputElement ||
        element instanceof HTMLTextAreaElement
          ? element.value
          : element.textContent?.trim(),
      ),
    )
    .toBe(value);
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
    "timeline.activity_synopsis_text",
    remoteValue,
    txnPrefix,
  );

  heldPrimaryPatch.release();
  const primaryPatch = await heldPrimaryPatch.waitForCompletion;
  expect(primaryPatch.status).toBe(409);
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
  if (expectConflictMarker && !expectEditedCellMounted) {
    await expect(
      page.getByTestId(
        conflictMarkerTestId(recordId, "timeline.activity_synopsis_text"),
      ),
    ).toBeVisible();
  }
  await expect(page.getByTestId("conflict-resolver")).toBeVisible();
  if (expectEditedCellMounted) {
    const editorTestId = timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId,
      surface: "grid",
    });
    await expect(page.getByTestId(editorTestId)).toHaveValue(localValue);
  }
  await expect(page.getByTestId("conflict-server-value")).toHaveValue(
    remoteValue,
  );
  await expect(page.getByTestId("conflict-local-value")).toHaveValue(
    localValue,
  );
}

export async function exerciseSameFieldResolver({
  action,
  expectedPrimary,
  incidentId,
  localValue,
  mergedValue,
  page,
  patchController,
  recordId,
  remotePage,
  remoteValue,
}: {
  action: "keep_saved" | "merged_value" | "use_unsaved";
  expectedPrimary: string;
  incidentId: string;
  localValue: string;
  mergedValue?: string;
  page: Page;
  patchController: Awaited<ReturnType<typeof installPatchController>>;
  recordId: string;
  remotePage: Page;
  remoteValue: string;
}) {
  await driveRealTimelineSummaryConflict({
    baseRowVersion: 1,
    localValue,
    page,
    patchController,
    recordId,
    remoteValue,
    remotePatchPage: remotePage,
    txnPrefix: `e602-remote-${recordId}`,
  });
  await expect(
    remotePage.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText(remoteValue);
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();

  if (action === "keep_saved") {
    await page.getByTestId("conflict-keep-saved").click();
  } else if (action === "use_unsaved") {
    await page.getByTestId("conflict-use-unsaved").click();
  } else {
    await page.getByTestId("conflict-merged-value").fill(mergedValue ?? "");
    await page.getByTestId("conflict-use-merged").click();
  }

  await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  await expect(page.getByTestId("conflict-resolver")).toHaveCount(0);
  await expectServerTimelineCells(page, incidentId, recordId, {
    "timeline.activity_synopsis_text": expectedPrimary,
  });
  await expect(
    page.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText(expectedPrimary);
  await expect(
    remotePage.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    ),
  ).toHaveText(expectedPrimary);
  await expect(
    page
      .getByTestId(rowCellTestId(recordId, "timeline.activity_synopsis_text"))
      .locator("xpath=ancestor::*[@role='gridcell'][1]"),
  ).toBeFocused();
}

export async function exerciseRevokedPendingReplay({
  createdBy,
  incidentKeyPrefix,
  localValues,
  page,
  scenario,
  sessionTracker,
  triggerRevocation,
}: SessionRecoveryScenario) {
  const replayValues = localValues ?? [
    `Collaboration ${createdBy} ${scenario} first local`,
    `Collaboration ${createdBy} ${scenario} second local`,
  ];
  if (replayValues.length < 2) {
    throw new Error(
      "revoked pending replay requires at least two local values",
    );
  }
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey(incidentKeyPrefix),
    `Collaboration ${createdBy} ${scenario} revocation recovery`,
  );
  const member = await createIncidentMemberUser(page, incidentId, {
    display_name: `Collaboration ${createdBy} ${scenario} Analyst`,
    email: uniqueEmail(`collaboration-${createdBy.toLowerCase()}-${scenario}`),
    initial_password: `Collaboration${createdBy.replaceAll("-", "")}${scenario}Pass!`,
    role: "editor",
  });
  const recordIds: string[] = [];
  for (const [index] of replayValues.entries()) {
    recordIds.push(
      requireRecordId(
        await createTimelineRow(
          page,
          incidentId,
          `Collaboration ${createdBy} ${scenario} ${index + 1} base`,
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
      purpose: `Collaboration ${createdBy} ${scenario} analyst runtime`,
      userId: member.user_id,
    });

    const socketMonitor = installIncidentSocketMonitor(page, incidentId);
    await page.goto(`/?incident_id=${incidentId}`);
    await socketMonitor.waitForAcceptedSocket();
    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page,
      recordId: firstReplayItem.recordId,
      surface: timelineViewSchemaId,
    });
    await expect(
      page.getByTestId(
        rowCellTestId(
          firstReplayItem.recordId,
          "timeline.activity_synopsis_text",
        ),
      ),
    ).toHaveText(`Collaboration ${createdBy} ${scenario} 1 base`);
    await expectCurrentIncidentRole(page, "Current incident role: editor");

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
      await scrollGridCellIntoView({
        cellKey: "timeline.activity_synopsis_text",
        page,
        recordId: item.recordId,
        surface: timelineViewSchemaId,
      });
      await expect(
        page.getByTestId(
          rowCellTestId(item.recordId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveText(item.value);
    }
    await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
      String(replayValues.length),
    );

    const messageStart = socketMonitor.messageCount();
    await sessionTracker.loginTrackedUser(page, {
      createdBy,
      email: member.email,
      password: member.initial_password,
      purpose: `Collaboration ${createdBy} ${scenario} analyst re-authentication`,
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

export function presenceDeltaMatches(
  message: SocketMessage,
  options: {
    fieldKey: string;
    mode: string;
    recordId: string;
  },
) {
  return presenceMessageMatches(message, options);
}

export function presenceMessageMatches(
  message: SocketMessage,
  options: {
    fieldKey: string;
    mode: string;
    recordId: string;
  },
) {
  if (message.type === "presence_delta") {
    return presenceRecordMatches(message.payload.presence, options);
  }
  if (message.type !== "presence_snapshot") {
    return false;
  }
  const presences = message.payload.presences;
  if (!Array.isArray(presences)) {
    return false;
  }
  return presences.some((presence) => presenceRecordMatches(presence, options));
}

function presenceRecordMatches(
  presence: unknown,
  options: {
    fieldKey: string;
    mode: string;
    recordId: string;
  },
) {
  return (
    presence !== null &&
    typeof presence === "object" &&
    !Array.isArray(presence) &&
    "record_id" in presence &&
    presence.record_id === options.recordId &&
    "field_key" in presence &&
    presence.field_key === options.fieldKey &&
    "mode" in presence &&
    presence.mode === options.mode
  );
}

export async function focusRemoteTimelineCellAndWaitForPresence({
  actorText,
  fieldKey,
  mode = "editing",
  primaryPage,
  recordId,
  remotePage,
  socketMonitor,
  timeoutMs,
}: {
  actorText?: string;
  fieldKey: string;
  mode?: string;
  primaryPage: Page;
  recordId: string;
  remotePage: Page;
  socketMonitor: ReturnType<typeof installIncidentSocketMonitor>;
  timeoutMs?: number;
}) {
  const markerStartAt = socketMonitor.messageCount();
  const markerPresence = socketMonitor.waitForMessageWhere(
    `matching presence ${recordId}:${fieldKey}:${mode}`,
    {
      matches: (message) =>
        presenceMessageMatches(message, {
          fieldKey,
          mode,
          recordId,
        }),
      startAt: markerStartAt,
      ...(timeoutMs === undefined ? {} : { timeoutMs }),
    },
  );
  await scrollGridCellIntoView({
    cellKey: fieldKey,
    page: remotePage,
    recordId,
    surface: timelineViewSchemaId,
  });
  await scrollGridCellIntoView({
    cellKey: fieldKey,
    page: primaryPage,
    recordId,
    surface: timelineViewSchemaId,
  });
  const remoteDisplay = remotePage.getByTestId(
    rowCellTestId(recordId, fieldKey),
  );
  await remoteDisplay.click();
  await expect(
    remotePage.getByTestId(
      timelineScalarEditorTestId({
        fieldKey,
        recordId,
        surface: "grid",
      }),
    ),
  ).toBeFocused();
  const presenceMessage = await markerPresence;
  const rowMarker = primaryPage.getByTestId(rowPresenceMarkerTestId(recordId));
  const cellMarker = primaryPage.getByTestId(
    cellPresenceMarkerTestId(recordId, fieldKey),
  );
  if (actorText === undefined) {
    await expect(rowMarker).toBeVisible();
    await expect(cellMarker).toBeVisible();
  } else {
    await expect(rowMarker).toContainText(actorText);
    await expect(cellMarker).toContainText(actorText);
  }
  return presenceMessage;
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
      change.field_key === "timeline.activity_synopsis_text",
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
        request_id: `collaboration-e2e-${code}`,
        retryable: false,
        status,
      },
    }),
    contentType: "application/json",
    status,
  });
}

function readTimelineSummary(row: Record<string, unknown>) {
  return readTimelineCell(row, "timeline.activity_synopsis_text");
}

function readTimelineCell(row: Record<string, unknown>, fieldKey: string) {
  const cells = row.cells;
  if (!cells || typeof cells !== "object" || Array.isArray(cells)) {
    return "";
  }
  const cell = (cells as Record<string, { value?: unknown }>)[fieldKey];
  return typeof cell?.value === "string" ? cell.value : "";
}
