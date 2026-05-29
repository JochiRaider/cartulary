import { Buffer } from "node:buffer";

import {
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  rowCellTestId,
  rowInspectButtonTestId,
  timelineEvidenceFileInputTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createViewRow,
  csrfHeaders,
  patchTimelineRecord,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
  collectionItems,
  evidenceViewSchemaId,
  timelineViewSchemaId,
  type ViewRow,
} from "./phase4Helpers";

test("E-5-01 attaches a screenshot to a selected Timeline row without leaving the workbook surface", async ({
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
      "timeline.summary": "Selected row screenshot",
    },
  )) as unknown as ViewRow;

  await openTimelineSurface(page, incidentId);
  await page.getByTestId(rowInspectButtonTestId(timelineRow.record_id)).click();
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
    .poll(async () => {
      const rows = (await queryViewRows(
        page,
        incidentId,
        timelineViewSchemaId,
      )) as unknown as ViewRow[];
      return rows.find((row) => row.record_id === timelineRow.record_id)?.cells[
        "timeline.evidence_count"
      ]?.value;
    })
    .toBe(1);
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
});

test("E-5-02 persists a screenshot-only Timeline row through the two-step evidence path", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E5DRAFT"),
    "Phase 5 draft screenshot attach",
  );

  await openTimelineSurface(page, incidentId);
  await page.getByTestId("timeline-evidence-file-draft").setInputFiles({
    name: "draft-screenshot.png",
    mimeType: "image/png",
    buffer: tinyPNG(),
  });

  await expect
    .poll(async () => {
      const rows = (await queryViewRows(
        page,
        incidentId,
        timelineViewSchemaId,
      )) as unknown as ViewRow[];
      return rows.find(
        (row) => row.cells["timeline.evidence_count"]?.value === 1,
      );
    })
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
  expect(row?.cells["timeline.summary"]?.value ?? "").toBe("");
  expect(row?.cells["timeline.capture_state"]?.value).toBe("rough");
  await expect(
    page.getByTestId(gridShellTestId(timelineViewSchemaId)),
  ).toBeVisible();
});

test("E-5-03 redeems inline-safe previews and shows explicit blocked-preview outcomes", async ({
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
  await page.getByTestId(evidencePreviewButtonTestId(safe.record_id)).click();
  await expect(
    page.getByTestId(evidencePreviewFrameTestId(safe.record_id)),
  ).toBeVisible();
  await expect(
    page
      .frameLocator(
        `[data-testid="${evidencePreviewFrameTestId(safe.record_id)}"]`,
      )
      .locator("body"),
  ).toContainText("safe preview body");

  await page.getByTestId(evidencePreviewButtonTestId(unsafe.record_id)).click();
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(unsafe.record_id)),
  ).toContainText("unsupported_preview");
});

test("E-5-04 tracks requested evidence before a blob exists and later advances it", async ({
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
      "timeline.summary": "Requested package tracking",
    },
  )) as unknown as ViewRow;

  await openEvidenceSurface(page, incidentId);
  await setGenericCreateField(page, "evidence.title", "Requested package");
  await setGenericCreateField(page, "evidence.storage_ref", "ticket://E5-04");
  await page
    .getByTestId(genericCreateSubmitTestId(evidenceViewSchemaId))
    .click();

  const requested = await waitForEvidenceRow(
    page,
    incidentId,
    "Requested package",
  );
  await expect(
    page.getByTestId(rowCellTestId(requested.record_id, "evidence.title")),
  ).toHaveText("Requested package");
  await expect(
    page.getByTestId(
      rowCellTestId(requested.record_id, "evidence.lifecycle_state"),
    ),
  ).toHaveText("requested");
  await expect(
    page.getByTestId(
      rowCellTestId(requested.record_id, "evidence.upload_state"),
    ),
  ).toHaveText("pending");

  const linkedTimeline = (await patchTimelineRecord(
    page,
    timelineRow.record_id,
    {
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
    },
  )) as unknown as ViewRow;
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
    page.getByTestId(
      rowCellTestId(timelineRow.record_id, "timeline.evidence_count"),
    ),
  ).toHaveText("0");
  await expect(
    page.getByTestId(
      rowCellTestId(timelineRow.record_id, "timeline.has_evidence"),
    ),
  ).toHaveText("false");

  await openEvidenceSurface(page, incidentId);
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
  await expect(
    page.getByTestId(
      rowCellTestId(advanced.record_id, "evidence.lifecycle_state"),
    ),
  ).toHaveText("available");
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
  const timelineRows = (await queryViewRows(
    page,
    incidentId,
    timelineViewSchemaId,
  )) as unknown as ViewRow[];
  expect(
    timelineRows.filter((row) => row.record_id === timelineRow.record_id),
  ).toHaveLength(1);
});

test("E-5-05 refreshes a second live workbook from the real evidence attach stream", async ({
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
      "timeline.summary": "Second workbook evidence refresh",
    },
  )) as unknown as ViewRow;

  const listenerContext = await browser.newContext({
    storageState: await page.context().storageState(),
  });
  const listener = await listenerContext.newPage();
  try {
    const socketMonitor = installPhase5IncidentSocketMonitor(
      listener,
      incidentId,
    );
    await openTimelineSurface(listener, incidentId);
    await socketMonitor.waitForMessage("hello_ack");
    const listenerURL = listener.url();
    await expect(
      listener.getByTestId(
        rowCellTestId(timelineRow.record_id, "timeline.evidence_count"),
      ),
    ).toHaveText("0");

    await openTimelineSurface(page, incidentId);
    await page
      .getByTestId(rowInspectButtonTestId(timelineRow.record_id))
      .click();
    await page
      .getByTestId(timelineEvidenceFileInputTestId(timelineRow.record_id))
      .setInputFiles({
        name: "socket-refresh.png",
        mimeType: "image/png",
        buffer: tinyPNG(),
      });

    await socketMonitor.waitForRecordChanged(timelineRow.record_id);
    await expect(
      listener.getByTestId(
        rowCellTestId(timelineRow.record_id, "timeline.evidence_count"),
      ),
    ).toHaveText("1");
    await expect(
      listener.getByTestId(
        rowCellTestId(timelineRow.record_id, "timeline.has_evidence"),
      ),
    ).toHaveText("true");
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
  await expect(page.getByRole("heading", { name: "Evidence" })).toBeVisible();
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
  await page.getByTestId(genericCreateFieldTestId(fieldKey)).fill(value);
}

async function waitForEvidenceRow(
  page: Page,
  incidentId: string,
  title: string,
): Promise<ViewRow> {
  await expect(page.locator("span", { hasText: title })).toBeVisible();
  const rows = (await queryViewRows(
    page,
    incidentId,
    evidenceViewSchemaId,
  )) as unknown as ViewRow[];
  const row = rows.find(
    (candidate) => candidate.cells["evidence.title"]?.value === title,
  );
  expect(row).toBeTruthy();
  return row as ViewRow;
}

function tinyPNG() {
  return Buffer.from(
    "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=",
    "base64",
  );
}

type Phase5SocketMessage = {
  type: string;
  payload: Record<string, unknown>;
};

function installPhase5IncidentSocketMonitor(page: Page, incidentId: string) {
  const messages: Phase5SocketMessage[] = [];
  const waiters: Array<{
    matches: (message: Phase5SocketMessage) => boolean;
    reject: (error: Error) => void;
    resolve: (message: Phase5SocketMessage) => void;
    timeout: ReturnType<typeof setTimeout>;
  }> = [];

  page.on("websocket", (socket) => {
    if (!socket.url().includes(`/ws/v1/incidents/${incidentId}`)) {
      return;
    }
    socket.on("framereceived", ({ payload }) => {
      const message = parsePhase5SocketPayload(payload);
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
    matches: (message: Phase5SocketMessage) => boolean,
    label: string,
  ) => {
    const existing = messages.find(matches);
    if (existing) {
      return Promise.resolve(existing);
    }
    return new Promise<Phase5SocketMessage>((resolve, reject) => {
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

function parsePhase5SocketPayload(
  payload: string | Buffer,
): Phase5SocketMessage | null {
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
  const upload = await page.request.put(blobData.upload_target.href, {
    data: options.body,
    headers: { "Content-Type": options.contentType },
  });
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
