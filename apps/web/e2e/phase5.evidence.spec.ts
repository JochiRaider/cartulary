import { Buffer } from "node:buffer";

import {
  gridShellTestId,
  rowCellTestId,
  rowInspectButtonTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createViewRow,
  csrfHeaders,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import {
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
    .getByTestId(`timeline-evidence-file-${timelineRow.record_id}`)
    .setInputFiles({
      name: "selected-screenshot.png",
      mimeType: "image/png",
      buffer: tinyPNG(),
    });

  await expect(page.getByTestId(gridShellTestId("timeline"))).toBeVisible();
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
  await expect(page.getByTestId(gridShellTestId("timeline"))).toBeVisible();
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
  await page.getByTestId(`evidence-preview-${safe.record_id}`).click();
  await expect(
    page.getByTestId(`evidence-preview-frame-${safe.record_id}`),
  ).toBeVisible();
  await expect(
    page
      .frameLocator(`[data-testid="evidence-preview-frame-${safe.record_id}"]`)
      .locator("body"),
  ).toContainText("safe preview body");

  await page.getByTestId(`evidence-preview-${unsafe.record_id}`).click();
  await expect(
    page.getByTestId(`evidence-access-message-${unsafe.record_id}`),
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

  await openEvidenceSurface(page, incidentId);
  await setGenericCreateField(page, "evidence.title", "Requested package");
  await setGenericCreateField(page, "evidence.storage_ref", "ticket://E5-04");
  await page
    .getByTestId(`generic-create-submit-${evidenceViewSchemaId}`)
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

  const payload = Buffer.from("phase5 requested evidence payload", "utf8");
  const digest = await crypto.subtle.digest("SHA-256", payload);
  const sha256Hex = [...new Uint8Array(digest)]
    .map((value) => value.toString(16).padStart(2, "0"))
    .join("");
  const createBlob = await page.request.post(`${apiBase}/api/v1/object-blobs`, {
    headers: await csrfHeaders(page),
    data: {
      incident_id: incidentId,
      client_txn_id: uniqueTxn("e5-blob"),
      byte_size: payload.byteLength,
      filename_hint: "requested.txt",
      content_type_hint: "text/plain",
      sha256_hex: sha256Hex,
    },
  });
  expect(createBlob.ok()).toBeTruthy();
  const blobData = (
    (await createBlob.json()) as {
      data: { object_blob_id: string; upload_target: { href: string } };
    }
  ).data;
  const uploadTarget = blobData.upload_target;
  const upload = await page.request.put(uploadTarget.href, {
    data: payload,
    headers: { "Content-Type": "text/plain" },
  });
  expect(upload.ok()).toBeTruthy();

  const attach = await page.request.post(
    `${apiBase}/api/v1/evidence-records/${requested.record_id}/attach-blob`,
    {
      headers: await csrfHeaders(page),
      data: {
        object_blob_id: blobData.object_blob_id,
        base_row_version: requested.row_version,
        client_txn_id: uniqueTxn("e5-attach"),
      },
    },
  );
  if (!attach.ok()) {
    throw new Error(`attach failed ${attach.status()}: ${await attach.text()}`);
  }

  await openEvidenceSurface(page, incidentId);
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
  await expect(page.getByTestId(gridShellTestId("timeline"))).toBeVisible();
}

async function setGenericCreateField(
  page: Page,
  fieldKey: string,
  value: string,
) {
  await page.getByTestId(`generic-create-field-${fieldKey}`).fill(value);
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
