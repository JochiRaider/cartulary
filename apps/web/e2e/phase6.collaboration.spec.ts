import {
  applyFilterChip,
  assertMarkerAnchoredToGridTarget,
  assertMountedGridRowCountAtMost,
  assertRecordFieldMutationAnchor,
  changeGrouping,
  scrollGridToBottom,
  scrollGridToOffset,
  sortByHeader,
} from "@cartulary/test-utils";
import {
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  rowCellTestId,
  rowPresenceMarkerTestId,
  saveStateTestId,
} from "@cartulary/ui-contracts";
import type { Page, Route, WebSocket } from "@playwright/test";
import { request } from "@playwright/test";

import { revokeAllSessions } from "./authRuntime";
import { expect, test } from "./fixtures";
import {
  apiBase,
  applyStorageState,
  createIncident,
  createIncidentMemberUser,
  createViewRow,
  csrfHeaders,
  openIncidentAsTrackedUser,
  queryViewRows,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";

const timelineViewSchemaId = "cartulary.view.timeline.v1";

type PatchCall = {
  body: Record<string, unknown>;
  recordId: string;
  status: number;
};

type PatchBehavior =
  | {
      hold: Promise<void>;
      release: () => void;
      resolveHit: (call: PatchCall) => void;
      type: "hold";
      waitForHit: Promise<PatchCall>;
    }
  | {
      code: string;
      status: number;
      type: "error";
    };

type SocketMessage = {
  payload: Record<string, unknown>;
  socketIndex: number;
  type: string;
};

type SessionTracker = {
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

test("E-6-01 shows two analysts each other's workbook presence within the expected interaction window", async ({
  browser,
  page,
  sessionTracker,
}) => {
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
    await expect(page.getByTestId(`row-${recordId}-summary`)).toHaveValue(
      "E-6-01 presence base",
    );

    remotePage = await openIncidentAsTrackedUser(browser, sessionTracker, {
      createdBy: "E-6-01",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "Phase 6 E-6-01 remote analyst",
      userId: remote.user_id,
    });
    await primarySocket.waitForMessage("presence_delta");
    await expect(page.getByTestId("presence-header")).toContainText("RA");

    const remoteInput = remotePage.getByTestId(`row-${recordId}-summary`);
    await remoteInput.focus();
    await expect(
      page.getByTestId(rowPresenceMarkerTestId(recordId)),
    ).toContainText("RA");
    await expect(
      page.getByTestId(cellPresenceMarkerTestId(recordId, "timeline.summary")),
    ).toContainText("RA");
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
    await expect(page.getByTestId(`row-${differentId}-summary`)).toHaveValue(
      "E-6-02 different base",
    );

    remotePage = await openIncidentAsTrackedUser(browser, sessionTracker, {
      createdBy: "E-6-02",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "Phase 6 E-6-02 remote analyst",
      userId: remote.user_id,
    });
    await expect(
      remotePage.getByTestId(`row-${differentId}-summary`),
    ).toHaveValue("E-6-02 different base");

    await patchTimelineField(
      page,
      differentId,
      1,
      "timeline.summary",
      "E-6-02 different primary",
      "e602-different-primary-summary",
    );
    await patchTimelineField(
      remotePage,
      differentId,
      1,
      "timeline.details",
      "E-6-02 different remote details",
      "e602-different-remote-details",
    );
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(page.getByTestId(`row-${differentId}-summary`)).toHaveValue(
      "E-6-02 different primary",
    );
    await expect(
      remotePage.getByTestId(`row-${differentId}-summary`),
    ).toHaveValue("E-6-02 different primary");
    await expectServerTimelineCells(page, incidentId, differentId, {
      "timeline.details": "E-6-02 different remote details",
      "timeline.summary": "E-6-02 different primary",
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

test("E-6-03 preserves unsaved local work after socket revocation and re-authentication", async ({
  page,
  sessionTracker,
  workerAdmin,
  workerAdminRequest,
}) => {
  await test.step("deployment-admin revoke-all preserves and replays local work", async () => {
    await exerciseRevokedPendingReplay({
      createdBy: "E-6-03",
      incidentKeyPrefix: "E603REVOKE",
      page,
      scenario: "revoke-all",
      sessionTracker,
      triggerRevocation: async ({ member }) => {
        await revokeAllSessions(
          workerAdminRequest,
          member.user_id,
          "Phase 6 E-6-03 browser revoke-all",
        );
      },
    });
    await applyStorageState(page, workerAdmin.storageState);
  });

  await test.step("current-session logout preserves and replays local work", async () => {
    await exerciseRevokedPendingReplay({
      createdBy: "E-6-03",
      incidentKeyPrefix: "E603LOGOUT",
      page,
      scenario: "logout",
      sessionTracker,
      triggerRevocation: async () => {
        const response = await page.request.post(
          `${apiBase}/api/v1/auth/logout`,
          {
            headers: await csrfHeaders(page),
            data: {},
          },
        );
        expect(response.ok()).toBeTruthy();
      },
    });
    await applyStorageState(page, workerAdmin.storageState);
  });

  await test.step("concurrency-limit revocation preserves and replays local work", async () => {
    await exerciseRevokedPendingReplay({
      createdBy: "E-6-03",
      incidentKeyPrefix: "E603CONCURRENCY",
      page,
      scenario: "concurrency",
      sessionTracker,
      triggerRevocation: async ({ member }) => {
        await createUntrackedLoginSessions(
          member.email,
          member.initial_password,
          5,
        );
      },
    });
    await applyStorageState(page, workerAdmin.storageState);
  });
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
  const betaRow = await createTimelineRow(page, incidentId, "E-6-04 Beta base");
  const alphaRow = await createTimelineRow(
    page,
    incidentId,
    "E-6-04 Zulu anchor base",
  );
  const alphaId = requireRecordId(alphaRow);
  const betaId = requireRecordId(betaRow);
  const patchController = await installPatchController(page);

  let remotePage: Page | null = null;
  try {
    const socketMonitor = installIncidentSocketMonitor(page, incidentId);
    await page.goto(`/?incident_id=${incidentId}`);
    await socketMonitor.waitForMessage("hello_ack");
    await sortByHeader(page, "timeline", "timeline.summary");
    await expect(page.getByTestId(`row-${betaId}-summary`)).toHaveValue(
      "E-6-04 Beta base",
    );
    await page.getByTestId(`row-${betaId}-mark-reviewed`).click();
    await expect(page.getByTestId(`row-${betaId}-capture-state`)).toHaveText(
      "reviewed",
    );
    await applyFilterChip(page, "timeline", "timeline.has_evidence", "false");
    await changeGrouping(page, "timeline", "timeline.capture_state");

    const betaPatch = await page.request.patch(
      `${apiBase}/api/v1/records/${betaId}`,
      {
        headers: await csrfHeaders(page),
        data: {
          view_schema_id: timelineViewSchemaId,
          base_row_version: 2,
          client_txn_id: uniqueTxn("e604-beta-live-patch"),
          changes: [
            {
              field_key: "timeline.details",
              value: "E-6-04 remote details",
            },
          ],
        },
      },
    );
    expect(betaPatch.ok()).toBeTruthy();
    await expect(page.getByTestId(`row-${betaId}-row-version`)).toHaveText("3");

    const betaInput = page.getByTestId(`row-${betaId}-summary`);
    const pastePatch = patchController.holdNextPatch();
    await betaInput.focus();
    await betaInput.fill("E-6-04 pasted beta anchor");
    await betaInput.dispatchEvent("paste");
    const pasteCall = await pastePatch.waitForHit;
    assertRecordFieldMutationAnchor({
      actualRecordId: pasteCall.recordId,
      body: pasteCall.body,
      expectedRecordId: betaId,
      expectedValue: "E-6-04 pasted beta anchor",
      fieldKey: "timeline.summary",
    });
    pastePatch.release();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expectServerSummaries(page, incidentId, {
      [betaId]: "E-6-04 pasted beta anchor",
    });

    await scrollGridToBottom(page, "timeline");
    await assertMountedGridRowCountAtMost({
      maxRows: 18,
      page,
      surface: "timeline",
    });
    const alphaInput = page.getByTestId(`row-${alphaId}-summary`);
    await expect(alphaInput).toHaveValue("E-6-04 Zulu anchor base");

    remotePage = await openIncidentAsTrackedUser(browser, sessionTracker, {
      createdBy: "E-6-04",
      email: remote.email,
      incidentId,
      password: remote.initial_password,
      purpose: "Phase 6 E-6-04 remote anchor analyst",
      userId: remote.user_id,
    });
    await sortByHeader(remotePage, "timeline", "timeline.summary");
    await applyFilterChip(
      remotePage,
      "timeline",
      "timeline.has_evidence",
      "false",
    );
    await changeGrouping(remotePage, "timeline", "timeline.capture_state");
    await scrollGridToBottom(remotePage, "timeline");
    await remotePage.getByTestId(`row-${alphaId}-summary`).focus();
    await socketMonitor.waitForMessage("presence_delta");
    await expect(
      page.getByTestId(rowPresenceMarkerTestId(alphaId)),
    ).toContainText("AA");
    await expect(
      page.getByTestId(cellPresenceMarkerTestId(alphaId, "timeline.summary")),
    ).toContainText("AA");

    const heldPatch = patchController.holdNextPatch();
    await alphaInput.fill("E-6-04 Alpha local");
    await alphaInput.press("Enter");
    await heldPatch.waitForHit;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");

    const alphaPatch = await page.request.patch(
      `${apiBase}/api/v1/records/${alphaId}`,
      {
        headers: await csrfHeaders(page),
        data: {
          view_schema_id: timelineViewSchemaId,
          base_row_version: 1,
          client_txn_id: uniqueTxn("e604-alpha-remote-conflict"),
          changes: [
            {
              field_key: "timeline.summary",
              value: "E-6-04 Alpha remote",
            },
          ],
        },
      },
    );
    expect(alphaPatch.ok()).toBeTruthy();
    await expect(alphaInput).toHaveValue("E-6-04 Alpha local");

    heldPatch.release();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
    await scrollGridToOffset(page, "timeline", 0);
    await expect(
      page.getByTestId(conflictMarkerTestId(alphaId, "timeline.summary")),
    ).toBeVisible();
    await assertMarkerAnchoredToGridTarget({
      markerTestId: conflictMarkerTestId(alphaId, "timeline.summary"),
      page,
      surface: "timeline",
      targetTestId: rowCellTestId(alphaId, "summary"),
    });
    await assertMarkerAnchoredToGridTarget({
      markerTestId: rowPresenceMarkerTestId(alphaId),
      page,
      surface: "timeline",
      targetTestId: rowCellTestId(alphaId, "capture-state"),
    });
    await assertMarkerAnchoredToGridTarget({
      markerTestId: cellPresenceMarkerTestId(alphaId, "timeline.summary"),
      page,
      surface: "timeline",
      targetTestId: rowCellTestId(alphaId, "summary"),
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
      await expect(page.getByTestId(`row-${firstId}-summary`)).toHaveValue(
        "E-6-05 FIFO A base",
      );
      await expect(
        page.getByText("Current incident role: admin"),
      ).toBeVisible();

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

  await test.step("halts on the first blocking non-retryable failure", async () => {
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("E605HALT"),
      "Phase 6 E-6-05 non-retryable halt",
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
    const sessionGate = await installAuthSessionGate(page);

    try {
      patchController.failNextPatch(401, "session_required");
      patchController.failNextPatch(400, "invalid_mutation_payload");
      await page.goto(`/?incident_id=${incidentId}`);
      await expect(page.getByTestId(`row-${firstId}-summary`)).toHaveValue(
        "E-6-05 halt A base",
      );
      await expect(
        page.getByText("Current incident role: admin"),
      ).toBeVisible();
      sessionGate.close();

      await editTimelineSummary(page, firstId, "E-6-05 halt A local");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();
      await editTimelineSummary(page, secondId, "E-6-05 halt B local");
      await editTimelineSummary(page, thirdId, "E-6-05 halt C local");
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "3",
      );

      sessionGate.open();
      await expect.poll(() => patchController.calls.length).toBe(2);
      expect(patchController.calls.map((call) => call.recordId)).toEqual([
        firstId,
        firstId,
      ]);
      expect(successfulPatchCalls(patchController.calls)).toHaveLength(0);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toContainText(
        "invalid_mutation_payload",
      );
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "3",
      );
      await expectServerSummaries(page, incidentId, {
        [firstId]: "E-6-05 halt A base",
        [secondId]: "E-6-05 halt B base",
        [thirdId]: "E-6-05 halt C base",
      });
    } finally {
      await patchController.dispose();
      await sessionGate.dispose();
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
    const patchController = await installPatchController(page);
    const sessionGate = await installAuthSessionGate(page);

    try {
      patchController.failNextPatch(401, "session_required");
      await page.goto(`/?incident_id=${incidentId}`);
      await expect(page.getByTestId(`row-${recordId}-summary`)).toHaveValue(
        "E-6-05 reload base",
      );
      await expect(
        page.getByText("Current incident role: admin"),
      ).toBeVisible();
      sessionGate.close();

      await editTimelineSummary(page, recordId, "E-6-05 reload local");
      await expect(page.getByTestId(`row-${recordId}-summary`)).toHaveValue(
        "E-6-05 reload local",
      );
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "1",
      );

      sessionGate.open();
      page.once("dialog", async (dialog) => {
        await dialog.accept();
      });
      await page.reload();

      await expect(page.getByTestId(`row-${recordId}-summary`)).toHaveValue(
        "E-6-05 reload base",
      );
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
      expect(successfulPatchCalls(patchController.calls)).toHaveLength(0);
      await expectServerSummaries(page, incidentId, {
        [recordId]: "E-6-05 reload base",
      });
    } finally {
      await patchController.dispose();
      await sessionGate.dispose();
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

async function createTimelineRow(
  page: Page,
  incidentId: string,
  summary: string,
) {
  return createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("phase6-timeline-row"),
    "timeline.summary": summary,
  });
}

async function editTimelineSummary(
  page: Page,
  recordId: string,
  value: string,
) {
  const input = page.getByTestId(`row-${recordId}-summary`);
  await input.fill(value);
  await input.press("Enter");
  await expect(input).toHaveValue(value);
}

async function exerciseRevokedPendingReplay({
  createdBy,
  incidentKeyPrefix,
  localValues,
  page,
  scenario,
  sessionTracker,
  triggerRevocation,
}: {
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
}) {
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
    await socketMonitor.waitForMessage("hello_ack");
    await expect(
      page.getByTestId(`row-${firstReplayItem.recordId}-summary`),
    ).toHaveValue(`Phase 6 ${createdBy} ${scenario} 1 base`);
    await expect(page.getByText("Current incident role: editor")).toBeVisible();

    const heldPatch = patchController.holdNextPatch();
    await editTimelineSummary(
      page,
      firstReplayItem.recordId,
      firstReplayItem.value,
    );
    await heldPatch.waitForHit;
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");

    await triggerRevocation({ incidentId, member });
    await socketMonitor.waitForMessage("session_revoked");
    await socketMonitor.waitForClose(0);

    heldPatch.release();
    await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();

    for (const item of replayItems.slice(1)) {
      await editTimelineSummary(page, item.recordId, item.value);
    }
    for (const item of replayItems) {
      await expect(
        page.getByTestId(`row-${item.recordId}-summary`),
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
    await socketMonitor.waitForMessage("hello_ack", {
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

async function patchTimelineField(
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

async function exerciseSameFieldResolver({
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
  const heldPrimaryPatch = patchController.holdNextPatch();
  await editTimelineSummary(page, recordId, localValue);
  await heldPrimaryPatch.waitForHit;
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
  await patchTimelineField(
    remotePage,
    recordId,
    1,
    "timeline.summary",
    remoteValue,
    `e602-remote-${recordId}`,
  );
  await expect(remotePage.getByTestId(`row-${recordId}-summary`)).toHaveValue(
    remoteValue,
  );
  heldPrimaryPatch.release();
  await expect.poll(() => patchController.calls.at(-1)?.status ?? 0).toBe(409);
  await expect(page.getByTestId(saveStateTestId())).toHaveText("Conflict");
  await expect(page.getByTestId("conflict-resolver")).toBeVisible();
  await expect(page.getByTestId(`row-${recordId}-summary`)).toHaveValue(
    localValue,
  );
  await expect(page.getByTestId("conflict-server-value")).toHaveValue(
    remoteValue,
  );
  await expect(page.getByTestId("conflict-local-value")).toHaveValue(
    localValue,
  );
  await expect(page.getByTestId("timeline-grid-shell")).toBeVisible();

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
    "timeline.summary": expectedPrimary,
  });
  await expect(page.getByTestId(`row-${recordId}-summary`)).toHaveValue(
    expectedPrimary,
  );
  await expect(remotePage.getByTestId(`row-${recordId}-summary`)).toHaveValue(
    expectedPrimary,
  );
  await expect(page.getByTestId(`row-${recordId}-summary`)).toBeFocused();
}

function installIncidentSocketMonitor(page: Page, incidentId: string) {
  const messages: SocketMessage[] = [];
  const closes: number[] = [];
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
      closes.push(socketIndex);
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
    waitForClose: (socketIndex: number, timeoutMs = 10_000) => {
      if (closes.includes(socketIndex)) {
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
              new Error(`timed out waiting for socket ${socketIndex} close`),
            );
          }, timeoutMs),
        };
        closeWaiters.push(waiter);
      });
    },
    waitForMessage: (
      type: string,
      options: { startAt?: number; timeoutMs?: number } = {},
    ) => {
      const startAt = options.startAt ?? 0;
      const existing = messages
        .slice(startAt)
        .find((message) => message.type === type);
      if (existing) {
        return Promise.resolve(existing);
      }
      return new Promise<SocketMessage>((resolve, reject) => {
        const waiter = {
          matches: (message: SocketMessage) =>
            messages.indexOf(message) >= startAt && message.type === type,
          reject,
          resolve,
          timeout: setTimeout(() => {
            messageWaiters.splice(messageWaiters.indexOf(waiter), 1);
            reject(new Error(`timed out waiting for socket message ${type}`));
          }, options.timeoutMs ?? 10_000),
        };
        messageWaiters.push(waiter);
      });
    },
  };
}

async function installPatchController(page: Page) {
  const calls: PatchCall[] = [];
  const behaviors: PatchBehavior[] = [];
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
    const behavior = behaviors.shift() ?? null;
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
  };

  await page.route(routePattern, handler);

  return {
    calls,
    dispose: async () => {
      await page.unroute(routePattern, handler);
    },
    failNextPatch: (status: number, code: string) => {
      behaviors.push({ code, status, type: "error" });
    },
    holdNextPatch: () => {
      let releaseHold!: () => void;
      let resolveHit!: (call: PatchCall) => void;
      const waitForHit = new Promise<PatchCall>((resolve) => {
        resolveHit = resolve;
      });
      const hold = new Promise<void>((resolve) => {
        releaseHold = resolve;
      });
      const behavior = {
        hold,
        release: releaseHold,
        resolveHit,
        type: "hold" as const,
        waitForHit,
      };
      behaviors.push(behavior);
      return {
        release: releaseHold,
        waitForHit,
      };
    },
  };
}

async function installPatchTransportFailureController(page: Page) {
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
      await page.unroute(routePattern, handler);
    },
  };
}

async function createUntrackedLoginSessions(
  email: string,
  password: string,
  count: number,
) {
  for (let index = 0; index < count; index += 1) {
    const anonymousRequests = await request.newContext({ baseURL: apiBase });
    try {
      const response = await anonymousRequests.post("/api/v1/auth/login", {
        data: {
          username: email,
          password,
        },
      });
      if (!response.ok()) {
        throw new Error(
          `untracked login ${index + 1} failed for ${email}: ${await response.text()}`,
        );
      }
    } finally {
      await anonymousRequests.dispose();
    }
  }
}

async function installAuthSessionGate(page: Page) {
  let open = true;
  const routePattern = "**/api/v1/auth/session";
  const handler = async (route: Route) => {
    if (route.request().method().toUpperCase() !== "GET" || open) {
      await route.fallback();
      return;
    }
    await fulfillJSONError(route, 401, "session_required");
  };
  await page.route(routePattern, handler);
  return {
    close: () => {
      open = false;
    },
    dispose: async () => {
      await page.unroute(routePattern, handler);
    },
    open: () => {
      open = true;
    },
  };
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
      socketIndex,
      type: parsed.type,
    };
  } catch {
    return null;
  }
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

function successfulPatchCalls(calls: PatchCall[]) {
  return calls.filter((call) => call.status >= 200 && call.status < 300);
}

function summaryPatchValue(body: Record<string, unknown>) {
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

function requireRecordId(row: Record<string, unknown>) {
  if (typeof row.record_id !== "string") {
    throw new Error(`missing record_id in row ${JSON.stringify(row)}`);
  }
  return row.record_id;
}

async function expectServerSummaries(
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

async function expectServerTimelineCells(
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
