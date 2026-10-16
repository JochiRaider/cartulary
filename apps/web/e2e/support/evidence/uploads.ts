import type { Buffer } from "node:buffer";

import type {
  CreateObjectBlobSlotRequest,
  CreateObjectBlobSlotResponse,
} from "@cartulary/protocol-ts/http";
import type { Locator, Page } from "@playwright/test";

import { csrfHeaders } from "../auth/browserSession";
import { apiBase } from "../runtime/configuration";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";

type ObjectUploadTarget = CreateObjectBlobSlotResponse["data"]["upload_target"];

type ObjectBlobUploadOptions = {
  readonly body: Buffer;
  readonly clientTxnId: string;
  readonly contentType: string;
  readonly filename: string;
  readonly incidentId: string;
};

type ResolvedObjectUploadTarget = {
  readonly headers: Readonly<Record<string, string>>;
  readonly href: string;
};

export function resolveObjectUploadTarget(
  uploadTarget: ObjectUploadTarget,
): ResolvedObjectUploadTarget {
  if (uploadTarget.method !== "PUT") {
    throw new Error("invalid_public_contract_response");
  }
  if (!/^\/api\/v1\/object-uploads\/[^/?#]+$/u.test(uploadTarget.href)) {
    throw new Error("invalid_public_contract_response");
  }
  const origin = new URL(apiBase);
  const resolved = new URL(uploadTarget.href, origin);
  if (
    resolved.origin !== origin.origin ||
    resolved.pathname !== uploadTarget.href ||
    resolved.search !== "" ||
    resolved.hash !== ""
  ) {
    throw new Error("invalid_public_contract_response");
  }
  return Object.freeze({
    headers: Object.freeze({ ...uploadTarget.headers }),
    href: `${origin.origin}${uploadTarget.href}`,
  });
}

export async function createAndUploadObjectBlob(
  page: Page,
  options: ObjectBlobUploadOptions,
): Promise<CreateObjectBlobSlotResponse["data"]> {
  const requestBody = {
    byte_size: options.body.byteLength,
    client_txn_id: options.clientTxnId,
    content_type_hint: options.contentType,
    filename_hint: options.filename,
    incident_id: options.incidentId,
  } satisfies CreateObjectBlobSlotRequest;
  const createResult = await publicHttpOperation({
    body: requestBody,
    headers: await csrfHeaders(page),
    operationID: "createObjectBlobSlot",
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!createResult.ok) {
    throw new Error(
      `create object blob slot failed with HTTP ${createResult.status}`,
    );
  }
  const blob = createResult.payload.data;
  const accepted = blob.accepted_contract;
  if (
    blob.incident_id !== requestBody.incident_id ||
    accepted.incident_id !== requestBody.incident_id ||
    accepted.byte_size !== requestBody.byte_size ||
    accepted.filename_hint !== requestBody.filename_hint ||
    accepted.content_type_hint !== requestBody.content_type_hint
  ) {
    throw new Error("invalid_public_contract_response");
  }
  const uploadTarget = resolveObjectUploadTarget(blob.upload_target);
  const uploadHeaders = {
    ...uploadTarget.headers,
    ...(await csrfHeaders(page)),
  };
  const upload = await page.request.put(uploadTarget.href, {
    data: options.body,
    headers: uploadHeaders,
  });
  if (!upload.ok()) {
    throw new Error(`object blob upload failed with HTTP ${upload.status()}`);
  }
  return blob;
}

/** Exercise chooser invocation before completion; direct change skips source capture. */
export async function chooseEvidenceFile(
  input: Locator,
  files: Parameters<Locator["setInputFiles"]>[0],
) {
  const page = input.page();
  const label = await input.getAttribute("aria-label");
  if (!label) throw new Error("Evidence picker must name its invoking action");
  const chooser = page.waitForEvent("filechooser");
  await page.getByRole("button", { name: label, exact: true }).first().click();
  await (await chooser).setFiles(files);
}

/** Recovery is local and explicitly disclosed; opening it is a user action. */
export async function openTimelineAttachmentFeedback(page: Page) {
  const trigger = page.getByRole("button", { name: /^Attachments:/ });
  await trigger.waitFor({ state: "visible" });
  if ((await trigger.getAttribute("aria-expanded")) === "false")
    await trigger.click();
}

/** Verify through the real Job owner before a fixture finalizes stored bytes. */
export async function verifyUploadedObjectBlob(
  page: Page,
  blobId: string,
  clientTxnId: string,
) {
  const admitted = await publicHttpOperation({
    operationID: "verifyObjectBlob",
    pathParameters: { object_blob_id: blobId },
    body: { client_txn_id: clientTxnId },
    headers: await csrfHeaders(page),
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!admitted.ok)
    throw new Error(`verification admission failed: ${admitted.status}`);
  let job = admitted.payload.data;
  const deadline = Date.now() + 10_000;
  while (job.status !== "succeeded") {
    if (job.status === "failed" || job.status === "canceled")
      throw new Error(`verification ended: ${job.status}`);
    if (Date.now() >= deadline)
      throw new Error("verification fixture observation expired");
    await new Promise((resolve) => setTimeout(resolve, 20));
    const observed = await publicHttpOperation({
      operationID: "getJob",
      pathParameters: { job_id: job.job_id },
      request: atJsonOrigin(page.request, apiBase),
    });
    if (!observed.ok)
      throw new Error(`verification observation failed: ${observed.status}`);
    job = observed.payload.data;
  }
  if (job.result_summary?.code !== "evidence_blob_verified")
    throw new Error("verification returned a different successful operation");
}
