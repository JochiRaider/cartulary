import { Buffer } from "node:buffer";
import { changeGrouping } from "@cartulary/test-utils";
import {
  gridGroupRowTestId,
  gridShellTestId,
  rowCellTestId,
  rowInspectButtonTestId,
} from "@cartulary/ui-contracts";
import type { Page, Route, TestInfo } from "@playwright/test";
import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  holdBrowserApiRequest,
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

const timelineViewSchemaId = "timeline";
const timelineApiViewSchemaId = "cartulary.view.timeline.v1";

type ViewRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, unknown>;
};

test.describe("Phase 3 workbook visual evidence", () => {
  test("V-3-GRID-01 captures the Timeline default viewport with stable row version and save-state strip", async ({
    page,
  }, testInfo) => {
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

    await captureScreenshot(
      page,
      testInfo,
      "v-3-grid-01-timeline-default",
      page.getByRole("main"),
    );
  });

  test("V-3-GRID-02 captures Timeline edit save-state visuals for active cell syncing saved and conflict states", async ({
    page,
  }, testInfo) => {
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
    await captureScreenshot(
      page,
      testInfo,
      "v-3-grid-02-active-edit-cell",
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
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
      await captureScreenshot(
        page,
        testInfo,
        "v-3-grid-02-syncing-strip",
        saveStateStrip,
      );
      await hold.release();
      await expect(saveState).toHaveText("Saved");
      await captureScreenshot(
        page,
        testInfo,
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
      await captureScreenshot(
        page,
        testInfo,
        "v-3-grid-02-conflict-strip",
        saveStateStrip,
      );
    } finally {
      await page.unroute(patchUrl, conflictHandler);
    }
  });

  test("V-3-GRID-03 captures Timeline grouped rows and currently exposed grid chrome", async ({
    page,
  }, testInfo) => {
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

    await captureScreenshot(
      page,
      testInfo,
      "v-3-grid-03-grouped-grid",
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    );
  });
});

test.describe("Phase 4 workbook visual evidence", () => {
  test("V-4-GRID-01 captures Timeline unresolved and resolved mention chips in the workbook grid", async ({
    page,
  }, testInfo) => {
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

    await captureScreenshot(
      page,
      testInfo,
      "v-4-grid-01-mention-chips",
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    );
  });

  test("V-4-GRID-02 captures Evidence access affordances on the required Evidence surface", async ({
    page,
  }, testInfo) => {
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

    await captureScreenshot(
      page,
      testInfo,
      "v-4-grid-02-evidence-access",
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
    );
  });

  test("V-4-GRID-03 captures Task Requests system view fields through the generic workbook grid", async ({
    page,
  }, testInfo) => {
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

    await captureScreenshot(
      page,
      testInfo,
      "v-4-grid-03-task-requests",
      page.getByTestId(gridShellTestId(taskRequestsViewSchemaId)),
    );
  });
});

test.describe("Phase 5 workbook visual evidence", () => {
  test("V-5-GRID-01 captures requested and available Evidence states on the required Evidence surface", async ({
    page,
  }, testInfo) => {
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
    await captureScreenshot(
      page,
      testInfo,
      "v-5-grid-01-requested-evidence",
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
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
    await captureScreenshot(
      page,
      testInfo,
      "v-5-grid-01-available-evidence",
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
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
    await captureScreenshot(
      page,
      testInfo,
      "v-5-grid-02-blocked-preview",
      page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
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
    await captureScreenshot(
      page,
      testInfo,
      "v-5-grid-02-timeline-evidence-badge",
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    );
  });
});

async function captureScreenshot(
  page: Page,
  testInfo: TestInfo,
  name: string,
  locator = page.getByRole("main"),
) {
  await expect(locator).toBeVisible();
  await page.evaluate(() => {
    document.documentElement.dataset.visualSnapshot = "true";
  });
  const screenshot = await locator.screenshot({
    animations: "disabled",
    caret: "hide",
  });
  await testInfo.attach(name, {
    body: screenshot,
    contentType: "image/png",
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
