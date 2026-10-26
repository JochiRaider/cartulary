import {
  dataTestIdSelector,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridShellTestId,
  timelineInspectorSectionTestId,
  timelineInspectorTestId,
  workbookFilterOperatorTestId,
  workbookFilterPopoverTriggerTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Request } from "@playwright/test";
import { expect, test } from "./fixtures";
import { observeEvidenceInspectorPaintInBrowser } from "./measurement/evidenceInspectorTimingSupport";

import { collectionItems } from "./support/entities/mentions";
import {
  linkEvidenceToTimeline,
  openTimelineSurface,
  tinyPNG,
} from "./support/evidence/browser";
import { createUploadedEvidenceFixture } from "./support/evidence/fixtures";
import { createIncident } from "./support/incidents/fixtures";
import { createIncidentMemberUser } from "./support/incidents/memberships";
import { webBase } from "./support/runtime/configuration";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  observePageRequests,
  rejectUnexpectedPageErrors,
} from "./support/runtime/observations";
import { readHttpOperationResponse } from "./support/transport/publicHttpOperationClient";
import { createViewRow } from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

test.beforeEach(({ page }) => {
  rejectUnexpectedPageErrors(page);
});

test("pages real linked Evidence metadata without issuing access handles", async ({
  page,
}) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-METADATA"),
    "Evidence metadata windows",
  );
  const source = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("metadata-source"),
    "timeline.activity_synopsis_text": "Metadata window source",
  });
  const ids: string[] = [];
  for (let i = 0; i < 7; i += 1) {
    const item = await createViewRow(page, incidentId, evidenceViewSchemaId, {
      client_txn_id: uniqueTxn("metadata-requested"),
      "evidence.title": `Requested evidence ${i + 1}`,
    });
    ids.push(item.record_id);
  }
  const linked = await linkEvidenceToTimeline(
    page,
    source.record_id,
    source.row_version,
    ids as [string, ...string[]],
  );
  const order = collectionItems(linked, "timeline.attached_evidence_ids").map(
    (item) => item.linked_record_id,
  );
  expect(order).toHaveLength(7);
  const handleRequests: string[] = [];
  let metadataRequests = 0;
  const observe = (request: Request) => {
    if (
      new URL(request.url()).pathname ===
      `/api/v1/incidents/${incidentId}/evidence-metadata`
    )
      metadataRequests++;
    if (
      /\/evidence-records\/[^/]+\/(preview|download)-handle$/u.test(
        new URL(request.url()).pathname,
      )
    )
      handleRequests.push(request.url());
  };
  const stopObserving = observePageRequests(page, observe);
  try {
    await openTimelineSurface(page, incidentId);
    const response = page.waitForResponse(
      (response) =>
        response.url().endsWith(`/incidents/${incidentId}/evidence-metadata`) &&
        response.request().method() === "POST",
    );
    await openTimelineInspector(page, source.record_id);
    const first = await readHttpOperationResponse(
      await response,
      "queryEvidenceMetadata",
    );
    expect(first.data.items.map((item) => item.record_id)).toEqual(
      order.slice(0, 5),
    );
    for (const item of first.data.items)
      expect(item).toMatchObject({
        lifecycle_state: "requested",
        has_blob: false,
        upload_state: null,
        preview_eligibility: {
          eligible: false,
          reason_code: "no_visible_blob",
        },
      });
    const panel = page.getByTestId(timelineInspectorSectionTestId("evidence"));
    await expect(
      panel.getByText("Linked evidence: 7", { exact: true }),
    ).toBeVisible();
    await expect(
      panel.getByText("Attached evidence count: 0", { exact: true }),
    ).toBeVisible();
    const list = panel.getByRole("list", { name: "Linked Evidence" });
    await expect(list.getByRole("listitem")).toHaveCount(5);
    await expect(
      list.locator('[data-evidence-metadata-state="ready"]'),
    ).toHaveCount(5);
    expect(metadataRequests).toBe(1);
    const next = panel.getByRole("button", { name: "Next evidence" });
    await next.focus();
    await expect(next).toBeFocused();
    const nextResponse = page.waitForResponse((response) =>
      response.url().endsWith(`/incidents/${incidentId}/evidence-metadata`),
    );
    await page.keyboard.press("Enter");
    const second = await readHttpOperationResponse(
      await nextResponse,
      "queryEvidenceMetadata",
    );
    expect(second.data.items.map((item) => item.record_id)).toEqual(
      order.slice(5),
    );
    await expect(list.getByRole("listitem")).toHaveCount(2);
    await expect(
      list.locator('[data-evidence-metadata-state="ready"]'),
    ).toHaveCount(2);
    expect(metadataRequests).toBe(2);
    await expect(
      panel.getByRole("button", { name: "Next evidence" }),
    ).toBeDisabled();
    await expect(
      page.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();
    await page
      .getByTestId(workbookInspectorCloseButtonTestId(timelineViewSchemaId))
      .click();
    await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
    const reopened = page.waitForResponse(
      (response) =>
        response.url().endsWith(`/incidents/${incidentId}/evidence-metadata`) &&
        response.request().method() === "POST",
    );
    await page.evaluate(() => performance.clearResourceTimings());
    await openTimelineInspector(page, source.record_id);
    const reopenedResponse = await reopened;
    await reopenedResponse.finished();
    await expect(
      list.locator('[data-evidence-metadata-state="ready"]'),
    ).toHaveCount(2);
    expect(metadataRequests).toBe(3);
    expect(
      await page.evaluate(
        (path) =>
          performance
            .getEntriesByType("resource")
            .filter((entry) => new URL(entry.name).pathname === path).length,
        `/api/v1/incidents/${incidentId}/evidence-metadata`,
      ),
    ).toBe(1);
    expect(handleRequests).toEqual([]);
  } finally {
    stopObserving();
  }
});

test("reopens fresh bounded metadata for a filtered Timeline row with 100 available Evidence links", async ({
  page: setupPage,
  browserScope,
  sessionTracker,
}) => {
  let page = setupPage;
  test.setTimeout(120_000);
  await page.setViewportSize({ width: 1920, height: 1080 });
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-METADATA-LARGE"),
    "Large linked metadata window",
  );
  const ids: string[] = [];
  for (let index = 0; index < 100; index++) {
    const evidence = await createUploadedEvidenceFixture(page, incidentId, {
      txnPrefix: "e5-preview",
      collectorPartyText: "Browser evidence",
      title: `Available evidence ${String(index + 1).padStart(3, "0")}`,
      filename: "synthetic.png",
      contentType: "image/png",
      body: tinyPNG(),
    });
    ids.push(evidence.record_id);
  }
  let row = await createViewRow(page, incidentId, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("large-metadata-target"),
    "timeline.activity_synopsis_text": "Large linked metadata target",
  });
  for (let offset = 0; offset < ids.length; offset += 64)
    row = await linkEvidenceToTimeline(
      page,
      row.record_id,
      row.row_version,
      ids.slice(offset, offset + 64) as [string, ...string[]],
    );
  const member = await createIncidentMemberUser(page, incidentId, {
    email: uniqueEmail("large-metadata-editor"),
    display_name: "Metadata editor",
    initial_password: "MetadataEditor1!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const context = await browserScope.newContext({
    viewport: { width: 1920, height: 1080 },
    deviceScaleFactor: 1,
  });
  {
    page = await context.newPage();
    await page.goto(webBase);
    await sessionTracker.loginTrackedUser(page, {
      createdBy: "large metadata regression",
      email: member.email,
      password: member.initial_password,
      purpose: "ordinary editor metadata lifetime",
      userId: member.user_id,
    });
    await openTimelineSurface(page, incidentId);
    const view = timelineViewSchemaId;
    await page.getByTestId(workbookFilterPopoverTriggerTestId(view)).click();
    await page
      .getByTestId(gridFilterFieldTestId(view))
      .selectOption("timeline.has_evidence");
    await page
      .getByTestId(workbookFilterOperatorTestId(view))
      .selectOption("eq");
    await page.getByTestId(gridFilterValueTestId(view)).selectOption("true");
    await page.getByTestId(gridFilterApplyTestId(view)).click();
    const filter = page.getByTestId(workbookFilterPopoverTriggerTestId(view));
    if ((await filter.getAttribute("aria-expanded")) === "true")
      await filter.click();
    const frame = page.locator(`[data-workbook-view-schema-id="${view}"]`);
    await expect(frame).toHaveAttribute("data-workbook-query-pending", "false");
    await expect(
      frame.locator('[role="row"][data-grid-record-id]'),
    ).toHaveCount(1);
    const metadataPath = `/api/v1/incidents/${incidentId}/evidence-metadata`;
    const counts = { metadata: 0, access: 0 };
    const observe = (request: Request) => {
      const path = new URL(request.url()).pathname;
      if (path === metadataPath) counts.metadata++;
      if (/\/evidence-records\/[^/]+\/(?:preview|download)-handle$/u.test(path))
        counts.access++;
    };
    const stopObserving = observePageRequests(page, observe);
    try {
      const initialRead = page.waitForResponse(
        (response) =>
          new URL(response.url()).pathname === metadataPath &&
          response.request().method() === "POST",
      );
      await openTimelineInspector(page, row.record_id);
      const initialMetadata = await readHttpOperationResponse(
        await initialRead,
        "queryEvidenceMetadata",
      );
      const panel = page.getByTestId(
        timelineInspectorSectionTestId("evidence"),
      );
      await expect(
        panel.locator('[data-evidence-metadata-state="ready"]'),
      ).toHaveCount(5);
      await expect(
        panel.getByText("Attached evidence count: 100", { exact: true }),
      ).toBeVisible();
      const evidenceSection = page.locator(
        '[data-inspector-navigation-panel="evidence"]',
      );
      if (await evidenceSection.isVisible()) await evidenceSection.click();
      else {
        await page.getByRole("button", { name: /^Sections:/u }).click();
        await page
          .getByRole("navigation", { name: "Inspector sections", exact: true })
          .getByRole("button", { name: "Evidence", exact: true })
          .click();
      }
      const selectedEvidence = page
        .getByRole("button", { name: "Sections: Evidence", exact: true })
        .or(
          page.locator(
            '[data-inspector-navigation-panel="evidence"][aria-current="location"]',
          ),
        );
      await expect(selectedEvidence).toBeVisible();
      expect(counts).toEqual({ metadata: 1, access: 0 });
      await page.getByTestId(workbookInspectorCloseButtonTestId(view)).click();
      await expect(page.getByTestId(timelineInspectorTestId())).toHaveCount(0);
      // Functional detector calibration against real rendering, not product
      // latency evidence: this marker deliberately precedes driver dispatch.
      await page.evaluate(() => {
        performance.clearMarks();
        performance.clearResourceTimings();
        performance.mark("evidence_detector_probe");
      });
      const paint = page.evaluate(observeEvidenceInspectorPaintInBrowser, {
        inspectorSelector: dataTestIdSelector(timelineInspectorTestId()),
        evidenceSelector: dataTestIdSelector(
          timelineInspectorSectionTestId("evidence"),
        ),
        metadataPath,
        target: {
          record_id: row.record_id,
          row_version: row.row_version,
          summary: "Large linked metadata target",
          items: initialMetadata.data.items.map((item) => ({
            record_id: item.record_id,
            row_version: item.row_version,
            title: item.title ?? "",
          })),
        },
        startMark: "evidence_detector_probe",
        stopMark: "evidence_detector_completed",
        timeoutMs: 10000,
      });
      const reopened = page.waitForResponse(
        (response) =>
          new URL(response.url()).pathname === metadataPath &&
          response.request().method() === "POST",
      );
      const [visible, response] = await Promise.all([
        paint,
        reopened,
        (async () => {
          await page.getByTestId(workbookInspectorToggleTestId(view)).click();
          await expect(selectedEvidence).toBeVisible();
        })(),
      ]);
      await response.finished();
      expect(visible.second_frame_at).toBeGreaterThan(visible.first_frame_at);
      await expect(
        panel.locator('[data-evidence-metadata-state="ready"]'),
      ).toHaveCount(5);
      expect(counts).toEqual({ metadata: 2, access: 0 });
      expect(
        await page.evaluate(
          (path) =>
            performance
              .getEntriesByType("resource")
              .filter((entry) => new URL(entry.name).pathname === path).length,
          metadataPath,
        ),
      ).toBe(1);
    } finally {
      stopObserving();
    }

    await browserScope.closeContext(context);
  }
});
