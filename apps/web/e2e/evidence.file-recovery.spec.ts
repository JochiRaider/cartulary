import { Buffer } from "node:buffer";
import type { ViewRow } from "@cartulary/protocol-ts/http";
import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  draftCellTestId,
  evidenceAttachFileInputTestId,
  evidencePreviewButtonTestId,
  gridShellTestId,
  surfaceTabTestId,
  timelineDraftEvidenceFileInputTestId,
  timelineEvidenceFileInputTestId,
  workbookInspectorToggleTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type { Page, TestInfo } from "@playwright/test";
import { expect, test } from "./fixtures";

import { collectionItems } from "./support/entities/mentions";
import {
  openEvidenceSurface,
  openTimelineSurface,
  tinyPNG,
} from "./support/evidence/browser";
import {
  chooseEvidenceFile,
  openTimelineAttachmentFeedback,
} from "./support/evidence/uploads";
import { createIncident } from "./support/incidents/fixtures";
import { fetchOwnedJobResponse } from "./support/jobs/ownership";
import { responseBarrier } from "./support/runtime/cleanup";
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
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import { openTimelineInspector } from "./support/workbook/rowMutations";

async function switchFileSurface(page: Page, view: string) {
  const tab = page.getByTestId(surfaceTabTestId(view));
  if (await tab.isVisible()) await tab.click();
  else {
    await page.getByTestId(workbookSurfacesMenuTriggerTestId()).click();
    await page.getByTestId(workbookSurfacesMenuOptionTestId(view)).click();
  }
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
}

async function reviewFileSource(page: Page, filename: string) {
  await openTimelineAttachmentFeedback(page);
  const recovery = page.getByRole("group", {
    includeHidden: true,
    name: `File recovery: ${filename}`,
    exact: true,
  });
  await recovery
    .getByRole("button", { name: "Review original source", exact: true })
    .click();
  await recovery
    .getByRole("button", { name: "Use reviewed source", exact: true })
    .click();
  await recovery.getByRole("button", { name: "Resume", exact: true }).click();
}

test.beforeEach(({ page }) => {
  rejectUnexpectedPageErrors(page);
});

test("recovers each uncertain file stage with exact requests after remount and refresh failure", async ({
  page,
}, info) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("EUR-STAGES"),
    "Stage recovery",
  );
  const source = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("source"),
    "timeline.activity_synopsis_text": "Original upload source",
  });
  const slotBodies: string[] = [],
    createBodies: string[] = [],
    linkBodies: string[] = [];
  let transfers = 0,
    failedReads = 0,
    failRefresh = false;
  await installOwnedRoute(page, "**/api/v1/object-blobs", async (route) => {
    slotBodies.push(route.request().postData() ?? "");
    const response = await fetchOwnedJobResponse(page, route);
    expect(response.ok()).toBeTruthy();
    if (slotBodies.length === 1) await route.abort("failed");
    else await route.fulfill({ response });
  });
  await installOwnedRoute(page, "**/api/v1/object-uploads/*", async (route) => {
    transfers++;
    const response = await fetchOwnedJobResponse(page, route);
    expect(response.status()).toBe(204);
    await route.abort("failed");
  });
  await installOwnedRoute(
    page,
    `**/views/${evidenceViewSchemaId}/rows`,
    async (route) => {
      createBodies.push(route.request().postData() ?? "");
      const response = await fetchOwnedJobResponse(page, route);
      expect(response.ok()).toBeTruthy();
      if (createBodies.length === 1) await route.abort("failed");
      else await route.fulfill({ response });
    },
  );
  await installOwnedRoute(
    page,
    `**/records/${source.record_id}`,
    async (route) => {
      if (route.request().method() !== "PATCH") return route.continue();
      linkBodies.push(route.request().postData() ?? "");
      const response = await fetchOwnedJobResponse(page, route);
      expect(response.ok()).toBeTruthy();
      if (linkBodies.length === 1) await route.abort("failed");
      else {
        failRefresh = true;
        await route.fulfill({ response });
      }
    },
  );
  await installOwnedRoute(page, "**/views/*/query", async (route) => {
    if (failRefresh) {
      failedReads++;
      await route.abort("failed");
    } else await route.continue();
  });
  await openTimelineSurface(page, incident);
  await openTimelineInspector(page, source.record_id);
  const filename = "stage-recovery.png";
  await chooseEvidenceFile(
    page.getByTestId(timelineEvidenceFileInputTestId(source.record_id)),
    {
      name: filename,
      mimeType: "image/png",
      buffer: tinyPNG(),
    },
  );
  await openTimelineAttachmentFeedback(page);
  const recovery = page.getByRole("group", {
    includeHidden: true,
    name: `File recovery: ${filename}`,
    exact: true,
  });
  await expect(recovery).toContainText("Upload preparation is uncertain");
  const attention = page.getByRole("button", {
    name: /^View:/,
    exact: true,
  });
  await expect(attention).toBeVisible();
  await attention.click();
  await page
    .getByRole("button", {
      name: "Upload preparation is uncertain. Recover the same request.",
      exact: true,
    })
    .click();
  expect(slotBodies).toHaveLength(1);
  expect(transfers).toBe(0);
  expect(createBodies).toHaveLength(0);
  expect(linkBodies).toHaveLength(0);
  await info.attach("inspector-evidence-uncertain", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await recovery.getByRole("button", { name: "Resume", exact: true }).click();
  await expect(recovery).toContainText("Upload acknowledgement is uncertain");
  await info.attach("uncertain-transfer", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await recovery.getByRole("button", { name: "Resume", exact: true }).click();
  await expect(recovery).toContainText("Evidence creation is uncertain");
  await info.attach("uncertain-finalization", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await switchFileSurface(page, evidenceViewSchemaId);
  await switchFileSurface(page, timelineViewSchemaId);
  await openTimelineAttachmentFeedback(page);
  await recovery.getByRole("button", { name: "Resume", exact: true }).click();
  await expect(recovery).toContainText("Review the original Timeline row");
  await info.attach("accepted-evidence-awaiting-association", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await reviewFileSource(page, filename);
  await expect(recovery).toContainText("Timeline attachment is uncertain");
  await recovery
    .getByRole("button", { name: "Discard retained file work", exact: true })
    .click();
  await expect(recovery).toContainText("Stopped.");
  await info.attach("stopped-uncertain-association", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await recovery.getByRole("button", { name: "Resume", exact: true }).click();
  await expect(recovery).toContainText("Evidence attached. Refresh pending.");
  await expect.poll(() => failedReads).toBeGreaterThan(0);
  await info.attach("file-recovery-local", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  failRefresh = false;
  await recovery.getByRole("button", { name: "Refresh", exact: true }).click();
  await expect(recovery).toContainText("Evidence attached.");
  expect(slotBodies).toHaveLength(2);
  expect(slotBodies[1]).toBe(slotBodies[0]);
  expect(createBodies).toHaveLength(2);
  expect(createBodies[1]).toBe(createBodies[0]);
  expect(linkBodies).toHaveLength(2);
  expect(linkBodies[1]).toBe(linkBodies[0]);
  expect(transfers).toBe(1);
  const evidence = await queryViewRows(page, incident, evidenceViewSchemaId);
  const timeline = await queryViewRows(page, incident, timelineViewSchemaId);
  expect(evidence).toHaveLength(1);
  expect(timeline).toHaveLength(1);
  expect(evidence[0]?.cells["evidence.lifecycle_state"]?.value).toBe(
    "requested",
  );
  expect(timeline[0]?.record_id).toBe(source.record_id);
  expect(timeline[0]?.cells["timeline.evidence_count"]?.value).toBe(1);
  expect(
    collectionItems(timeline[0] as ViewRow, "timeline.attached_evidence_ids"),
  ).toHaveLength(1);
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
});

test("retains an existing Evidence attachment receipt after lost acknowledgement and navigation", async ({
  page,
}) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("EUR-ATTACH"),
    "Existing attachment recovery",
  );
  const source = await createViewRow(page, incident, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("evidence"),
    "evidence.title": "Original requested Evidence",
  });
  const bodies: string[] = [];
  let slots = 0,
    transfers = 0,
    patches = 0;
  observePageRequests(page, (request) => {
    if (request.url().endsWith("/object-blobs")) slots++;
    if (request.method() === "PUT") transfers++;
    if (request.method() === "PATCH") patches++;
  });
  await installOwnedRoute(
    page,
    `**/evidence-records/${source.record_id}/attach-blob`,
    async (route) => {
      bodies.push(route.request().postData() ?? "");
      const response = await fetchOwnedJobResponse(page, route);
      expect(response.ok()).toBeTruthy();
      if (bodies.length === 1) await route.abort("failed");
      else await route.fulfill({ response });
    },
  );
  await openEvidenceSurface(page, incident);
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: evidencePreviewButtonTestId(source.record_id),
  });
  await chooseEvidenceFile(
    page.getByTestId(evidenceAttachFileInputTestId(source.record_id)),
    {
      name: "existing-empty.txt",
      mimeType: "text/plain",
      buffer: Buffer.from(""),
    },
  );
  const recovery = page.getByRole("group", {
    name: "File recovery: existing-empty.txt",
    exact: true,
  });
  await expect(recovery).toContainText(
    "Attachment acknowledgement is uncertain",
  );
  await switchFileSurface(page, timelineViewSchemaId);
  await switchFileSurface(page, evidenceViewSchemaId);
  await recovery.getByRole("button", { name: "Resume", exact: true }).click();
  await expect(recovery).toContainText("File attached. Custody unchanged.");
  expect(bodies).toHaveLength(2);
  expect(bodies[1]).toBe(bodies[0]);
  expect([slots, transfers, patches]).toEqual([1, 1, 0]);
  const rows = await queryViewRows(page, incident, evidenceViewSchemaId);
  expect(rows).toHaveLength(1);
  expect(rows[0]?.row_version).toBe(source.row_version + 1);
  expect(rows[0]?.cells["evidence.lifecycle_state"]?.value).toBe("requested");
});

async function exerciseDraftCreationOrder(
  page: Page,
  info: TestInfo,
  ordinaryFirst: true | false,
) {
  test.setTimeout(120_000);

  const incident = await createIncident(
    page,
    uniqueIncidentKey("EUR-DRAFT-ORDER"),
    "Shared draft creation",
  );
  const gateBarrier = responseBarrier(page);
  const release = gateBarrier.release;
  const gate = gateBarrier.promise;
  const createBodies: string[] = [];
  let slots = 0,
    transfers = 0;
  const heldPath = ordinaryFirst
    ? "**/api/v1/object-uploads/*"
    : `**/views/${timelineViewSchemaId}/rows`;
  await installOwnedRoute(page, heldPath, async (route) => {
    if (!(await awaitResponseBarrier(route, gate))) return;
    await route.continue();
  });
  const watch = (request: import("@playwright/test").Request) => {
    if (request.url().endsWith("/object-blobs")) slots++;
    if (request.method() === "PUT") transfers++;
    if (request.url().endsWith(`/views/${timelineViewSchemaId}/rows`))
      createBodies.push(request.postData() ?? "");
  };
  const stopObserving = observePageRequests(page, watch);
  await openTimelineSurface(page, incident);
  const draftId = draftCellTestId("timeline.activity_synopsis_text");
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: draftId,
  });
  const grid = page.getByRole("region", {
    name: "Timeline file work area",
    exact: true,
  });
  await grid.evaluate((element) => {
    const data = new DataTransfer();
    data.items.add(new File(["one"], "one.txt"));
    data.items.add(new File(["two"], "two.txt"));
    element.dispatchEvent(
      new DragEvent("drop", { bubbles: true, dataTransfer: data }),
    );
  });
  await expect(
    page
      .getByRole("region", { name: "Timeline attachments", exact: true })
      .getByText("Choose one file at a time.", { exact: true }),
  ).toBeVisible();
  expect(slots).toBe(0);
  const filename = ordinaryFirst ? "clipboard-draft.png" : "picker-draft.png";
  if (ordinaryFirst) {
    await page.getByTestId(draftId).click();
    await page.getByTestId(draftId).evaluate((element, bytes) => {
      const data = new DataTransfer();
      data.items.add(
        new File([new Uint8Array(bytes)], "clipboard-draft.png", {
          type: "image/png",
        }),
      );
      element.dispatchEvent(
        new ClipboardEvent("paste", { bubbles: true, clipboardData: data }),
      );
    }, Array.from(tinyPNG()));
  } else {
    await page
      .getByTestId(workbookInspectorToggleTestId(timelineViewSchemaId))
      .click();
    await chooseEvidenceFile(
      page.getByTestId(timelineDraftEvidenceFileInputTestId()),
      {
        name: filename,
        mimeType: "image/png",
        buffer: tinyPNG(),
      },
    );
  }
  await openTimelineAttachmentFeedback(page);
  const recovery = page.getByRole("group", {
    includeHidden: true,
    name: `File recovery: ${filename}`,
    exact: true,
  });
  if (ordinaryFirst) await expect(recovery).toContainText("Uploading file");
  else await expect.poll(() => createBodies.length).toBe(1);
  await scrollGridTargetIntoView({
    page,
    surface: timelineViewSchemaId,
    targetTestId: draftId,
  });
  await page.getByTestId(draftId).click();
  await page.keyboard.type("Typed during pending file work");
  await page.keyboard.press("Enter");
  if (ordinaryFirst) {
    await expect
      .poll(
        async () =>
          (await queryViewRows(page, incident, timelineViewSchemaId)).length,
      )
      .toBe(1);
    expect(
      await queryViewRows(page, incident, evidenceViewSchemaId),
    ).toHaveLength(0);
  }
  release?.();
  await expect
    .poll(
      async () =>
        (await queryViewRows(page, incident, evidenceViewSchemaId)).length,
    )
    .toBe(1);
  if (ordinaryFirst) {
    await expect(
      recovery.getByRole("button", { name: "Resume", exact: true }),
    ).toBeVisible();
    if (
      await recovery
        .getByRole("button", { name: "Review original source", exact: true })
        .isVisible()
    )
      await reviewFileSource(page, filename);
    else
      await recovery
        .getByRole("button", { name: "Resume", exact: true })
        .click();
  }
  await expect
    .poll(async () => {
      const rows = await queryViewRows(page, incident, timelineViewSchemaId);
      return [
        rows.length,
        rows[0]?.cells["timeline.activity_synopsis_text"]?.value,
        rows[0]?.cells["timeline.evidence_count"]?.value,
      ];
    })
    .toEqual([1, "Typed during pending file work", 1]);
  expect(createBodies).toHaveLength(1);
  expect(slots).toBe(1);
  expect(transfers).toBe(1);
  await info.attach(
    ordinaryFirst
      ? "typing-during-transfer"
      : "typing-during-screenshot-create",
    { body: await page.screenshot(), contentType: "image/png" },
  );
  await removeOwnedRoute(page, heldPath);
  stopObserving();
}

test("keeps draft typing and one creation when ordinary authoring finishes first", async ({
  page,
}, info) => {
  await exerciseDraftCreationOrder(page, info, true);
});

test("keeps draft typing and one creation when file creation finishes first", async ({
  page,
}, info) => {
  await exerciseDraftCreationOrder(page, info, false);
});

test("reviews the original source after a rejected file link while preserving unrelated selection and edits", async ({
  page,
}, info) => {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("EUR-SOURCE"),
    "Original source review",
  );
  const source = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("source"),
    "timeline.activity_synopsis_text": "Original file source",
  });
  const other = await createViewRow(page, incident, timelineViewSchemaId, {
    client_txn_id: uniqueTxn("other"),
    "timeline.activity_synopsis_text": "Unrelated selected row",
  });
  const bodies: string[] = [];
  let creates = 0,
    transfers = 0;
  observePageRequests(page, (request) => {
    if (request.url().endsWith(`/views/${evidenceViewSchemaId}/rows`))
      creates++;
    if (request.method() === "PUT") transfers++;
  });
  await installOwnedRoute(
    page,
    `**/records/${source.record_id}`,
    async (route) => {
      if (route.request().method() !== "PATCH") return route.continue();
      bodies.push(route.request().postData() ?? "");
      if (bodies.length === 1)
        await patchRecord(page, source.record_id, {
          view_schema_id: timelineViewSchemaId,
          base_row_version: source.row_version,
          client_txn_id: uniqueTxn("other-editor"),
          changes: [
            {
              field_key: "timeline.activity_synopsis_text",
              value: "Reviewed concurrent source edit",
            },
          ],
        });
      if (bodies.length === 1) {
        await route.fulfill({
          status: 409,
          contentType: "application/json",
          body: JSON.stringify({
            error: {
              code: "row_version_conflict",
              status: 409,
              message: "Source changed",
              retryable: false,
              request_id: "source-conflict",
              details: {},
            },
          }),
        });
      } else {
        const response = await fetchOwnedJobResponse(page, route);
        expect(response.status()).toBe(200);
        await route.fulfill({ response });
      }
    },
  );
  await openTimelineSurface(page, incident);
  await openTimelineInspector(page, source.record_id);
  await page
    .getByRole("region", {
      name: "File attachment for this Timeline record",
      exact: true,
    })
    .evaluate((element, bytes) => {
      const data = new DataTransfer();
      data.items.add(
        new File([new Uint8Array(bytes)], "review-source.png", {
          type: "image/png",
        }),
      );
      element.dispatchEvent(
        new DragEvent("drop", { bubbles: true, dataTransfer: data }),
      );
    }, Array.from(tinyPNG()));
  await openTimelineAttachmentFeedback(page);
  const recovery = page.getByRole("group", {
    includeHidden: true,
    name: "File recovery: review-source.png",
    exact: true,
  });
  await expect(recovery).toContainText("Timeline attachment needs recovery");
  await openTimelineInspector(page, other.record_id);
  await recovery
    .getByRole("button", { name: "Review original source", exact: true })
    .click();
  await expect(recovery).toContainText("Reviewed concurrent source edit");
  await info.attach("original-source-review", {
    body: await page.screenshot(),
    contentType: "image/png",
  });
  await recovery
    .getByRole("button", { name: "Use reviewed source", exact: true })
    .click();
  await recovery.getByRole("button", { name: "Resume", exact: true }).click();
  await expect(recovery).toContainText("Evidence attached.");
  const rows = await queryViewRows(page, incident, timelineViewSchemaId);
  expect(
    rows.find((row) => row.record_id === source.record_id)?.cells[
      "timeline.activity_synopsis_text"
    ]?.value,
  ).toBe("Reviewed concurrent source edit");
  expect(
    rows.find((row) => row.record_id === source.record_id)?.cells[
      "timeline.evidence_count"
    ]?.value,
  ).toBe(1);
  expect(
    rows.find((row) => row.record_id === other.record_id)?.cells[
      "timeline.evidence_count"
    ]?.value,
  ).toBe(0);
  expect(bodies).toHaveLength(2);
  expect(bodies[1]).not.toBe(bodies[0]);
  expect([creates, transfers]).toEqual([1, 1]);
});
