import { Buffer } from "node:buffer";
import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import {
  assertMarkerAnchoredToGridTarget,
  changeGrouping,
} from "@cartulary/test-utils";
import {
  cartularyDefaultThemeId,
  cellPresenceMarkerTestId,
  conflictMarkerTestId,
  dataTestIdSelector,
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidencePreviewButtonTestId,
  gridActionsHeaderTestId,
  gridGroupRowTestId,
  gridRowGutterTestId,
  gridSavedRowsSelector,
  gridScrollportSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  incidentControlsPanelTestId,
  incidentMembershipListTestId,
  pendingQueueCountTestId,
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
  timelineInspectorSectionTestId,
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
  gridSavedRows,
  holdBrowserApiRequest,
  queryViewRows,
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
  installIncidentSocketMonitor,
  installPatchController,
  installPatchTransportFailureController,
  openIncidentAsTrackedUserReady,
  successfulPatchCalls,
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

function tagActionsPayload(tagNames: string[]) {
  return {
    kind: "collection_actions_v1",
    actions: tagNames.map((tagName) => ({
      op: "add_tag",
      tag_name: tagName,
    })),
  };
}

type GridVisualRegressionOptions = {
  maxDiffPixels?: number;
  testInfo?: TestInfo;
} & (
  | { scroll: GridVisualScrollState; anchor?: never }
  | { anchor: GridVisualAnchor; scroll?: never }
);

test.describe("FE-P2 workbook visual readiness", () => {
  test("FE-V-P2-01 Capture Default Timeline workbook shell with top bar, view bar, dense Timeline grid, row-context inspector, and status strip.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEV2SHELL"),
      "FE-P2 visual default shell",
    );

    const rows: ViewRow[] = [];
    const fixtureRows = [
      "Login attempt with valid user",
      "Password spray from single source",
      "Failed MFA challenge",
      "Suspicious automation execution",
      "Outbound connection to uncommon provider",
      "New service installed",
      "Potential credential access",
      "User accessed sensitive share",
      "Data archived to temporary directory",
      "Archive staged for exfiltration",
      "Alert from endpoint rule triggered",
      "Host isolated by containment playbook",
      "Scheduled task removed",
      "Credential reset completed",
      "Investigation opened",
      "Containment review assigned",
      "Remote shell attempt blocked",
      "Cloud sign-in risk elevated",
      "Analyst comment added",
      "Final verification queued",
    ];
    for (const [index, summary] of fixtureRows.entries()) {
      rows.push(
        (await createViewRow(page, incidentId, timelineViewSchemaId, {
          client_txn_id: uniqueTxn(`FEV2SHELL-ROW-${index + 1}`),
          "timeline.occurred_at": new Date(
            Date.UTC(2026, 3, 18, 14, 12 + index * 2, 34),
          ).toISOString(),
          "timeline.summary": summary,
          "timeline.details": `Default Timeline workbook shell fixture row ${
            index + 1
          }`,
          "timeline.host_refs": collectionActionsPayload([
            index % 3 === 0 ? "host-gamma" : "host-alpha",
          ]),
          "timeline.identity_refs": collectionActionsPayload([
            index % 2 === 0
              ? "identity-alpha@example.test"
              : "identity-beta@example.test",
          ]),
          "timeline.tags": tagActionsPayload([
            index % 4 === 0 ? "review" : "triage",
            index % 5 === 0 ? "evidence" : "timeline",
          ]),
        })) as ViewRow,
      );
    }
    const rowSummariesById = new Map(
      rows.map((row, index) => [row.record_id, fixtureRows[index] ?? ""]),
    );

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);

    const shell = page.getByTestId(workbookShellReadyTestId());
    await expect(shell).toBeVisible();
    await expect(shell).toHaveAttribute(
      "data-active-view-schema-id",
      timelineViewSchemaId,
    );
    for (const slot of workbookShellSlots) {
      await expect(
        shell.locator(dataTestIdSelector(workbookShellSlotTestId(slot))),
      ).toBeVisible();
    }
    await expect(page.getByTestId(incidentControlsPanelTestId())).toHaveCount(
      0,
    );
    await expect(page.getByTestId("incident-summary-key")).toHaveCount(0);
    await expect(page.getByTestId("incident-patch-button")).toHaveCount(0);
    await expect(page.getByTestId(incidentMembershipListTestId())).toHaveCount(
      0,
    );
    await expect(page.getByText("Phase 3 workbook")).toHaveCount(0);
    await expect(page.getByText(/Timeline mutation substrate/u)).toHaveCount(0);
    await expect(
      page.getByTestId(surfaceTabTestId(timelineViewSchemaId)),
    ).toHaveAttribute("aria-current", "page");
    await expect(
      page.getByTestId(systemViewSwitcherTriggerTestId()),
    ).toBeVisible();
    await expect(
      page.getByTestId(savedViewSelectorTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

    const grid = page.getByTestId(gridShellTestId(timelineViewSchemaId));
    await expect(grid).toBeVisible();
    await expect
      .poll(
        async () =>
          (await queryViewRows(page, incidentId, timelineViewSchemaId)).length,
      )
      .toBe(rows.length);
    await expect
      .poll(async () => gridSavedRows(page, timelineViewSchemaId).count())
      .toBeGreaterThanOrEqual(12);
    const renderedRecordIds = await grid
      .locator(gridSavedRowsSelector())
      .evaluateAll((rowElements) =>
        rowElements.map(
          (rowElement) => rowElement.getAttribute("data-grid-record-id") ?? "",
        ),
      );
    expect(rows.map((row) => row.record_id)).toEqual(
      expect.arrayContaining(renderedRecordIds),
    );
    const selectedRowId = renderedRecordIds[0];
    const selectedRow = rows.find((row) => row.record_id === selectedRowId);
    if (selectedRow === undefined) {
      throw new Error(`FE-V-P2 fixture selected unknown row ${selectedRowId}`);
    }
    const selectedGridRow = grid.locator(
      `[data-grid-record-id="${selectedRow.record_id}"]`,
    );
    await selectedGridRow.click();
    await expect(selectedGridRow).toHaveAttribute("aria-selected", "true");
    await expect(
      selectedGridRow.locator(
        '[role="rowheader"][data-grid-field-key="__cartulary_row_gutter__"]',
      ),
    ).toBeVisible();

    const defaultTimelineFields = [
      "timeline.occurred_at",
      "timeline.summary",
      "timeline.host_refs",
      "timeline.identity_refs",
      "timeline.evidence_count",
      "timeline.tags",
      "timeline.edited_at",
    ];
    const headerFieldKeys = await grid
      .locator('[role="columnheader"] [data-grid-field-key]')
      .evaluateAll((headers) =>
        headers.map((header) => header.getAttribute("data-grid-field-key")),
      );
    expect(headerFieldKeys).toEqual(defaultTimelineFields);
    for (const fieldKey of defaultTimelineFields) {
      await expect(
        grid.locator(
          `[data-grid-record-id="${selectedRow.record_id}"] [data-grid-field-key="${fieldKey}"]`,
        ),
      ).toHaveCount(1);
    }

    await expect(
      page.getByTestId(
        rowCellTestId(selectedRow.record_id, "timeline.summary"),
      ),
    ).toHaveValue(rowSummariesById.get(selectedRow.record_id) ?? "");

    await expect(page.getByTestId("timeline-inspector")).toBeVisible();
    await expect(page.getByTestId("timeline-inspector")).toContainText(
      selectedRow.record_id,
    );
    for (const section of [
      "details",
      "relationships",
      "evidence",
      "history",
    ] as const) {
      await expect(
        page.getByTestId(timelineInspectorSectionTestId(section)),
      ).toBeVisible();
    }
    await page
      .getByTestId(timelineEvidenceFileInputTestId(selectedRow.record_id))
      .setInputFiles({
        name: "default-timeline-workbook-shell.png",
        mimeType: "image/png",
        buffer: tinyPNG(),
      });
    await expect(
      page.getByTestId(
        rowCellTestId(selectedRow.record_id, "timeline.evidence_count"),
      ),
    ).toHaveText("1");

    await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
      scroll: { top: 0, left: "left" },
    });
    const summaryCell = page.getByTestId(
      rowCellTestId(selectedRow.record_id, "timeline.summary"),
    );
    await summaryCell.focus();
    await expect(summaryCell).toBeFocused();
    const timelineScrollportSelector = `${dataTestIdSelector(
      gridShellTestId(timelineViewSchemaId),
    )} ${gridScrollportSelector()}`;
    await expect
      .poll(() =>
        page.evaluate(
          (selector) => ({
            gridLeft:
              document.querySelector<HTMLElement>(selector)?.scrollLeft ?? -1,
            windowY: window.scrollY,
          }),
          timelineScrollportSelector,
        ),
      )
      .toEqual({ gridLeft: 0, windowY: 0 });

    await assertViewportVisualRegression(
      page,
      "fe-v-p2-01-default-timeline-workbook-shell",
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
              conflict_token: "visual-conflict-token",
              record_id: timelineRow.record_id,
              field_key: "timeline.summary",
              conflict_resolution_class: "text_compare_merge",
              base_row_version: timelineRow.row_version,
              current_row_version: timelineRow.row_version + 1,
              base_value: "Active visual edit",
              server_value: "Server visual edit",
              client_value: "Conflict visual edit",
              server_updated_by: "visual-remote-user",
              server_updated_at: "2026-06-02T12:00:00Z",
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

test.describe("FE-P3 visual readiness", () => {
  test("FE-V-P3-01 Capture frozen column, resize handle, fill-down handle, edit cell, group outline row, and empty successful query grid-adapter fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEV3GRID"),
      "FE-P3 grid adapter visual fixture",
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("FEV3GRID-ROW"),
      "timeline.occurred_at": "2026-05-31T10:00:00Z",
      "timeline.summary": "FE-P3 visual adapter row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await injectFeP3GridAdapterVisualFixture(page);

    const fixture = page.locator("[data-design-fixture='fe-p3-grid-adapter']");
    await expect(fixture).toBeVisible();
    for (const fixtureId of [
      "FE-VFIX-09",
      "FE-VFIX-10",
      "FE-VFIX-11",
      "FE-VFIX-12",
      "FE-VFIX-13",
      "FE-VFIX-15",
    ]) {
      await expect(
        fixture.locator(`[data-fixture-id='${fixtureId}']`),
      ).toBeVisible();
    }
    await assertVisualRegression(
      page,
      "fe-v-p3-01-grid-adapter-fixtures",
      fixture,
    );
  });
});

test.describe("FE-P4 visual readiness", () => {
  test("FE-V-P4-01 Capture save-state strip, pending replay indication, inline edit cell, and empty successful Timeline query fixtures.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1440, height: 900 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEV4VISUAL"),
      "FE-P4 visual readiness",
    );
    const timelineRow = (await createViewRow(
      page,
      incidentId,
      timelineViewSchemaId,
      {
        client_txn_id: uniqueTxn("FEV4VISUAL-ROW"),
        "timeline.occurred_at": "2026-06-03T10:00:00Z",
        "timeline.summary": "FE-P4 visual editable row",
      },
    )) as ViewRow;

    await page.goto(`/?incident_id=${incidentId}`);
    await maskIncidentIdentity(page, incidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");

    const summaryInput = page.getByTestId(
      rowCellTestId(timelineRow.record_id, "timeline.summary"),
    );
    await expect(summaryInput).toHaveValue("FE-P4 visual editable row");
    await summaryInput.focus();
    await summaryInput.fill("FE-P4 active visual edit");
    await assertWorkbookGridVisualRegression(
      page,
      "fe-v-p4-01-active-edit-cell",
      timelineViewSchemaId,
      { scroll: { top: 0, left: "left" } },
    );

    const patchController = await installPatchTransportFailureController(page);
    try {
      patchController.disconnect();
      await summaryInput.press("Enter");
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Syncing");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toBeVisible();
      await expect(page.getByTestId(pendingQueueCountTestId())).toContainText(
        "1",
      );
      await normalizeWorkbookGridVisualState(page, timelineViewSchemaId, {
        scroll: { top: 0, left: "left" },
      });
      await assertVisualRegression(
        page,
        "fe-v-p4-01-pending-replay-status",
      );

      patchController.connect();
      await expect
        .poll(() => successfulPatchCalls(patchController.calls).length)
        .toBe(1);
      await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
      await expect(page.getByTestId(pendingQueueNoticeTestId())).toHaveCount(0);
    } finally {
      patchController.connect();
      await patchController.dispose();
    }

    const emptyIncidentId = await createIncident(
      page,
      uniqueIncidentKey("FEV4EMPTY"),
      "FE-P4 empty Timeline query",
    );
    await page.goto(`/?incident_id=${emptyIncidentId}`);
    await maskIncidentIdentity(page, emptyIncidentId);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expect(page.getByTestId(saveStateTestId())).toHaveText("Saved");
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await expect
      .poll(
        async () =>
          (await queryViewRows(page, emptyIncidentId, timelineViewSchemaId))
            .length,
      )
      .toBe(0);
    await expect(gridSavedRows(page, timelineViewSchemaId)).toHaveCount(0);
    await expect(page.getByText("Draft timeline row")).toBeVisible();
    await expect(
      page.getByRole("button", { name: "Create blank row" }),
    ).toBeVisible();
    await assertWorkbookGridVisualRegression(
      page,
      "fe-v-p4-01-empty-timeline-query",
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
    const primarySocket = installIncidentSocketMonitor(page, incidentId);

    let remotePage: Page | null = null;
    try {
      await page.goto(`/?incident_id=${incidentId}`);
      await primarySocket.waitForAcceptedSocket();
      await maskIncidentIdentity(page, incidentId);
      await expect(
        page.getByTestId(
          rowCellTestId(timelineRow.record_id, "timeline.summary"),
        ),
      ).toHaveValue("Presence visual row");

      const remoteSession = await openIncidentAsTrackedUserReady(
        browser,
        sessionTracker,
        {
          createdBy: "V-6-GRID-01",
          email: remote.email,
          incidentId,
          password: remote.initial_password,
          purpose: "Phase 6 visual presence analyst",
          readyRecordId: timelineRow.record_id,
          userId: remote.user_id,
        },
      );
      remotePage = remoteSession.page;
      const markerStartAt = primarySocket.messageCount();
      const fieldKey = "timeline.summary";
      const markerDelta = primarySocket.waitForMessage("presence_delta", {
        matches: (message) => {
          const presence = message.payload.presence;
          return (
            presence !== null &&
            typeof presence === "object" &&
            !Array.isArray(presence) &&
            "record_id" in presence &&
            presence.record_id === timelineRow.record_id &&
            "field_key" in presence &&
            presence.field_key === fieldKey &&
            "mode" in presence &&
            presence.mode === "editing"
          );
        },
        startAt: markerStartAt,
      });
      await remotePage
        .getByTestId(rowCellTestId(timelineRow.record_id, fieldKey))
        .focus();
      await markerDelta;
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
        targetTestId: gridRowGutterTestId(
          timelineViewSchemaId,
          timelineRow.record_id,
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

test.describe("FE-P11 visual readiness", () => {
  test("FE-V-P11-03 Capture exposed dark_graphite token and theme states with deterministic density, color, component, focus, and semantic-state samples.", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 1280, height: 720 });
    const incidentId = await createIncident(
      page,
      uniqueIncidentKey("FEV11THEME"),
      "FE-P11 exposed theme visual fixture",
    );
    await createViewRow(page, incidentId, timelineViewSchemaId, {
      client_txn_id: uniqueTxn("FEV11THEME-ROW"),
      "timeline.occurred_at": "2026-05-31T11:00:00Z",
      "timeline.summary": "FE-P11 exposed theme fixture row",
    });

    await page.goto(`/?incident_id=${incidentId}`);
    await expect(page.getByTestId(workbookShellReadyTestId())).toBeVisible();
    await expect(page.locator("main.cartulary-shell").first()).toHaveAttribute(
      "data-cartulary-theme",
      cartularyDefaultThemeId,
    );
    await assertExposedThemeCssVariables(page);
    await injectExposedThemeVisualFixture(page);

    const fixture = page.locator("[data-design-fixture='exposed-theme']");
    await expect(fixture).toBeVisible();
    await assertVisualRegression(
      page,
      "fe-v-p11-03-exposed-theme-states",
      fixture,
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

const exposedThemeCssVars = [
  "--ct-colors-accent",
  "--ct-colors-canvas",
  "--ct-colors-surface-1",
  "--ct-colors-surface-2",
  "--ct-colors-ink",
  "--ct-colors-ink-muted",
  "--ct-colors-semantic-success",
  "--ct-colors-semantic-caution",
  "--ct-colors-semantic-conflict",
  "--ct-colors-semantic-destructive",
  "--ct-density-default-rowHeight",
  "--ct-density-default-cellPadding",
  "--ct-component-button-primary-backgroundColor",
  "--ct-component-button-primary-textColor",
  "--ct-component-button-secondary-backgroundColor",
  "--ct-component-button-secondary-border",
  "--ct-component-button-danger-textColor",
  "--ct-component-text-input-backgroundColor",
  "--ct-component-chip-backgroundColor",
  "--ct-component-grid-cell-padding",
  "--ct-component-focus-ring-border",
] as const;

async function assertExposedThemeCssVariables(page: Page) {
  const missingVars = await page.evaluate(
    (varNames) => {
      const styles = window.getComputedStyle(document.documentElement);
      return varNames.filter((name) => !styles.getPropertyValue(name).trim());
    },
    [...exposedThemeCssVars],
  );
  expect(missingVars).toEqual([]);
}

async function injectFeP3GridAdapterVisualFixture(page: Page) {
  await page.evaluate(() => {
    document
      .querySelector("style[data-design-fixture-style='fe-p3-grid-adapter']")
      ?.remove();
    document
      .querySelector("[data-design-fixture='fe-p3-grid-adapter']")
      ?.remove();

    const main = document.querySelector("main.cartulary-shell");
    if (!(main instanceof HTMLElement)) {
      throw new Error("Expected workbook shell main before FE-P3 fixture");
    }

    const style = document.createElement("style");
    style.dataset.designFixtureStyle = "fe-p3-grid-adapter";
    style.textContent = `
      [data-design-fixture='fe-p3-grid-adapter'] {
        position: fixed;
        inset: var(--ct-spacing-xl);
        box-sizing: border-box;
        display: grid;
        grid-template-columns: minmax(0, 1fr) 20rem;
        gap: var(--ct-spacing-md);
        overflow: hidden;
        background: var(--ct-colors-canvas);
        color: var(--ct-colors-ink);
        border: var(--ct-border-strong);
        border-radius: var(--ct-rounded-lg);
        padding: var(--ct-spacing-lg);
        box-shadow: var(--ct-elevation-panel);
        font-family: var(--ct-typography-ui-fontFamily);
        font-size: var(--ct-typography-ui-fontSize);
        font-weight: var(--ct-typography-ui-fontWeight);
        letter-spacing: var(--ct-typography-ui-letterSpacing);
        line-height: var(--ct-typography-ui-lineHeight);
        z-index: 1000;
      }

      [data-design-fixture='fe-p3-grid-adapter'] * {
        box-sizing: border-box;
      }

      .fe-p3-grid-fixture-table {
        min-width: 0;
        overflow: hidden;
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-md);
        background: var(--ct-colors-surface-1);
      }

      .fe-p3-grid-fixture-row {
        display: grid;
        grid-template-columns: 10rem 16rem 12rem 14rem 14rem;
        min-width: 66rem;
      }

      .fe-p3-grid-fixture-head,
      .fe-p3-grid-fixture-cell {
        position: relative;
        min-width: 0;
        min-height: 3.75rem;
        display: flex;
        align-items: center;
        gap: var(--ct-spacing-xs);
        overflow: hidden;
        padding: var(--ct-density-default-cellPadding);
        border-inline-end: var(--ct-border-hairline);
        border-block-end: var(--ct-border-hairline);
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink);
      }

      .fe-p3-grid-fixture-head {
        min-height: 3rem;
        background: var(--ct-colors-surface-2);
        color: var(--ct-colors-ink-muted);
        font-family: var(--ct-typography-metadata-fontFamily);
        font-size: var(--ct-typography-metadata-fontSize);
        font-weight: var(--ct-typography-metadata-fontWeight);
      }

      .fe-p3-grid-fixture-frozen {
        position: sticky;
        left: 0;
        z-index: 2;
        background: var(--ct-colors-surface-2);
        box-shadow: 0.75rem 0 1rem rgba(0, 0, 0, 0.28);
      }

      .fe-p3-grid-fixture-resize-handle {
        position: absolute;
        inset-block: 0.45rem;
        inset-inline-end: 0.2rem;
        width: 0.25rem;
        border-radius: var(--ct-rounded-sm);
        background: var(--ct-colors-hairline-focus);
      }

      .fe-p3-grid-fixture-active {
        outline: var(--ct-component-focus-ring-border);
        outline-offset: -0.2rem;
        background: var(--ct-colors-surface-3);
      }

      .fe-p3-grid-fixture-editor {
        width: 100%;
        min-width: 0;
        border: var(--ct-border-strong);
        border-radius: var(--ct-rounded-sm);
        padding: 0.45rem 0.55rem;
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink);
        font: inherit;
      }

      .fe-p3-grid-fixture-fill {
        position: absolute;
        right: 0.2rem;
        bottom: 0.2rem;
        width: 0.65rem;
        height: 0.65rem;
        border: 0.15rem solid var(--ct-colors-hairline-focus);
        border-radius: var(--ct-rounded-sm);
        background: var(--ct-colors-surface-1);
      }

      .fe-p3-grid-fixture-group {
        grid-column: 1 / -1;
        min-height: 3.5rem;
        background: var(--ct-colors-surface-2);
        color: var(--ct-colors-ink-muted);
        font-weight: 600;
      }

      .fe-p3-grid-fixture-tree-toggle {
        display: inline-grid;
        place-items: center;
        width: 1.35rem;
        height: 1.35rem;
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-sm);
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink);
      }

      .fe-p3-grid-fixture-side {
        display: grid;
        grid-template-rows: auto 1fr;
        gap: var(--ct-spacing-sm);
        min-width: 0;
      }

      .fe-p3-grid-fixture-caption {
        margin: 0;
        color: var(--ct-colors-ink-muted);
        font-family: var(--ct-typography-metadata-fontFamily);
        font-size: var(--ct-typography-metadata-fontSize);
      }

      .fe-p3-grid-fixture-empty {
        display: grid;
        place-items: center;
        min-height: 16rem;
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-md);
        background: var(--ct-colors-surface-1);
        color: var(--ct-colors-ink-muted);
        text-align: center;
      }

      .fe-p3-grid-fixture-empty strong {
        display: block;
        margin-block-end: var(--ct-spacing-xs);
        color: var(--ct-colors-ink);
      }
    `;
    document.head.append(style);

    const fixture = document.createElement("section");
    fixture.dataset.designFixture = "fe-p3-grid-adapter";
    fixture.setAttribute("aria-label", "FE-P3 grid adapter visual fixtures");
    fixture.innerHTML = `
      <div class="fe-p3-grid-fixture-table" role="grid" aria-label="Adapter fixture grid">
        <div class="fe-p3-grid-fixture-row" role="row">
          <div class="fe-p3-grid-fixture-head fe-p3-grid-fixture-frozen" role="columnheader" data-fixture-id="FE-VFIX-09">Record</div>
          <div class="fe-p3-grid-fixture-head" role="columnheader" data-fixture-id="FE-VFIX-10">Summary<span class="fe-p3-grid-fixture-resize-handle" aria-hidden="true"></span></div>
          <div class="fe-p3-grid-fixture-head" role="columnheader">State</div>
          <div class="fe-p3-grid-fixture-head" role="columnheader">Assignee</div>
          <div class="fe-p3-grid-fixture-head" role="columnheader">Last edit</div>
        </div>
        <div class="fe-p3-grid-fixture-row" role="row">
          <div class="fe-p3-grid-fixture-cell fe-p3-grid-fixture-group" role="rowheader" data-fixture-id="FE-VFIX-13"><span class="fe-p3-grid-fixture-tree-toggle" aria-hidden="true">v</span> reviewed group, 2 rows</div>
        </div>
        <div class="fe-p3-grid-fixture-row" role="row">
          <div class="fe-p3-grid-fixture-cell fe-p3-grid-fixture-frozen" role="rowheader">record-1</div>
          <div class="fe-p3-grid-fixture-cell fe-p3-grid-fixture-active" role="gridcell" data-fixture-id="FE-VFIX-12"><input class="fe-p3-grid-fixture-editor" value="Edit cell adapter" aria-label="Summary editor" readonly><span class="fe-p3-grid-fixture-fill" data-fixture-id="FE-VFIX-11" aria-hidden="true"></span></div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">reviewed</div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">Analyst</div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">saved</div>
        </div>
        <div class="fe-p3-grid-fixture-row" role="row">
          <div class="fe-p3-grid-fixture-cell fe-p3-grid-fixture-frozen" role="rowheader">record-2</div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">Frozen column remains pinned</div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">rough</div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">Unassigned</div>
          <div class="fe-p3-grid-fixture-cell" role="gridcell">clean</div>
        </div>
      </div>
      <aside class="fe-p3-grid-fixture-side" aria-label="Empty successful query fixture">
        <p class="fe-p3-grid-fixture-caption">Adapter-owned visual states only. Row-gutter presence remains FE-P7 and grouped-result query ownership remains FE-P8.</p>
        <div class="fe-p3-grid-fixture-empty" data-fixture-id="FE-VFIX-15">
          <span><strong>No rows match this query</strong>Successful empty result</span>
        </div>
      </aside>
    `;
    main.append(fixture);
  });
}

async function injectExposedThemeVisualFixture(page: Page) {
  await page.evaluate(() => {
    const existingStyle = document.querySelector(
      "style[data-design-fixture-style='exposed-theme']",
    );
    existingStyle?.remove();
    const existingFixture = document.querySelector(
      "[data-design-fixture='exposed-theme']",
    );
    existingFixture?.remove();

    const main = document.querySelector("main.cartulary-shell");
    if (!(main instanceof HTMLElement)) {
      throw new Error("Expected workbook shell main before theme fixture");
    }

    const style = document.createElement("style");
    style.dataset.designFixtureStyle = "exposed-theme";
    style.textContent = `
      [data-design-fixture='exposed-theme'] {
        position: fixed;
        inset: var(--ct-spacing-xl);
        box-sizing: border-box;
        display: grid;
        grid-template-columns: 1.1fr 0.9fr;
        gap: var(--ct-spacing-md);
        overflow: hidden;
        background: var(--ct-colors-canvas);
        color: var(--ct-colors-ink);
        border: var(--ct-border-strong);
        border-radius: var(--ct-rounded-lg);
        padding: var(--ct-spacing-lg);
        box-shadow: var(--ct-elevation-panel);
        font-family: var(--ct-typography-ui-fontFamily);
        font-size: var(--ct-typography-ui-fontSize);
        font-weight: var(--ct-typography-ui-fontWeight);
        letter-spacing: var(--ct-typography-ui-letterSpacing);
        line-height: var(--ct-typography-ui-lineHeight);
        z-index: 1000;
      }

      [data-design-fixture='exposed-theme'] * {
        box-sizing: border-box;
      }

      .theme-fixture-panel {
        display: grid;
        gap: var(--ct-spacing-sm);
        align-content: start;
        min-width: 0;
        background: var(--ct-colors-surface-1);
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-md);
        padding: var(--ct-spacing-md);
      }

      .theme-fixture-title {
        margin: 0;
        color: var(--ct-colors-ink);
        font-family: var(--ct-typography-surface-title-fontFamily);
        font-size: var(--ct-typography-surface-title-fontSize);
        font-weight: var(--ct-typography-surface-title-fontWeight);
        letter-spacing: var(--ct-typography-surface-title-letterSpacing);
        line-height: var(--ct-typography-surface-title-lineHeight);
      }

      .theme-fixture-note {
        margin: 0;
        color: var(--ct-colors-ink-muted);
        font-family: var(--ct-typography-metadata-fontFamily);
        font-size: var(--ct-typography-metadata-fontSize);
        font-weight: var(--ct-typography-metadata-fontWeight);
        letter-spacing: var(--ct-typography-metadata-letterSpacing);
        line-height: var(--ct-typography-metadata-lineHeight);
      }

      .theme-fixture-swatches,
      .theme-fixture-components,
      .theme-fixture-states {
        display: grid;
        grid-template-columns: repeat(2, minmax(0, 1fr));
        gap: var(--ct-spacing-xs);
      }

      .theme-swatch,
      .theme-state {
        min-height: var(--ct-density-default-rowHeight);
        display: flex;
        align-items: center;
        gap: var(--ct-spacing-xs);
        border: var(--ct-border-hairline);
        border-radius: var(--ct-rounded-sm);
        padding: var(--ct-density-default-cellPadding);
        background: var(--ct-colors-surface-2);
        color: var(--ct-colors-ink);
      }

      .theme-swatch::before,
      .theme-state::before {
        content: "";
        inline-size: var(--ct-component-icon-inline-size);
        block-size: var(--ct-component-icon-inline-size);
        border-radius: var(--ct-rounded-pill);
        border: var(--ct-border-hairline);
        flex: 0 0 auto;
      }

      .theme-swatch[data-token='accent']::before {
        background: var(--ct-colors-accent);
      }

      .theme-swatch[data-token='surface']::before {
        background: var(--ct-colors-surface-3);
      }

      .theme-swatch[data-token='ink']::before {
        background: var(--ct-colors-ink);
      }

      .theme-swatch[data-token='hairline']::before {
        background: var(--ct-colors-hairline-strong);
      }

      .theme-state[data-state='success']::before {
        background: var(--ct-colors-semantic-success);
      }

      .theme-state[data-state='caution']::before {
        background: var(--ct-colors-semantic-caution);
      }

      .theme-state[data-state='conflict']::before {
        background: var(--ct-colors-semantic-conflict);
      }

      .theme-state[data-state='destructive']::before {
        background: var(--ct-colors-semantic-destructive);
      }

      .theme-button {
        min-height: var(--ct-density-default-rowHeight);
        border: 0;
        border-radius: var(--ct-component-button-primary-rounded);
        padding: var(--ct-component-button-primary-padding);
        font-family: var(--ct-typography-button-fontFamily);
        font-size: var(--ct-typography-button-fontSize);
        font-weight: var(--ct-typography-button-fontWeight);
        letter-spacing: var(--ct-typography-button-letterSpacing);
        line-height: var(--ct-typography-button-lineHeight);
      }

      .theme-button-primary {
        background: var(--ct-component-button-primary-backgroundColor);
        color: var(--ct-component-button-primary-textColor);
      }

      .theme-button-secondary {
        background: var(--ct-component-button-secondary-backgroundColor);
        color: var(--ct-component-button-secondary-textColor);
        border: var(--ct-component-button-secondary-border);
      }

      .theme-button-danger {
        background: var(--ct-component-button-danger-backgroundColor);
        color: var(--ct-component-button-danger-textColor);
        border: var(--ct-border-hairline);
      }

      .theme-input,
      .theme-grid-cell {
        min-height: var(--ct-density-default-rowHeight);
        width: 100%;
        border: var(--ct-component-text-input-border);
        border-radius: var(--ct-component-text-input-rounded);
        background: var(--ct-component-text-input-backgroundColor);
        color: var(--ct-component-text-input-textColor);
        padding: var(--ct-component-text-input-padding);
        font: inherit;
      }

      .theme-chip {
        display: inline-flex;
        align-items: center;
        width: max-content;
        border: var(--ct-component-chip-border);
        border-radius: var(--ct-component-chip-rounded);
        background: var(--ct-component-chip-backgroundColor);
        color: var(--ct-component-chip-textColor);
        padding: var(--ct-component-chip-padding);
      }

      .theme-grid-cell {
        display: flex;
        align-items: center;
        background: var(--ct-component-grid-cell-backgroundColor);
        color: var(--ct-component-grid-cell-textColor);
        padding: var(--ct-component-grid-cell-padding);
        font-family: var(--ct-typography-grid-cell-fontFamily);
        font-size: var(--ct-typography-grid-cell-fontSize);
        font-weight: var(--ct-typography-grid-cell-fontWeight);
        letter-spacing: var(--ct-typography-grid-cell-letterSpacing);
        line-height: var(--ct-typography-grid-cell-lineHeight);
      }

      .theme-focus-sample {
        outline: var(--ct-component-focus-ring-border);
        outline-offset: var(--ct-component-focus-ring-offset);
      }
    `;
    document.head.append(style);

    const fixture = document.createElement("section");
    fixture.dataset.designFixture = "exposed-theme";
    fixture.setAttribute("aria-label", "Exposed theme token state fixture");
    fixture.innerHTML = `
      <div class="theme-fixture-panel">
        <h2 class="theme-fixture-title">dark_graphite token states</h2>
        <p class="theme-fixture-note">Generated CSS variables applied through the workbook runtime.</p>
        <div class="theme-fixture-swatches" aria-label="Color token samples">
          <div class="theme-swatch" data-token="accent">Accent</div>
          <div class="theme-swatch" data-token="surface">Surface</div>
          <div class="theme-swatch" data-token="ink">Ink</div>
          <div class="theme-swatch" data-token="hairline">Hairline</div>
        </div>
        <div class="theme-fixture-states" aria-label="Semantic state samples">
          <div class="theme-state" data-state="success">Success state</div>
          <div class="theme-state" data-state="caution">Caution state</div>
          <div class="theme-state" data-state="conflict">Conflict state</div>
          <div class="theme-state" data-state="destructive">Destructive state</div>
        </div>
      </div>
      <div class="theme-fixture-panel">
        <h2 class="theme-fixture-title">Component and density states</h2>
        <p class="theme-fixture-note">Default density, buttons, input, chip, focus, and grid-cell tokens.</p>
        <div class="theme-fixture-components">
          <button class="theme-button theme-button-primary" type="button">Primary</button>
          <button class="theme-button theme-button-secondary theme-focus-sample" type="button">Secondary focus</button>
          <button class="theme-button theme-button-danger" type="button">Danger</button>
          <span class="theme-chip">Evidence chip</span>
        </div>
        <input class="theme-input" value="Readonly token input" readonly />
        <div class="theme-grid-cell">Grid cell typography and default density</div>
      </div>
    `;
    main.append(fixture);
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
