import type { Buffer } from "node:buffer";
import type { Page } from "@playwright/test";

import { expect } from "./fixtures";
import {
  apiBase,
  createViewRow,
  csrfHeaders,
  queryViewRows,
  uniqueTxn,
} from "./helpers";
import { evidenceViewSchemaId, type ViewRow } from "./phase4Helpers";

export type EvidenceUploadOptions = {
  body: Buffer;
  contentType: string;
  filename: string;
  requestedAt: string;
  title: string;
  txnPrefix: string;
};

type EvidenceFixtureOptions = {
  collectorPartyText: string;
  lifecycleState: string;
  requestedAt: string;
  storageRef: string;
  title: string;
  txnPrefix: string;
};

type EvidenceUploadFixtureOptions = EvidenceUploadOptions & {
  collectorPartyText: string;
  txnSuffixes?: {
    attach: string;
    blob: string;
    row: string;
  };
};

export async function createEvidenceFixtureRow(
  page: Page,
  incidentId: string,
  options: EvidenceFixtureOptions,
): Promise<ViewRow> {
  return (await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn(options.txnPrefix),
    "evidence.collector_party_text": options.collectorPartyText,
    "evidence.lifecycle_state": options.lifecycleState,
    "evidence.requested_at": options.requestedAt,
    "evidence.storage_ref": options.storageRef,
    "evidence.title": options.title,
  })) as ViewRow;
}

export async function createUploadedEvidenceFixture(
  page: Page,
  incidentId: string,
  options: EvidenceUploadFixtureOptions,
): Promise<ViewRow> {
  const txnSuffixes = options.txnSuffixes ?? {
    attach: "attach",
    blob: "blob",
    row: "row",
  };
  const row = (await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn(`${options.txnPrefix}-${txnSuffixes.row}`),
    "evidence.collector_party_text": options.collectorPartyText,
    "evidence.requested_at": options.requestedAt,
    "evidence.title": options.title,
  })) as ViewRow;
  const createBlob = await page.request.post(`${apiBase}/api/v1/object-blobs`, {
    headers: await csrfHeaders(page),
    data: {
      byte_size: options.body.byteLength,
      client_txn_id: uniqueTxn(`${options.txnPrefix}-${txnSuffixes.blob}`),
      content_type_hint: options.contentType,
      filename_hint: options.filename,
      incident_id: incidentId,
    },
  });
  expect(createBlob.ok()).toBeTruthy();
  const blobEnvelope = (await createBlob.json()) as {
    data: {
      object_blob_id: string;
      upload_target: {
        href: string;
        method?: string;
      };
    };
  };
  expect(blobEnvelope.data.upload_target.method ?? "PUT").toBe("PUT");

  const upload = await page.request.put(
    resolveAPIHref(blobEnvelope.data.upload_target.href),
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
        base_row_version: row.row_version,
        client_txn_id: uniqueTxn(`${options.txnPrefix}-${txnSuffixes.attach}`),
        object_blob_id: blobEnvelope.data.object_blob_id,
      },
    },
  );
  expect(attach.ok()).toBeTruthy();
  return waitForEvidenceFixtureState(page, incidentId, row.record_id, {
    lifecycleState: "available",
    uploadState: "available",
  });
}

async function waitForEvidenceFixtureState(
  page: Page,
  incidentId: string,
  recordId: string,
  state: { lifecycleState: string; uploadState: string },
): Promise<ViewRow> {
  let matchingRow: ViewRow | null = null;
  await expect
    .poll(
      async () => {
        const rows = (await queryViewRows(
          page,
          incidentId,
          evidenceViewSchemaId,
        )) as ViewRow[];
        matchingRow =
          rows.find((candidate) => candidate.record_id === recordId) ?? null;
        return {
          lifecycleState: cellValue(
            matchingRow?.cells["evidence.lifecycle_state"],
          ),
          uploadState: cellValue(matchingRow?.cells["evidence.upload_state"]),
        };
      },
      { timeout: 30_000 },
    )
    .toEqual(state);
  if (matchingRow === null) {
    throw new Error(`missing evidence row ${recordId}`);
  }
  return matchingRow;
}

function resolveAPIHref(href: string): string {
  return href.startsWith("/") ? `${apiBase}${href}` : href;
}

function cellValue(cell: unknown): unknown {
  if (cell !== null && typeof cell === "object" && "value" in cell) {
    return (cell as { value: unknown }).value;
  }
  return cell;
}
