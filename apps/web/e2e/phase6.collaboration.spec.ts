import { performance } from "node:perf_hooks";

import {
  applyFilterChip,
  assertMarkerAnchoredToGridTarget,
  assertMountedGridRowCountAtMost,
  changeGrouping,
  gridAnchorCommandScenarios,
  scrollGridCellIntoView,
  scrollGridToOffset,
  sortByHeader,
} from "@cartulary/test-utils";
import {
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  currentIncidentRoleTestId,
  gridRowGutterTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  rowCellTestId,
  rowPresenceMarkerTestId,
  saveStateTestId,
  timelineRowVersionTestId,
} from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";

import { revokeAllSessions } from "./authRuntime";
import { expect, test } from "./fixtures";
import {
  apiBase,
  applyStorageState,
  createIncident,
  createIncidentMemberUser,
  csrfHeaders,
  openIncidentAsTrackedUser,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  createTimelineRow,
  driveRealTimelineSummaryConflict,
  editTimelineSummary,
  exerciseRevokedPendingReplay,
  exerciseSameFieldResolver,
  expectServerSummaries,
  expectServerTimelineCells,
  focusRemoteTimelineCellAndWaitForPresence,
  installIncidentSocketMonitor,
  installPatchController,
  installPatchTransportFailureController,
  openIncidentAsTrackedUserReady,
  patchTimelineField,
  presenceDeltaMatches,
  requireRecordId,
  successfulPatchCalls,
  summaryPatchValue,
  timelineViewSchemaId,
} from "./phase6Harness";
import { assertRecordFieldMutationAnchor } from "./workbookRequestSupport";

const presenceInteractionThresholdMs = 1000;

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

test("FE-I-P7-01 Verify conflict resolver actions submit public mutations and refresh rows without losing focus or pending queue ordering.", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEIP701"),
    "FE-I-P7-01 conflict resolver public route",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "FE-I-P7 Remote",
    email: uniqueEmail("fe-i-p7-remote"),
    initial_password: "FeIP7RemotePass!",
    role: "editor",
  });
  const conflictId = requireRecordId(
    await createTimelineRow(page, incidentId, "FE-I-P7 conflict base"),
  );
  const queuedAId = requireRecordId(
    await createTimelineRow(page, incidentId, "FE-I-P7 queued A base"),
  );
  const queuedBId = requireRecordId(
    await createTimelineRow(page, incidentId, "FE-I-P7 queued B base"),
  );
  const patchController = await installPatchController(page);

  let remotePage: Page | null = null;
  try {
    await page.goto(`/?incident_id=${incidentId}`);
    await expect(
      page.getByTestId(
        rowCellTestId(conflictId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveValue("FE-I-P7 conflict base");

    remotePage = await openIncidentAsTrackedUser(browser, sessionTracker, {
      createdBy: "FE-I-P7-01",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "FE-I-P7 public resolver remote analyst",
      userId: remote.user_id,
    });
    await expect(
      remotePage.getByTestId(
        rowCellTestId(conflictId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveValue("FE-I-P7 conflict base");

    await driveRealTimelineSummaryConflict({
      baseRowVersion: 1,
      localValue: "FE-I-P7 local draft",
      page,
      patchController,
      recordId: conflictId,
      remotePatchPage: remotePage,
      remoteValue: "FE-I-P7 remote saved",
      txnPrefix: "feip701-remote-conflict",
    });

    await patchTimelineField(
      remotePage,
      conflictId,
      2,
      "timeline.activity_synopsis_text",
      "FE-I-P7 newer saved",
      "feip701-newer-saved",
    );

    await page
      .getByTestId("conflict-merged-value")
      .fill("FE-I-P7 stale merged draft");
    const staleResolveRequest = page.waitForRequest(
      (request) =>
        request.method() === "POST" &&
        request.url().includes(`/api/v1/records/${conflictId}/conflicts/`) &&
        request.url().endsWith("/resolve"),
    );
    const staleResolveResponse = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        response.url().includes(`/api/v1/records/${conflictId}/conflicts/`) &&
        response.url().endsWith("/resolve"),
    );
    await page.getByTestId("conflict-use-merged").click();
    const staleRequest = await staleResolveRequest;
    const staleResponse = await staleResolveResponse;
    expect(staleResponse.status()).toBe(409);
    expect(JSON.parse(staleRequest.postData() ?? "{}")).toMatchObject({
      conflict_token: expect.any(String),
      resolution_kind: "merged_value",
      resolved_value: "FE-I-P7 stale merged draft",
    });
    const staleEnvelope = await staleResponse.json();
    expect(staleEnvelope.error.code).toBe("same_field_conflict");
    expect(staleEnvelope.error.conflict.server_value).toBe(
      "FE-I-P7 newer saved",
    );
    await expect(page.getByTestId("conflict-server-value")).toHaveValue(
      "FE-I-P7 newer saved",
    );
    await expect(page.getByTestId("conflict-local-value")).toHaveValue(
      "FE-I-P7 stale merged draft",
    );

    await page
      .getByTestId("conflict-merged-value")
      .fill("FE-I-P7 final merged");
    const successResolveRequest = page.waitForRequest(
      (request) =>
        request.method() === "POST" &&
        request.url().includes(`/api/v1/records/${conflictId}/conflicts/`) &&
        request.url().endsWith("/resolve"),
    );
    const successResolveResponse = page.waitForResponse(
      (response) =>
        response.request().method() === "POST" &&
        response.url().includes(`/api/v1/records/${conflictId}/conflicts/`) &&
        response.url().endsWith("/resolve"),
    );
    await page.getByTestId("conflict-use-merged").click();
    const successRequest = await successResolveRequest;
    const successResponse = await successResolveResponse;
    expect(successResponse.ok()).toBeTruthy();
    expect(JSON.parse(successRequest.postData() ?? "{}")).toMatchObject({
      conflict_token: expect.any(String),
      resolution_kind: "merged_value",
      resolved_value: "FE-I-P7 final merged",
    });
    await expect(page.getByTestId("conflict-resolver")).toHaveCount(0);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expectServerTimelineCells(page, incidentId, conflictId, {
      "timeline.activity_synopsis_text": "FE-I-P7 final merged",
    });

    const heldQueuePatch = patchController.holdNextPatch({
      recordId: queuedAId,
    });
    await editTimelineSummary(page, queuedAId, "FE-I-P7 queued A local");
    await heldQueuePatch.waitForHit;
    await editTimelineSummary(page, queuedBId, "FE-I-P7 queued B local");
    await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
      "2",
    );
    heldQueuePatch.release();
    await expect
      .poll(
        () =>
          successfulPatchCalls(patchController.calls).filter((call) =>
            [queuedAId, queuedBId].includes(call.recordId),
          ).length,
      )
      .toBe(2);
    const replayed = successfulPatchCalls(patchController.calls).filter(
      (call) => [queuedAId, queuedBId].includes(call.recordId),
    );
    expect(replayed.map((call) => call.recordId)).toEqual([
      queuedAId,
      queuedBId,
    ]);
    expect(replayed.map((call) => summaryPatchValue(call.body))).toEqual([
      "FE-I-P7 queued A local",
      "FE-I-P7 queued B local",
    ]);
  } finally {
    await patchController.dispose();
    await remotePage?.context().close();
  }
});

test("FE-E-P7-01 Verify multi-client live row update, presence anchoring, reset/invalidate handling, stale-row requery, and same-field conflict resolver through /ws/v1/ and /api/v1/.", async ({
  browser,
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEEP701"),
    "FE-E-P7-01 live collaboration",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "FE-E-P7 Remote",
    email: uniqueEmail("fe-e-p7-remote"),
    initial_password: "FeEP7RemotePass!",
    role: "editor",
  });
  const liveId = requireRecordId(
    await createTimelineRow(page, incidentId, "FE-E-P7 live base"),
  );
  const invalidateId = requireRecordId(
    await createTimelineRow(page, incidentId, "FE-E-P7 invalidate base"),
  );
  const conflictId = requireRecordId(
    await createTimelineRow(page, incidentId, "FE-E-P7 Zulu conflict base"),
  );
  const patchController = await installPatchController(page);
  const socketMonitor = installIncidentSocketMonitor(page, incidentId);

  let remotePage: Page | null = null;
  try {
    await page.goto(`/?incident_id=${incidentId}`);
    await socketMonitor.waitForMessage("hello_ack");
    await sortByHeader(
      page,
      timelineViewSchemaId,
      "timeline.activity_synopsis_text",
    );
    await applyFilterChip(
      page,
      timelineViewSchemaId,
      "timeline.has_evidence",
      "false",
    );
    await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");

    remotePage = await openIncidentAsTrackedUser(browser, sessionTracker, {
      createdBy: "FE-E-P7-01",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "FE-E-P7 remote analyst",
      userId: remote.user_id,
    });
    await sortByHeader(
      remotePage,
      timelineViewSchemaId,
      "timeline.activity_synopsis_text",
    );
    await applyFilterChip(
      remotePage,
      timelineViewSchemaId,
      "timeline.has_evidence",
      "false",
    );
    await changeGrouping(
      remotePage,
      timelineViewSchemaId,
      "timeline.capture_state",
    );

    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page: remotePage,
      recordId: liveId,
      surface: timelineViewSchemaId,
    });
    await focusRemoteTimelineCellAndWaitForPresence({
      actorText: "FR",
      fieldKey: "timeline.activity_synopsis_text",
      primaryPage: page,
      recordId: liveId,
      remotePage,
      socketMonitor,
    });
    await assertMarkerAnchoredToGridTarget({
      anchorKind: "row-gutter",
      markerTestId: rowPresenceMarkerTestId(liveId),
      page,
      surface: timelineViewSchemaId,
      targetTestId: gridRowGutterTestId(timelineViewSchemaId, liveId),
    });
    await assertMarkerAnchoredToGridTarget({
      anchorKind: "cell",
      markerTestId: cellPresenceMarkerTestId(
        liveId,
        "timeline.activity_synopsis_text",
      ),
      page,
      surface: timelineViewSchemaId,
      targetTestId: rowCellTestId(liveId, "timeline.activity_synopsis_text"),
    });

    const liveStartAt = socketMonitor.messageCount();
    await patchTimelineField(
      remotePage,
      liveId,
      1,
      "timeline.activity_synopsis_text",
      "FE-E-P7 live remote update",
      "feep701-live-remote",
    );
    await socketMonitor.waitForMessage("record_changed", {
      matches: (message) =>
        message.payload.record_id === liveId &&
        message.payload.row_version === 2,
      startAt: liveStartAt,
    });
    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page,
      recordId: liveId,
      surface: timelineViewSchemaId,
    });
    await expect(
      page.getByTestId(
        rowCellTestId(liveId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveValue("FE-E-P7 live remote update");

    const removeStartAt = socketMonitor.messageCount();
    const deleteResponse = await page.request.delete(
      `${apiBase}/api/v1/records/${invalidateId}`,
      {
        headers: await csrfHeaders(page),
        data: {
          base_row_version: 1,
          client_txn_id: uniqueTxn("feep701-delete"),
        },
      },
    );
    expect(deleteResponse.ok()).toBeTruthy();
    const removeMessage = await socketMonitor.waitForMessage("record_changed", {
      matches: (message) =>
        message.payload.record_id === invalidateId &&
        Array.isArray(message.payload.affected_views) &&
        message.payload.affected_views.some(
          (view: { change_kind?: string }) => view.change_kind === "remove",
        ),
      startAt: removeStartAt,
    });
    const tombstoneVersion = Number(removeMessage.payload.row_version);
    expect(tombstoneVersion).toBeGreaterThan(1);

    const invalidateStartAt = socketMonitor.messageCount();
    const restoreResponse = await page.request.post(
      `${apiBase}/api/v1/records/${invalidateId}/restore`,
      {
        headers: await csrfHeaders(page),
        data: {
          base_row_version: tombstoneVersion,
          client_txn_id: uniqueTxn("feep701-restore"),
        },
      },
    );
    expect(restoreResponse.ok()).toBeTruthy();
    await socketMonitor.waitForMessage("record_changed", {
      matches: (message) =>
        message.payload.record_id === invalidateId &&
        Array.isArray(message.payload.affected_views) &&
        message.payload.affected_views.some(
          (view: { change_kind?: string; view_schema_id?: string }) =>
            view.view_schema_id === timelineViewSchemaId &&
            view.change_kind === "invalidate",
        ),
      startAt: invalidateStartAt,
    });
    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page,
      recordId: invalidateId,
      surface: timelineViewSchemaId,
    });
    await expect(
      page.getByTestId(
        rowCellTestId(invalidateId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveValue("FE-E-P7 invalidate base");

    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page,
      recordId: conflictId,
      surface: timelineViewSchemaId,
    });
    await driveRealTimelineSummaryConflict({
      baseRowVersion: 1,
      localValue: "FE-E-P7 local conflict",
      page,
      patchController,
      recordId: conflictId,
      remotePatchPage: remotePage,
      remoteValue: "FE-E-P7 remote conflict",
      txnPrefix: "feep701-conflict",
    });
    await assertMarkerAnchoredToGridTarget({
      anchorKind: "cell",
      markerTestId: conflictMarkerTestId(
        conflictId,
        "timeline.activity_synopsis_text",
      ),
      page,
      surface: timelineViewSchemaId,
      targetTestId: rowCellTestId(
        conflictId,
        "timeline.activity_synopsis_text",
      ),
    });
    await page.getByTestId("conflict-use-unsaved").click();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expectServerTimelineCells(page, incidentId, conflictId, {
      "timeline.activity_synopsis_text": "FE-E-P7 local conflict",
    });
  } finally {
    await patchController.dispose();
    await remotePage?.context().close();
  }

  await exerciseRevokedPendingReplay({
    createdBy: "FE-E-P7-01",
    incidentKeyPrefix: "FEEP701REVOKE",
    page,
    scenario: "revoked",
    sessionTracker,
    triggerRevocation: async ({ member }) => {
      await revokeAllSessions(
        workerAdminRequest,
        member.user_id,
        "FE-E-P7-01 browser revoke-all",
      );
    },
  });
});

test("E-6-01 shows two analysts each other's workbook presence within the expected interaction window", async ({
  browser,
  page,
  sessionTracker,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E601"),
    "Phase 6 E-6-01 browser presence",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "Remote Analyst",
    email: uniqueEmail("phase6-e601-remote"),
    initial_password: "Phase6E601Remote!",
    role: "editor",
  });
  const row = await createTimelineRow(page, incidentId, "E-6-01 presence base");
  const recordId = requireRecordId(row);
  const primarySocket = installIncidentSocketMonitor(page, incidentId);

  let remotePage: Page | null = null;
  try {
    await page.goto(`/?incident_id=${incidentId}`);
    await primarySocket.waitForMessage("hello_ack");
    await expect(
      page.getByTestId(
        rowCellTestId(recordId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveValue("E-6-01 presence base");

    const remoteSession = await openIncidentAsTrackedUserReady(
      browser,
      sessionTracker,
      {
        createdBy: "E-6-01",
        email: remote.email,
        incidentId,
        password: remote.initial_password,
        purpose: "Phase 6 E-6-01 remote analyst",
        readyRecordId: recordId,
        userId: remote.user_id,
      },
    );
    remotePage = remoteSession.page;
    await primarySocket.waitForMessage("presence_delta");
    await expect(page.getByTestId("presence-header")).toContainText("RA");

    const fieldKey = "timeline.activity_synopsis_text";
    const remoteInput = remotePage.getByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    );
    const markerStartAt = primarySocket.messageCount();
    const interactionStartedAtMs = performance.now();
    const markerDelta = primarySocket.waitForMessage("presence_delta", {
      matches: (message) =>
        presenceDeltaMatches(message, {
          fieldKey,
          mode: "editing",
          recordId,
        }),
      startAt: markerStartAt,
      timeoutMs: presenceInteractionThresholdMs,
    });
    await remoteInput.focus();
    const [presenceDelta, markerTiming] = await Promise.all([
      markerDelta,
      waitForPresenceMarkerTiming(page, {
        actorText: "RA",
        fieldKey,
        recordId,
        startedAtMs: interactionStartedAtMs,
        timeoutMs: presenceInteractionThresholdMs,
      }),
    ]);
    const socketDeltaDurationMs =
      presenceDelta.receivedAtMs - interactionStartedAtMs;
    const timingArtifact = {
      actor_text: "RA",
      cell_marker_duration_ms: markerTiming.cellMarkerDurationMs,
      field_key: fieldKey,
      interaction_threshold_ms: presenceInteractionThresholdMs,
      record_id: recordId,
      row_marker_duration_ms: markerTiming.rowMarkerDurationMs,
      remote_outbound_presence_updates: remoteSession.socketMonitor
        .sentMessages()
        .filter((message) => {
          return message.type === "presence_update";
        }).length,
      remote_socket_index: remoteSession.acceptedSocket.socketIndex,
      socket_delta_duration_ms: socketDeltaDurationMs,
      ui_render_duration_ms: markerTiming.renderDurationMs,
    };
    await testInfo.attach("phase6-e-6-01-presence-timing.json", {
      body: JSON.stringify(timingArtifact, null, 2),
      contentType: "application/json",
    });
    expect(socketDeltaDurationMs).toBeLessThanOrEqual(
      presenceInteractionThresholdMs,
    );
    expect(markerTiming.renderDurationMs).toBeLessThanOrEqual(
      presenceInteractionThresholdMs,
    );
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
  } finally {
    await remotePage?.context().close();
  }
});

test("E-6-02 auto-merges different-field concurrent edits and requires explicit same-field resolution", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E602"),
    "Phase 6 E-6-02 concurrent resolver UX",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "Phase 6 E-6-02 Remote",
    email: uniqueEmail("phase6-e602-remote"),
    initial_password: "Phase6E602Remote!",
    role: "editor",
  });
  const differentId = requireRecordId(
    await createTimelineRow(page, incidentId, "E-6-02 different base"),
  );
  const keepId = requireRecordId(
    await createTimelineRow(page, incidentId, "E-6-02 keep base"),
  );
  const useId = requireRecordId(
    await createTimelineRow(page, incidentId, "E-6-02 use base"),
  );
  const mergedId = requireRecordId(
    await createTimelineRow(page, incidentId, "E-6-02 merged base"),
  );
  const patchController = await installPatchController(page);

  let remotePage: Page | null = null;
  try {
    const socketMonitor = installIncidentSocketMonitor(page, incidentId);
    await page.goto(`/?incident_id=${incidentId}`);
    await socketMonitor.waitForMessage("hello_ack");
    await expect(
      page.getByTestId(
        rowCellTestId(differentId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveValue("E-6-02 different base");

    remotePage = await openIncidentAsTrackedUser(browser, sessionTracker, {
      createdBy: "E-6-02",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "Phase 6 E-6-02 remote analyst",
      userId: remote.user_id,
    });
    await expect(
      remotePage.getByTestId(
        rowCellTestId(differentId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveValue("E-6-02 different base");

    await patchTimelineField(
      page,
      differentId,
      1,
      "timeline.activity_synopsis_text",
      "E-6-02 different primary",
      "e602-different-primary-summary",
    );
    await patchTimelineField(
      remotePage,
      differentId,
      1,
      "timeline.raw_activity_text",
      "E-6-02 different remote details",
      "e602-different-remote-details",
    );
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(
      page.getByTestId(
        rowCellTestId(differentId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveValue("E-6-02 different primary");
    await expect(
      remotePage.getByTestId(
        rowCellTestId(differentId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveValue("E-6-02 different primary");
    await expectServerTimelineCells(page, incidentId, differentId, {
      "timeline.raw_activity_text": "E-6-02 different remote details",
      "timeline.activity_synopsis_text": "E-6-02 different primary",
    });

    await exerciseSameFieldResolver({
      action: "keep_saved",
      expectedPrimary: "E-6-02 keep remote",
      localValue: "E-6-02 keep primary",
      page,
      incidentId,
      patchController,
      recordId: keepId,
      remotePage,
      remoteValue: "E-6-02 keep remote",
    });

    await exerciseSameFieldResolver({
      action: "use_unsaved",
      expectedPrimary: "E-6-02 use primary",
      localValue: "E-6-02 use primary",
      page,
      incidentId,
      patchController,
      recordId: useId,
      remotePage,
      remoteValue: "E-6-02 use remote",
    });

    await exerciseSameFieldResolver({
      action: "merged_value",
      expectedPrimary: "E-6-02 merged final",
      localValue: "E-6-02 merged primary",
      mergedValue: "E-6-02 merged final",
      page,
      incidentId,
      patchController,
      recordId: mergedId,
      remotePage,
      remoteValue: "E-6-02 merged remote",
    });
  } finally {
    await patchController.dispose();
    await remotePage?.context().close();
  }
});

test("E-6-04 keeps live updates conflict markers and presence markers anchored to record_id and field_key", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E604"),
    "Phase 6 E-6-04 live-cell anchoring",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "Anchor Analyst",
    email: uniqueEmail("phase6-e604-anchor"),
    initial_password: "Phase6E604Anchor!",
    role: "editor",
  });
  for (let index = 0; index < 24; index += 1) {
    await createTimelineRow(
      page,
      incidentId,
      `E-6-04 Filler ${String(index).padStart(2, "0")}`,
    );
  }
  const commandRows = [];
  for (const scenario of gridAnchorCommandScenarios(timelineViewSchemaId)) {
    const sortLabel = String.fromCharCode(65 + commandRows.length);
    commandRows.push({
      baseSummary: `E-6-04 ${sortLabel} ${scenario.name} command base`,
      recordId: requireRecordId(
        await createTimelineRow(
          page,
          incidentId,
          `E-6-04 ${sortLabel} ${scenario.name} command base`,
        ),
      ),
      scenario,
      sortLabel,
    });
  }
  const alphaRow = await createTimelineRow(
    page,
    incidentId,
    "E-6-04 Zulu anchor base",
  );
  const alphaId = requireRecordId(alphaRow);
  const patchController = await installPatchController(page);

  let remotePage: Page | null = null;
  try {
    const socketMonitor = installIncidentSocketMonitor(page, incidentId);
    await page.goto(`/?incident_id=${incidentId}`);
    await socketMonitor.waitForMessage("hello_ack");
    await sortByHeader(
      page,
      timelineViewSchemaId,
      "timeline.activity_synopsis_text",
    );
    await applyFilterChip(
      page,
      timelineViewSchemaId,
      "timeline.has_evidence",
      "false",
    );
    await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");

    for (const { baseSummary, recordId, scenario, sortLabel } of commandRows) {
      await scrollGridCellIntoView({
        cellKey: "timeline.activity_synopsis_text",
        page,
        recordId,
        surface: timelineViewSchemaId,
      });
      const input = page.getByTestId(
        rowCellTestId(recordId, "timeline.activity_synopsis_text"),
      );
      await expect(input).toHaveValue(baseSummary);

      await patchTimelineField(
        page,
        recordId,
        1,
        "timeline.raw_activity_text",
        `E-6-04 ${scenario.name} remote details`,
        `e604-${scenario.name}-live-patch`,
      );
      await expect(
        page.getByTestId(timelineRowVersionTestId(recordId)),
      ).toHaveText("2");

      const expectedValue = `E-6-04 ${sortLabel} ${scenario.name} anchored local`;
      const heldPatch = patchController.holdNextPatch();
      await input.focus();
      await input.fill(expectedValue);
      await scenario.commit({ input, page, surface: timelineViewSchemaId });
      const call = await heldPatch.waitForHit;
      assertRecordFieldMutationAnchor({
        actualRecordId: call.recordId,
        body: call.body,
        expectedRecordId: recordId,
        expectedValue,
        fieldKey: "timeline.activity_synopsis_text",
      });
      heldPatch.release();
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expectServerSummaries(page, incidentId, {
        [recordId]: expectedValue,
      });
    }

    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page,
      recordId: alphaId,
      surface: timelineViewSchemaId,
    });
    await assertMountedGridRowCountAtMost({
      maxRows: 30,
      page,
      surface: timelineViewSchemaId,
    });
    const alphaInput = page.getByTestId(
      rowCellTestId(alphaId, "timeline.activity_synopsis_text"),
    );
    await expect(alphaInput).toHaveValue("E-6-04 Zulu anchor base");

    remotePage = await openIncidentAsTrackedUser(browser, sessionTracker, {
      createdBy: "E-6-04",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "Phase 6 E-6-04 remote anchor analyst",
      userId: remote.user_id,
    });
    await sortByHeader(
      remotePage,
      timelineViewSchemaId,
      "timeline.activity_synopsis_text",
    );
    await applyFilterChip(
      remotePage,
      timelineViewSchemaId,
      "timeline.has_evidence",
      "false",
    );
    await changeGrouping(
      remotePage,
      timelineViewSchemaId,
      "timeline.capture_state",
    );
    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page: remotePage,
      recordId: alphaId,
      surface: timelineViewSchemaId,
    });
    await focusRemoteTimelineCellAndWaitForPresence({
      actorText: "AA",
      fieldKey: "timeline.activity_synopsis_text",
      primaryPage: page,
      recordId: alphaId,
      remotePage,
      socketMonitor,
    });

    await driveRealTimelineSummaryConflict({
      baseRowVersion: 1,
      expectConflictMarker: false,
      expectEditedCellMounted: false,
      localValue: "E-6-04 Alpha local",
      page,
      patchController,
      recordId: alphaId,
      remoteValue: "E-6-04 Alpha remote",
      txnPrefix: "e604-alpha-remote-conflict",
    });
    await scrollGridToOffset(page, timelineViewSchemaId, 0);
    await expect(
      page.getByTestId(
        conflictMarkerTestId(alphaId, "timeline.activity_synopsis_text"),
      ),
    ).toBeVisible();
    await assertMarkerAnchoredToGridTarget({
      anchorKind: "cell",
      markerTestId: conflictMarkerTestId(
        alphaId,
        "timeline.activity_synopsis_text",
      ),
      page,
      surface: timelineViewSchemaId,
      targetTestId: rowCellTestId(alphaId, "timeline.activity_synopsis_text"),
    });
    await assertMarkerAnchoredToGridTarget({
      anchorKind: "row-gutter",
      markerTestId: rowPresenceMarkerTestId(alphaId),
      page,
      surface: timelineViewSchemaId,
      targetTestId: gridRowGutterTestId(timelineViewSchemaId, alphaId),
    });
    await assertMarkerAnchoredToGridTarget({
      anchorKind: "cell",
      markerTestId: cellPresenceMarkerTestId(
        alphaId,
        "timeline.activity_synopsis_text",
      ),
      page,
      surface: timelineViewSchemaId,
      targetTestId: rowCellTestId(alphaId, "timeline.activity_synopsis_text"),
    });
    await expect(alphaInput).toHaveValue("E-6-04 Alpha local");
  } finally {
    await patchController.dispose();
    await remotePage?.context().close();
  }
});

test("E-6-05 replays queued unsent writes after re-authentication without silent reload restore", async ({
  page,
  sessionTracker,
  workerAdmin,
  workerAdminRequest,
}) => {
  await test.step("replays transient browser request failures in FIFO order after transport recovery", async () => {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("E605FIFO"),
      "Phase 6 E-6-05 FIFO recovery",
    );
    const firstId = requireRecordId(
      await createTimelineRow(page, incidentId, "E-6-05 FIFO A base"),
    );
    const secondId = requireRecordId(
      await createTimelineRow(page, incidentId, "E-6-05 FIFO B base"),
    );
    const thirdId = requireRecordId(
      await createTimelineRow(page, incidentId, "E-6-05 FIFO C base"),
    );
    const patchController = await installPatchTransportFailureController(page);

    try {
      patchController.disconnect();
      await page.goto(`/?incident_id=${incidentId}`);
      await expect(
        page.getByTestId(
          rowCellTestId(firstId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveValue("E-6-05 FIFO A base");
      await expectCurrentIncidentRole(page, "Current incident role: admin");

      await editTimelineSummary(page, firstId, "E-6-05 FIFO A local");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();
      await editTimelineSummary(page, secondId, "E-6-05 FIFO B local");
      await editTimelineSummary(page, thirdId, "E-6-05 FIFO C local");

      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "3",
      );
      expect(successfulPatchCalls(patchController.calls)).toHaveLength(0);

      patchController.connect();
      await expect
        .poll(() => successfulPatchCalls(patchController.calls).length)
        .toBe(3);
      const replayed = successfulPatchCalls(patchController.calls);
      expect(replayed.map((call) => call.recordId)).toEqual([
        firstId,
        secondId,
        thirdId,
      ]);
      expect(replayed.map((call) => summaryPatchValue(call.body))).toEqual([
        "E-6-05 FIFO A local",
        "E-6-05 FIFO B local",
        "E-6-05 FIFO C local",
      ]);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expectServerSummaries(page, incidentId, {
        [firstId]: "E-6-05 FIFO A local",
        [secondId]: "E-6-05 FIFO B local",
        [thirdId]: "E-6-05 FIFO C local",
      });
    } finally {
      await patchController.dispose();
    }
  });

  await test.step("replays queued writes in FIFO order after real HTTP auth failure and re-authentication", async () => {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("E605HTTPAUTH"),
      "Phase 6 E-6-05 HTTP auth recovery",
    );
    const member = await createIncidentMemberUser(page, incidentId, {
      display_name: "Phase 6 E-6-05 HTTP Auth Analyst",
      email: uniqueEmail("phase6-e605-http-auth"),
      initial_password: "Phase6E605HttpAuth!",
      role: "editor",
    });
    const firstId = requireRecordId(
      await createTimelineRow(page, incidentId, "E-6-05 auth A base"),
    );
    const secondId = requireRecordId(
      await createTimelineRow(page, incidentId, "E-6-05 auth B base"),
    );
    const patchController = await installPatchController(page);

    try {
      await sessionTracker.loginTrackedUser(page, {
        createdBy: "E-6-05",
        email: member.email,
        password: member.initial_password,
        purpose: "Phase 6 E-6-05 HTTP auth analyst runtime",
        userId: member.user_id,
      });
      await page.goto(`/?incident_id=${incidentId}`);
      await expect(
        page.getByTestId(
          rowCellTestId(firstId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveValue("E-6-05 auth A base");
      await expectCurrentIncidentRole(page, "Current incident role: editor");

      await page.context().clearCookies();
      await editTimelineSummary(page, firstId, "E-6-05 auth A local");
      await expect
        .poll(() => patchController.calls.at(-1)?.status ?? 0)
        .toBe(401);
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();
      await editTimelineSummary(page, secondId, "E-6-05 auth B local");
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "2",
      );

      await sessionTracker.loginTrackedUser(page, {
        createdBy: "E-6-05",
        email: member.email,
        password: member.initial_password,
        purpose: "Phase 6 E-6-05 HTTP auth analyst re-authentication",
        userId: member.user_id,
      });
      await expect
        .poll(() => successfulPatchCalls(patchController.calls).length)
        .toBe(2);
      const replayed = successfulPatchCalls(patchController.calls);
      expect(replayed.map((call) => call.recordId)).toEqual([
        firstId,
        secondId,
      ]);
      expect(replayed.map((call) => summaryPatchValue(call.body))).toEqual([
        "E-6-05 auth A local",
        "E-6-05 auth B local",
      ]);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expectServerSummaries(page, incidentId, {
        [firstId]: "E-6-05 auth A local",
        [secondId]: "E-6-05 auth B local",
      });
    } finally {
      await patchController.dispose();
      await applyStorageState(page, workerAdmin.storageState);
    }
  });

  await test.step("halts on the first real same-field conflict and keeps later writes queued", async () => {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("E605HALT"),
      "Phase 6 E-6-05 same-field conflict halt",
    );
    const firstId = requireRecordId(
      await createTimelineRow(page, incidentId, "E-6-05 halt A base"),
    );
    const secondId = requireRecordId(
      await createTimelineRow(page, incidentId, "E-6-05 halt B base"),
    );
    const thirdId = requireRecordId(
      await createTimelineRow(page, incidentId, "E-6-05 halt C base"),
    );
    const patchController = await installPatchController(page);

    try {
      await page.goto(`/?incident_id=${incidentId}`);
      await expect(
        page.getByTestId(
          rowCellTestId(firstId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveValue("E-6-05 halt A base");
      await expectCurrentIncidentRole(page, "Current incident role: admin");

      const heldPatch = patchController.holdNextPatch();
      await editTimelineSummary(page, firstId, "E-6-05 halt A local");
      await heldPatch.waitForHit;
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
      await editTimelineSummary(page, secondId, "E-6-05 halt B local");
      await editTimelineSummary(page, thirdId, "E-6-05 halt C local");
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "3",
      );

      await patchTimelineField(
        page,
        firstId,
        1,
        "timeline.activity_synopsis_text",
        "E-6-05 halt A remote",
        "e605-halt-remote-conflict",
      );

      heldPatch.release();
      await expect
        .poll(() => patchController.calls.at(-1)?.status ?? 0)
        .toBe(409);
      expect(
        patchController.calls
          .filter((call) => call.status !== 0)
          .map((call) => call.recordId),
      ).toEqual([firstId]);
      expect(successfulPatchCalls(patchController.calls)).toHaveLength(0);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
      await expect(page.getByTestId("conflict-resolver")).toBeVisible();
      await expect(page.getByTestId("conflict-server-value")).toHaveValue(
        "E-6-05 halt A remote",
      );
      await expect(page.getByTestId("conflict-local-value")).toHaveValue(
        "E-6-05 halt A local",
      );
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "2",
      );
      await expectServerSummaries(page, incidentId, {
        [firstId]: "E-6-05 halt A remote",
        [secondId]: "E-6-05 halt B base",
        [thirdId]: "E-6-05 halt C base",
      });
    } finally {
      await patchController.dispose();
    }
  });

  await test.step("does not restore the in-memory queue after a full reload", async () => {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("E605RELOAD"),
      "Phase 6 E-6-05 reload boundary",
    );
    const recordId = requireRecordId(
      await createTimelineRow(page, incidentId, "E-6-05 reload base"),
    );
    const patchController = await installPatchTransportFailureController(page);

    try {
      patchController.disconnect();
      await page.goto(`/?incident_id=${incidentId}`);
      await expect(
        page.getByTestId(
          rowCellTestId(recordId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveValue("E-6-05 reload base");
      await expectCurrentIncidentRole(page, "Current incident role: admin");

      await editTimelineSummary(page, recordId, "E-6-05 reload local");
      await expect(
        page.getByTestId(
          rowCellTestId(recordId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveValue("E-6-05 reload local");
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "1",
      );

      patchController.connect();
      page.once("dialog", async (dialog) => {
        await dialog.accept();
      });
      await page.reload();

      await expect(
        page.getByTestId(
          rowCellTestId(recordId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveValue("E-6-05 reload base");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
      expect(successfulPatchCalls(patchController.calls)).toHaveLength(0);
      await expectServerSummaries(page, incidentId, {
        [recordId]: "E-6-05 reload base",
      });
    } finally {
      await patchController.dispose();
    }
  });

  await test.step("replays queued writes in FIFO order after real session revocation and re-authentication", async () => {
    await exerciseRevokedPendingReplay({
      createdBy: "E-6-05",
      incidentKeyPrefix: "E605REVOKE",
      localValues: [
        "E-6-05 revoked A local",
        "E-6-05 revoked B local",
        "E-6-05 revoked C local",
      ],
      page,
      scenario: "revoked",
      sessionTracker,
      triggerRevocation: async ({ member }) => {
        await revokeAllSessions(
          workerAdminRequest,
          member.user_id,
          "Phase 6 E-6-05 browser revoke-all",
        );
      },
    });
  });
});

async function waitForPresenceMarkerTiming(
  page: Page,
  options: {
    actorText: string;
    fieldKey: string;
    recordId: string;
    startedAtMs: number;
    timeoutMs: number;
  },
) {
  const rowMarker = page.getByTestId(rowPresenceMarkerTestId(options.recordId));
  const cellMarker = page.getByTestId(
    cellPresenceMarkerTestId(options.recordId, options.fieldKey),
  );
  let rowMarkerDurationMs: number | null = null;
  let cellMarkerDurationMs: number | null = null;
  let lastRowText = "";
  let lastCellText = "";
  const deadline = options.startedAtMs + options.timeoutMs;

  while (performance.now() <= deadline) {
    if (rowMarkerDurationMs === null) {
      lastRowText = await locatorText(rowMarker);
      if (lastRowText.includes(options.actorText)) {
        rowMarkerDurationMs = performance.now() - options.startedAtMs;
      }
    }
    if (cellMarkerDurationMs === null) {
      lastCellText = await locatorText(cellMarker);
      if (lastCellText.includes(options.actorText)) {
        cellMarkerDurationMs = performance.now() - options.startedAtMs;
      }
    }
    if (rowMarkerDurationMs !== null && cellMarkerDurationMs !== null) {
      return {
        cellMarkerDurationMs,
        renderDurationMs: Math.max(rowMarkerDurationMs, cellMarkerDurationMs),
        rowMarkerDurationMs,
      };
    }
    await new Promise((resolve) => setTimeout(resolve, 20));
  }

  throw new Error(
    `timed out waiting for presence markers: row=${JSON.stringify(lastRowText)} cell=${JSON.stringify(lastCellText)}`,
  );
}

async function locatorText(locator: Locator) {
  return (await locator.textContent({ timeout: 100 }).catch(() => "")) ?? "";
}
