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
} from "@cartulary/test-utils/grid";
import {
  cellPresenceMarkerTestId,
  currentIncidentRoleTestId,
  gridRowGutterTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  rowCellTestId,
  rowPresenceMarkerTestId,
  saveStateTestId,
  timelineRowVersionTestId,
  timelineScalarEditorTestId,
} from "@cartulary/ui-contracts";
import type { Locator, Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { openIncidentAsTrackedUser } from "./pages/incidentDirectory";
import { applyStorageState, csrfHeaders } from "./support/auth/browserSession";
import { revokeAllSessions } from "./support/auth/sessions";
import {
  createTimelineRow,
  driveRealTimelineSummaryConflict,
  editTimelineSummary,
  exerciseRevokedPendingReplay,
  exerciseSameFieldResolver,
  expectServerSummaries,
  expectServerTimelineCells,
  focusRemoteTimelineCellAndWaitForPresence,
  installPatchController,
  installPatchTransportFailureController,
  openIncidentAsTrackedUserReady,
  patchTimelineField,
  presenceDeltaMatches,
  requireRecordId,
  successfulPatchCalls,
  summaryPatchValue,
} from "./support/collaboration/replay";
import { timelineViewSchemaId } from "./support/contracts/workbookSurfaces";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { assertRecordFieldMutationAnchor } from "./support/workbook/mutationAnchors";

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

test("Verify conflict resolver actions submit public mutations and refresh rows without losing focus or pending queue ordering.", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("COLLABORATIONINTEGRATION"),
    "integration.collaboration.row-01 conflict resolver public route",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "integration.collaboration Remote",
    email: uniqueEmail("integration.collaboration-remote"),
    initial_password: "FeIP7RemotePass!",
    role: "editor",
  });
  const conflictId = requireRecordId(
    await createTimelineRow(
      page,
      incidentId,
      "integration.collaboration conflict base",
    ),
  );
  const queuedAId = requireRecordId(
    await createTimelineRow(
      page,
      incidentId,
      "integration.collaboration queued A base",
    ),
  );
  const queuedBId = requireRecordId(
    await createTimelineRow(
      page,
      incidentId,
      "integration.collaboration queued B base",
    ),
  );
  const patchController = await installPatchController(page);

  let remotePage: Page | null = null;
  try {
    await page.goto(`/?incident_id=${incidentId}`);
    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page,
      recordId: conflictId,
      surface: timelineViewSchemaId,
    });
    await expect(
      page.getByTestId(
        rowCellTestId(conflictId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("integration.collaboration conflict base");

    remotePage = await openIncidentAsTrackedUser(browser, sessionTracker, {
      createdBy: "integration.collaboration.row-01",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "integration.collaboration public resolver remote analyst",
      userId: remote.user_id,
    });
    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page: remotePage,
      recordId: conflictId,
      surface: timelineViewSchemaId,
    });
    await expect(
      remotePage.getByTestId(
        rowCellTestId(conflictId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("integration.collaboration conflict base");

    await driveRealTimelineSummaryConflict({
      baseRowVersion: 1,
      localValue: "integration.collaboration local draft",
      page,
      patchController,
      recordId: conflictId,
      remotePatchPage: remotePage,
      remoteValue: "integration.collaboration remote saved",
      txnPrefix: "collaboration-integration-remote-conflict",
    });

    await patchTimelineField(
      remotePage,
      conflictId,
      2,
      "timeline.activity_synopsis_text",
      "integration.collaboration newer saved",
      "collaboration-integration-newer-saved",
    );

    await page
      .getByTestId("conflict-merged-value")
      .fill("integration.collaboration stale merged draft");
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
      resolved_value: "integration.collaboration stale merged draft",
    });
    const staleEnvelope = await staleResponse.json();
    expect(staleEnvelope.error.code).toBe("same_field_conflict");
    expect(staleEnvelope.error.conflict.server_value).toBe(
      "integration.collaboration newer saved",
    );
    await expect(page.getByTestId("conflict-server-value")).toHaveValue(
      "integration.collaboration newer saved",
    );
    await expect(page.getByTestId("conflict-local-value")).toHaveValue(
      "integration.collaboration stale merged draft",
    );

    await page
      .getByTestId("conflict-merged-value")
      .fill("integration.collaboration final merged");
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
      resolved_value: "integration.collaboration final merged",
    });
    await expect(page.getByTestId("conflict-resolver")).toHaveCount(0);
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expectServerTimelineCells(page, incidentId, conflictId, {
      "timeline.activity_synopsis_text":
        "integration.collaboration final merged",
    });

    const queueTransportController =
      await installPatchTransportFailureController(page);
    try {
      queueTransportController.disconnect();
      await editTimelineSummary(
        page,
        queuedAId,
        "integration.collaboration queued A local",
      );
      await editTimelineSummary(
        page,
        queuedBId,
        "integration.collaboration queued B local",
      );
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "2",
      );
      queueTransportController.connect();
      await expect
        .poll(
          () =>
            successfulPatchCalls(queueTransportController.calls).filter(
              (call) => [queuedAId, queuedBId].includes(call.recordId),
            ).length,
        )
        .toBe(2);
      const replayed = successfulPatchCalls(
        queueTransportController.calls,
      ).filter((call) => [queuedAId, queuedBId].includes(call.recordId));
      expect(replayed.map((call) => call.recordId)).toEqual([
        queuedAId,
        queuedBId,
      ]);
      expect(replayed.map((call) => summaryPatchValue(call.body))).toEqual([
        "integration.collaboration queued A local",
        "integration.collaboration queued B local",
      ]);
    } finally {
      await queueTransportController.dispose();
    }
  } finally {
    await patchController.dispose();
    await remotePage?.context().close();
  }
});

test("Verify multi-client live row update, presence anchoring, reset/invalidate handling, stale-row requery, and same-field conflict resolver through /ws/v1/ and /api/v1/.", async ({
  browser,
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("COLLABORATION"),
    "end-to-end.collaboration.row-01 live collaboration",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "end-to-end.collaboration Remote",
    email: uniqueEmail("end-to-end.collaboration-remote"),
    initial_password: "FeEP7RemotePass!",
    role: "editor",
  });
  const liveId = requireRecordId(
    await createTimelineRow(
      page,
      incidentId,
      "end-to-end.collaboration live base",
    ),
  );
  const invalidateId = requireRecordId(
    await createTimelineRow(
      page,
      incidentId,
      "end-to-end.collaboration invalidate base",
    ),
  );
  const conflictId = requireRecordId(
    await createTimelineRow(
      page,
      incidentId,
      "end-to-end.collaboration Zulu conflict base",
    ),
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
      createdBy: "end-to-end.collaboration.row-01",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "end-to-end.collaboration remote analyst",
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
      actorText: "ER",
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
      "end-to-end.collaboration live remote update",
      "collaboration-live-remote",
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
    ).toHaveText("end-to-end.collaboration live remote update");

    const removeStartAt = socketMonitor.messageCount();
    const deleteResponse = await page.request.delete(
      `${apiBase}/api/v1/records/${invalidateId}`,
      {
        headers: await csrfHeaders(page),
        data: {
          base_row_version: 1,
          client_txn_id: uniqueTxn("collaboration-delete"),
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
          client_txn_id: uniqueTxn("collaboration-restore"),
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
    ).toHaveText("end-to-end.collaboration invalidate base");

    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page,
      recordId: conflictId,
      surface: timelineViewSchemaId,
    });
    await driveRealTimelineSummaryConflict({
      baseRowVersion: 1,
      expectConflictMarker: false,
      expectEditedCellMounted: false,
      localValue: "end-to-end.collaboration local conflict",
      page,
      patchController,
      recordId: conflictId,
      remotePatchPage: remotePage,
      remoteValue: "end-to-end.collaboration remote conflict",
      txnPrefix: "collaboration-conflict",
    });
    await page.getByTestId("conflict-use-unsaved").click();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expectServerTimelineCells(page, incidentId, conflictId, {
      "timeline.activity_synopsis_text":
        "end-to-end.collaboration local conflict",
    });
  } finally {
    await patchController.dispose();
    await remotePage?.context().close();
  }

  await exerciseRevokedPendingReplay({
    createdBy: "end-to-end.collaboration.row-01",
    incidentKeyPrefix: "COLLABORATIONREVOKE",
    page,
    scenario: "revoked",
    sessionTracker,
    triggerRevocation: async ({ member }) => {
      await revokeAllSessions(
        workerAdminRequest,
        member.user_id,
        "end-to-end.collaboration.row-01 browser revoke-all",
      );
    },
  });
});

test("shows two analysts each other's workbook presence within the expected interaction window", async ({
  browser,
  page,
  sessionTracker,
}, testInfo) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("COLLABORATION-PRESENCE"),
    "Collaboration collaboration-conflict browser presence",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "Remote Analyst",
    email: uniqueEmail("collaboration-e601-remote"),
    initial_password: "CollaborationE601Remote!",
    role: "editor",
  });
  const row = await createTimelineRow(
    page,
    incidentId,
    "collaboration-conflict presence base",
  );
  const recordId = requireRecordId(row);
  const primarySocket = installIncidentSocketMonitor(page, incidentId);

  let remotePage: Page | null = null;
  try {
    await page.goto(`/?incident_id=${incidentId}`);
    await primarySocket.waitForMessage("hello_ack");
    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page,
      recordId,
      surface: timelineViewSchemaId,
    });
    await expect(
      page.getByTestId(
        rowCellTestId(recordId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("collaboration-conflict presence base");

    const remoteSession = await openIncidentAsTrackedUserReady(
      browser,
      sessionTracker,
      {
        createdBy: "collaboration-conflict",
        email: remote.email,
        incidentId,
        password: remote.initial_password,
        purpose: "Collaboration collaboration-conflict remote analyst",
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
    await remoteInput.click();
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
    await testInfo.attach("collaboration-e-6-01-presence-timing.json", {
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

test("auto-merges different-field concurrent edits and requires explicit same-field resolution", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("COLLABORATION-CONFLICT"),
    "Collaboration collaboration-conflict concurrent resolver UX",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "Collaboration collaboration-conflict Remote",
    email: uniqueEmail("collaboration-e602-remote"),
    initial_password: "CollaborationE602Remote!",
    role: "editor",
  });
  const differentId = requireRecordId(
    await createTimelineRow(
      page,
      incidentId,
      "collaboration-conflict different base",
    ),
  );
  const keepId = requireRecordId(
    await createTimelineRow(
      page,
      incidentId,
      "collaboration-conflict keep base",
    ),
  );
  const useId = requireRecordId(
    await createTimelineRow(
      page,
      incidentId,
      "collaboration-conflict use base",
    ),
  );
  const mergedId = requireRecordId(
    await createTimelineRow(
      page,
      incidentId,
      "collaboration-conflict merged base",
    ),
  );
  const patchController = await installPatchController(page);

  let remotePage: Page | null = null;
  try {
    const socketMonitor = installIncidentSocketMonitor(page, incidentId);
    await page.goto(`/?incident_id=${incidentId}`);
    await socketMonitor.waitForMessage("hello_ack");
    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page,
      recordId: differentId,
      surface: timelineViewSchemaId,
    });
    await expect(
      page.getByTestId(
        rowCellTestId(differentId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("collaboration-conflict different base");

    remotePage = await openIncidentAsTrackedUser(browser, sessionTracker, {
      createdBy: "collaboration-conflict",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "Collaboration collaboration-conflict remote analyst",
      userId: remote.user_id,
    });
    await scrollGridCellIntoView({
      cellKey: "timeline.activity_synopsis_text",
      page: remotePage,
      recordId: differentId,
      surface: timelineViewSchemaId,
    });
    await expect(
      remotePage.getByTestId(
        rowCellTestId(differentId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("collaboration-conflict different base");

    await patchTimelineField(
      page,
      differentId,
      1,
      "timeline.activity_synopsis_text",
      "collaboration-conflict different primary",
      "e602-different-primary-summary",
    );
    await patchTimelineField(
      remotePage,
      differentId,
      1,
      "timeline.raw_activity_text",
      "collaboration-conflict different remote details",
      "e602-different-remote-details",
    );
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(
      page.getByTestId(
        rowCellTestId(differentId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("collaboration-conflict different primary");
    await expect(
      remotePage.getByTestId(
        rowCellTestId(differentId, "timeline.activity_synopsis_text"),
      ),
    ).toHaveText("collaboration-conflict different primary");
    await expectServerTimelineCells(page, incidentId, differentId, {
      "timeline.raw_activity_text":
        "collaboration-conflict different remote details",
      "timeline.activity_synopsis_text":
        "collaboration-conflict different primary",
    });

    await exerciseSameFieldResolver({
      action: "keep_saved",
      expectedPrimary: "collaboration-conflict keep remote",
      localValue: "collaboration-conflict keep primary",
      page,
      incidentId,
      patchController,
      recordId: keepId,
      remotePage,
      remoteValue: "collaboration-conflict keep remote",
    });

    await exerciseSameFieldResolver({
      action: "use_unsaved",
      expectedPrimary: "collaboration-conflict use primary",
      localValue: "collaboration-conflict use primary",
      page,
      incidentId,
      patchController,
      recordId: useId,
      remotePage,
      remoteValue: "collaboration-conflict use remote",
    });

    await exerciseSameFieldResolver({
      action: "merged_value",
      expectedPrimary: "collaboration-conflict merged final",
      localValue: "collaboration-conflict merged primary",
      mergedValue: "collaboration-conflict merged final",
      page,
      incidentId,
      patchController,
      recordId: mergedId,
      remotePage,
      remoteValue: "collaboration-conflict merged remote",
    });
  } finally {
    await patchController.dispose();
    await remotePage?.context().close();
  }
});

test("keeps live updates conflict markers and presence markers anchored to record_id and field_key", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("COLLABORATION-ANCHOR"),
    "Collaboration collaboration-conflict live-cell anchoring",
  );
  const remote = await createIncidentMemberUser(page, incidentId, {
    display_name: "Anchor Analyst",
    email: uniqueEmail("collaboration-e604-anchor"),
    initial_password: "CollaborationE604Anchor!",
    role: "editor",
  });
  for (let index = 0; index < 24; index += 1) {
    await createTimelineRow(
      page,
      incidentId,
      `collaboration-conflict Filler ${String(index).padStart(2, "0")}`,
    );
  }
  const commandRows = [];
  for (const scenario of gridAnchorCommandScenarios(timelineViewSchemaId)) {
    const sortLabel = String.fromCharCode(65 + commandRows.length);
    commandRows.push({
      baseSummary: `collaboration-conflict ${sortLabel} ${scenario.name} command base`,
      recordId: requireRecordId(
        await createTimelineRow(
          page,
          incidentId,
          `collaboration-conflict ${sortLabel} ${scenario.name} command base`,
        ),
      ),
      scenario,
      sortLabel,
    });
  }
  const alphaRow = await createTimelineRow(
    page,
    incidentId,
    "collaboration-conflict Zulu anchor base",
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
      await expect(input).toHaveText(baseSummary);

      await patchTimelineField(
        page,
        recordId,
        1,
        "timeline.raw_activity_text",
        `collaboration-conflict ${scenario.name} remote details`,
        `e604-${scenario.name}-live-patch`,
      );
      await expect(
        page.getByTestId(timelineRowVersionTestId(recordId)),
      ).toHaveText("2");

      const expectedValue = `collaboration-conflict ${sortLabel} ${scenario.name} anchored local`;
      const heldPatch = patchController.holdNextPatch();
      await input.click();
      const editor = page.getByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.activity_synopsis_text",
          recordId,
          surface: "grid",
        }),
      );
      await editor.fill(expectedValue);
      await scenario.commit({
        input: editor,
        page,
        surface: timelineViewSchemaId,
      });
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
    await expect(alphaInput).toHaveText(
      "collaboration-conflict Zulu anchor base",
    );

    remotePage = await openIncidentAsTrackedUser(browser, sessionTracker, {
      createdBy: "collaboration-conflict",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "Collaboration collaboration-conflict remote anchor analyst",
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
      localValue: "collaboration-conflict Alpha local",
      page,
      patchController,
      recordId: alphaId,
      remoteValue: "collaboration-conflict Alpha remote",
      txnPrefix: "e604-alpha-remote-conflict",
    });
    await scrollGridToOffset(page, timelineViewSchemaId, 0);
    await assertMarkerAnchoredToGridTarget({
      anchorKind: "row-gutter",
      markerTestId: rowPresenceMarkerTestId(alphaId),
      page,
      surface: timelineViewSchemaId,
      targetTestId: gridRowGutterTestId(timelineViewSchemaId, alphaId),
    });
    await expect(
      page.getByTestId(
        timelineScalarEditorTestId({
          fieldKey: "timeline.activity_synopsis_text",
          recordId: alphaId,
          surface: "grid",
        }),
      ),
    ).toHaveValue("collaboration-conflict Alpha local");
  } finally {
    await patchController.dispose();
    await remotePage?.context().close();
  }
});

test("replays queued unsent writes after re-authentication without silent reload restore", async ({
  page,
  sessionTracker,
  workerAdmin,
  workerAdminRequest,
}) => {
  await test.step("replays transient browser request failures in FIFO order after transport recovery", async () => {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("COLLABORATION-FIFO"),
      "Collaboration collaboration-conflict FIFO recovery",
    );
    const firstId = requireRecordId(
      await createTimelineRow(
        page,
        incidentId,
        "collaboration-conflict FIFO A base",
      ),
    );
    const secondId = requireRecordId(
      await createTimelineRow(
        page,
        incidentId,
        "collaboration-conflict FIFO B base",
      ),
    );
    const thirdId = requireRecordId(
      await createTimelineRow(
        page,
        incidentId,
        "collaboration-conflict FIFO C base",
      ),
    );
    const patchController = await installPatchTransportFailureController(page);

    try {
      patchController.disconnect();
      await page.goto(`/?incident_id=${incidentId}`);
      await scrollGridCellIntoView({
        cellKey: "timeline.activity_synopsis_text",
        page,
        recordId: firstId,
        surface: timelineViewSchemaId,
      });
      await expect(
        page.getByTestId(
          rowCellTestId(firstId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveText("collaboration-conflict FIFO A base");
      await expectCurrentIncidentRole(page, "Current incident role: admin");

      await editTimelineSummary(
        page,
        firstId,
        "collaboration-conflict FIFO A local",
      );
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();
      await editTimelineSummary(
        page,
        secondId,
        "collaboration-conflict FIFO B local",
      );
      await editTimelineSummary(
        page,
        thirdId,
        "collaboration-conflict FIFO C local",
      );

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
        "collaboration-conflict FIFO A local",
        "collaboration-conflict FIFO B local",
        "collaboration-conflict FIFO C local",
      ]);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expectServerSummaries(page, incidentId, {
        [firstId]: "collaboration-conflict FIFO A local",
        [secondId]: "collaboration-conflict FIFO B local",
        [thirdId]: "collaboration-conflict FIFO C local",
      });
    } finally {
      await patchController.dispose();
    }
  });

  await test.step("replays queued writes in FIFO order after real HTTP auth failure and re-authentication", async () => {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("COLLABORATION-AUTH-RECOVERY"),
      "Collaboration collaboration-conflict HTTP auth recovery",
    );
    const member = await createIncidentMemberUser(page, incidentId, {
      display_name: "Collaboration collaboration-conflict HTTP Auth Analyst",
      email: uniqueEmail("collaboration-e605-http-auth"),
      initial_password: "CollaborationE605HttpAuth!",
      role: "editor",
    });
    const firstId = requireRecordId(
      await createTimelineRow(
        page,
        incidentId,
        "collaboration-conflict auth A base",
      ),
    );
    const secondId = requireRecordId(
      await createTimelineRow(
        page,
        incidentId,
        "collaboration-conflict auth B base",
      ),
    );
    const patchController = await installPatchController(page);

    try {
      await sessionTracker.loginTrackedUser(page, {
        createdBy: "collaboration-conflict",
        email: member.email,
        password: member.initial_password,
        purpose:
          "Collaboration collaboration-conflict HTTP auth analyst runtime",
        userId: member.user_id,
      });
      await page.goto(`/?incident_id=${incidentId}`);
      await scrollGridCellIntoView({
        cellKey: "timeline.activity_synopsis_text",
        page,
        recordId: firstId,
        surface: timelineViewSchemaId,
      });
      await expect(
        page.getByTestId(
          rowCellTestId(firstId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveText("collaboration-conflict auth A base");
      await expectCurrentIncidentRole(page, "Current incident role: editor");

      await page.context().clearCookies();
      await editTimelineSummary(
        page,
        firstId,
        "collaboration-conflict auth A local",
      );
      await expect
        .poll(() => patchController.calls.at(-1)?.status ?? 0)
        .toBe(401);
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();
      await editTimelineSummary(
        page,
        secondId,
        "collaboration-conflict auth B local",
      );
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "2",
      );

      await sessionTracker.loginTrackedUser(page, {
        createdBy: "collaboration-conflict",
        email: member.email,
        password: member.initial_password,
        purpose:
          "Collaboration collaboration-conflict HTTP auth analyst re-authentication",
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
        "collaboration-conflict auth A local",
        "collaboration-conflict auth B local",
      ]);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expectServerSummaries(page, incidentId, {
        [firstId]: "collaboration-conflict auth A local",
        [secondId]: "collaboration-conflict auth B local",
      });
    } finally {
      await patchController.dispose();
      await applyStorageState(page, workerAdmin.storageState);
    }
  });

  await test.step("halts replay on the first real same-field conflict and retains later queued writes", async () => {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("COLLABORATION-CONFLICT-HALT"),
      "Collaboration collaboration-conflict same-field conflict halt",
    );
    const firstId = requireRecordId(
      await createTimelineRow(
        page,
        incidentId,
        "collaboration-conflict halt A base",
      ),
    );
    const secondId = requireRecordId(
      await createTimelineRow(
        page,
        incidentId,
        "collaboration-conflict halt B base",
      ),
    );
    const thirdId = requireRecordId(
      await createTimelineRow(
        page,
        incidentId,
        "collaboration-conflict halt C base",
      ),
    );
    const transportController =
      await installPatchTransportFailureController(page);

    try {
      transportController.disconnect();
      await page.goto(`/?incident_id=${incidentId}`);
      await scrollGridCellIntoView({
        cellKey: "timeline.activity_synopsis_text",
        page,
        recordId: firstId,
        surface: timelineViewSchemaId,
      });
      await expect(
        page.getByTestId(
          rowCellTestId(firstId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveText("collaboration-conflict halt A base");
      await expectCurrentIncidentRole(page, "Current incident role: admin");

      await editTimelineSummary(
        page,
        firstId,
        "collaboration-conflict halt A local",
      );
      await editTimelineSummary(
        page,
        secondId,
        "collaboration-conflict halt B local",
      );
      await editTimelineSummary(
        page,
        thirdId,
        "collaboration-conflict halt C local",
      );
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "3",
      );

      const conflictController = await installPatchController(page);
      try {
        const heldPatch = conflictController.holdNextPatch();
        await transportController.dispose();
        await heldPatch.waitForHit;
        await patchTimelineField(
          page,
          firstId,
          1,
          "timeline.activity_synopsis_text",
          "collaboration-conflict halt A remote",
          "e605-halt-remote-conflict",
        );
        heldPatch.release();
        await expect
          .poll(() => conflictController.calls.at(-1)?.status ?? 0)
          .toBe(409);
        expect(successfulPatchCalls(conflictController.calls)).toHaveLength(0);
        await expect(page.getByTestId(saveStateTestId())).toHaveText(
          "Conflict",
        );
        await expect(page.getByTestId("conflict-resolver")).toBeVisible();
        await expect(page.getByTestId("conflict-server-value")).toHaveValue(
          "collaboration-conflict halt A remote",
        );
        await expect(page.getByTestId("conflict-local-value")).toHaveValue(
          "collaboration-conflict halt A local",
        );
        await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
          "2",
        );
        await expectServerSummaries(page, incidentId, {
          [firstId]: "collaboration-conflict halt A remote",
          [secondId]: "collaboration-conflict halt B base",
          [thirdId]: "collaboration-conflict halt C base",
        });
      } finally {
        await conflictController.dispose();
      }
    } finally {
      await transportController.dispose();
    }
  });

  await test.step("does not restore the in-memory queue after a full reload", async () => {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("COLLABORATION-RELOAD"),
      "Collaboration collaboration-conflict reload boundary",
    );
    const recordId = requireRecordId(
      await createTimelineRow(
        page,
        incidentId,
        "collaboration-conflict reload base",
      ),
    );
    const patchController = await installPatchTransportFailureController(page);

    try {
      patchController.disconnect();
      await page.goto(`/?incident_id=${incidentId}`);
      await scrollGridCellIntoView({
        cellKey: "timeline.activity_synopsis_text",
        page,
        recordId,
        surface: timelineViewSchemaId,
      });
      await expect(
        page.getByTestId(
          rowCellTestId(recordId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveText("collaboration-conflict reload base");
      await expectCurrentIncidentRole(page, "Current incident role: admin");

      await editTimelineSummary(
        page,
        recordId,
        "collaboration-conflict reload local",
      );
      await expect(
        page.getByTestId(
          rowCellTestId(recordId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveText("collaboration-conflict reload local");
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "1",
      );

      patchController.connect();
      page.once("dialog", async (dialog) => {
        await dialog.accept();
      });
      await page.reload();

      await scrollGridCellIntoView({
        cellKey: "timeline.activity_synopsis_text",
        page,
        recordId,
        surface: timelineViewSchemaId,
      });
      await expect(
        page.getByTestId(
          rowCellTestId(recordId, "timeline.activity_synopsis_text"),
        ),
      ).toHaveText("collaboration-conflict reload base");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
      expect(successfulPatchCalls(patchController.calls)).toHaveLength(0);
      await expectServerSummaries(page, incidentId, {
        [recordId]: "collaboration-conflict reload base",
      });
    } finally {
      await patchController.dispose();
    }
  });

  await test.step("replays queued writes in FIFO order after real session revocation and re-authentication", async () => {
    await exerciseRevokedPendingReplay({
      createdBy: "collaboration-conflict",
      incidentKeyPrefix: "E605REVOKE",
      localValues: [
        "collaboration-conflict revoked A local",
        "collaboration-conflict revoked B local",
        "collaboration-conflict revoked C local",
      ],
      page,
      scenario: "revoked",
      sessionTracker,
      triggerRevocation: async ({ member }) => {
        await revokeAllSessions(
          workerAdminRequest,
          member.user_id,
          "Collaboration collaboration-conflict browser revoke-all",
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
