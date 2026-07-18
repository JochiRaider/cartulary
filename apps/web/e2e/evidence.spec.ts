import { Buffer } from "node:buffer";
import { scrollGridTargetIntoView } from "@cartulary/test-utils/grid";
import {
  dataTestIdSelector,
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  rowCellTestId,
  timelineDraftEvidenceFileInputTestId,
  timelineEvidenceFileInputTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "./support/contracts/workbookSurfaces";
import { collectionItems, type ViewRow } from "./support/entities/mentions";
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
import { openTimelineInspector } from "./support/workbook/rowMutations";

test.beforeEach(({ page }) => {
  failOnUnexpectedPageError(page);
});

test("attaches a screenshot to a selected Timeline row without leaving the workbook surface", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E5SELECTED"),
    "Phase 5 selected screenshot attach",
  );
  const timelineRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e5-selected-timeline"),
      "timeline.activity_synopsis_text": "Selected row screenshot",
    },
  )) as unknown as ViewRow;
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
        const rows = (await queryViewRows(
          page,
          incidentId,
          timelineViewSchemaId,
        )) as unknown as ViewRow[];
        return rows.find((row) => row.record_id === timelineRow.record_id)
          ?.cells["timeline.evidence_count"]?.value;
      },
      { timeout: 30_000 },
    )
    .toBe(1);
  expect(objectUploadRoutes.length).toBeGreaterThan(0);
});

test("persists a screenshot-only Timeline row through the two-step evidence path", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E5DRAFT"),
    "Phase 5 draft screenshot attach",
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
        const rows = (await queryViewRows(
          page,
          incidentId,
          timelineViewSchemaId,
        )) as unknown as ViewRow[];
        return rows.find(
          (row) => row.cells["timeline.evidence_count"]?.value === 1,
        );
      },
      { timeout: 30_000 },
    )
    .not.toBeUndefined();

  const rows = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as unknown as ViewRow[];
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
    uniqueIncidentKey("E5PREVIEW"),
    "Phase 5 evidence preview",
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
  ).toContainText("unsupported_preview");
});

test("tracks requested evidence before a blob exists and later advances it", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E5EVIDENCE"),
    "Phase 5 requested evidence",
  );
  const timelineRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e5-requested-timeline"),
      "timeline.activity_synopsis_text": "Requested package tracking",
    },
  )) as unknown as ViewRow;

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

  const linkedTimeline = (await patchRecord(page, timelineRow.record_id, {
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
  })) as unknown as ViewRow;
  expect(
    collectionItems(linkedTimeline, "timeline.attached_evidence_ids").some(
      (item) => item.linked_record_id === requested.record_id,
    ),
  ).toBe(true);
  expect(linkedTimeline.cells["timeline.evidence_count"]?.value).toBe(0);
  expect(linkedTimeline.cells["timeline.has_evidence"]?.value).toBe(false);
  await expect
    .poll(async () => {
      const rows = (await queryViewRows(
        page,
        incidentId,
        timelineViewSchemaId,
      )) as unknown as ViewRow[];
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
    targetTestId: evidenceAttachFileInputTestId(requested.record_id),
  });
  await page
    .getByTestId(evidenceAttachFileInputTestId(requested.record_id))
    .setInputFiles({
      name: "requested.txt",
      mimeType: "text/plain",
      buffer: Buffer.from("phase5 requested evidence payload", "utf8"),
    });

  const advanced = await waitForEvidenceRow(
    page,
    incidentId,
    "Requested package",
  );
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: rowCellTestId(advanced.record_id, "evidence.lifecycle_state"),
  });
  await expect(
    page.getByTestId(
      rowCellTestId(advanced.record_id, "evidence.lifecycle_state"),
    ),
  ).toHaveText("available");
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
  expect(
    (await queryViewRows(page, incidentId, evidenceViewSchemaId)).filter(
      (row) => row.record_id === requested.record_id,
    ),
  ).toHaveLength(1);
  await expect
    .poll(async () => {
      const rows = (await queryViewRows(
        page,
        incidentId,
        timelineViewSchemaId,
      )) as unknown as ViewRow[];
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
  const timelineRows = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as unknown as ViewRow[];
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
    uniqueIncidentKey("E5SOCKET"),
    "Phase 5 socket evidence refresh",
  );
  const timelineRow = (await createViewRow(
    page,
    incidentId,
    timelineViewSchemaId,
    {
      client_txn_id: uniqueTxn("e5-socket-timeline"),
      "timeline.activity_synopsis_text": "Second workbook evidence refresh",
    },
  )) as unknown as ViewRow;

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
        const rows = (await queryViewRows(
          listener,
          incidentId,
          timelineViewSchemaId,
        )) as unknown as ViewRow[];
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
    const rows = (await queryViewRows(
      page,
      incidentId,
      evidenceViewSchemaId,
    )) as unknown as ViewRow[];
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

function resolveUploadHref(href: string): string {
  return href.startsWith("/") ? `${apiBase}${href}` : href;
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
  const row = (await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("e5-preview-evidence"),
    "evidence.title": options.title,
    "evidence.collector_party_text": "Browser evidence",
  })) as unknown as ViewRow;
  const createBlob = await page.request.post(`${apiBase}/api/v1/object-blobs`, {
    headers: await csrfHeaders(page),
    data: {
      incident_id: incidentId,
      client_txn_id: uniqueTxn("e5-preview-blob"),
      byte_size: options.body.byteLength,
      filename_hint: options.filename,
      content_type_hint: options.contentType,
    },
  });
  expect(createBlob.ok()).toBeTruthy();
  const blobData = (
    (await createBlob.json()) as {
      data: { object_blob_id: string; upload_target: { href: string } };
    }
  ).data;
  const upload = await page.request.put(
    resolveUploadHref(blobData.upload_target.href),
    {
      data: options.body,
      headers: { "Content-Type": options.contentType },
    },
  );
  expect(upload.ok()).toBeTruthy();

  const attach = await page.request.post(
    `${apiBase}/api/v1/evidence-records/${row.record_id}/attach-blob`,
    {
      headers: await csrfHeaders(page),
      data: {
        object_blob_id: blobData.object_blob_id,
        base_row_version: row.row_version,
        client_txn_id: uniqueTxn("e5-preview-attach"),
      },
    },
  );
  expect(attach.ok()).toBeTruthy();
  const rows = (await queryViewRows(
    page,
    incidentId,
    evidenceViewSchemaId,
  )) as unknown as ViewRow[];
  const attached = rows.find(
    (candidate) => candidate.record_id === row.record_id,
  );
  expect(attached).toBeTruthy();
  return attached as ViewRow;
}
