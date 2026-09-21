import { Buffer } from "node:buffer";
import type {
  AttachBlobToEvidenceRecordResponse,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  dataTestIdSelector,
  draftCellTestId,
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  rowCellTestId,
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
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { collectionItems } from "./support/entities/mentions";
import { createAndUploadObjectBlob } from "./support/evidence/uploads";
import { createIncident } from "./support/incidents/fixtures";
import { apiBase } from "./support/runtime/configuration";
import {
  uniqueIncidentKey,
  uniqueTxn,
} from "./support/runtime/fixtureIdentity";
import {
  createViewRow,
  patchRecord,
  queryViewRows,
} from "./support/workbook/query";
import {
  editGenericCell,
  openTimelineInspector,
} from "./support/workbook/rowMutations";

test.beforeEach(({ page }) => {
  failOnUnexpectedPageError(page);
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
  await page
    .getByTestId(timelineEvidenceFileInputTestId(timelineRow.record_id))
    .setInputFiles({
      name: "selected-screenshot.png",
      mimeType: "image/png",
      buffer: tinyPNG(),
    });

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
  await page.getByTestId(timelineDraftEvidenceFileInputTestId()).setInputFiles({
    name: "draft-screenshot.png",
    mimeType: "image/png",
    buffer: tinyPNG(),
  });

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

test("redeems inline-safe previews and shows explicit blocked-preview outcomes", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-PREVIEW"),
    "Evidence evidence preview",
  );
  const safe = await createUploadedEvidence(page, incidentId, {
    title: "Safe text preview",
    filename: "safe.txt",
    contentType: "text/plain",
    body: Buffer.from("safe preview body", "utf8"),
  });
  const unsafe = await createUploadedEvidence(page, incidentId, {
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

test("tracks requested evidence before a blob exists and later advances it", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("EVIDENCE-LIFECYCLE"),
    "Evidence requested evidence",
  );
  const timelineRow = await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e5-requested-timeline"),
      "timeline.activity_synopsis_text": "Requested package tracking",
    },
  );

  await openEvidenceSurface(page, incidentId);
  await setGenericCreateField(page, "evidence.title", "Requested package");
  await setGenericCreateField(page, "evidence.storage_ref", "ticket://E5-04");
  const submitTestId = genericCreateSubmitTestId(evidenceViewSchemaId);
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: submitTestId,
  });
  await page.getByTestId(submitTestId).click();

  const requested = await waitForEvidenceRow(
    page,
    incidentId,
    "Requested package",
  );
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: rowCellTestId(requested.record_id, "evidence.title"),
  });
  await page
    .getByTestId(rowCellTestId(requested.record_id, "evidence.title"))
    .focus();
  await expect(
    page.getByTestId(rowCellTestId(requested.record_id, "evidence.title")),
  ).toHaveText("Requested package");
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: rowCellTestId(
      requested.record_id,
      "evidence.lifecycle_state",
    ),
  });
  await expect(
    page.getByTestId(
      rowCellTestId(requested.record_id, "evidence.lifecycle_state"),
    ),
  ).toHaveText("requested");
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: rowCellTestId(requested.record_id, "evidence.upload_state"),
  });
  await expect(
    page.getByTestId(
      rowCellTestId(requested.record_id, "evidence.upload_state"),
    ),
  ).toHaveText("pending");

  const linkedTimeline = await patchRecord(page, timelineRow.record_id, {
    view_schema_id: timelineViewSchemaId,
    base_row_version: timelineRow.row_version,
    client_txn_id: uniqueTxn("e5-requested-link"),
    changes: [
      {
        field_key: "timeline.attached_evidence_ids",
        action_payload: {
          kind: "collection_actions_v1",
          actions: [
            {
              op: "add_record_ref",
              linked_record_id: requested.record_id,
            },
          ],
        },
      },
    ],
  });
  expect(
    collectionItems(linkedTimeline, "timeline.attached_evidence_ids").some(
      (item) => item.linked_record_id === requested.record_id,
    ),
  ).toBe(true);
  expect(linkedTimeline.cells["timeline.evidence_count"]?.value).toBe(0);
  expect(linkedTimeline.cells["timeline.has_evidence"]?.value).toBe(false);
  await expect
    .poll(async () => {
      const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
      const row = rows.find(
        (candidate) => candidate.record_id === timelineRow.record_id,
      );
      return [
        row?.cells["timeline.evidence_count"]?.value,
        row?.cells["timeline.has_evidence"]?.value,
      ];
    })
    .toEqual([0, false]);

  await openTimelineSurface(page, incidentId);
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();

  await openEvidenceSurface(page, incidentId);
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: evidencePreviewButtonTestId(requested.record_id),
  });
  const attachedResponse = page.waitForResponse(
    (response) =>
      response.request().method() === "POST" &&
      response
        .url()
        .endsWith(`/evidence-records/${requested.record_id}/attach-blob`),
  );
  await page
    .getByTestId(evidenceAttachFileInputTestId(requested.record_id))
    .setInputFiles({
      name: "requested.txt",
      mimeType: "text/plain",
      buffer: Buffer.from(
        "evidence_lifecycle requested evidence payload",
        "utf8",
      ),
    });

  const attached = await attachedResponse;
  expect(attached.ok()).toBe(true);
  const advanced = (
    (await attached.json()) as AttachBlobToEvidenceRecordResponse
  ).data.row;
  expect(advanced.record_id).toBe(requested.record_id);
  expect(advanced.row_version).toBeGreaterThan(requested.row_version);
  expect(advanced.cells["evidence.lifecycle_state"]?.value).toBe("requested");
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: rowCellTestId(advanced.record_id, "evidence.lifecycle_state"),
  });
  await expect(
    page.getByTestId(
      rowCellTestId(advanced.record_id, "evidence.lifecycle_state"),
    ),
  ).toHaveText("requested");
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: rowCellTestId(advanced.record_id, "evidence.upload_state"),
  });
  await expect(
    page.getByTestId(
      rowCellTestId(advanced.record_id, "evidence.upload_state"),
    ),
  ).toHaveText("available");
  await editGenericCell(
    page,
    evidenceViewSchemaId,
    advanced.record_id,
    "evidence.lifecycle_state",
    "available",
  );
  await expect(
    page.getByTestId(
      rowCellTestId(advanced.record_id, "evidence.lifecycle_state"),
    ),
  ).toHaveText("available");
  expect(
    (await queryViewRows(page, incidentId, evidenceViewSchemaId)).filter(
      (row) => row.record_id === requested.record_id,
    ),
  ).toHaveLength(1);
  await expect
    .poll(async () => {
      const rows = await queryViewRows(page, incidentId, timelineViewSchemaId);
      const row = rows.find(
        (candidate) => candidate.record_id === timelineRow.record_id,
      );
      return [
        row?.cells["timeline.evidence_count"]?.value,
        row?.cells["timeline.has_evidence"]?.value,
      ];
    })
    .toEqual([1, true]);
  await openTimelineSurface(page, incidentId);
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
  const timelineRows = await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  );
  expect(
    timelineRows.filter((row) => row.record_id === timelineRow.record_id),
  ).toHaveLength(1);
});

test("refreshes a second live workbook from the real evidence attach stream", async ({
  browser,
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

  const listenerContext = await browser.newContext({
    storageState: await page.context().storageState(),
  });
  const listener = await listenerContext.newPage();
  failOnUnexpectedPageError(listener);
  try {
    const socketMonitor = installIncidentSocketMonitor(listener, incidentId);
    await openTimelineSurface(listener, incidentId);
    await socketMonitor.waitForMessage("hello_ack");
    const listenerURL = listener.url();
    await expect(
      listener.getByTestId(gridShellTestId(timelineViewSchemaId)),
    ).toBeVisible();

    await openTimelineSurface(page, incidentId);
    await openTimelineInspector(page, timelineRow.record_id);
    await page
      .getByTestId(timelineEvidenceFileInputTestId(timelineRow.record_id))
      .setInputFiles({
        name: "socket-refresh.png",
        mimeType: "image/png",
        buffer: tinyPNG(),
      });

    await socketMonitor.waitForRecordChanged(timelineRow.record_id);
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
  } finally {
    await listenerContext.close();
  }
});

async function openEvidenceSurface(page: Page, incidentId: string) {
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      evidenceViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
  ).toBeVisible();
}

async function openTimelineSurface(page: Page, incidentId: string) {
  await page.goto(
    `/?incident_id=${incidentId}&view_schema_id=${encodeURIComponent(
      timelineViewSchemaId,
    )}`,
  );
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
}

async function setGenericCreateField(
  page: Page,
  fieldKey: string,
  value: string,
) {
  const fieldTestId = genericCreateFieldTestId(fieldKey);
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: fieldTestId,
  });
  await page.getByTestId(fieldTestId).fill(value);
}

async function waitForEvidenceRow(
  page: Page,
  incidentId: string,
  title: string,
): Promise<ViewRow> {
  const deadline = Date.now() + 5_000;
  while (Date.now() <= deadline) {
    const rows = await queryViewRows(page, incidentId, evidenceViewSchemaId);
    const row = rows.find(
      (candidate) => candidate.cells["evidence.title"]?.value === title,
    );
    if (row !== undefined) {
      await scrollGridTargetIntoView({
        page,
        surface: evidenceViewSchemaId,
        targetTestId: rowCellTestId(row.record_id, "evidence.title"),
      });
      await expect(
        page.getByTestId(rowCellTestId(row.record_id, "evidence.title")),
      ).toBeVisible();
      return row;
    }
    await page.waitForTimeout(50);
  }
  throw new Error(`timed out waiting for evidence row ${title}`);
}

function tinyPNG() {
  return Buffer.from(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=",
    "base64",
  );
}

function failOnUnexpectedPageError(page: Page) {
  page.on("pageerror", (error) => {
    throw new Error(`Unexpected page error: ${error.stack ?? error.message}`);
  });
}

type SocketMessage = {
  type: string;
  payload: Record<string, unknown>;
};

function installIncidentSocketMonitor(page: Page, incidentId: string) {
  const messages: SocketMessage[] = [];
  const waiters: Array<{
    matches: (message: SocketMessage) => boolean;
    reject: (error: Error) => void;
    resolve: (message: SocketMessage) => void;
    timeout: ReturnType<typeof setTimeout>;
  }> = [];

  page.on("websocket", (socket) => {
    if (!socket.url().includes(`/ws/v1/incidents/${incidentId}`)) {
      return;
    }
    socket.on("framereceived", ({ payload }) => {
      const message = parseSocketPayload(payload);
      if (!message) {
        return;
      }
      messages.push(message);
      for (const waiter of [...waiters]) {
        if (!waiter.matches(message)) {
          continue;
        }
        clearTimeout(waiter.timeout);
        waiters.splice(waiters.indexOf(waiter), 1);
        waiter.resolve(message);
      }
    });
  });

  const waitFor = (
    matches: (message: SocketMessage) => boolean,
    label: string,
  ) => {
    const existing = messages.find(matches);
    if (existing) {
      return Promise.resolve(existing);
    }
    return new Promise<SocketMessage>((resolve, reject) => {
      const waiter = {
        matches,
        reject,
        resolve,
        timeout: setTimeout(() => {
          waiters.splice(waiters.indexOf(waiter), 1);
          reject(new Error(`timed out waiting for ${label}`));
        }, 10_000),
      };
      waiters.push(waiter);
    });
  };

  return {
    waitForMessage: (type: string) =>
      waitFor((message) => message.type === type, `socket message ${type}`),
    waitForRecordChanged: (recordId: string) =>
      waitFor(
        (message) =>
          message.type === "record_changed" &&
          message.payload.record_id === recordId,
        `record_changed for ${recordId}`,
      ),
  };
}

function parseSocketPayload(payload: string | Buffer): SocketMessage | null {
  const text = typeof payload === "string" ? payload : payload.toString("utf8");
  let parsed: unknown;
  try {
    parsed = JSON.parse(text);
  } catch {
    return null;
  }
  if (!parsed || typeof parsed !== "object") {
    return null;
  }
  const candidate = parsed as { payload?: unknown; type?: unknown };
  if (typeof candidate.type !== "string") {
    return null;
  }
  return {
    type: candidate.type,
    payload:
      candidate.payload && typeof candidate.payload === "object"
        ? (candidate.payload as Record<string, unknown>)
        : {},
  };
}

function collectObjectUploadRoutes(page: Page): string[] {
  const routes: string[] = [];
  page.on("request", (request) => {
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

async function createUploadedEvidence(
  page: Page,
  incidentId: string,
  options: {
    title: string;
    filename: string;
    contentType: string;
    body: Buffer;
  },
) {
  const row = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("e5-preview-evidence"),
    "evidence.title": options.title,
    "evidence.collector_party_text": "Browser evidence",
  });
  const blob = await createAndUploadObjectBlob(page, {
    body: options.body,
    clientTxnId: uniqueTxn("e5-preview-blob"),
    contentType: options.contentType,
    filename: options.filename,
    incidentId,
  });

  const attach = await page.request.post(
    `${apiBase}/api/v1/evidence-records/${row.record_id}/attach-blob`,
    {
      headers: await csrfHeaders(page),
      data: {
        object_blob_id: blob.object_blob_id,
        base_row_version: row.row_version,
        client_txn_id: uniqueTxn("e5-preview-attach"),
      },
    },
  );
  expect(attach.ok()).toBeTruthy();
  const attachEnvelope =
    (await attach.json()) as AttachBlobToEvidenceRecordResponse;
  return patchRecord(page, row.record_id, {
    view_schema_id: evidenceViewSchemaId,
    base_row_version: attachEnvelope.data.row.row_version,
    client_txn_id: uniqueTxn("e5-preview-available"),
    changes: [{ field_key: "evidence.lifecycle_state", value: "available" }],
  });
}

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
  const recovery = page.getByRole("group", {
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
  await page.route("**/api/v1/object-blobs", async (route) => {
    slotBodies.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.ok()).toBeTruthy();
    if (slotBodies.length === 1) await route.abort("failed");
    else await route.fulfill({ response });
  });
  await page.route("**/api/v1/object-uploads/*", async (route) => {
    transfers++;
    const response = await route.fetch();
    expect(response.status()).toBe(204);
    await route.abort("failed");
  });
  await page.route(`**/views/${evidenceViewSchemaId}/rows`, async (route) => {
    createBodies.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.ok()).toBeTruthy();
    if (createBodies.length === 1) await route.abort("failed");
    else await route.fulfill({ response });
  });
  await page.route(`**/records/${source.record_id}`, async (route) => {
    if (route.request().method() !== "PATCH") return route.continue();
    linkBodies.push(route.request().postData() ?? "");
    const response = await route.fetch();
    expect(response.ok()).toBeTruthy();
    if (linkBodies.length === 1) await route.abort("failed");
    else {
      failRefresh = true;
      await route.fulfill({ response });
    }
  });
  await page.route("**/views/*/query", async (route) => {
    if (failRefresh) {
      failedReads++;
      await route.abort("failed");
    } else await route.continue();
  });
  await openTimelineSurface(page, incident);
  await openTimelineInspector(page, source.record_id);
  const filename = "stage-recovery.png";
  await page
    .getByTestId(timelineEvidenceFileInputTestId(source.record_id))
    .setInputFiles({
      name: filename,
      mimeType: "image/png",
      buffer: tinyPNG(),
    });
  const recovery = page.getByRole("group", {
    name: `File recovery: ${filename}`,
    exact: true,
  });
  await expect(recovery).toContainText("Upload preparation is uncertain");
  await recovery.getByRole("button", { name: "Resume", exact: true }).click();
  await expect(recovery).toContainText("Upload acknowledgement is uncertain");
  await recovery.getByRole("button", { name: "Resume", exact: true }).click();
  await expect(recovery).toContainText("Evidence creation is uncertain");
  await switchFileSurface(page, evidenceViewSchemaId);
  await switchFileSurface(page, timelineViewSchemaId);
  await recovery.getByRole("button", { name: "Resume", exact: true }).click();
  await expect(recovery).toContainText("Review the original Timeline row");
  await reviewFileSource(page, filename);
  await expect(recovery).toContainText("Timeline attachment is uncertain");
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
  page.on("request", (request) => {
    if (request.url().endsWith("/object-blobs")) slots++;
    if (request.method() === "PUT") transfers++;
    if (request.method() === "PATCH") patches++;
  });
  await page.route(
    `**/evidence-records/${source.record_id}/attach-blob`,
    async (route) => {
      bodies.push(route.request().postData() ?? "");
      const response = await route.fetch();
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
  await page
    .getByTestId(evidenceAttachFileInputTestId(source.record_id))
    .setInputFiles({
      name: "existing-empty.txt",
      mimeType: "text/plain",
      buffer: Buffer.from(""),
    });
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

test("keeps ordinary draft typing available during file transfer and shares both creation orderings", async ({
  page,
}, info) => {
  test.setTimeout(120_000);
  for (const ordinaryFirst of [true, false]) {
    const incident = await createIncident(
      page,
      uniqueIncidentKey("EUR-DRAFT-ORDER"),
      "Shared draft creation",
    );
    let release: (() => void) | undefined;
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    const createBodies: string[] = [];
    let slots = 0,
      transfers = 0;
    const heldPath = ordinaryFirst
      ? "**/api/v1/object-uploads/*"
      : `**/views/${timelineViewSchemaId}/rows`;
    await page.route(heldPath, async (route) => {
      await gate;
      await route.continue();
    });
    const watch = (request: import("@playwright/test").Request) => {
      if (request.url().endsWith("/object-blobs")) slots++;
      if (request.method() === "PUT") transfers++;
      if (request.url().endsWith(`/views/${timelineViewSchemaId}/rows`))
        createBodies.push(request.postData() ?? "");
    };
    page.on("request", watch);
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
      page.getByText("Choose one file at a time.", { exact: true }),
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
      await page
        .getByTestId(timelineDraftEvidenceFileInputTestId())
        .setInputFiles({
          name: filename,
          mimeType: "image/png",
          buffer: tinyPNG(),
        });
    }
    const recovery = page.getByRole("group", {
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
    await page.unroute(heldPath);
    page.off("request", watch);
  }
});

test("reviews the original source after a rejected file link while preserving unrelated selection and edits", async ({
  page,
}) => {
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
  page.on("request", (request) => {
    if (request.url().endsWith(`/views/${evidenceViewSchemaId}/rows`))
      creates++;
    if (request.method() === "PUT") transfers++;
  });
  await page.route(`**/records/${source.record_id}`, async (route) => {
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
      const response = await route.fetch();
      expect(response.status()).toBe(200);
      await route.fulfill({ response });
    }
  });
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
  const recovery = page.getByRole("group", {
    name: "File recovery: review-source.png",
    exact: true,
  });
  await expect(recovery).toContainText("Timeline attachment needs recovery");
  await openTimelineInspector(page, other.record_id);
  await recovery
    .getByRole("button", { name: "Review original source", exact: true })
    .click();
  await expect(recovery).toContainText("Reviewed concurrent source edit");
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
