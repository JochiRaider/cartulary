import type { Buffer } from "node:buffer";
import type {
  AttachBlobToEvidenceRecordRequest,
  EvidenceCreateRequest,
  ViewRow,
} from "@cartulary/protocol-ts/http";
import { evidenceViewSchemaId } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { csrfHeaders } from "../auth/browserSession";

import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";
import { createViewRow, queryViewRows } from "../workbook/query";
import { createAndUploadObjectBlob } from "./uploads";

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
  lifecycleState: NonNullable<
    EvidenceCreateRequest["evidence.lifecycle_state"]
  >;
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
  return await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn(options.txnPrefix),
    "evidence.collector_party_text": options.collectorPartyText,
    "evidence.lifecycle_state": options.lifecycleState,
    "evidence.requested_at": options.requestedAt,
    "evidence.storage_ref": options.storageRef,
    "evidence.title": options.title,
  });
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
  const row = await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn(`${options.txnPrefix}-${txnSuffixes.row}`),
    "evidence.collector_party_text": options.collectorPartyText,
    "evidence.requested_at": options.requestedAt,
    "evidence.title": options.title,
  });
  const blob = await createAndUploadObjectBlob(page, {
    body: options.body,
    clientTxnId: uniqueTxn(`${options.txnPrefix}-${txnSuffixes.blob}`),
    contentType: options.contentType,
    filename: options.filename,
    incidentId,
  });

  const attachBody = {
    base_row_version: row.row_version,
    client_txn_id: uniqueTxn(`${options.txnPrefix}-${txnSuffixes.attach}`),
    object_blob_id: blob.object_blob_id,
  } satisfies AttachBlobToEvidenceRecordRequest;
  const attach = await publicHttpOperation({
    body: attachBody,
    headers: await csrfHeaders(page),
    operationID: "attachBlobToEvidenceRecord",
    pathParameters: { record_id: row.record_id },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!attach.ok) {
    throw new Error(
      `attachBlobToEvidenceRecord failed with HTTP ${attach.status}: ${JSON.stringify(attach.payload)}`,
    );
  }
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
        const rows = await queryViewRows(
          page,
          incidentId,
          evidenceViewSchemaId,
        );
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

function cellValue(cell: unknown): unknown {
  if (cell !== null && typeof cell === "object" && "value" in cell) {
    return (cell as { value: unknown }).value;
  }
  return cell;
}
