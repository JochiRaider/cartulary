import { Buffer } from "node:buffer";
import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import {
  assertMarkerAnchoredToGridTarget,
  changeGrouping,
} from "@cartulary/test-utils";
import {
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  dataTestIdSelector,
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidencePreviewButtonTestId,
  gridActionsHeaderTestId,
  gridGroupRowTestId,
  gridScrollportSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  pendingQueueNoticeTestId,
  relationshipItemsTestId,
  rowCellTestId,
  rowInspectButtonTestId,
  rowPresenceMarkerTestId,
  savedViewSelectorTestId,
  saveStateTestId,
  surfaceTabTestId,
  systemViewSwitcherTriggerTestId,
  timelineEvidenceFileInputTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowVersionTestId,
  workbookShellReadyTestId,
  workbookShellSlots,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import type { Locator, Page, Route, TestInfo } from "@playwright/test";
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
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "./phase4Helpers";
import {
  driveRealTimelineSummaryConflict,
  editTimelineSummary,
  installPatchController,
} from "./phase6Harness";

type ViewRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, unknown>;
};

type GridVisualScrollLeft = "left" | number;

type GridVisualScrollState = {
  top: number;
  left: GridVisualScrollLeft;
};

type WorkbookGridVisualScrollSnapshot = {
  shellTop: number;
  shellLeft: number;
  scrollportTop: number;
  scrollportLeft: number;
};

type GridVisualAnchor = {
  kind: "timelineEvidenceActions";
  rowId: string;
  top: number;
};

function gridShellSelector(surface: string): string {
  return dataTestIdSelector(gridShellTestId(surface));
}

type GridVisualRegressionOptions = {
  maxDiffPixels?: number;
  testInfo?: TestInfo;
} & (
  | { scroll: GridVisualScrollState; anchor?: never }
  | { anchor: GridVisualAnchor; scroll?: never }
);

test.describe("FE-P2 workbook visual readiness", () => {
  test("FE-V-P2-01 Capture default workbook shell with top bar, tabs, System views, view bar, primary grid slot, inspector, and status strip using deterministic visual harness settings.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEV2SHELL"),
      "FE-P2 visual default shell",
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("FEV2SHELL-ROW"),
        "timeline.occurred_at": "2026-05-31T09:00:00Z",
        "timeline.summary": "FE-P2 visual shell row",
        "timeline.details": "Default shell inspector visual fixture",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);

    const shell = page.getByTestId(workbookShellReadyTestId());
    await expect(shell).toBeVisible();
    for (const slot of workbookShellSlots) {
      await expect(
        shell.locator(dataTestIdSelector(workbookShellSlotTestId(slot))),
      ).toBeVisible();
    }
    await expect(
      page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
    ).toHaveAttribute("aria-current", "page");
    await expect(page.getByTestId(systemViewSwitcherTriggerTestId()))
      .toBeVisible();
    await expect(
      page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(
      page.getByTestId(
        rowCellTestId(timelineRow.record_id, "timeline.summary"),
      ),
    ).toHaveValue("FE-P2 visual shell row");

    await page.getByTestId(rowInspectButtonTestId(timelineRow.record_id)).click();
    await expect(page.getByTestId("timeline-inspector")).toBeVisible();
    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });

    await assertViewportVisualRegression(
      page,
      "fe-v-p2-01-default-workbook-shell",
    );
  });
});

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
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V3GRID01-ROW"),
        "timeline.occurred_at": "2025-02-17T09:12:00Z",
        "timeline.summary": "Default visual row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);

    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(
      page.getByTestId(timelineRowVersionTestId(timelineRow.record_id)),
    ).toHaveText(String(timelineRow.row_version));
    await expect(
      page.getByTestId(
        rowCellTestId(timelineRow.record_id, "timeline.summary"),
      ),
    ).toHaveValue("Default visual row");

    await assertViewportVisualRegression(page, "v-3-grid-01-timeline-default");
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
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V3GRID02-ROW"),
        "timeline.occurred_at": "2025-01-01T00:00:00Z",
        "timeline.summary": "Editable visual row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);

    const saveState = page.getByTestId(saveStateTestId());
    const summaryInput = page.getByTestId(
      rowCellTestId(timelineRow.record_id, "timeline.summary"),
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
      await assertStatusStripVisualRegression(
        page,
        "v-3-grid-02-syncing-strip",
      );
      await hold.release();
      await expect(saveState).toHaveText("Saved");
      await assertStatusStripVisualRegression(page, "v-3-grid-02-saved-strip");
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
      await assertStatusStripVisualRegression(
        page,
        "v-3-grid-02-conflict-strip",
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
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V3GRID03-ROWA"),
        "timeline.occurred_at": "2025-02-17T11:00:00Z",
        "timeline.summary": "Alpha grouped row",
      },
    )) as ViewRow;
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("V3GRID03-ROWB"),
      "timeline.occurred_at": "2025-02-17T11:05:00Z",
      "timeline.summary": "Beta grouped row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await page
      .getByTestId(timelineRowMarkReviewedButtonTestId(firstRow.record_id))
      .click();
    await expect(
      page.getByTestId(
        rowCellTestId(firstRow.record_id, "timeline.capture_state"),
      ),
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
    await expect(
      page.getByTestId(
        gridGroupRowTestId(
          timelineViewSchemaId,
          "timeline.capture_state",
          "rough",
        ),
      ),
    ).toBeVisible();
    await expect(
      page
        .getByTestId(gridShellTestId(timelineViewSchemaId))
        .getByText("Unassigned", { exact: true }),
    ).toHaveCount(0);

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
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V4GRID01-UNRESOLVED"),
        "timeline.summary": "Unresolved mention visual row",
        [hostRefsFieldKey]: collectionActionsPayload(["WS-023?"]),
      },
    )) as ViewRow;
    const resolvedRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V4GRID01-RESOLVED"),
        "timeline.summary": "Resolved mention visual row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await expect(
      page
        .getByTestId(
          relationshipItemsTestId(unresolvedRow.record_id, hostRefsFieldKey),
        )
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
        .getByTestId(
          relationshipItemsTestId(resolvedRow.record_id, hostRefsFieldKey),
        )
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
      page.getByTestId(evidencePreviewButtonTestId(evidenceRow.record_id)),
    ).toBeVisible();
    await expect(
      page.getByTestId(evidenceAccessMessageTestId(evidenceRow.record_id)),
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
      .getByTestId(evidenceAttachFileInputTestId(evidenceRow.record_id))
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
  }, testInfo) => {
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
      timelineViewSchemaId,
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
      page.getByTestId(evidenceAccessMessageTestId(blocked.record_id)),
    ).toContainText("Blocked");
    await assertWorkbookGridVisualRegression(
      page,
      "v-5-grid-02-blocked-preview",
      evidenceViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    await page.goto(
      `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
        timelineViewSchemaId,
      )}`,
    );
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await page
      .getByTestId(rowInspectButtonTestId(timelineRow.record_id))
      .click();
    await page
      .getByTestId(timelineEvidenceFileInputTestId(timelineRow.record_id))
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
    await expect(page.getByTestId("timeline-inspector")).toContainText(
      timelineRow.record_id,
    );
    await page.evaluate(() => {
      if (document.activeElement instanceof HTMLElement) {
        document.activeElement.blur();
      }
    });
    await assertWorkbookGridVisualRegression(
      page,
      "v-5-grid-02-timeline-evidence-badge",
      timelineViewSchemaId,
      {
        anchor: {
          kind: "timelineEvidenceActions",
          rowId: timelineRow.record_id,
          top: 0,
        },
        maxDiffPixels: 8_000,
        testInfo,
      },
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
      timelineViewSchemaId,
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
        page.getByTestId(
          rowCellTestId(timelineRow.record_id, "timeline.summary"),
        ),
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
        .getByTestId(rowCellTestId(timelineRow.record_id, "timeline.summary"))
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
        targetTestId: rowCellTestId(
          timelineRow.record_id,
          "timeline.capture_state",
        ),
      });
      await assertMarkerAnchoredToGridTarget({
        anchorKind: "cell",
        markerTestId: cellPresenceMarkerTestId(
          timelineRow.record_id,
          "timeline.summary",
        ),
        page,
        surface: timelineViewSchemaId,
        targetTestId: rowCellTestId(timelineRow.record_id, "timeline.summary"),
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
      timelineViewSchemaId,
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
        targetTestId: rowCellTestId(timelineRow.record_id, "timeline.summary"),
      });

      await assertViewportVisualRegression(
        page,
        "v-6-grid-02-conflict-resolver",
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
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID03-ROW"),
        "timeline.occurred_at": "2025-03-06T10:00:00Z",
        "timeline.summary": "Pending visual base",
      },
    )) as ViewRow;
    const conflictRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID03-CONFLICT-ROW"),
        "timeline.occurred_at": "2025-03-06T10:05:00Z",
        "timeline.summary": "Pending conflict visual base",
      },
    )) as ViewRow;
    const queuedRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("V6GRID03-QUEUED-ROW"),
        "timeline.occurred_at": "2025-03-06T10:10:00Z",
        "timeline.summary": "Pending queued visual base",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    const summaryInput = page.getByTestId(
      rowCellTestId(syncRow.record_id, "timeline.summary"),
    );
    const saveState = page.getByTestId(saveStateTestId());
    const patchController = await installPatchController(page);
    const hold = patchController.holdNextPatch({ recordId: syncRow.record_id });

    try {
      await summaryInput.fill("Pending visual syncing");
      await summaryInput.press("Enter");
      await hold.waitForHit;
      await expect(saveState).toHaveText("Syncing");
      await assertStatusStripVisualRegression(
        page,
        "v-6-grid-03-syncing-strip",
      );
      await hold.release();
      await hold.waitForCompletion;
      await expect(saveState).toHaveText("Saved");
      await assertStatusStripVisualRegression(page, "v-6-grid-03-saved-strip");

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
      await assertViewportVisualRegression(
        page,
        "v-6-grid-03-blocked-conflict",
      );

      await page.getByTestId("conflict-keep-saved").click();
      await expect(saveState).toHaveText("Saved");
    } finally {
      await patchController.dispose();
    }

    await expect(saveState).toHaveText("Saved");
    await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
    await assertStatusStripVisualRegression(
      page,
      "v-6-grid-03-recovered-saved-strip",
    );
  });
});

async function assertVisualRegression(
  page: Page,
  name: string,
  locator = page.getByRole("main"),
  options: { maxDiffPixels?: number } = {},
) {
  await expect(locator).toBeVisible();
  await prepareVisualRegressionState(page);
  await expect(locator).toHaveScreenshot(`${name}.png`, {
    animations: "disabled",
    caret: "hide",
    ...(options.maxDiffPixels === undefined
      ? {}
      : { maxDiffPixels: options.maxDiffPixels }),
  });
}

async function assertViewportVisualRegression(page: Page, name: string) {
  await prepareVisualRegressionState(page);
  await expect(page).toHaveScreenshot(`${name}.png`, {
    animations: "disabled",
    caret: "hide",
    fullPage: false,
  });
}

async function assertStatusStripVisualRegression(page: Page, name: string) {
  const statusStrip = page.getByTestId(workbookShellSlotTestId("status-strip"));
  await expectStatusStripFocusAnchorVisuallyHidden(statusStrip);
  await assertVisualRegression(page, name, statusStrip);
}

async function expectStatusStripFocusAnchorVisuallyHidden(
  statusStrip: Locator,
) {
  const focusAnchor = statusStrip.getByTestId("workbook-focus-anchor");
  await expect(focusAnchor).toHaveCount(1);
  await expect
    .poll(
      async () =>
        focusAnchor.evaluate((node) => {
          const rect = node.getBoundingClientRect();
          const style = window.getComputedStyle(node);
          return {
            blockSize: Math.round(rect.height),
            clipPath: style.clipPath,
            inlineSize: Math.round(rect.width),
            overflow: style.overflow,
            position: style.position,
          };
        }),
      {
        message:
          "Expected status-strip focus anchor to remain present but visually hidden",
      },
    )
    .toEqual({
      blockSize: 1,
      clipPath: "inset(50%)",
      inlineSize: 1,
      overflow: "hidden",
      position: "absolute",
    });
}

async function assertWorkbookGridVisualRegression(
  page: Page,
  name: string,
  surface: string,
  options: GridVisualRegressionOptions,
) {
  try {
    await prepareVisualRegressionState(page);
    await normalizeWorkbookGridVisualState(page, surface, options);
    await assertVisualRegression(
      page,
      name,
      page.getByTestId(gridShellTestId(surface)),
      options.maxDiffPixels === undefined
        ? {}
        : { maxDiffPixels: options.maxDiffPixels },
    );
  } catch (error) {
    try {
      await attachWorkbookGridVisualDiagnostics(page, name, surface, options);
    } catch {
      // Preserve the assertion failure when the page is already torn down.
    }
    throw error;
  }
}

async function prepareVisualRegressionState(page: Page) {
  await page.evaluate(() => {
    document.documentElement.dataset.visualSnapshot = "true";
  });
  await waitForVendoredFonts(page);
  await attachFontManifestDigest();
  await maskVisualDynamicText(page);
}

async function waitForVendoredFonts(page: Page) {
  await page.evaluate(async () => {
    await Promise.all([
      document.fonts.load('400 12px "Inter"'),
      document.fonts.load('400 12px "JetBrains Mono"'),
    ]);
    await document.fonts.ready;
    const faces = Array.from(document.fonts);
    for (const family of ["Inter", "JetBrains Mono"]) {
      const familyFaces = faces.filter((face) => face.family === family);
      if (familyFaces.length === 0) {
        throw new Error(`missing vendored font-face for ${family}`);
      }
      const failedFace = familyFaces.find((face) => face.status === "error");
      if (failedFace) {
        throw new Error(`vendored font ${family} failed to load`);
      }
      if (!document.fonts.check(`400 12px "${family}"`)) {
        throw new Error(`vendored font ${family} is not ready`);
      }
    }
  });
}

async function attachFontManifestDigest() {
  const manifest = readFileSync(
    new URL("../public/assets/fonts/FONT_MANIFEST.json", import.meta.url),
  );
  const sha256 = createHash("sha256").update(manifest).digest("hex");
  await test.info().attach("font-manifest-sha256", {
    body: Buffer.from(`${sha256}\n`, "utf8"),
    contentType: "text/plain",
  });
}

async function normalizeWorkbookGridVisualState(
  page: Page,
  surface: string,
  options: GridVisualRegressionOptions,
) {
  if ("anchor" in options) {
    await normalizeWorkbookGridAnchorVisualState(page, surface, options.anchor);
    return;
  }
  const { scroll } = options;
  await setWorkbookGridScroll(page, surface, scroll);
  await waitForVisualLayoutFrame(page);
  const expected = await setWorkbookGridScroll(page, surface, scroll);
  await expect
    .poll(() => readWorkbookGridScroll(page, surface), {
      message: `Expected ${surface} grid visual scroll to normalize shell and scrollport state`,
    })
    .toEqual(expected);
  await waitForVisualLayoutFrame(page);
}

async function normalizeWorkbookGridAnchorVisualState(
  page: Page,
  surface: string,
  anchor: GridVisualAnchor,
) {
  await setWorkbookGridAnchor(page, surface, anchor);
  await waitForVisualLayoutFrame(page);
  await expect
    .poll(
      async () => {
        const expected = await setWorkbookGridAnchor(page, surface, anchor);
        const state = await readWorkbookGridAnchorState(page, surface, anchor);
        return (
          state.ready &&
          state.diagnostics.scroll.shell.left === expected.shellLeft &&
          state.diagnostics.scroll.shell.top === expected.shellTop &&
          state.diagnostics.scroll.scrollport.left ===
            expected.scrollportLeft &&
          state.diagnostics.scroll.scrollport.top === expected.scrollportTop
        );
      },
      {
        message: `Expected ${surface} grid visual anchor ${anchor.kind} to reach stable geometry with normalized shell and scrollport state`,
        timeout: 6_000,
      },
    )
    .toBe(true);
  await waitForVisualLayoutFrame(page);
}

async function waitForVisualLayoutFrame(page: Page) {
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
): Promise<WorkbookGridVisualScrollSnapshot> {
  return page.evaluate(
    ({ left, scrollportSelector, shellSelector, surface, top }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
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
      const expectedLeft = left === "left" ? 0 : left;
      const expectedTop = Math.min(Math.max(0, top), maxTop);
      shell.scrollTop = 0;
      shell.scrollLeft = 0;
      scrollport.scrollTop = expectedTop;
      scrollport.scrollLeft = Math.min(Math.max(0, expectedLeft), maxLeft);
      shell.scrollTop = 0;
      shell.scrollLeft = 0;
      return {
        shellTop: Math.round(shell.scrollTop),
        shellLeft: Math.round(shell.scrollLeft),
        scrollportTop: Math.round(scrollport.scrollTop),
        scrollportLeft: Math.round(scrollport.scrollLeft),
      };
    },
    {
      left: scroll.left,
      scrollportSelector: gridScrollportSelector(),
      shellSelector: gridShellSelector(surface),
      surface,
      top: scroll.top,
    },
  );
}

function buildWorkbookGridAnchorSelectors(
  surface: string,
  anchor: GridVisualAnchor,
) {
  switch (anchor.kind) {
    case "timelineEvidenceActions":
      return {
        fieldKeys: [
          "timeline.evidence_count",
          "timeline.tags",
          "timeline.edited_at",
        ],
        requiredTestIds: {
          actionButton: rowInspectButtonTestId(anchor.rowId),
          actionsHeader: gridActionsHeaderTestId(surface),
          editedCell: rowCellTestId(anchor.rowId, "timeline.edited_at"),
          editedHeader: gridSortHeaderTestId(surface, "timeline.edited_at"),
          evidenceCell: rowCellTestId(anchor.rowId, "timeline.evidence_count"),
          evidenceHeader: gridSortHeaderTestId(
            surface,
            "timeline.evidence_count",
          ),
          hasEvidenceBadge: rowCellTestId(
            anchor.rowId,
            "timeline.has_evidence",
          ),
          tagsHeader: gridSortHeaderTestId(surface, "timeline.tags"),
        },
      };
  }
}

async function setWorkbookGridAnchor(
  page: Page,
  surface: string,
  anchor: GridVisualAnchor,
): Promise<WorkbookGridVisualScrollSnapshot> {
  return page.evaluate(
    ({ anchor, scrollportSelector, selectors, shellSelector, surface }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
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
      const expectedTop = Math.min(Math.max(0, anchor.top), maxTop);
      const byTestId = (testId: string) =>
        Array.from(shell.querySelectorAll<HTMLElement>("[data-testid]")).find(
          (element) => element.getAttribute("data-testid") === testId,
        ) ?? null;

      shell.scrollTop = 0;
      shell.scrollLeft = 0;
      scrollport.scrollTop = expectedTop;
      const actionButton = byTestId(selectors.requiredTestIds.actionButton);
      if (actionButton === null) {
        shell.scrollLeft = Math.max(0, shell.scrollWidth - shell.clientWidth);
        scrollport.scrollLeft = maxLeft;
      } else {
        actionButton.scrollIntoView({ block: "nearest", inline: "end" });
      }
      shell.scrollTop = 0;

      const requiredElements = Object.values(selectors.requiredTestIds)
        .map((testId) => byTestId(testId))
        .filter((element): element is HTMLElement => element !== null);
      if (requiredElements.length > 0) {
        const shellRect = shell.getBoundingClientRect();
        const leftMost = Math.min(
          ...requiredElements.map(
            (element) => element.getBoundingClientRect().left,
          ),
        );
        const rightMost = Math.max(
          ...requiredElements.map(
            (element) => element.getBoundingClientRect().right,
          ),
        );
        const padding = 8;
        if (leftMost < shellRect.left + padding) {
          shell.scrollLeft = Math.max(
            0,
            shell.scrollLeft - (shellRect.left + padding - leftMost),
          );
        } else if (rightMost > shellRect.right - padding) {
          shell.scrollLeft = Math.min(
            Math.max(0, shell.scrollWidth - shell.clientWidth),
            shell.scrollLeft + (rightMost - (shellRect.right - padding)),
          );
        }
      }
      shell.scrollTop = 0;

      return {
        shellTop: Math.round(shell.scrollTop),
        shellLeft: Math.round(shell.scrollLeft),
        scrollportTop: Math.round(scrollport.scrollTop),
        scrollportLeft: Math.round(scrollport.scrollLeft),
      };
    },
    {
      anchor,
      scrollportSelector: gridScrollportSelector(),
      selectors: buildWorkbookGridAnchorSelectors(surface, anchor),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

async function readWorkbookGridAnchorState(
  page: Page,
  surface: string,
  anchor: GridVisualAnchor,
) {
  return page.evaluate(
    async ({
      anchor,
      scrollportSelector,
      selectors,
      shellSelector,
      surface,
    }) => {
      const readDiagnostics = () => {
        const shell = document.querySelector<HTMLElement>(shellSelector);
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
        const shellRect = shell.getBoundingClientRect();
        const byTestId = (testId: string) =>
          Array.from(shell.querySelectorAll<HTMLElement>("[data-testid]")).find(
            (element) => element.getAttribute("data-testid") === testId,
          ) ?? null;
        const roundedRect = (element: HTMLElement | null) => {
          if (element === null) {
            return null;
          }
          const rect = element.getBoundingClientRect();
          const visible =
            rect.width > 0 &&
            rect.height > 0 &&
            rect.right <= shellRect.right - 1 &&
            rect.left >= shellRect.left + 1 &&
            rect.bottom >= shellRect.top + 1 &&
            rect.top <= shellRect.bottom - 1;
          return {
            bottom: Math.round(rect.bottom),
            height: Math.round(rect.height),
            left: Math.round(rect.left),
            right: Math.round(rect.right),
            top: Math.round(rect.top),
            visible,
            width: Math.round(rect.width),
          };
        };
        const requiredRects = Object.fromEntries(
          Object.entries(selectors.requiredTestIds).map(([key, testId]) => [
            key,
            roundedRect(byTestId(testId)),
          ]),
        );
        const visibleFieldKeys = Array.from(
          shell.querySelectorAll<HTMLElement>("[data-grid-field-key]"),
        )
          .filter((element) => {
            const rect = element.getBoundingClientRect();
            return (
              rect.width > 0 &&
              rect.height > 0 &&
              rect.right >= shellRect.left + 1 &&
              rect.left <= shellRect.right - 1 &&
              rect.bottom >= shellRect.top + 1 &&
              rect.top <= shellRect.bottom - 1
            );
          })
          .map((element) => element.getAttribute("data-grid-field-key") ?? "")
          .filter((fieldKey, index, fieldKeys) => {
            return fieldKey !== "" && fieldKeys.indexOf(fieldKey) === index;
          });
        const missingTestIds = Object.entries(selectors.requiredTestIds)
          .filter(([, testId]) => byTestId(testId) === null)
          .map(([key]) => key);
        const hiddenTestIds = Object.entries(requiredRects)
          .filter(([, rect]) => rect === null || !rect.visible)
          .map(([key]) => key);
        const missingFieldKeys = selectors.fieldKeys.filter(
          (fieldKey) => !visibleFieldKeys.includes(fieldKey),
        );
        return {
          activeElementTestId:
            document.activeElement instanceof HTMLElement
              ? document.activeElement.getAttribute("data-testid")
              : null,
          anchorKind: anchor.kind,
          inspectorOpen:
            document.querySelector('[data-testid="timeline-inspector"]') !==
            null,
          missingFieldKeys,
          missingTestIds,
          ready:
            missingTestIds.length === 0 &&
            hiddenTestIds.length === 0 &&
            missingFieldKeys.length === 0,
          requiredRects,
          screenshotTargetTestId: `${surface}-grid-shell`,
          scroll: {
            shell: {
              clientHeight: shell.clientHeight,
              clientWidth: shell.clientWidth,
              left: Math.round(shell.scrollLeft),
              maxLeft: Math.max(0, shell.scrollWidth - shell.clientWidth),
              maxTop: Math.max(0, shell.scrollHeight - shell.clientHeight),
              scrollHeight: shell.scrollHeight,
              scrollWidth: shell.scrollWidth,
              top: Math.round(shell.scrollTop),
            },
            scrollport: {
              clientHeight: scrollport.clientHeight,
              clientWidth: scrollport.clientWidth,
              left: Math.round(scrollport.scrollLeft),
              maxLeft: Math.max(
                0,
                scrollport.scrollWidth - scrollport.clientWidth,
              ),
              maxTop: Math.max(
                0,
                scrollport.scrollHeight - scrollport.clientHeight,
              ),
              scrollHeight: scrollport.scrollHeight,
              scrollWidth: scrollport.scrollWidth,
              top: Math.round(scrollport.scrollTop),
            },
          },
          surface,
          visibleFieldKeys,
        };
      };
      const nextFrame = () =>
        new Promise<void>((resolve) => {
          requestAnimationFrame(() => resolve());
        });
      const samples = [readDiagnostics()];
      await nextFrame();
      samples.push(readDiagnostics());
      await nextFrame();
      samples.push(readDiagnostics());
      const signature = (sample: (typeof samples)[number]) =>
        JSON.stringify({
          rects: sample.requiredRects,
          scroll: sample.scroll,
          visibleFieldKeys: sample.visibleFieldKeys,
        });
      const firstSample = samples[0];
      if (firstSample === undefined) {
        throw new Error("Expected grid visual anchor diagnostics to sample");
      }
      const lastSample = samples[samples.length - 1];
      if (lastSample === undefined) {
        throw new Error(
          "Expected grid visual anchor diagnostics to retain a final sample",
        );
      }
      return {
        diagnostics: lastSample,
        ready:
          samples.every((sample) => sample.ready) &&
          samples.every(
            (sample) => signature(sample) === signature(firstSample),
          ),
        samples,
      };
    },
    {
      anchor,
      scrollportSelector: gridScrollportSelector(),
      selectors: buildWorkbookGridAnchorSelectors(surface, anchor),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

async function attachWorkbookGridVisualDiagnostics(
  page: Page,
  name: string,
  surface: string,
  options: GridVisualRegressionOptions,
) {
  const testInfo = options.testInfo ?? test.info();
  const diagnostics =
    "anchor" in options
      ? await readWorkbookGridAnchorState(page, surface, options.anchor)
      : await readWorkbookGridDiagnostics(page, surface);
  await testInfo.attach(`${name}-grid-diagnostics`, {
    body: JSON.stringify(diagnostics, null, 2),
    contentType: "application/json",
  });
}

async function readWorkbookGridDiagnostics(page: Page, surface: string) {
  return page.evaluate(
    ({ scrollportSelector, shellSelector, surface }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
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
      const scrollportRect = scrollport.getBoundingClientRect();
      const visibleFieldKeys = Array.from(
        shell.querySelectorAll<HTMLElement>("[data-grid-field-key]"),
      )
        .filter((element) => {
          const rect = element.getBoundingClientRect();
          return (
            rect.width > 0 &&
            rect.height > 0 &&
            rect.right >= scrollportRect.left + 1 &&
            rect.left <= scrollportRect.right - 1 &&
            rect.bottom >= scrollportRect.top + 1 &&
            rect.top <= scrollportRect.bottom - 1
          );
        })
        .map((element) => element.getAttribute("data-grid-field-key") ?? "")
        .filter((fieldKey, index, fieldKeys) => {
          return fieldKey !== "" && fieldKeys.indexOf(fieldKey) === index;
        });
      return {
        activeElementTestId:
          document.activeElement instanceof HTMLElement
            ? document.activeElement.getAttribute("data-testid")
            : null,
        inspectorOpen:
          document.querySelector('[data-testid="timeline-inspector"]') !== null,
        ready: true,
        requiredRects: {},
        screenshotTargetTestId: `${surface}-grid-shell`,
        scroll: {
          shell: {
            clientHeight: shell.clientHeight,
            clientWidth: shell.clientWidth,
            left: Math.round(shell.scrollLeft),
            maxLeft: Math.max(0, shell.scrollWidth - shell.clientWidth),
            maxTop: Math.max(0, shell.scrollHeight - shell.clientHeight),
            scrollHeight: shell.scrollHeight,
            scrollWidth: shell.scrollWidth,
            top: Math.round(shell.scrollTop),
          },
          scrollport: {
            clientHeight: scrollport.clientHeight,
            clientWidth: scrollport.clientWidth,
            left: Math.round(scrollport.scrollLeft),
            maxLeft: Math.max(
              0,
              scrollport.scrollWidth - scrollport.clientWidth,
            ),
            maxTop: Math.max(
              0,
              scrollport.scrollHeight - scrollport.clientHeight,
            ),
            scrollHeight: scrollport.scrollHeight,
            scrollWidth: scrollport.scrollWidth,
            top: Math.round(scrollport.scrollTop),
          },
        },
        surface,
        visibleFieldKeys,
      };
    },
    {
      scrollportSelector: gridScrollportSelector(),
      shellSelector: gridShellSelector(surface),
      surface,
    },
  );
}

async function readWorkbookGridScroll(
  page: Page,
  surface: string,
): Promise<WorkbookGridVisualScrollSnapshot> {
  return page.evaluate(
    ({ scrollportSelector, shellSelector, surface }) => {
      const shell = document.querySelector<HTMLElement>(shellSelector);
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
        shellTop: Math.round(shell.scrollTop),
        shellLeft: Math.round(shell.scrollLeft),
        scrollportTop: Math.round(scrollports[0].scrollTop),
        scrollportLeft: Math.round(scrollports[0].scrollLeft),
      };
    },
    {
      scrollportSelector: gridScrollportSelector(),
      shellSelector: gridShellSelector(surface),
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
    const timestampReplacement: [RegExp, string] = [
      /\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?Z\b/g,
      "2025-01-01T00:00:00Z",
    ];
    const replacements: Array<[RegExp, string]> = [
      [
        /\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b/gi,
        "00000000-0000-0000-0000-000000000000",
      ],
      timestampReplacement,
      [/\bIR-[A-Z0-9-]+\b/g, "IR-VISUAL-FIXTURE"],
      [/Playwright Worker Admin \d+/g, "Playwright Worker Admin"],
    ];
    const formControlReplacements = replacements.filter(
      (replacement) => replacement !== timestampReplacement,
    );
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
      // Controlled inputs repaint their fixture values; do not race React by
      // replacing timestamp values in form controls during screenshot prep.
      for (const [pattern, replacement] of formControlReplacements) {
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
