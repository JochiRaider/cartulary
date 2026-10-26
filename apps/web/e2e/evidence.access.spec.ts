import { Buffer } from "node:buffer";

import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  dataTestIdSelector,
  evidenceAccessMessageTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
  gridShellTestId,
  rowCellTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  timelineScalarEditorTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Locator } from "@playwright/test";
import { expect, test } from "./fixtures";
import {
  expectRevealedInInspectorBody,
  linkEvidenceToTimeline,
  linkedTextReviewFixture,
  openEvidenceSurface,
  openTimelineSurface,
} from "./support/evidence/browser";
import { createUploadedEvidenceFixture } from "./support/evidence/fixtures";
import { createIncident } from "./support/incidents/fixtures";
import { browserScopeFor, responseBarrier } from "./support/runtime/cleanup";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  observePageRequests,
  rejectUnexpectedPageErrors,
} from "./support/runtime/observations";
import {
  awaitResponseBarrier,
  installOwnedRoute,
  removeOwnedRoute,
} from "./support/transport/requestInterception";
import { createViewRow } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

async function focusTimelineGridCell(cell: Locator) {
  await cell.evaluate((element) => {
    const gridCell = element.closest<HTMLElement>('[role="gridcell"]');
    if (!gridCell) throw new Error("Missing committed grid cell");
    gridCell.focus();
  });
  await expect(
    cell.locator('xpath=ancestor::*[@role="gridcell"]'),
  ).toBeFocused();
}

function evidenceTransientResponse() {
  return JSON.stringify({
    error: {
      code: "object_store_unavailable",
      status: 503,
      retryable: true,
      request_id: "request-object-store",
      message: "private",
      details: { reason_code: "endpoint_unreachable" },
    },
  });
}

test.beforeEach(({ page }) => {
  rejectUnexpectedPageErrors(page);
});

test("redeems inline-safe previews and shows explicit blocked-preview outcomes", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-PREVIEW"),
    "Evidence evidence preview",
  );
  const safe = await createUploadedEvidenceFixture(page, incidentId, {
    txnPrefix: "e5-preview",
    collectorPartyText: "Browser evidence",
    title: "Safe text preview",
    filename: "safe.txt",
    contentType: "text/plain",
    body: Buffer.from("safe preview body", "utf8"),
  });
  const unsafe = await createUploadedEvidenceFixture(page, incidentId, {
    txnPrefix: "e5-preview",
    collectorPartyText: "Browser evidence",
    title: "Unsafe HTML preview",
    filename: "unsafe.html",
    contentType: "text/html",
    body: Buffer.from(
      "<script>window.__unsafe_preview = true</script>",
      "utf8",
    ),
  });

  await openEvidenceSurface(page, incidentId);
  const safePreviewButtonTestId = evidencePreviewButtonTestId(safe.record_id);
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: safePreviewButtonTestId,
  });
  await page.getByTestId(safePreviewButtonTestId).click();
  await expect(
    page.getByTestId(evidencePreviewFrameTestId(safe.record_id)),
  ).toBeVisible();
  await expect(
    page
      .frameLocator(
        dataTestIdSelector(evidencePreviewFrameTestId(safe.record_id)),
      )
      .locator("body"),
  ).toContainText("safe preview body");

  const unsafePreviewButtonTestId = evidencePreviewButtonTestId(
    unsafe.record_id,
  );
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: unsafePreviewButtonTestId,
  });
  await page.getByTestId(unsafePreviewButtonTestId).click();
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(unsafe.record_id)),
  ).toHaveText("No preview");
});

test("reviews linked text Evidence from Timeline with Space and fresh explicit actions", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const { incidentId, source, title, evidence } =
    await linkedTextReviewFixture(page);
  const requests: Array<{ path: string; body: string | null }> = [];
  observePageRequests(page, (request) => {
    if (
      request.method() === "POST" &&
      /\/evidence-records\/[^/]+\/(preview|download)-handle$/u.test(
        new URL(request.url()).pathname,
      )
    )
      requests.push({
        path: new URL(request.url()).pathname,
        body: request.postData(),
      });
  });
  await openTimelineSurface(page, incidentId);
  const cell = page.getByTestId(
    rowCellTestId(source.record_id, "timeline.activity_synopsis_text"),
  );
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: rowCellTestId(
      source.record_id,
      "timeline.activity_synopsis_text",
    ),
  });
  await focusTimelineGridCell(cell);
  const gridShell = page.getByTestId(gridShellTestId(timelineViewSchemaId));
  const gridScrollBefore = await gridShell.evaluate((element) => ({
    left: element.scrollLeft,
    top: element.scrollTop,
  }));
  const documentScrollBefore = await page.evaluate(() => window.scrollY);
  await page.keyboard.press("Space");
  await expect(
    page.getByRole("list", { name: "Linked Evidence" }),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(timelineInspectorSectionTestId("evidence"))
      .getByText(title, { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByTestId(evidencePreviewFrameTestId(evidence.record_id)),
  ).toBeVisible();
  await expect(
    page
      .frameLocator(
        dataTestIdSelector(evidencePreviewFrameTestId(evidence.record_id)),
      )
      .locator("body"),
  ).toContainText("linked timeline preview body");
  expect(requests.map((entry) => entry.path)).toEqual([
    `/api/v1/evidence-records/${evidence.record_id}/preview-handle`,
  ]);
  expect(requests[0]?.body).toBe("{}");
  await page.keyboard.press("Escape");
  await expect(page.getByTestId(evidencePreviewPanelTestId())).toHaveCount(0);
  await expect(
    page.getByRole("list", { name: "Linked Evidence" }),
  ).toBeFocused();
  const timelineInspector = page.getByTestId(timelineInspectorTestId());
  await expect(
    timelineInspector.locator('[data-inspector-navigation-panel="evidence"]'),
  ).toHaveAttribute("aria-current", "location");
  await expect(
    page.getByRole("list", { name: "Linked Evidence" }),
  ).toBeInViewport();
  await expectRevealedInInspectorBody(
    page.getByRole("list", { name: "Linked Evidence" }),
  );
  await expect(
    timelineInspector.getByRole("button", { name: "Close inspector" }),
  ).toBeInViewport();
  expect(await page.evaluate(() => window.scrollY)).toBe(documentScrollBefore);
  expect(
    await gridShell.evaluate((element) => ({
      left: element.scrollLeft,
      top: element.scrollTop,
    })),
  ).toEqual(gridScrollBefore);
  const previewButton = page.getByTestId(
    evidencePreviewButtonTestId(evidence.record_id, "inspector"),
  );
  await page.getByRole("button", { name: "Edit Activity Synopsis" }).click();
  const rawEditor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId: source.record_id,
      surface: "inspector",
    }),
  );
  await rawEditor.fill("  unfinished review note Ω  ");
  await previewButton.click();
  await expect(
    page.getByTestId(evidencePreviewFrameTestId(evidence.record_id)),
  ).toBeVisible();
  await page.getByRole("button", { name: "Close", exact: true }).click();
  await expect(previewButton).toBeFocused();
  await expect(
    timelineInspector.locator('[data-inspector-navigation-panel="evidence"]'),
  ).toHaveAttribute("aria-current", "location");
  await expectRevealedInInspectorBody(previewButton);
  await expect(rawEditor).toHaveValue("  unfinished review note Ω  ");
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  await openTimelineInspector(page, source.record_id);
  if (!(await rawEditor.count()))
    await page
      .getByRole("button", { name: "Resume draft for Activity Synopsis" })
      .click();
  await expect(rawEditor).toHaveValue("  unfinished review note Ω  ");
  const downloadButton = page.getByTestId(
    evidenceDownloadButtonTestId(evidence.record_id, "inspector"),
  );
  await downloadButton.click();
  await expect
    .poll(
      () =>
        requests.filter((entry) => entry.path.endsWith("/download-handle"))
          .length,
    )
    .toBe(1);
  await downloadButton.click();
  await expect
    .poll(
      () =>
        requests.filter((entry) => entry.path.endsWith("/download-handle"))
          .length,
    )
    .toBe(2);
  expect(
    requests
      .filter((entry) => entry.path.endsWith("/download-handle"))
      .every((entry) => entry.body === "{}"),
  ).toBe(true);
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
});

test("keeps a blocked linked preview local and checks a mixed Space selection", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1024, height: 720 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("TIMELINE-MIXED-REVIEW"),
    "Timeline mixed Evidence review",
  );
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("timeline-mixed-source"),
    "timeline.activity_synopsis_text": "Mixed linked Evidence source",
  });
  const safe = await createUploadedEvidenceFixture(page, incidentId, {
    txnPrefix: "e5-preview",
    collectorPartyText: "Browser evidence",
    title: "Plain text notes",
    filename: "plain.txt",
    contentType: "text/plain",
    body: Buffer.from("mixed safe body", "utf8"),
  });
  const blocked = await createUploadedEvidenceFixture(page, incidentId, {
    txnPrefix: "e5-preview",
    collectorPartyText: "Browser evidence",
    title: "HTML evidence that cannot be previewed",
    filename: "blocked.html",
    contentType: "text/html",
    body: Buffer.from("<p>blocked</p>", "utf8"),
  });
  await linkEvidenceToTimeline(page, source.record_id, source.row_version, [
    safe.record_id,
    blocked.record_id,
  ]);
  const requests: string[] = [];
  observePageRequests(page, (request) => {
    if (
      request.method() === "POST" &&
      /\/evidence-records\/[^/]+\/(preview|download)-handle$/u.test(
        new URL(request.url()).pathname,
      )
    )
      requests.push(new URL(request.url()).pathname);
  });
  await openTimelineSurface(page, incidentId);
  await openTimelineInspector(page, source.record_id);
  await expect(
    page
      .getByTestId(timelineInspectorSectionTestId("evidence"))
      .getByText("Plain text notes", { exact: true }),
  ).toBeVisible();
  await expect(
    page
      .getByTestId(timelineInspectorSectionTestId("evidence"))
      .getByText("HTML evidence that cannot be previewed", { exact: true }),
  ).toBeVisible();
  await expect(
    page.getByTestId(
      evidencePreviewButtonTestId(blocked.record_id, "inspector"),
    ),
  ).toHaveAttribute("aria-disabled", "true");
  expect(requests).toHaveLength(0);
  await expect(
    page.getByText("This file type cannot be previewed."),
  ).toBeVisible();
  await expect(page.getByTestId(evidencePreviewPanelTestId())).toHaveCount(0);
  expect(
    requests.filter((entry) => entry.endsWith("/download-handle")),
  ).toHaveLength(0);
  await page
    .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
    .click();
  const cell = page.getByTestId(
    rowCellTestId(source.record_id, "timeline.activity_synopsis_text"),
  );
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: rowCellTestId(
      source.record_id,
      "timeline.activity_synopsis_text",
    ),
  });
  await focusTimelineGridCell(cell);
  await page.keyboard.press("Space");
  await expect(
    page.getByTestId(evidencePreviewFrameTestId(safe.record_id)),
  ).toBeVisible();
  await expect(
    page
      .frameLocator(
        dataTestIdSelector(evidencePreviewFrameTestId(safe.record_id)),
      )
      .locator("body"),
  ).toContainText("mixed safe body");
  expect(
    requests.filter(
      (entry) =>
        entry === `/api/v1/evidence-records/${safe.record_id}/preview-handle`,
    ).length,
  ).toBe(2);
  expect(
    requests.filter(
      (entry) =>
        entry ===
        `/api/v1/evidence-records/${blocked.record_id}/preview-handle`,
    ).length,
  ).toBe(1);
  expect(
    requests.filter((entry) => entry.endsWith("/download-handle")),
  ).toHaveLength(0);
});

test("recovers linked Preview after transient storage failure without submitting an inspector draft", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-LINK-RETRY"),
    "Linked Evidence retry",
  );
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("linked-retry-source"),
    "timeline.activity_synopsis_text": "Source with unfinished review",
  });
  const evidence = await createUploadedEvidenceFixture(page, incidentId, {
    txnPrefix: "e5-preview",
    collectorPartyText: "Browser evidence",
    title: "Recoverable linked notes",
    filename: "retry.txt",
    contentType: "text/plain",
    body: Buffer.from("fresh linked preview", "utf8"),
  });
  await linkEvidenceToTimeline(page, source.record_id, source.row_version, [
    evidence.record_id,
  ]);
  const path = `**/evidence-records/${evidence.record_id}/preview-handle`;
  const bodies: string[] = [];
  let attempts = 0;
  let writes = 0;
  observePageRequests(page, (request) => {
    if (
      request.method() === "PATCH" &&
      new URL(request.url()).pathname.endsWith(`/records/${source.record_id}`)
    )
      writes++;
  });
  await installOwnedRoute(page, path, async (route) => {
    bodies.push(route.request().postData() ?? "");
    if (attempts++ === 0)
      await route.fulfill({
        status: 503,
        contentType: "application/json",
        body: evidenceTransientResponse(),
      });
    else await route.continue();
  });
  await openTimelineSurface(page, incidentId);
  await openTimelineInspector(page, source.record_id);
  await page.getByRole("button", { name: "Edit Activity Synopsis" }).click();
  const editor = page.getByTestId(
    timelineScalarEditorTestId({
      fieldKey: "timeline.activity_synopsis_text",
      recordId: source.record_id,
      surface: "inspector",
    }),
  );
  await editor.fill("  unfinished review note Ω  ");
  await editor.evaluate((element: HTMLInputElement | HTMLTextAreaElement) =>
    element.setSelectionRange(3, 10),
  );
  const caret = await editor.evaluate(
    (element: HTMLInputElement | HTMLTextAreaElement) => ({
      start: element.selectionStart,
      end: element.selectionEnd,
    }),
  );
  const grid = page.getByTestId(gridShellTestId(timelineViewSchemaId));
  const scroll = await grid.evaluate((element) => ({
    left: element.scrollLeft,
    top: element.scrollTop,
  }));
  const preview = page.getByTestId(
    evidencePreviewButtonTestId(evidence.record_id, "inspector"),
  );
  await preview.click();
  await expect(preview).toHaveText("Retry Preview");
  await expect(
    page.getByText(/Object storage is temporarily unavailable/u),
  ).toBeVisible();
  await preview.click();
  await expect(
    page.getByTestId(evidencePreviewFrameTestId(evidence.record_id)),
  ).toBeVisible();
  await page.getByRole("button", { name: "Close", exact: true }).click();
  await expect(preview).toBeFocused();
  await expect(editor).toHaveValue("  unfinished review note Ω  ");
  expect(
    await editor.evaluate(
      (element: HTMLInputElement | HTMLTextAreaElement) => ({
        start: element.selectionStart,
        end: element.selectionEnd,
      }),
    ),
  ).toEqual(caret);
  expect(
    await grid.evaluate((element) => ({
      left: element.scrollLeft,
      top: element.scrollTop,
    })),
  ).toEqual(scroll);
  await expect(
    page
      .getByTestId(timelineInspectorSectionTestId("evidence"))
      .getByText("Recoverable linked notes", { exact: true }),
  ).toBeVisible();
  expect(bodies).toEqual(["{}", "{}"]);
  expect(writes).toBe(0);
});

test("bounds mixed-link Space discovery and leaves explicit linked actions usable", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-SPACE-BOUND"),
    "Bounded Space discovery",
  );
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("space-bound-source"),
    "timeline.activity_synopsis_text": "Mixed preview source",
  });
  const safe = await createUploadedEvidenceFixture(page, incidentId, {
    txnPrefix: "e5-preview",
    collectorPartyText: "Browser evidence",
    title: "Available text",
    filename: "available.txt",
    contentType: "text/plain",
    body: Buffer.from("available preview", "utf8"),
  });
  const other = await createUploadedEvidenceFixture(page, incidentId, {
    txnPrefix: "e5-preview",
    collectorPartyText: "Browser evidence",
    title: "Held text",
    filename: "held.txt",
    contentType: "text/plain",
    body: Buffer.from("held preview", "utf8"),
  });
  await linkEvidenceToTimeline(page, source.record_id, source.row_version, [
    safe.record_id,
    other.record_id,
  ]);
  await page.clock.install();
  let release = () => {};
  const gateBarrier = responseBarrier(page);
  release = gateBarrier.release;
  const gate = gateBarrier.promise;
  let observeHeld = () => {};
  const heldObservedBarrier = responseBarrier(page);
  observeHeld = heldObservedBarrier.release;
  const heldObserved = heldObservedBarrier.promise;
  let settleHeld = () => {};
  const heldSettledBarrier = responseBarrier(page);
  settleHeld = heldSettledBarrier.release;
  const heldSettled = heldSettledBarrier.promise;
  const heldPath = `**/evidence-records/${other.record_id}/preview-handle`;
  await installOwnedRoute(page, heldPath, async (route) => {
    observeHeld();
    if (!(await awaitResponseBarrier(route, gate))) return;
    try {
      await route.continue();
    } catch {
      /* The local deadline may abort this request. */
    } finally {
      settleHeld();
    }
  });
  try {
    await openTimelineSurface(page, incidentId);
    await scrollGridTargetIntoView({
      page,
      surface: timelineViewSchemaId,
      targetTestId: rowCellTestId(
        source.record_id,
        "timeline.activity_synopsis_text",
      ),
    });
    const cell = page.getByTestId(
      rowCellTestId(source.record_id, "timeline.activity_synopsis_text"),
    );
    await focusTimelineGridCell(cell);
    await page.keyboard.press("Space");
    await expect(page.getByText("Checking linked previews…")).toBeVisible();
    await heldObserved;
    // Metadata admission has its own deadline. Establish it before advancing
    // the clock so this scenario expires only the held preview discovery.
    await expect(
      page
        .getByRole("list", { name: "Linked Evidence" })
        .locator('[data-evidence-metadata-state="ready"]'),
    ).toHaveCount(2);
    await page.clock.fastForward(30_000);
    await expect(
      page.getByText(/Preview availability could not be fully checked/u),
    ).toBeVisible();
    await expect(
      page.getByRole("list", { name: "Linked Evidence" }),
    ).toBeVisible();
    await expect(page.getByTestId(evidencePreviewPanelTestId())).toHaveCount(0);
    const direct = page.getByTestId(
      evidencePreviewButtonTestId(safe.record_id, "inspector"),
    );
    await direct.click();
    await expect(
      page.getByTestId(evidencePreviewFrameTestId(safe.record_id)),
    ).toBeVisible();
    await page.getByRole("button", { name: "Close", exact: true }).click();
    release();
    await heldSettled;
    await page.clock.fastForward(1);
    await expect(page.getByTestId(evidencePreviewPanelTestId())).toHaveCount(0);

    release();

    await removeOwnedRoute(page, heldPath);
  } finally {
    gateBarrier.cancel();
  }
});

test("recovers Evidence sheet Preview and Download with fresh handle requests", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-SHEET-RETRY"),
    "Evidence sheet retry",
  );
  const evidence = await createUploadedEvidenceFixture(page, incidentId, {
    txnPrefix: "e5-preview",
    collectorPartyText: "Browser evidence",
    title: "Recoverable sheet file",
    filename: "sheet.txt",
    contentType: "text/plain",
    body: Buffer.from("sheet preview", "utf8"),
  });
  const bodies: string[] = [];
  const attempts = { preview: 0, download: 0 };
  for (const kind of ["preview", "download"] as const) {
    await installOwnedRoute(
      page,
      `**/evidence-records/${evidence.record_id}/${kind}-handle`,
      async (route) => {
        bodies.push(route.request().postData() ?? "");
        if (attempts[kind]++ === 0)
          await route.fulfill({
            status: 503,
            contentType: "application/json",
            body: evidenceTransientResponse(),
          });
        else await route.continue();
      },
    );
  }
  await openEvidenceSurface(page, incidentId);
  const preview = page.getByTestId(
    evidencePreviewButtonTestId(evidence.record_id),
  );
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: evidencePreviewButtonTestId(evidence.record_id),
  });
  await preview.click();
  await expect(preview).toHaveText("Retry Preview");
  await preview.click();
  await expect(
    page.getByTestId(evidencePreviewFrameTestId(evidence.record_id)),
  ).toBeVisible();
  await page.getByRole("button", { name: "Close", exact: true }).click();
  const download = page.getByTestId(
    evidenceDownloadButtonTestId(evidence.record_id),
  );
  await download.click();
  await expect(download).toHaveText("Retry Download");
  await download.click();
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(evidence.record_id)),
  ).toHaveText("Download requested");
  expect(attempts).toEqual({ preview: 2, download: 2 });
  expect(bodies).toEqual(["{}", "{}", "{}", "{}"]);
});

test("fences a pending linked Download when newer Preview loses current authorization", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-AUTH-FENCE"),
    "Evidence authority fence",
  );
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("auth-fence-source"),
    "timeline.activity_synopsis_text": "Authority fence source",
  });
  const evidence = await createUploadedEvidenceFixture(page, incidentId, {
    txnPrefix: "e5-preview",
    collectorPartyText: "Browser evidence",
    title: "Authority fence file",
    filename: "fence.txt",
    contentType: "text/plain",
    body: Buffer.from("fence preview", "utf8"),
  });
  await linkEvidenceToTimeline(page, source.record_id, source.row_version, [
    evidence.record_id,
  ]);
  let release = () => {};
  const gateBarrier = responseBarrier(page);
  release = gateBarrier.release;
  const gate = gateBarrier.promise;
  let observeHeld = () => {};
  const heldObservedBarrier = responseBarrier(page);
  observeHeld = heldObservedBarrier.release;
  const heldObserved = heldObservedBarrier.promise;
  let settleHeld = () => {};
  const heldSettledBarrier = responseBarrier(page);
  settleHeld = heldSettledBarrier.release;
  const heldSettled = heldSettledBarrier.promise;
  let downloads = 0;
  const observeDownload = () => {
    downloads++;
  };
  page.on("download", observeDownload);
  browserScopeFor(page).settleBeforeClosing(async () => {
    page.off("download", observeDownload);
  }, page.context());
  const heldPath = `**/evidence-records/${evidence.record_id}/download-handle`;
  await installOwnedRoute(page, heldPath, async (route) => {
    observeHeld();
    if (!(await awaitResponseBarrier(route, gate))) return;
    try {
      await route.continue();
    } catch {
      /* Superseded intent may abort transport. */
    } finally {
      settleHeld();
    }
  });
  await installOwnedRoute(
    page,
    `**/evidence-records/${evidence.record_id}/preview-handle`,
    async (route) => {
      await route.fulfill({
        status: 403,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "authorization_denied",
            status: 403,
            retryable: false,
            request_id: "request-authority",
            message: "private",
            details: {},
          },
        }),
      });
    },
  );
  try {
    await openTimelineSurface(page, incidentId);
    await openTimelineInspector(page, source.record_id);
    await page
      .getByTestId(
        evidenceDownloadButtonTestId(evidence.record_id, "inspector"),
      )
      .click();
    await heldObserved;
    await page
      .getByTestId(evidencePreviewButtonTestId(evidence.record_id, "inspector"))
      .click();
    await expect(
      page.getByTestId(
        evidencePreviewButtonTestId(evidence.record_id, "inspector"),
      ),
    ).toBeDisabled();
    release();
    await heldSettled;
    await expect(page.getByTestId(evidencePreviewPanelTestId())).toHaveCount(0);
    expect(downloads).toBe(0);

    release();

    await removeOwnedRoute(page, heldPath);
  } finally {
    gateBarrier.cancel();
  }
});
