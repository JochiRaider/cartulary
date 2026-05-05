import { Buffer } from "node:buffer";

import { rowCellTestId } from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  csrfHeaders,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";
import { evidenceViewSchemaId, type ViewRow } from "./phase4Helpers";

test("E-5-01 attaches a screenshot to a selected Timeline row without leaving the workbook surface", async () => {
  // Pending Sprint 5 browser attach implementation.
});

test("E-5-02 persists a screenshot-only Timeline row through the two-step evidence path", async () => {
  // Pending Sprint 5 screenshot-only evidence implementation.
});

test("E-5-03 redeems inline-safe previews and shows explicit blocked-preview outcomes", async () => {
  // Pending Sprint 5 browser preview implementation.
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
