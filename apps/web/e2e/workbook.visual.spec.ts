import { Buffer } from "node:buffer";
import {
  assertMarkerAnchoredToGridTarget,
  changeGrouping,
} from "@cartulary/test-utils";
import {
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  gridGroupRowTestId,
  gridScrollportSelector,
  gridShellTestId,
  pendingQueueNoticeTestId,
  rowCellTestId,
  rowInspectButtonTestId,
  rowPresenceMarkerTestId,
  saveStateTestId,
} from "@cartulary/ui-contracts";
import type { Page, Route } from "@playwright/test";
import { expect, test } from "./fixtures";
import {
  createIncident,
  createIncidentMemberUser,
  createViewRow,
  holdBrowserApiRequest,
  openIncidentAsTrackedUser,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  addRelationshipTokenViaUI,
  collectionActionsPayload,
  evidenceViewSchemaId,
  hostRefsFieldKey,
  hostsViewSchemaId,
  timelineViewSchemaId as phase4TimelineViewSchemaId,
  taskRequestsViewSchemaId,
} from "./phase4Helpers";
import {
  driveRealTimelineSummaryConflict,
  editTimelineSummary,
  installPatchController,
} from "./phase6Harness";

const timelineViewSchemaId = "timeline";
const timelineApiViewSchemaId = "cartulary.view.timeline.v1";

type ViewRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, unknown>;
};

type GridVisualScrollLeft = "left" | "right" | number;

type GridVisualScrollState = {
  top: number;
  left: GridVisualScrollLeft;
};

test.describe("Phase 3 workbook visual evidence", () => {
  test("V-3-GRID-01 captures the Timeline default viewport with stable row version and save-state strip", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V3GRID01"),
      "Phase 3 visual default",
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineApiViewSchemaId,
      {
        client_txn_id: uniqueTxn("V3GRID01-ROW"),
        "timeline.occurred_at": "2025-02-17T09:12:00Z",
        "timeline.summary": "Default visual row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);

    await expect(page.getByTestId("save-state")).toHaveText("Saved");
    await expect(
      page.getByTestId(`row-${timelineRow.record_id}-row-version`),
    ).toHaveText(String(timelineRow.row_version));
    await expect(
      page.getByTestId(`row-${timelineRow.record_id}-summary`),
    ).toHaveValue("Default visual row");

    await assertVisualRegression(
      page,
      "v-3-grid-01-timeline-default",
      page.getByRole("main"),
    );
  });

  test("V-3-GRID-02 captures Timeline edit save-state visuals for active cell syncing saved and conflict states", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V3GRID02"),
      "Phase 3 visual edit state",
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineApiViewSchemaId,
      {
        client_txn_id: uniqueTxn("V3GRID02-ROW"),
        "timeline.occurred_at": "2025-02-17T10:30:00Z",
        "timeline.summary": "Editable visual row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);

    const saveState = page.getByTestId("save-state");
    const saveStateStrip = saveState.locator("..");
    const summaryInput = page.getByTestId(
      `row-${timelineRow.record_id}-summary`,
    );

    await expect(saveState).toHaveText("Saved");
    await summaryInput.focus();
    await summaryInput.fill("Active visual edit");
    await assertWorkbookGridVisualRegression(
      page,
      "v-3-grid-02-active-edit-cell",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    const patchUrl = `**/api/v1/records/${timelineRow.record_id}`;
    const hold = await holdBrowserApiRequest(page, {
      method: "PATCH",
      path: `/api/v1/records/${timelineRow.record_id}`,
    });

    try {
      await summaryInput.press("Enter");
      await hold.waitForHit;
      await expect(saveState).toHaveText("Syncing");
      await assertVisualRegression(
        page,
        "v-3-grid-02-syncing-strip",
        saveStateStrip,
      );
      await hold.release();
      await expect(saveState).toHaveText("Saved");
      await assertVisualRegression(
        page,
        "v-3-grid-02-saved-strip",
        saveStateStrip,
      );
    } finally {
      await hold.dispose();
    }

    const conflictHandler = async (route: Route) => {
      if (route.request().method().toUpperCase() !== "PATCH") {
        await route.fallback();
        return;
      }
      await route.fulfill({
        status: 409,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            status: 409,
            code: "same_field_conflict",
            message: "same-field conflict",
            request_id: "visual-conflict",
            retryable: false,
            details: {},
            conflict: {
              field_key: "timeline.summary",
              base_row_version: timelineRow.row_version,
              current_row_version: timelineRow.row_version + 1,
              base_value: "Active visual edit",
              server_value: "Server visual edit",
              client_value: "Conflict visual edit",
            },
          },
        }),
      });
    };

    await page.route(patchUrl, conflictHandler);
    try {
      await summaryInput.fill("Conflict visual edit");
      await summaryInput.press("Enter");
      await expect(saveState).toHaveText("Conflict");
      await assertVisualRegression(
        page,
        "v-3-grid-02-conflict-strip",
        saveStateStrip,
      );
    } finally {
      await page.unroute(patchUrl, conflictHandler);
    }
  });

  test("V-3-GRID-03 captures Timeline grouped rows and currently exposed grid chrome", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V3GRID03"),
      "Phase 3 visual grouped rows",
    );
    const firstRow = (await createViewRow(
      page,
      incidentId,
      timelineApiViewSchemaId,
      {
        client_txn_id: uniqueTxn("V3GRID03-ROWA"),
        "timeline.occurred_at": "2025-02-17T11:00:00Z",
        "timeline.summary": "Alpha grouped row",
      },
    )) as ViewRow;
    await createViewRow(page, incidentId, timelineApiViewSchemaId, {
      client_txn_id: uniqueTxn("V3GRID03-ROWB"),
      "timeline.occurred_at": "2025-02-17T11:05:00Z",
      "timeline.summary": "Beta grouped row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await page.getByTestId(`row-${firstRow.record_id}-mark-reviewed`).click();
    await expect(
      page.getByTestId(`row-${firstRow.record_id}-capture-state`),
    ).toHaveText("reviewed");

    await changeGrouping(page, timelineViewSchemaId, "timeline.capture_state");
    await expect(
      page.getByTestId(
        gridGroupRowTestId(
          timelineViewSchemaId,
          "timeline.capture_state",
          "reviewed",
        ),
      ),
    ).toBeVisible();

    await assertWorkbookGridVisualRegression(
      page,
      "v-3-grid-03-grouped-grid",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });
});

test.describe("Phase 4 workbook visual evidence", () => {
  test("V-4-GRID-01 captures Timeline unresolved and resolved mention chips in the workbook grid", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V4GRID01"),
      "Phase 4 visual mention chips",
    );
    await createViewRow(page, incidentId, hostsViewSchemaId, {
      client_txn_id: uniqueTxn("V4GRID01-HOST"),
      "host.display_name": "WS-023",
      "host.hostname": "ws-023.visual.example.test",
    });
    const unresolvedRow = (await createViewRow(
      page,
      incidentId,
      phase4TimelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V4GRID01-UNRESOLVED"),
        "timeline.summary": "Unresolved mention visual row",
        [hostRefsFieldKey]: collectionActionsPayload(["WS-023?"]),
      },
    )) as ViewRow;
    const resolvedRow = (await createViewRow(
      page,
      incidentId,
      phase4TimelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V4GRID01-RESOLVED"),
        "timeline.summary": "Resolved mention visual row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await expect(
      page
        .getByTestId(`row-${unresolvedRow.record_id}-hostRefs-items`)
        .getByLabel("Unresolved WS-023?"),
    ).toBeVisible();
    await addRelationshipTokenViaUI(
      page,
      resolvedRow.record_id,
      "hostRefs",
      "WS-023",
    );
    await expect(
      page
        .getByTestId(`row-${resolvedRow.record_id}-hostRefs-items`)
        .getByLabel("Resolved WS-023"),
    ).toBeVisible();

    await assertWorkbookGridVisualRegression(
      page,
      "v-4-grid-01-mention-chips",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });

  test("V-4-GRID-02 captures Evidence access affordances on the required Evidence surface", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V4GRID02"),
      "Phase 4 visual evidence access",
    );
    const evidenceRow = (await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("V4GRID02-EVIDENCE"),
        "evidence.title": "Visual evidence package",
        "evidence.storage_ref": "slot/visual",
      },
    )) as ViewRow;

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        evidenceViewSchemaId,
      )}`,
    );
    await maskIncidentIdentity(page, incidentId);
    await expect(
      page.getByTestId(rowCellTestId(evidenceRow.record_id, "evidence.title")),
    ).toHaveText("Visual evidence package");
    await expect(
      page.getByTestId(`evidence-preview-${evidenceRow.record_id}`),
    ).toBeVisible();
    await expect(
      page.getByTestId(`evidence-access-message-${evidenceRow.record_id}`),
    ).toContainText("Blocked");

    await assertWorkbookGridVisualRegression(
      page,
      "v-4-grid-02-evidence-access",
      evidenceViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });

  test("V-4-GRID-03 captures Task Requests system view fields through the generic workbook grid", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V4GRID03"),
      "Phase 4 visual task requests",
    );
    const taskRow = (await createViewRow(
      page,
      incidentId,
      taskRequestsViewSchemaId,
      {
        client_txn_id: uniqueTxn("V4GRID03-TASK"),
        "task.title": "Visual task request",
        "task.task_kind": "collection",
      },
    )) as ViewRow;

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        taskRequestsViewSchemaId,
      )}`,
    );
    await maskIncidentIdentity(page, incidentId);
    await expect(
      page.getByTestId(rowCellTestId(taskRow.record_id, "task.title")),
    ).toHaveText("Visual task request");
    await expect(
      page.getByTestId(rowCellTestId(taskRow.record_id, "task.status")),
    ).toHaveText("open");

    await assertWorkbookGridVisualRegression(
      page,
      "v-4-grid-03-task-requests",
      taskRequestsViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });
});

test.describe("Phase 5 workbook visual evidence", () => {
  test("V-5-GRID-01 captures requested and available Evidence states on the required Evidence surface", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V5GRID01"),
      "Phase 5 visual evidence states",
    );
    const evidenceRow = (await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("V5GRID01-EVIDENCE"),
        "evidence.title": "Requested visual package",
        "evidence.storage_ref": "ticket://visual-request",
      },
    )) as ViewRow;

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        evidenceViewSchemaId,
      )}`,
    );
    await maskIncidentIdentity(page, incidentId);
    await expect(
      page.getByTestId(
        rowCellTestId(evidenceRow.record_id, "evidence.lifecycle_state"),
      ),
    ).toHaveText("requested");
    await assertWorkbookGridVisualRegression(
      page,
      "v-5-grid-01-requested-evidence",
      evidenceViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    await page
      .getByTestId(`evidence-attach-file-${evidenceRow.record_id}`)
      .setInputFiles({
        name: "visual-request.txt",
        mimeType: "text/plain",
        buffer: Buffer.from("phase5 visual evidence", "utf8"),
      });
    await expect(
      page.getByTestId(
        rowCellTestId(evidenceRow.record_id, "evidence.lifecycle_state"),
      ),
    ).toHaveText("available");
    await expect(
      page.getByTestId(
        rowCellTestId(evidenceRow.record_id, "evidence.upload_state"),
      ),
    ).toHaveText("available");
    await assertWorkbookGridVisualRegression(
      page,
      "v-5-grid-01-available-evidence",
      evidenceViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );
  });

  test("V-5-GRID-02 captures blocked preview feedback and Timeline evidence badges", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V5GRID02"),
      "Phase 5 visual evidence badges",
    );
    const blocked = (await createViewRow(
      page,
      incidentId,
      evidenceViewSchemaId,
      {
        client_txn_id: uniqueTxn("V5GRID02-BLOCKED"),
        "evidence.title": "Blocked visual package",
        "evidence.storage_ref": "ticket://visual-blocked",
      },
    )) as ViewRow;
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineApiViewSchemaId,
      {
        client_txn_id: uniqueTxn("V5GRID02-TIMELINE"),
        "timeline.summary": "Visual evidence badge row",
      },
    )) as ViewRow;

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        evidenceViewSchemaId,
      )}`,
    );
    await maskIncidentIdentity(page, incidentId);
    await expect(
      page.getByTestId(`evidence-access-message-${blocked.record_id}`),
    ).toContainText("Blocked");
    await assertWorkbookGridVisualRegression(
      page,
      "v-5-grid-02-blocked-preview",
      evidenceViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        timelineApiViewSchemaId,
      )}`,
    );
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await page
      .getByTestId(rowInspectButtonTestId(timelineRow.record_id))
      .click();
    await page
      .getByTestId(`timeline-evidence-file-${timelineRow.record_id}`)
      .setInputFiles({
        name: "visual-badge.png",
        mimeType: "image/png",
        buffer: tinyPNG(),
      });
    await expect(
      page.getByTestId(
        rowCellTestId(timelineRow.record_id, "timeline.evidence_count"),
      ),
    ).toHaveText("1");
    await expect(
      page.getByTestId(
        rowCellTestId(timelineRow.record_id, "timeline.has_evidence"),
      ),
    ).toHaveText("true");
    await assertWorkbookGridVisualRegression(
      page,
      "v-5-grid-02-timeline-evidence-badge",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "right" } },
    );
  });
});

test.describe("Phase 6 workbook visual evidence", () => {
  test("V-6-GRID-01 regresses Phase 6 row-gutter and same-cell presence markers", async ({
    browser,
    page,
    sessionTracker,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V6GRID01"),
      "Phase 6 visual presence markers",
    );
    const remote = await createIncidentMemberUser(page, incidentId, {
      display_name: "Visual Analyst",
      email: uniqueEmail("phase6-v6grid01-remote"),
      initial_password: "Phase6V6Grid01!",
      role: "editor",
    });
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineApiViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID01-ROW"),
        "timeline.summary": "Presence visual row",
      },
    )) as ViewRow;

    let remotePage: Page | null = null;
    try {
      await page.goto(`/?incident_id=${incidentId}`);
      await maskIncidentIdentity(page, incidentId);
      await expect(
        page.getByTestId(`row-${timelineRow.record_id}-summary`),
      ).toHaveValue("Presence visual row");

      remotePage = await openIncidentAsTrackedUser(browser, sessionTracker, {
        createdBy: "V-6-GRID-01",
        email: remote.email,
        incidentId,
        password: remote.initial_password,
        purpose: "Phase 6 visual presence analyst",
        userId: remote.user_id,
      });
      await remotePage
        .getByTestId(`row-${timelineRow.record_id}-summary`)
        .focus();
      await expect(
        page.getByTestId(rowPresenceMarkerTestId(timelineRow.record_id)),
      ).toContainText("VA");
      await expect(
        page.getByTestId(
          cellPresenceMarkerTestId(timelineRow.record_id, "timeline.summary"),
        ),
      ).toContainText("VA");
      await assertMarkerAnchoredToGridTarget({
        anchorKind: "row-gutter",
        markerTestId: rowPresenceMarkerTestId(timelineRow.record_id),
        page,
        surface: timelineViewSchemaId,
        targetTestId: rowCellTestId(timelineRow.record_id, "capture-state"),
      });
      await assertMarkerAnchoredToGridTarget({
        anchorKind: "cell",
        markerTestId: cellPresenceMarkerTestId(
          timelineRow.record_id,
          "timeline.summary",
        ),
        page,
        surface: timelineViewSchemaId,
        targetTestId: rowCellTestId(timelineRow.record_id, "summary"),
      });

      await assertWorkbookGridVisualRegression(
        page,
        "v-6-grid-01-presence-markers",
        timelineViewSchemaId,
        { scroll: { top: 0, left: "left" } },
      );
    } finally {
      await remotePage?.context().close();
    }
  });

  test("V-6-GRID-02 regresses Phase 6 same-field conflict marker resolver and Conflict strip", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V6GRID02"),
      "Phase 6 visual conflict resolver",
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineApiViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID02-ROW"),
        "timeline.summary": "Conflict visual base",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    const patchController = await installPatchController(page);
    try {
      await driveRealTimelineSummaryConflict({
        baseRowVersion: timelineRow.row_version,
        localValue: "Conflict visual local",
        page,
        patchController,
        recordId: timelineRow.record_id,
        remoteValue: "Conflict visual server",
        txnPrefix: "visual-phase6-conflict",
      });
      await expect(
        page.getByTestId(
          conflictMarkerTestId(timelineRow.record_id, "timeline.summary"),
        ),
      ).toBeVisible();
      await assertMarkerAnchoredToGridTarget({
        anchorKind: "cell",
        markerTestId: conflictMarkerTestId(
          timelineRow.record_id,
          "timeline.summary",
        ),
        page,
        surface: timelineViewSchemaId,
        targetTestId: rowCellTestId(timelineRow.record_id, "summary"),
      });

      await assertVisualRegression(
        page,
        "v-6-grid-02-conflict-resolver",
        page.getByRole("main"),
      );
    } finally {
      await patchController.dispose();
    }
  });

  test("V-6-GRID-03 regresses Phase 6 pending-queue save-state transitions", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("V6GRID03"),
      "Phase 6 visual pending queue",
    );
    const syncRow = (await createViewRow(
      page,
      incidentId,
      timelineApiViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID03-ROW"),
        "timeline.summary": "Pending visual base",
      },
    )) as ViewRow;
    const conflictRow = (await createViewRow(
      page,
      incidentId,
      timelineApiViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID03-CONFLICT-ROW"),
        "timeline.summary": "Pending conflict visual base",
      },
    )) as ViewRow;
    const queuedRow = (await createViewRow(
      page,
      incidentId,
      timelineApiViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID03-QUEUED-ROW"),
        "timeline.summary": "Pending queued visual base",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    const summaryInput = page.getByTestId(`row-${syncRow.record_id}-summary`);
    const saveState = page.getByTestId(saveStateTestId());
    const saveStateStrip = saveState.locator("..");
    const hold = await holdBrowserApiRequest(page, {
      method: "PATCH",
      path: `/api/v1/records/${syncRow.record_id}`,
    });

    try {
      await summaryInput.fill("Pending visual syncing");
      await summaryInput.press("Enter");
      await hold.waitForHit;
      await expect(saveState).toHaveText("Syncing");
      await assertVisualRegression(
        page,
        "v-6-grid-03-syncing-strip",
        saveStateStrip,
      );
      await hold.release();
      await expect(saveState).toHaveText("Saved");
      await assertVisualRegression(
        page,
        "v-6-grid-03-saved-strip",
        saveStateStrip,
      );
    } finally {
      await hold.dispose();
    }

    const patchController = await installPatchController(page);
    try {
      await driveRealTimelineSummaryConflict({
        afterLocalPatchHeld: async () => {
          await editTimelineSummary(
            page,
            queuedRow.record_id,
            "Pending visual queued replay",
          );
          await expect(
            page.getByTestId(pendingQueueNoticeTestId()),
          ).toBeVisible();
        },
        baseRowVersion: conflictRow.row_version,
        localValue: "Pending visual blocked",
        page,
        patchController,
        recordId: conflictRow.record_id,
        remoteValue: "Pending visual server",
        txnPrefix: "visual-phase6-pending-conflict",
      });
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();
      await assertVisualRegression(
        page,
        "v-6-grid-03-blocked-conflict",
        page.getByRole("main"),
      );

      await page.getByTestId("conflict-keep-saved").click();
      await expect(saveState).toHaveText("Saved");
    } finally {
      await patchController.dispose();
    }

    await expect(saveState).toHaveText("Saved");
    await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
    await assertVisualRegression(
      page,
      "v-6-grid-03-recovered-saved-strip",
      saveStateStrip,
    );
  });
});

async function assertVisualRegression(
  page: Page,
  name: string,
  locator = page.getByRole("main"),
) {
  await expect(locator).toBeVisible();
  await page.evaluate(() => {
    document.documentElement.dataset.visualSnapshot = "true";
  });
  await maskVisualDynamicText(page);
  await expect(locator).toHaveScreenshot(`${name}.png`, {
    animations: "disabled",
    caret: "hide",
  });
}

async function assertWorkbookGridVisualRegression(
  page: Page,
  name: string,
  surface: string,
  options: { scroll: GridVisualScrollState },
) {
  await normalizeWorkbookGridVisualState(page, surface, options.scroll);
  await assertVisualRegression(
    page,
    name,
    page.getByTestId(gridShellTestId(surface)),
  );
}

async function normalizeWorkbookGridVisualState(
  page: Page,
  surface: string,
  scroll: GridVisualScrollState,
) {
  const expected = await setWorkbookGridScroll(page, surface, scroll);
  await expect
    .poll(() => readWorkbookGridScroll(page, surface))
    .toEqual(expected);
  await page.evaluate(() => {
    return new Promise<void>((resolve) => {
      requestAnimationFrame(() => {
        requestAnimationFrame(() => resolve());
      });
    });
  });
}

async function setWorkbookGridScroll(
  page: Page,
  surface: string,
  scroll: GridVisualScrollState,
) {
  return page.evaluate(
    ({ left, scrollportSelector, surface, top }) => {
      const shell = document.querySelector<HTMLElement>(
        `[data-testid="${surface}-grid-shell"]`,
      );
      if (shell === null) {
        throw new Error(`Expected ${surface} grid shell to exist`);
      }
      const scrollports = Array.from(
        shell.querySelectorAll<HTMLElement>(scrollportSelector),
      );
      if (scrollports.length !== 1 || scrollports[0] === undefined) {
        throw new Error(
          `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
        );
      }
      const scrollport = scrollports[0];
      const maxLeft = Math.max(
        0,
        scrollport.scrollWidth - scrollport.clientWidth,
      );
      const maxTop = Math.max(
        0,
        scrollport.scrollHeight - scrollport.clientHeight,
      );
      const expectedLeft =
        left === "left" ? 0 : left === "right" ? maxLeft : left;
      const expectedTop = Math.min(Math.max(0, top), maxTop);
      scrollport.scrollTop = expectedTop;
      scrollport.scrollLeft = Math.min(Math.max(0, expectedLeft), maxLeft);
      return {
        top: scrollport.scrollTop,
        left: scrollport.scrollLeft,
      };
    },
    {
      left: scroll.left,
      scrollportSelector: gridScrollportSelector(),
      surface,
      top: scroll.top,
    },
  );
}

async function readWorkbookGridScroll(page: Page, surface: string) {
  return page.evaluate(
    ({ scrollportSelector, surface }) => {
      const shell = document.querySelector<HTMLElement>(
        `[data-testid="${surface}-grid-shell"]`,
      );
      if (shell === null) {
        throw new Error(`Expected ${surface} grid shell to exist`);
      }
      const scrollports = Array.from(
        shell.querySelectorAll<HTMLElement>(scrollportSelector),
      );
      if (scrollports.length !== 1 || scrollports[0] === undefined) {
        throw new Error(
          `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
        );
      }
      return {
        top: scrollports[0].scrollTop,
        left: scrollports[0].scrollLeft,
      };
    },
    {
      scrollportSelector: gridScrollportSelector(),
      surface,
    },
  );
}

async function maskVisualDynamicText(page: Page) {
  await page.evaluate(() => {
    const styleId = "visual-dynamic-input-mask";
    if (!document.getElementById(styleId)) {
      const style = document.createElement("style");
      style.id = styleId;
      style.textContent = `
        [data-testid="conflict-server-actor"],
        [data-testid="conflict-server-updated-at"] {
          color: transparent !important;
          caret-color: transparent !important;
        }
      `;
      document.head.append(style);
    }
    const replacements: Array<[RegExp, string]> = [
      [
        /\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b/gi,
        "00000000-0000-0000-0000-000000000000",
      ],
      [
        /\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z\b/g,
        "2025-01-01T00:00:00Z",
      ],
      [/\bIR-[A-Z0-9-]+\b/g, "IR-VISUAL-FIXTURE"],
      [/Playwright Worker Admin \d+/g, "Playwright Worker Admin"],
    ];
    const walker = document.createTreeWalker(
      document.body,
      NodeFilter.SHOW_TEXT,
    );
    for (let node = walker.nextNode(); node; node = walker.nextNode()) {
      let text = node.textContent ?? "";
      for (const [pattern, replacement] of replacements) {
        text = text.replace(pattern, replacement);
      }
      node.textContent = text;
    }
    for (const element of document.querySelectorAll("input, textarea")) {
      if (
        !(element instanceof HTMLInputElement) &&
        !(element instanceof HTMLTextAreaElement)
      ) {
        continue;
      }
      let value = element.value;
      for (const [pattern, replacement] of replacements) {
        value = value.replace(pattern, replacement);
      }
      element.value = value;
    }
  });
}

async function maskIncidentIdentity(page: Page, incidentId: string) {
  await page.evaluate((id) => {
    for (const node of document.querySelectorAll("p")) {
      if (node.textContent?.includes(id)) {
        node.textContent = "Incident visual-fixture";
      }
    }
  }, incidentId);
}

function tinyPNG() {
  return Buffer.from(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=",
    "base64",
  );
}
