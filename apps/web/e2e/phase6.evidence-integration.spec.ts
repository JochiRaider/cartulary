import { Buffer } from "node:buffer";

import type {
  EvidenceAttachBlobRequest,
  EvidenceHandleEnvelope,
  EvidenceHandleIssueRequest,
  ObjectBlobCreateEnvelope,
  ObjectBlobCreateRequest,
} from "@cartulary/protocol-ts";
import { scrollGridTargetIntoView } from "@cartulary/test-utils";
import {
  dataTestIdSelector,
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  gridShellTestId,
} from "@cartulary/ui-contracts";
import type { APIRequestContext, Page, Request } from "@playwright/test";

import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createIncidentMemberUser,
  createViewRow,
  csrfHeaders,
  queryViewRows,
  uniqueEmail,
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

test("FE-E-P6-01 Verify evidence attach, preview, download, blocked preview, and authorization denial through same-origin public handles.", async ({
  page,
  browser,
  sessionTracker,
  workerAdminRequest,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("FEEP601"),
    "FE-E-P6-01 evidence handles",
  );
  const safeBody = "FE-E-P6 safe preview body";
  const safeRow = (await createViewRow(page, incidentId, evidenceViewSchemaId, {
    client_txn_id: uniqueTxn("fee-p6-safe-evidence"),
    "evidence.title": "FE-E-P6 safe evidence",
    "evidence.collector_party_text": "Browser evidence",
  })) as unknown as ViewRow;
  const blockedRow = await createUploadedEvidence(page, incidentId, {
    title: "FE-E-P6 blocked preview evidence",
    filename: "fe-e-p6-blocked.html",
    contentType: "text/html",
    body: Buffer.from(
      "<script>window.__fee_p6_blocked = true</script>",
      "utf8",
    ),
  });
  const authRow = await createUploadedEvidence(page, incidentId, {
    title: "FE-E-P6 current authorization evidence",
    filename: "fe-e-p6-current-auth.txt",
    contentType: "text/plain",
    body: Buffer.from("FE-E-P6 current authorization body", "utf8"),
  });
  const observed = collectEvidenceRouteRequests(page);

  await openEvidenceSurface(page, incidentId);
  const workbookURL = page.url();
  await expectStableEvidenceControls(page, safeRow.record_id);
  await expectStableEvidenceActionControls(page, blockedRow.record_id);
  await page
    .getByTestId(evidenceAttachFileInputTestId(safeRow.record_id))
    .setInputFiles({
      name: "fe-e-p6-safe.txt",
      mimeType: "text/plain",
      buffer: Buffer.from(safeBody, "utf8"),
    });

  await expect(
    page.getByTestId(evidenceAccessMessageTestId(safeRow.record_id)),
  ).toHaveText("Evidence attached.");
  const attachedSafeRow = await waitForEvidenceState(
    page,
    incidentId,
    safeRow.record_id,
    {
      lifecycleState: "available",
      uploadState: "available",
    },
  );

  const createBlobRequest = await observed.requirePost(
    (request) => new URL(request.url()).pathname === "/api/v1/object-blobs",
    "FE-E-P6 object blob create request",
  );
  const createBlobBody =
    createBlobRequest.postDataJSON() as ObjectBlobCreateRequest;
  expect(createBlobBody).toEqual({
    incident_id: incidentId,
    client_txn_id: createBlobBody.client_txn_id,
    byte_size: Buffer.byteLength(safeBody),
    filename_hint: "fe-e-p6-safe.txt",
    content_type_hint: "text/plain",
  } satisfies ObjectBlobCreateRequest);
  const createBlobEnvelope =
    await observed.requireJsonResponse<ObjectBlobCreateEnvelope>(
      createBlobRequest,
      "FE-E-P6 object blob create envelope",
    );
  expectSameOriginObjectUploadTarget(
    createBlobEnvelope.data.upload_target.href,
  );
  expect(createBlobEnvelope.data.upload_target.method).toBe("PUT");

  const uploadRequest = await observed.requirePut(
    (request) =>
      new URL(request.url()).pathname.startsWith("/api/v1/object-uploads/"),
    "FE-E-P6 object upload request",
  );
  expect(new URL(uploadRequest.url()).pathname).toBe(
    createBlobEnvelope.data.upload_target.href,
  );

  const attachRequest = await observed.requirePost(
    (request) =>
      new URL(request.url()).pathname ===
      `/api/v1/evidence-records/${safeRow.record_id}/attach-blob`,
    "FE-E-P6 evidence attach request",
  );
  const attachBody = attachRequest.postDataJSON() as EvidenceAttachBlobRequest;
  expect(attachBody).toEqual({
    object_blob_id: createBlobEnvelope.data.object_blob_id,
    base_row_version: safeRow.row_version,
    client_txn_id: attachBody.client_txn_id,
  } satisfies EvidenceAttachBlobRequest);
  await observed.requireJsonResponse(
    attachRequest,
    "FE-E-P6 evidence attach envelope",
  );

  await page
    .getByTestId(evidencePreviewButtonTestId(safeRow.record_id))
    .click();
  const safePreviewFrame = page.getByTestId(
    evidencePreviewFrameTestId(safeRow.record_id),
  );
  await expect(safePreviewFrame).toBeVisible();
  await expect(
    page
      .frameLocator(
        dataTestIdSelector(evidencePreviewFrameTestId(safeRow.record_id)),
      )
      .locator("body"),
  ).toContainText(safeBody);
  const safePreviewSrc = await safePreviewFrame.getAttribute("src");
  expectSameOriginEvidenceHandle(safePreviewSrc ?? "");
  const safePreviewHandleRequest = await observed.requirePost(
    (request) =>
      new URL(request.url()).pathname ===
      `/api/v1/evidence-records/${safeRow.record_id}/preview-handle`,
    "FE-E-P6 preview handle request",
  );
  const safePreviewHandleBody =
    safePreviewHandleRequest.postDataJSON() as EvidenceHandleIssueRequest;
  expect(safePreviewHandleBody).toEqual({});
  const safePreviewHandleEnvelope =
    await observed.requireJsonResponse<EvidenceHandleEnvelope>(
      safePreviewHandleRequest,
      "FE-E-P6 preview handle envelope",
    );
  expectSameOriginEvidenceHandle(safePreviewHandleEnvelope.data.href);
  expect(safePreviewHandleEnvelope.data.handle_kind).toBe("preview");
  expect(safePreviewHandleEnvelope.data.method).toBe("GET");
  await expectActiveEvidenceSurface(page, workbookURL);

  const downloadPromise = page.waitForEvent("download");
  await page
    .getByTestId(evidenceDownloadButtonTestId(safeRow.record_id))
    .click();
  const download = await downloadPromise;
  expect(download.suggestedFilename()).toBe("fe-e-p6-safe.txt");
  const safeDownloadHandleRequest = await observed.requirePost(
    (request) =>
      new URL(request.url()).pathname ===
      `/api/v1/evidence-records/${safeRow.record_id}/download-handle`,
    "FE-E-P6 download handle request",
  );
  const safeDownloadHandleBody =
    safeDownloadHandleRequest.postDataJSON() as EvidenceHandleIssueRequest;
  expect(safeDownloadHandleBody).toEqual({});
  const safeDownloadHandleEnvelope =
    await observed.requireJsonResponse<EvidenceHandleEnvelope>(
      safeDownloadHandleRequest,
      "FE-E-P6 download handle envelope",
    );
  expectSameOriginEvidenceHandle(safeDownloadHandleEnvelope.data.href);
  expect(safeDownloadHandleEnvelope.data.handle_kind).toBe("download");
  expect(safeDownloadHandleEnvelope.data.method).toBe("GET");
  expectSameOriginEvidenceHandle(download.url());
  expect(new URL(download.url()).pathname).toBe(
    new URL(safeDownloadHandleEnvelope.data.href, webBase).pathname,
  );
  await expectActiveEvidenceSurface(page, workbookURL);

  await page
    .getByTestId(evidencePreviewButtonTestId(blockedRow.record_id))
    .click();
  await expect(
    page.getByTestId(evidenceAccessMessageTestId(blockedRow.record_id)),
  ).toContainText("evidence_access_unavailable: unsupported_preview");
  await expect(
    page.getByTestId(evidencePreviewFrameTestId(blockedRow.record_id)),
  ).toHaveCount(0);
  const blockedPreviewRequest = await observed.requirePost(
    (request) =>
      new URL(request.url()).pathname ===
      `/api/v1/evidence-records/${blockedRow.record_id}/preview-handle`,
    "FE-E-P6 blocked preview handle request",
  );
  const blockedPreviewEnvelope = await observed.requireJsonErrorResponse(
    blockedPreviewRequest,
    "FE-E-P6 blocked preview public error envelope",
    409,
    "evidence_access_unavailable",
  );
  expect(blockedPreviewEnvelope.error.details.reason_code).toBe(
    "unsupported_preview",
  );
  await expectActiveEvidenceSurface(page, workbookURL);

  const memberPassword = "MemberEvidence1!";
  const member = await createIncidentMemberUser(page, incidentId, {
    email: uniqueEmail("fe-e-p6-member"),
    display_name: "FE-E-P6 member",
    initial_password: memberPassword,
    role: "editor",
  });
  const memberContext = await browser.newContext();
  const memberPage = await memberContext.newPage();
  failOnUnexpectedPageError(memberPage);
  try {
    await sessionTracker.loginTrackedUser(memberPage, {
      createdBy: "FE-E-P6-01",
      email: member.email,
      password: memberPassword,
      purpose: "current evidence handle authorization denial",
      userId: member.user_id,
    });
    const memberObserved = collectEvidenceRouteRequests(memberPage);
    await openEvidenceSurface(memberPage, incidentId);
    const memberWorkbookURL = memberPage.url();
    await expectStableEvidenceActionControls(memberPage, authRow.record_id);
    await memberPage
      .getByTestId(evidencePreviewButtonTestId(authRow.record_id))
      .click();
    const authPreviewFrame = memberPage.getByTestId(
      evidencePreviewFrameTestId(authRow.record_id),
    );
    await expect(authPreviewFrame).toBeVisible();
    await expect(
      memberPage
        .frameLocator(
          dataTestIdSelector(evidencePreviewFrameTestId(authRow.record_id)),
        )
        .locator("body"),
    ).toContainText("FE-E-P6 current authorization body");
    const authPreviewHandleRequest = await memberObserved.requirePost(
      (request) =>
        new URL(request.url()).pathname ===
        `/api/v1/evidence-records/${authRow.record_id}/preview-handle`,
      "FE-E-P6 member preview handle request",
    );
    const authPreviewHandleEnvelope =
      await memberObserved.requireJsonResponse<EvidenceHandleEnvelope>(
        authPreviewHandleRequest,
        "FE-E-P6 member preview handle envelope",
      );
    const currentAuthHref = authPreviewHandleEnvelope.data.href;
    expectSameOriginEvidenceHandle(currentAuthHref);
    await expectActiveEvidenceSurface(memberPage, memberWorkbookURL);

    const memberMembership = await loadIncidentMembership(
      workerAdminRequest,
      incidentId,
      member.user_id,
    );
    await deleteIncidentMembership(
      workerAdminRequest,
      incidentId,
      member.user_id,
      memberMembership.membership_version,
    );

    const redemptionDenied = await fetchPublicJSONFromPage(memberPage, {
      href: currentAuthHref,
      method: "GET",
    });
    expectSameOriginEvidenceHandle(redemptionDenied.url);
    expectPublicErrorEnvelope(
      redemptionDenied,
      404,
      "handle_not_found_or_revoked",
    );

    const issuanceDenied = await fetchPublicJSONFromPage(memberPage, {
      href: `/api/v1/evidence-records/${authRow.record_id}/preview-handle`,
      method: "POST",
      data: {},
    });
    expectSameOriginPublicRouteURL(
      issuanceDenied.url,
      new RegExp(
        `^/api/v1/evidence-records/${escapeRegExp(
          authRow.record_id,
        )}/preview-handle$`,
        "u",
      ),
    );
    expectPublicErrorEnvelope(issuanceDenied, 404, "evidence_record_not_found");
    await expectActiveEvidenceSurface(memberPage, memberWorkbookURL);
    await expectNoRawStorageDetails(memberPage, [
      currentAuthHref,
      redemptionDenied.url,
      redemptionDenied.bodyText,
      issuanceDenied.url,
      issuanceDenied.bodyText,
      ...memberObserved.requests().map((request) => request.url()),
    ]);
  } finally {
    await memberContext.close();
  }

  const handlePaths = observed
    .evidenceHandleRequests()
    .map((request) => new URL(request.url()).pathname);
  expect(handlePaths).toContain(
    new URL(safePreviewSrc ?? "", webBase).pathname,
  );
  for (const path of handlePaths) {
    expectSameOriginEvidenceHandle(path);
  }
  await expectNoRawStorageDetails(page, [
    await page.locator("body").innerText(),
    await page
      .getByTestId(evidenceAccessMessageTestId(safeRow.record_id))
      .innerText(),
    await page
      .getByTestId(evidenceAccessMessageTestId(blockedRow.record_id))
      .innerText(),
    createBlobEnvelope.data.upload_target.href,
    safePreviewSrc ?? "",
    download.url(),
    safePreviewHandleEnvelope.data.href,
    safeDownloadHandleEnvelope.data.href,
    JSON.stringify(blockedPreviewEnvelope),
    ...observed.requests().map((request) => request.url()),
  ]);
  expect(attachedSafeRow.cells["evidence.lifecycle_state"]?.value).toBe(
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
  await expectStableEvidenceActionControls(page, recordId);
  const testId = evidenceAccessMessageTestId(recordId);
  await scrollGridTargetIntoView({
    page,
    surface: evidenceViewSchemaId,
    targetTestId: testId,
  });
  await expect(page.getByTestId(testId)).toHaveAttribute("data-testid", testId);
}

async function expectStableEvidenceActionControls(
  page: Page,
  recordId: string,
) {
  for (const testId of [
    evidenceAttachFileInputTestId(recordId),
    evidencePreviewButtonTestId(recordId),
    evidenceDownloadButtonTestId(recordId),
  ]) {
    await scrollGridTargetIntoView({
      page,
      surface: evidenceViewSchemaId,
      targetTestId: testId,
    });
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
          uploadState:
            matchingRow?.cells["evidence.upload_state"]?.value ?? null,
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
    requireJsonErrorResponse: async (
      request: Request,
      label: string,
      status: number,
      code: string,
    ) => {
      const response = await request.response();
      if (!response) {
        throw new Error(`missing ${label} response`);
      }
      expect(response.status(), `${label} should use public error status`).toBe(
        status,
      );
      const envelope = (await response.json()) as PublicErrorEnvelope;
      expectPublicErrorEnvelopePayload(envelope, status, code);
      return envelope;
    },
    requirePost: (predicate: (request: Request) => boolean, label: string) =>
      requireObservedRequest(seenRequests, "POST", predicate, label),
    requirePut: (predicate: (request: Request) => boolean, label: string) =>
      requireObservedRequest(seenRequests, "PUT", predicate, label),
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
    client_txn_id: uniqueTxn("fee-p6-uploaded-evidence"),
    "evidence.title": options.title,
    "evidence.collector_party_text": "Browser evidence",
  })) as unknown as ViewRow;
  const createBlob = await page.request.post(`${apiBase}/api/v1/object-blobs`, {
    headers: await csrfHeaders(page),
    data: {
      incident_id: incidentId,
      client_txn_id: uniqueTxn("fee-p6-uploaded-blob"),
      byte_size: options.body.byteLength,
      filename_hint: options.filename,
      content_type_hint: options.contentType,
    } satisfies ObjectBlobCreateRequest,
  });
  expect(createBlob.ok()).toBeTruthy();
  const blobEnvelope = (await createBlob.json()) as ObjectBlobCreateEnvelope;
  expectSameOriginObjectUploadTarget(blobEnvelope.data.upload_target.href);
  expect(blobEnvelope.data.upload_target.method).toBe("PUT");

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
        object_blob_id: blobEnvelope.data.object_blob_id,
        base_row_version: row.row_version,
        client_txn_id: uniqueTxn("fee-p6-uploaded-attach"),
      } satisfies EvidenceAttachBlobRequest,
    },
  );
  expect(attach.ok()).toBeTruthy();
  return waitForEvidenceState(page, incidentId, row.record_id, {
    lifecycleState: "available",
    uploadState: "available",
  });
}

async function expectActiveEvidenceSurface(page: Page, expectedURL: string) {
  expect(page.url()).toBe(expectedURL);
  await expect(
    page.getByTestId(gridShellTestId(evidenceViewSchemaId)),
  ).toBeVisible();
}

function resolveAPIHref(href: string): string {
  return href.startsWith("/") ? `${apiBase}${href}` : href;
}

function expectSameOriginEvidenceHandle(href: string) {
  const parsed = new URL(href, webBase);
  expect(parsed.origin).toBe(new URL(webBase).origin);
  expect(parsed.pathname).toMatch(/^\/api\/v1\/evidence-handles\/[^/]+$/u);
  expectNoRawStorageDetailsInText(href);
}

function expectSameOriginObjectUploadTarget(href: string) {
  const parsed = new URL(href, webBase);
  expect(parsed.origin).toBe(new URL(webBase).origin);
  expect(parsed.pathname).toMatch(/^\/api\/v1\/object-uploads\/[^/]+$/u);
  expectNoRawStorageDetailsInText(href);
}

function expectSameOriginPublicRouteURL(href: string, pathPattern: RegExp) {
  const parsed = new URL(href, webBase);
  expect(parsed.origin).toBe(new URL(webBase).origin);
  expect(parsed.pathname).toMatch(pathPattern);
  expectNoRawStorageDetailsInText(href);
}

type IncidentMembershipRecord = {
  membership_version: number;
  role: string;
  user_id: string;
};

async function loadIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  userId: string,
) {
  const response = await authRequests.get(
    `/api/v1/incidents/${incidentId}/memberships`,
  );
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as {
    data: { memberships: IncidentMembershipRecord[] };
  };
  const membership =
    body.data.memberships.find((candidate) => candidate.user_id === userId) ??
    null;
  if (membership === null) {
    throw new Error(`missing incident membership for ${userId}`);
  }
  return membership;
}

async function deleteIncidentMembership(
  authRequests: APIRequestContext,
  incidentId: string,
  userId: string,
  baseMembershipVersion: number,
) {
  const response = await authRequests.delete(
    `/api/v1/incidents/${incidentId}/memberships/${userId}`,
    {
      data: {
        base_membership_version: baseMembershipVersion,
      },
    },
  );
  expect(response.status()).toBe(204);
}

type PublicBrowserFetchResult = {
  bodyText: string;
  json: unknown;
  status: number;
  url: string;
};

async function fetchPublicJSONFromPage(
  page: Page,
  options: {
    data?: unknown;
    href: string;
    method: "GET" | "POST";
  },
): Promise<PublicBrowserFetchResult> {
  return page.evaluate(async ({ data, href, method }) => {
    const csrfCookie =
      document.cookie
        .split("; ")
        .find((cookie) => cookie.startsWith("cartulary_csrf="))
        ?.split("=")[1] ?? "";
    const headers: Record<string, string> = {
      Accept: "application/json",
    };
    if (method !== "GET") {
      headers["Content-Type"] = "application/json";
      headers["X-CSRF-Token"] = decodeURIComponent(csrfCookie);
    }
    const requestInit: RequestInit = {
      credentials: "same-origin",
      headers,
      method,
    };
    if (data !== undefined) {
      requestInit.body = JSON.stringify(data);
    }
    const response = await fetch(href, requestInit);
    const bodyText = await response.text();
    let json: unknown = null;
    try {
      json = bodyText === "" ? null : JSON.parse(bodyText);
    } catch {
      json = null;
    }
    return {
      bodyText,
      json,
      status: response.status,
      url: response.url,
    };
  }, options);
}

type PublicErrorEnvelope = {
  error: {
    code: string;
    details: Record<string, unknown>;
    request_id: string;
    retryable: boolean;
    status: number;
  };
};

function expectPublicErrorEnvelope(
  result: PublicBrowserFetchResult,
  status: number,
  code: string,
) {
  expect(result.status).toBe(status);
  expectPublicErrorEnvelopePayload(
    result.json as PublicErrorEnvelope,
    status,
    code,
  );
}

function expectPublicErrorEnvelopePayload(
  envelope: PublicErrorEnvelope,
  status: number,
  code: string,
) {
  expect(envelope.error.status).toBe(status);
  expect(envelope.error.code).toBe(code);
  expect(envelope.error.request_id).toEqual(expect.any(String));
  expect(envelope.error.retryable).toBe(false);
  expect(envelope.error.details).toEqual(expect.any(Object));
}

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");
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
