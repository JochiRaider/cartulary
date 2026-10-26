import {
  gridShellTestId,
  timelineDraftEvidenceFileInputTestId,
  timelineEvidenceFileInputTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { openTimelineSurface, tinyPNG } from "./support/evidence/browser";
import { chooseEvidenceFile } from "./support/evidence/uploads";
import { createIncident } from "./support/incidents/fixtures";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  observePageRequests,
  rejectUnexpectedPageErrors,
} from "./support/runtime/observations";
import { installIncidentSocketMonitor } from "./support/transport/incidentSocket";
import { createViewRow, queryViewRows } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

function collectObjectUploadRoutes(page: Page): string[] {
  const routes: string[] = [];
  observePageRequests(page, (request) => {
    if (request.method() !== "PUT") {
      return;
    }
    const url = new URL(request.url());
    if (url.pathname.startsWith("/api/v1/object-uploads/")) {
      routes.push(url.pathname);
    }
  });
  return routes;
}

test.beforeEach(({ page }) => {
  rejectUnexpectedPageErrors(page);
});

test("attaches a screenshot to a selected Timeline row without leaving the workbook surface", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-SELECTED"),
    "Evidence selected screenshot attach",
  );
  const timelineRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e5-selected-timeline"),
      "timeline.activity_synopsis_text": "Selected row screenshot",
    },
  );
  const objectUploadRoutes = collectObjectUploadRoutes(page);

  await openTimelineSurface(page, incidentId);
  await openTimelineInspector(page, timelineRow.record_id);
  await chooseEvidenceFile(
    page.getByTestId(timelineEvidenceFileInputTestId(timelineRow.record_id)),
    {
      name: "selected-screenshot.png",
      mimeType: "image/png",
      buffer: tinyPNG(),
    },
  );

  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await expect
    .poll(
      async () => {
        const rows = await queryViewRows(
          page,
          incidentId,
          timelineViewSchemaId,
        );
        return rows.find((row) => row.record_id === timelineRow.record_id)
          ?.cells["timeline.evidence_count"]?.value;
      },
      { timeout: 30_000 },
    )
    .toBe(1);
  expect(objectUploadRoutes.length).toBeGreaterThan(0);
});

test("persists a screenshot-only Timeline row through atomic Evidence create", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-DRAFT"),
    "Evidence draft screenshot attach",
  );
  const objectUploadRoutes = collectObjectUploadRoutes(page);

  await openTimelineSurface(page, incidentId);
  await page
    .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
    .click();
  await chooseEvidenceFile(
    page.getByTestId(timelineDraftEvidenceFileInputTestId()),
    {
      name: "draft-screenshot.png",
      mimeType: "image/png",
      buffer: tinyPNG(),
    },
  );

  await expect
    .poll(
      async () => {
        const rows = await queryViewRows(
          page,
          incidentId,
          timelineViewSchemaId,
        );
        return rows.find(
          (row) => row.cells["timeline.evidence_count"]?.value === 1,
        );
      },
      { timeout: 30_000 },
    )
    .not.toBeUndefined();

  const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
  const row = rows.find(
    (candidate) => candidate.cells["timeline.evidence_count"]?.value === 1,
  );
  expect(row).toBeTruthy();
  const rowRecordId = row?.record_id;
  expect(rowRecordId).toBeTruthy();
  expect(row?.cells["timeline.activity_synopsis_text"]?.value ?? "").toBe("");
  expect(row?.cells["timeline.capture_state"]?.value).toBe("rough");
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  if (!rowRecordId) {
    throw new Error("missing screenshot-only Timeline row id");
  }
  expect(objectUploadRoutes.length).toBeGreaterThan(0);
});

test("refreshes a second live workbook from the real evidence attach stream", async ({
  browserScope,
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-SOCKET"),
    "Evidence socket evidence refresh",
  );
  const timelineRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e5-socket-timeline"),
      "timeline.activity_synopsis_text": "Second workbook evidence refresh",
    },
  );

  const listenerContext = await browserScope.newContext({
    storageState: await page.context().storageState(),
  });
  const listener = await listenerContext.newPage();
  rejectUnexpectedPageErrors(listener);
  {
    const socketMonitor = installIncidentSocketMonitor(listener, incidentId);
    await openTimelineSurface(listener, incidentId);
    await socketMonitor.waitForMessage("hello_ack");
    const listenerURL = listener.url();
    await expect(
      listener.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();

    await openTimelineSurface(page, incidentId);
    await openTimelineInspector(page, timelineRow.record_id);
    await chooseEvidenceFile(
      page.getByTestId(timelineEvidenceFileInputTestId(timelineRow.record_id)),
      {
        name: "socket-refresh.png",
        mimeType: "image/png",
        buffer: tinyPNG(),
      },
    );

    await socketMonitor.waitForMessage("record_changed", {
      matches: (message) => message.payload.record_id === timelineRow.record_id,
    });
    await expect
      .poll(async () => {
        const rows = await queryViewRows(
          listener,
          incidentId,
          timelineViewSchemaId,
        );
        const row = rows.find(
          (candidate) => candidate.record_id === timelineRow.record_id,
        );
        return [
          row?.cells["timeline.evidence_count"]?.value,
          row?.cells["timeline.has_evidence"]?.value,
        ];
      })
      .toEqual([1, true]);
    expect(listener.url()).toBe(listenerURL);

    await browserScope.closeContext(listenerContext);
  }
});
