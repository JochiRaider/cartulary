import { Buffer } from "node:buffer";

import type {
  EvidenceAttachBlobRequest,
  EvidenceHandleEnvelope,
  EvidenceHandleIssueRequest,
  ObjectBlobCreateEnvelope,
  ObjectBlobCreateRequest,
} from "@cartulary/protocol-ts";
import {
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  gridShellTestId,
} from "@cartulary/ui-contracts";
import type { Page, Request } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  createIncident,
  createViewRow,
  queryViewRows,
  uniqueIncidentKey,
  uniqueTxn,
  webBase,
} from "./helpers";
import { evidenceViewSchemaId, type ViewRow } from "./phase4Helpers";

test.beforeEach(({ page }) => {
  failOnUnexpectedPageError(page);
});

test("FE-I-P6-01 Verify attach flow uses generated protocol types, public error envelopes, and stable evidence selectors without raw object URLs or paths.", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEIP601"),
    "FE-I-P6-01 evidence integration",
  );
  const evidenceRow = (await createViewRow(
    page,
    incidentId,
    evidenceViewSchemaId,
    {
      client_txn_id: uniqueTxn("fei-p6-evidence"),
      "evidence.title": "FE-I-P6 attach target",
      "evidence.collector_party_text": "Browser evidence",
    },
  )) as unknown as ViewRow;
  const observed = collectEvidenceRouteRequests(page);

  await openEvidenceSurface(page, incidentId);
  await expectStableEvidenceControls(page, evidenceRow.record_id);
  await page
    .getByTestId(evidenceAttachFileInputTestId(evidenceRow.record_id))
    .setInputFiles({
      name: "fe-i-p6-evidence.txt",
      mimeType: "text/plain",
      buffer: Buffer.from("FE-I-P6 evidence body", "utf8"),
    });

  await expect(
    page.getByTestId(evidenceAccessMessageTestId(evidenceRow.record_id)),
  ).toHaveText("Evidence attached.");
  const attachedRow = await waitForEvidenceState(
    page,
    incidentId,
    evidenceRow.record_id,
    {
      lifecycleState: "available",
      uploadState: "available",
    },
  );

  const createBlobRequest = await observed.requirePost(
    (request) => new URL(request.url()).pathname === "/api/v1/object-blobs",
    "object blob create request",
  );
  const createBlobBody =
    createBlobRequest.postDataJSON() as ObjectBlobCreateRequest;
  const expectedCreateBlobBody = {
    incident_id: incidentId,
    client_txn_id: createBlobBody.client_txn_id,
    byte_size: Buffer.byteLength("FE-I-P6 evidence body"),
    filename_hint: "fe-i-p6-evidence.txt",
    content_type_hint: "text/plain",
  } satisfies ObjectBlobCreateRequest;
  expect(createBlobBody).toEqual(expectedCreateBlobBody);
  expect(Object.keys(createBlobBody).sort()).toEqual([
    "byte_size",
    "client_txn_id",
    "content_type_hint",
    "filename_hint",
    "incident_id",
  ]);

  const createBlobEnvelope =
    await observed.requireJsonResponse<ObjectBlobCreateEnvelope>(
      createBlobRequest,
      "object blob create envelope",
    );
  expect(createBlobEnvelope.data.upload_target.href).toMatch(
    /^\/api\/v1\/object-uploads\//u,
  );
  expect(createBlobEnvelope.data.upload_target.method).toBe("PUT");

  const uploadRequest = await observed.requirePut(
    (request) =>
      new URL(request.url()).pathname.startsWith("/api/v1/object-uploads/"),
    "object upload request",
  );
  expect(new URL(uploadRequest.url()).pathname).toBe(
    createBlobEnvelope.data.upload_target.href,
  );

  const attachRequest = await observed.requirePost(
    (request) =>
      new URL(request.url()).pathname ===
      `/api/v1/evidence-records/${evidenceRow.record_id}/attach-blob`,
    "evidence attach request",
  );
  const attachBody = attachRequest.postDataJSON() as EvidenceAttachBlobRequest;
  const expectedAttachBody = {
    object_blob_id: createBlobEnvelope.data.object_blob_id,
    base_row_version: evidenceRow.row_version,
    client_txn_id: attachBody.client_txn_id,
  } satisfies EvidenceAttachBlobRequest;
  expect(attachBody).toEqual(expectedAttachBody);
  expect(Object.keys(attachBody).sort()).toEqual([
    "base_row_version",
    "client_txn_id",
    "object_blob_id",
  ]);

  await page
    .getByTestId(evidencePreviewButtonTestId(evidenceRow.record_id))
    .click();
  const previewFrame = page.getByTestId(
    evidencePreviewFrameTestId(evidenceRow.record_id),
  );
  await expect(previewFrame).toBeVisible();
  const previewSrc = await previewFrame.getAttribute("src");
  expectSameOriginEvidenceHandle(previewSrc ?? "");

  const previewHandleRequest = await observed.requirePost(
    (request) =>
      new URL(request.url()).pathname ===
      `/api/v1/evidence-records/${evidenceRow.record_id}/preview-handle`,
    "preview handle request",
  );
  const previewHandleBody =
    previewHandleRequest.postDataJSON() as EvidenceHandleIssueRequest;
  expect(previewHandleBody).toEqual({});
  const previewHandleEnvelope =
    await observed.requireJsonResponse<EvidenceHandleEnvelope>(
      previewHandleRequest,
      "preview handle envelope",
    );
  expectSameOriginEvidenceHandle(previewHandleEnvelope.data.href);
  expect(previewHandleEnvelope.data.handle_kind).toBe("preview");
  expect(previewHandleEnvelope.data.method).toBe("GET");

  const downloadPromise = page.waitForEvent("download");
  await page
    .getByTestId(evidenceDownloadButtonTestId(evidenceRow.record_id))
    .click();
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toBe("fe-i-p6-evidence.txt");
  const downloadHandleRequest = await observed.requirePost(
    (request) =>
      new URL(request.url()).pathname ===
      `/api/v1/evidence-records/${evidenceRow.record_id}/download-handle`,
    "download handle request",
  );
  const downloadHandleBody =
    downloadHandleRequest.postDataJSON() as EvidenceHandleIssueRequest;
  expect(downloadHandleBody).toEqual({});
  const downloadHandleEnvelope =
    await observed.requireJsonResponse<EvidenceHandleEnvelope>(
      downloadHandleRequest,
      "download handle envelope",
    );
  expectSameOriginEvidenceHandle(downloadHandleEnvelope.data.href);
  expect(downloadHandleEnvelope.data.handle_kind).toBe("download");
  expect(downloadHandleEnvelope.data.method).toBe("GET");

  const handlePaths = observed
    .evidenceHandleRequests()
    .map((request) => new URL(request.url()).pathname);
  expect(handlePaths).toContain(new URL(previewSrc ?? "", webBase).pathname);
  expectSameOriginEvidenceHandle(download.url());
  for (const path of handlePaths) {
    expectSameOriginEvidenceHandle(path);
  }
  await expectNoRawStorageDetails(page, [
    await page.locator("body").innerText(),
    await page
      .getByTestId(evidenceAccessMessageTestId(evidenceRow.record_id))
      .innerText(),
    previewSrc ?? "",
    download.url(),
    previewHandleEnvelope.data.href,
    downloadHandleEnvelope.data.href,
    ...observed.requests().map((request) => request.url()),
  ]);
  expect(attachedRow.cells["evidence.lifecycle_state"]?.value).toBe(
    "available",
  );
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

async function expectStableEvidenceControls(page: Page, recordId: string) {
  for (const testId of [
    evidenceAttachFileInputTestId(recordId),
    evidencePreviewButtonTestId(recordId),
    evidenceDownloadButtonTestId(recordId),
    evidenceAccessMessageTestId(recordId),
  ]) {
    await expect(page.getByTestId(testId)).toHaveAttribute(
      "data-testid",
      testId,
    );
  }
}

async function waitForEvidenceState(
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
        )) as unknown as ViewRow[];
        matchingRow =
          rows.find((candidate) => candidate.record_id === recordId) ?? null;
        return {
          lifecycleState:
            matchingRow?.cells["evidence.lifecycle_state"]?.value ?? null,
          uploadState: matchingRow?.cells["evidence.upload_state"]?.value ?? null,
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

function collectEvidenceRouteRequests(page: Page) {
  const seenRequests: Request[] = [];
  page.on("request", (request) => {
    const url = new URL(request.url());
    if (!url.pathname.startsWith("/api/v1/")) {
      return;
    }
    if (
      url.pathname === "/api/v1/object-blobs" ||
      url.pathname.startsWith("/api/v1/object-uploads/") ||
      url.pathname.startsWith("/api/v1/evidence-records/") ||
      url.pathname.startsWith("/api/v1/evidence-handles/")
    ) {
      seenRequests.push(request);
    }
  });

  return {
    evidenceHandleRequests: () =>
      seenRequests.filter(
        (request) =>
          request.method() === "GET" &&
          new URL(request.url()).pathname.startsWith(
            "/api/v1/evidence-handles/",
          ),
      ),
    requireJsonResponse: async <T>(request: Request, label: string) => {
      const response = await request.response();
      if (!response) {
        throw new Error(`missing ${label} response`);
      }
      expect(response.ok(), `${label} should be public success envelope`).toBe(
        true,
      );
      return (await response.json()) as T;
    },
    requirePost: (
      predicate: (request: Request) => boolean,
      label: string,
    ) => requireObservedRequest(seenRequests, "POST", predicate, label),
    requirePut: (
      predicate: (request: Request) => boolean,
      label: string,
    ) => requireObservedRequest(seenRequests, "PUT", predicate, label),
    requests: () => [...seenRequests],
  };
}

async function requireObservedRequest(
  requests: Request[],
  method: string,
  predicate: (request: Request) => boolean,
  label: string,
): Promise<Request> {
  await expect
    .poll(() =>
      requests.some(
        (request) => request.method() === method && predicate(request),
      ),
    )
    .toBe(true);
  const request = requests.find(
    (candidate) => candidate.method() === method && predicate(candidate),
  );
  if (!request) {
    throw new Error(`missing ${label}`);
  }
  return request;
}

function expectSameOriginEvidenceHandle(href: string) {
  const parsed = new URL(href, webBase);
  expect(parsed.origin).toBe(new URL(webBase).origin);
  expect(parsed.pathname).toMatch(/^\/api\/v1\/evidence-handles\/[^/]+$/u);
  expectNoRawStorageDetailsInText(href);
}

const rawStorageLeakPatterns = [
  /https?:\/\/(?:minio|s3|seaweedfs|object-store)[^\s"')]+/iu,
  /\bs3:\/\//iu,
  /\bobject:\/\//iu,
  /\bminio\b/iu,
  /\bseaweedfs?\b/iu,
  /\bobject[-_ ]store(?:\b|_)/iu,
  /\bstorage[-_ ]backend(?:\b|_)/iu,
  /\bbucket(?:\b|_)/iu,
  /\bobject[-_ ](?:key|blob[-_ ]storage[-_ ]key)(?:\b|_)/iu,
  /\/(?:var|srv|mnt|data|tmp|home|app|workspace)\//iu,
] as const;

async function expectNoRawStorageDetails(page: Page, values: string[]) {
  for (const value of values) {
    expectNoRawStorageDetailsInText(value);
  }
  await expect(page.locator("body")).not.toContainText(
    /minio|seaweedfs|bucket|object[-_ ]store|storage[-_ ]backend|object[-_ ]key|\/var\//iu,
  );
}

function expectNoRawStorageDetailsInText(value: string) {
  for (const pattern of rawStorageLeakPatterns) {
    expect(value).not.toMatch(pattern);
  }
}

function failOnUnexpectedPageError(page: Page) {
  page.on("pageerror", (error) => {
    throw new Error(`Unexpected page error: ${error.stack ?? error.message}`);
  });
}
