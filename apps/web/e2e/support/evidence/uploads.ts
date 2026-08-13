import type { Buffer } from "node:buffer";

import type {
  CreateObjectBlobSlotRequest,
  CreateObjectBlobSlotResponse,
} from "@cartulary/protocol-ts/http";
import type { Page } from "@playwright/test";

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
