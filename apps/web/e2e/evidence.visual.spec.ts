import { Buffer } from "node:buffer";
import type {
  EvidenceCreateRequest,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import {
  evidenceAccessMessageTestId,
  evidenceAccessStateTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
  genericCreateFieldTestId,
  gridShellTestId,
  timelineEvidenceFileInputTestId,
  timelineInspectorSectionTestId,
  workbookInspectorPanelTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect } from "./fixtures";
import {
  createEvidenceFixtureRow,
  createUploadedEvidenceFixture,
  type EvidenceUploadOptions,
} from "./support/evidence/fixtures";
import { chooseEvidenceFile } from "./support/evidence/uploads";
import { createIncident } from "./support/incidents/fixtures";
import { navigateVisualApplication } from "./support/runtime/applicationReadiness";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import { installPublicErrorResponse } from "./support/transport/requestInterception";
import {
  assertViewportVisualRegression,
  assertVisualRegression,
  type VisualAnchor,
} from "./support/visual/capture";
import { test } from "./support/visual/fixtures";
import {
  createRichTimelineIncident,
  createRichTimelineSource,
  createVisualTimelineRow,
} from "./support/visual/timelineVisualFixture";
import { assertWorkbookGridVisualRegression } from "./support/visual/workbookCaptures";
import {
  attachWorkbookGridVisualDiagnostics,
  readWorkbookGridScroll,
  setWorkbookGridScroll,
  type WorkbookGridVisualScrollSnapshot,
  waitForVisualLayoutFrame,
} from "./support/visual/workbookGeometry";
import {
  mountedGridCell,
  mountedGridTarget,
} from "./support/workbook/keyboardGrid";
import { createViewRow, patchRecord } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";
import {
  openTimelineEvidenceFixture,
  retainTimelineEvidencePartialResult,
} from "./support/workbook/timelineRelatedEvidence";

type VisualEvidenceRowStateKey =
  | "available"
  | "quarantined"
  | "pending_receipt"
  | "requested";

async function assertEvidenceAccessVisualRegression(
  page: Page,
  name: string,
  actionRecordId: string,
) {
  let expectedScroll: WorkbookGridVisualScrollSnapshot;
  try {
    const evidenceFixture = page.getByRole("region", {
      name: "Workbook work area",
      exact: true,
    });
    await assertVisualRegression(page, name, evidenceFixture, {
      renderSurface: evidenceViewSchemaId,
      anchor: {
        locator: page.getByTestId(
          workbookInspectorPanelTestId(evidenceViewSchemaId, "evidence"),
        ),
        align: "start",
        scrollportSelector: `aside[data-view-schema-id="${evidenceViewSchemaId}"] [data-inspector-scroll-body]`,
      },
      verifyFraming: async () => {
        await expect
          .poll(() => readWorkbookGridScroll(page, evidenceViewSchemaId))
          .toEqual(expectedScroll);
      },
      prepareState: async () => {
        await setWorkbookGridScroll(page, evidenceViewSchemaId, {
          top: 0,
          left: "right",
        });
        const actionButton = await mountedGridTarget(
          page,
          evidenceViewSchemaId,
          evidencePreviewButtonTestId(actionRecordId),
        );
        await waitForVisualLayoutFrame(page);
        await expect(actionButton).toBeVisible();
        expectedScroll = await readWorkbookGridScroll(
          page,
          evidenceViewSchemaId,
        );
      },
    });
  } catch (error) {
    try {
      await attachWorkbookGridVisualDiagnostics(
        page,
        name,
        evidenceViewSchemaId,
        { scroll: { top: 0, left: "left" } },
      );
    } catch {
      // Preserve the assertion failure when the page is already torn down.
    }
    throw error;
  }
}

async function expectVisualEvidenceState(
  page: Page,
  recordId: string,
  stateKey: VisualEvidenceRowStateKey,
) {
  const previewButton = await mountedGridTarget(
    page,
    evidenceViewSchemaId,
    evidencePreviewButtonTestId(recordId),
  );
  const stateContainer = page.getByTestId(evidenceAccessStateTestId(recordId));
  await expect(stateContainer).toHaveAttribute(
    "data-evidence-state-key",
    stateKey,
  );
  if (stateKey === "available") {
    await expect(previewButton).toBeEnabled();
    await expect(
      page.getByTestId(evidenceDownloadButtonTestId(recordId)),
    ).toBeEnabled();
    return;
  }
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(recordId)),
  ).toBeVisible();
  await expect(
    page.getByTestId(evidencePreviewButtonTestId(recordId)),
  ).toBeDisabled();
  await expect(
    page.getByTestId(evidenceDownloadButtonTestId(recordId)),
  ).toBeDisabled();
}

async function createUploadedVisualEvidence(
  page: Page,
  incidentId: string,
  options: EvidenceUploadOptions,
): Promise<ViewRow> {
  return createUploadedEvidenceFixture(page, incidentId, {
    collectorPartyText: "browser.evidence-workflow visual fixture",
    ...options,
    receivedAt: "2026-05-01T10:45:00Z",
    txnSuffixes: {
      attach: "ATTACH",
      blob: "BLOB",
      row: "ROW",
    },
  });
}

async function createVisualEvidenceRow(
  page: Page,
  incidentId: string,
  options: {
    lifecycleState: NonNullable<
      EvidenceCreateRequest["evidence.lifecycle_state"]
    >;
    requestedAt: string;
    storageRef: string;
    title: string;
    txnPrefix: string;
  },
): Promise<ViewRow> {
  return createEvidenceFixtureRow(page, incidentId, {
    collectorPartyText: "browser.evidence-workflow visual fixture",
    ...options,
  });
}

function tinyPNG() {
  return Buffer.from(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=",
    "base64",
  );
}

test("captures Evidence access affordances on the required Evidence surface", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("VISUALEVIDENCEACCESS"),
    "Entity linking visual evidence access",
  );
  const evidenceRow = await createViewRow(
    page,
    incidentId,
    evidenceViewSchemaId,
    {
      client_txn_id: uniqueTxn("VISUALEVIDENCEACCESS-EVIDENCE"),
      "evidence.title": "Visual evidence package",
      "evidence.requested_at": "2025-03-01T10:00:00Z",
      "evidence.storage_ref": "slot/visual",
    },
  );

  await navigateVisualApplication(
    page,
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      evidenceViewSchemaId,
    )}`,
  );

  await expect(
    await mountedGridCell(
      page,
      evidenceViewSchemaId,
      evidenceRow.record_id,
      "evidence.title",
    ),
  ).toHaveText("Visual evidence package");
  await expect(
    await mountedGridTarget(
      page,
      evidenceViewSchemaId,
      evidencePreviewButtonTestId(evidenceRow.record_id),
    ),
  ).toBeVisible();
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(evidenceRow.record_id)),
  ).toContainText("Requested");

  await assertWorkbookGridVisualRegression(
    page,
    "record-relationships-evidence-access",
    evidenceViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );
});

test("captures requested and available Evidence states on the required Evidence surface", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("VISUALEVIDENCEAVAILABLE"),
    "Evidence lifecycle visual evidence states",
  );
  const evidenceRow = await createViewRow(
    page,
    incidentId,
    evidenceViewSchemaId,
    {
      client_txn_id: uniqueTxn("VISUALEVIDENCEAVAILABLE-EVIDENCE"),
      "evidence.title": "Requested visual package",
      "evidence.requested_at": "2025-03-01T10:00:00Z",
      "evidence.storage_ref": "ticket://visual-request",
    },
  );

  await navigateVisualApplication(
    page,
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      evidenceViewSchemaId,
    )}`,
  );

  await expect(
    await mountedGridCell(
      page,
      evidenceViewSchemaId,
      evidenceRow.record_id,
      "evidence.lifecycle_state",
    ),
  ).toHaveText("requested");
  await assertWorkbookGridVisualRegression(
    page,
    "evidence-grid-requested-evidence",
    evidenceViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );

  await mountedGridTarget(
    page,
    evidenceViewSchemaId,
    evidencePreviewButtonTestId(evidenceRow.record_id),
  );
  await chooseEvidenceFile(
    page.getByTestId(evidenceAttachFileInputTestId(evidenceRow.record_id)),
    {
      name: "visual-request.txt",
      mimeType: "text/plain",
      buffer: Buffer.from("evidence_lifecycle visual evidence", "utf8"),
    },
  );
  await expect(
    page.getByRole("group", {
      name: "File recovery: visual-request.txt",
      exact: true,
    }),
  ).toContainText("File attached. Custody unchanged.");
  await patchRecord(page, evidenceRow.record_id, {
    view_schema_id: evidenceViewSchemaId,
    base_row_version: evidenceRow.row_version + 1,
    client_txn_id: uniqueTxn("visual-explicit-custody"),
    changes: [
      { field_key: "evidence.lifecycle_state", value: "available" },
      { field_key: "evidence.received_at", value: "2025-03-01T10:30:00Z" },
    ],
  });
  await page
    .getByRole("button", { name: "Discard retained file work", exact: true })
    .last()
    .click();
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(evidenceRow.record_id)),
  ).toHaveText("Available", { timeout: 30_000 });
  await expect(
    await mountedGridCell(
      page,
      evidenceViewSchemaId,
      evidenceRow.record_id,
      "evidence.lifecycle_state",
    ),
  ).toHaveText("available");
  await expect(
    await mountedGridCell(
      page,
      evidenceViewSchemaId,
      evidenceRow.record_id,
      "evidence.upload_state",
    ),
  ).toHaveText("available");
  await assertWorkbookGridVisualRegression(
    page,
    "evidence-grid-available-evidence",
    evidenceViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );
});

test("captures blocked preview feedback and Timeline evidence badges", async ({
  page,
}, testInfo) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("VISUALEVIDENCEBLOCKED"),
    "Evidence lifecycle visual evidence badges",
  );
  const blocked = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("VISUALEVIDENCEBLOCKED-BLOCKED"),
    "evidence.title": "Blocked visual package",
    "evidence.requested_at": "2025-03-01T10:00:00Z",
    "evidence.storage_ref": "ticket://visual-blocked",
  });
  const timelineIncidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUAL-TIMELINE-EVIDENCE"),
    "Service-account evidence review",
  );
  const timelineRow = await createVisualTimelineRow(
    page,
    timelineIncidentId,
    "visualevidenceblocked-timeline",
    {
      client_txn_id: uniqueTxn("VISUALEVIDENCEBLOCKED-TIMELINE"),
      "timeline.activity_synopsis_text": "Visual evidence badge row",
    },
  );

  await navigateVisualApplication(
    page,
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      evidenceViewSchemaId,
    )}`,
  );

  await expect(
    await mountedGridTarget(
      page,
      evidenceViewSchemaId,
      evidenceAccessMessageTestId(blocked.record_id),
    ),
  ).toContainText("Requested");
  await assertWorkbookGridVisualRegression(
    page,
    "evidence-grid-blocked-preview",
    evidenceViewSchemaId,
    { scroll: { top: 0, left: "left" } },
  );

  await navigateVisualApplication(
    page,
    `/?incident_id=${timelineIncidentId}&view_schema_id=${encodeURIComponent(
      timelineViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await openTimelineInspector(page, timelineRow.record_id);
  await chooseEvidenceFile(
    page.getByTestId(timelineEvidenceFileInputTestId(timelineRow.record_id)),
    {
      name: "visual-badge.png",
      mimeType: "image/png",
      buffer: tinyPNG(),
    },
  );
  await expect(
    page.getByTestId(timelineInspectorSectionTestId("evidence")),
  ).toContainText("Attached evidence count: 1");
  await expect(page.getByRole("button", { name: /^Attachments:/ })).toHaveText(
    "Attachments: 0 need attention, 0 in progress, 1 completed",
  );
  await page.evaluate(() => {
    if (document.activeElement instanceof HTMLElement) {
      document.activeElement.blur();
    }
  });
  await assertWorkbookGridVisualRegression(
    page,
    "evidence-grid-timeline-evidence-badge",
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

test("Capture evidence count, affordance, available, requested, pending, blocked, failed, inconsistent, preview, and download-handle state fixtures.", async ({
  page,
}, testInfo) => {
  await page.setViewportSize({ width: 1280, height: 720 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("VISUALEVIDENCEWORKFLOW"),
    "browser.evidence-workflow visual evidence affordance",
  );
  const requested = await createVisualEvidenceRow(page, incidentId, {
    lifecycleState: "requested",
    requestedAt: "2026-05-01T10:00:00Z",
    storageRef: "case://evidence-workflow/requested",
    title: "01 requested evidence",
    txnPrefix: "VISUALEVIDENCEWORKFLOW-REQUESTED",
  });
  const pending = await createVisualEvidenceRow(page, incidentId, {
    lifecycleState: "pending_receipt",
    requestedAt: "2026-05-01T10:05:00Z",
    storageRef: "case://evidence-workflow/pending",
    title: "02 pending evidence",
    txnPrefix: "VISUALEVIDENCEWORKFLOW-PENDING",
  });
  const blocked = await createVisualEvidenceRow(page, incidentId, {
    lifecycleState: "quarantined",
    requestedAt: "2026-05-01T10:10:00Z",
    storageRef: "case://evidence-workflow/quarantined",
    title: "03 quarantined evidence",
    txnPrefix: "VISUALEVIDENCEWORKFLOW-BLOCKED",
  });
  const availablePreview = await createUploadedVisualEvidence(
    page,
    incidentId,
    {
      body: Buffer.from(
        "visual.evidence-workflow preview visual evidence\n",
        "utf8",
      ),
      contentType: "text/plain",
      filename: "evidence-preview.txt",
      requestedAt: "2026-05-01T10:15:00Z",
      title: "04 available preview evidence",
      txnPrefix: "VISUALEVIDENCEWORKFLOW-PREVIEW",
    },
  );
  const downloadHandle = await createUploadedVisualEvidence(page, incidentId, {
    body: Buffer.from(
      "visual.evidence-workflow download handle visual evidence\n",
      "utf8",
    ),
    contentType: "text/plain",
    filename: "evidence-download-handle.txt",
    requestedAt: "2026-05-01T10:20:00Z",
    title: "05 download handle evidence",
    txnPrefix: "VISUALEVIDENCEWORKFLOW-DOWNLOAD",
  });
  const previewBlocked = await createUploadedVisualEvidence(page, incidentId, {
    body: Buffer.from(
      "<!doctype html><title>visual.evidence-workflow unsupported preview</title>",
      "utf8",
    ),
    contentType: "text/html",
    filename: "evidence-preview-blocked.html",
    requestedAt: "2026-05-01T10:25:00Z",
    title: "06 preview blocked evidence",
    txnPrefix: "VISUALEVIDENCEWORKFLOW-PREVIEW-BLOCKED",
  });
  const failedHandle = await createUploadedVisualEvidence(page, incidentId, {
    body: Buffer.from(
      "visual.evidence-workflow failed handle visual evidence\n",
      "utf8",
    ),
    contentType: "text/plain",
    filename: "evidence-failed-handle.txt",
    requestedAt: "2026-05-01T10:30:00Z",
    title: "07 failed handle evidence",
    txnPrefix: "VISUALEVIDENCEWORKFLOW-FAILED",
  });
  const inconsistentHandle = await createUploadedVisualEvidence(
    page,
    incidentId,
    {
      body: Buffer.from(
        "visual.evidence-workflow inconsistent handle visual evidence\n",
        "utf8",
      ),
      contentType: "text/plain",
      filename: "evidence-inconsistent-handle.txt",
      requestedAt: "2026-05-01T10:35:00Z",
      title: "08 inconsistent handle evidence",
      txnPrefix: "VISUALEVIDENCEWORKFLOW-INCONSISTENT",
    },
  );
  const timelineIncidentId = await createRichTimelineIncident(
    page,
    uniqueIncidentKey("VISUAL-TIMELINE-EVIDENCE"),
    "Service-account evidence review",
  );
  const timelineRow = await createVisualTimelineRow(
    page,
    timelineIncidentId,
    "visualevidenceworkflow-timeline",
    {
      client_txn_id: uniqueTxn("VISUALEVIDENCEWORKFLOW-TIMELINE"),
      "timeline.activity_utc_text": "2026-05-01T11:00:00Z",
      "timeline.activity_synopsis_text":
        "browser.evidence-workflow timeline evidence count",
    },
  );

  await navigateVisualApplication(
    page,
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      evidenceViewSchemaId,
    )}`,
  );

  await expect(
    page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
  ).toBeVisible();
  await expectVisualEvidenceState(page, requested.record_id, "requested");
  await expectVisualEvidenceState(page, pending.record_id, "pending_receipt");
  await expectVisualEvidenceState(page, blocked.record_id, "quarantined");
  await expectVisualEvidenceState(
    page,
    availablePreview.record_id,
    "available",
  );
  await expectVisualEvidenceState(page, downloadHandle.record_id, "available");
  await expectVisualEvidenceState(page, previewBlocked.record_id, "available");
  await expectVisualEvidenceState(page, failedHandle.record_id, "available");
  await expectVisualEvidenceState(
    page,
    inconsistentHandle.record_id,
    "available",
  );

  await (
    await mountedGridTarget(
      page,
      evidenceViewSchemaId,
      evidencePreviewButtonTestId(availablePreview.record_id),
    )
  ).click();
  await expect(
    page.getByTestId(evidencePreviewFrameTestId(availablePreview.record_id)),
  ).toBeVisible();
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(availablePreview.record_id)),
  ).toHaveText("Preview open");
  await page
    .getByTestId(evidencePreviewPanelTestId())
    .getByRole("button", { name: "Close" })
    .click();

  const downloadPromise = page.waitForEvent("download");
  await (
    await mountedGridTarget(
      page,
      evidenceViewSchemaId,
      evidenceDownloadButtonTestId(downloadHandle.record_id),
    )
  ).click();
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toBe("evidence-download-handle.txt");
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(downloadHandle.record_id)),
  ).toHaveText("Download requested");

  await (
    await mountedGridTarget(
      page,
      evidenceViewSchemaId,
      evidencePreviewButtonTestId(previewBlocked.record_id),
    )
  ).click();
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(previewBlocked.record_id)),
  ).toContainText("No preview");

  await installPublicErrorResponse(page, {
    method: "POST",
    status: 409,
    code: "evidence_access_unavailable",
    path: `/api/v1/evidence-records/${failedHandle.record_id}/preview-handle`,
    details: { reason_code: "blob_failed" },
  });
  await (
    await mountedGridTarget(
      page,
      evidenceViewSchemaId,
      evidencePreviewButtonTestId(failedHandle.record_id),
    )
  ).click();
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(failedHandle.record_id)),
  ).toContainText("Upload failed");

  await installPublicErrorResponse(page, {
    method: "POST",
    status: 409,
    code: "evidence_access_unavailable",
    path: `/api/v1/evidence-records/${inconsistentHandle.record_id}/preview-handle`,
    details: { reason_code: "evidence_inconsistent" },
  });
  await (
    await mountedGridTarget(
      page,
      evidenceViewSchemaId,
      evidencePreviewButtonTestId(inconsistentHandle.record_id),
    )
  ).click();
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(inconsistentHandle.record_id)),
  ).toContainText("Inconsistent");

  await page
    .getByTestId(evidenceAccessMessageTestId(failedHandle.record_id))
    .click();
  await page
    .getByTestId(workbookInspectorPanelTestId(evidenceViewSchemaId, "evidence"))
    .evaluate((element) =>
      element.scrollIntoView({ block: "start", inline: "nearest" }),
    );
  await expect(
    page.getByTestId(
      evidenceAccessMessageTestId(failedHandle.record_id, "inspector"),
    ),
  ).toContainText("upload failed");
  await assertEvidenceAccessVisualRegression(
    page,
    "evidence-affordance-states",
    availablePreview.record_id,
  );

  await navigateVisualApplication(
    page,
    `/?incident_id=${timelineIncidentId}&view_schema_id=${encodeURIComponent(
      timelineViewSchemaId,
    )}`,
  );

  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  await openTimelineInspector(page, timelineRow.record_id);
  await expect(
    page.getByTestId(timelineInspectorSectionTestId("evidence")),
  ).toHaveAttribute("aria-label", "Timeline evidence");
  await chooseEvidenceFile(
    page.getByTestId(timelineEvidenceFileInputTestId(timelineRow.record_id)),
    {
      buffer: tinyPNG(),
      mimeType: "image/png",
      name: "timeline-evidence.png",
    },
  );
  await expect(
    page.getByTestId(timelineInspectorSectionTestId("evidence")),
  ).toContainText("Attached evidence count: 1");
  await expect(page.getByRole("button", { name: /^Attachments:/ })).toHaveText(
    "Attachments: 0 need attention, 0 in progress, 1 completed",
  );
  await page.evaluate(() => {
    if (document.activeElement instanceof HTMLElement) {
      document.activeElement.blur();
    }
  });
  await assertWorkbookGridVisualRegression(
    page,
    "evidence-timeline-evidence-count",
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

test("Capture Timeline Evidence metadata Party selection and retained partial success", async ({
  page,
}) => {
  const capture = async (
    page: Page,
    name: string,
    options: { anchor?: VisualAnchor } = {},
  ) => {
    await assertViewportVisualRegression(page, name, options);
    await test.info().attach(`${name}-review`, {
      body: await page.screenshot({ animations: "disabled", caret: "hide" }),
      contentType: "image/png",
    });
  };
  await page.setViewportSize({ width: 1280, height: 720 });
  const supplied = await createRichTimelineSource(page, "evidence-source", {
    client_txn_id: uniqueTxn("evidence-source"),
    "timeline.activity_synopsis_text": "Preserved investigation source",
    "timeline.raw_activity_text": "Original source text remains unchanged.",
  });
  const { form } = await openTimelineEvidenceFixture(
    page,
    (url) => navigateVisualApplication(page, url),
    supplied,
  );
  // This capture frames the beginning of the form; give that frame an explicit
  // visible focus target instead of retaining focus in a later seeded field.
  await form.getByTestId(genericCreateFieldTestId("evidence.title")).focus();
  const anchor: VisualAnchor = {
    locator: form,
    align: "start",
    scrollportSelector: `aside[data-view-schema-id="${timelineViewSchemaId}"] [data-inspector-scroll-body]`,
  };
  await capture(page, "timeline-related-evidence-authoring", { anchor });
  await page.setViewportSize({ width: 768, height: 640 });
  await capture(page, "timeline-related-evidence-authoring-narrow", { anchor });
  await form
    .getByRole("button", { name: "Choose Collector Party", exact: true })
    .click();
  const picker = form.getByRole("region", {
    name: "Choose Collector Party",
    exact: true,
  });
  await expect(
    picker.getByRole("button", { name: "Apply Party", exact: true }),
  ).toBeEnabled();
  await capture(page, "timeline-related-evidence-party-narrow", {
    anchor: { ...anchor, locator: picker },
  });
  await picker
    .getByRole("button", { name: "Cancel Party selection", exact: true })
    .click();
  const { recovery } = await retainTimelineEvidencePartialResult(page);
  await capture(page, "timeline-related-evidence-partial-narrow");
  await page.setViewportSize({ width: 1280, height: 720 });
  await capture(page, "timeline-related-evidence-partial");
  await test.info().attach("timeline-related-evidence-recovery-tree", {
    body: await recovery.ariaSnapshot(),
    contentType: "text/plain",
  });
});
