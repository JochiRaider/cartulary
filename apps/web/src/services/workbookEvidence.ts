import type {
  CreateObjectBlobSlotRequest,
  CreateObjectBlobSlotResponse,
  ErrorEnvelope,
} from "@cartulary/protocol-ts";
import { publicErrorStatusText } from "../shared/publicError";
import { apiPath, fetchHTTPOperation } from "./browserApi";

export type EvidenceUploadOutcome =
  | { readonly kind: "accepted" }
  | {
      readonly kind: "rejected";
      readonly retryable: boolean;
      readonly message: string;
    };

export type EvidenceObjectUploadTarget =
  CreateObjectBlobSlotResponse["data"]["upload_target"];

export async function uploadEvidenceObjectBlobTarget(
  apiBase: string | undefined,
  uploadTarget: EvidenceObjectUploadTarget,
  file: File,
): Promise<EvidenceUploadOutcome> {
  const uploadHref =
    uploadTarget.href.startsWith("/") && apiBase
      ? apiPath(apiBase, uploadTarget.href)
      : uploadTarget.href;
  const headers = new Headers();
  for (const [key, value] of Object.entries(uploadTarget.headers)) {
    headers.set(key, value);
  }
  let hasContentType = false;
  headers.forEach((_value, key) => {
    if (key.toLowerCase() === "content-type") {
      hasContentType = true;
    }
  });
  if (!hasContentType) {
    headers.set("Content-Type", file.type || "application/octet-stream");
  }
  for (let attempt = 0; attempt < 3; attempt += 1) {
    let upload: Response;
    try {
      upload = await fetch(uploadHref, {
        method: uploadTarget.method,
        credentials: "omit",
        headers,
        body: file,
      });
    } catch {
      if (attempt < 2) {
        await sleep(200 * (attempt + 1));
        continue;
      }
      return {
        kind: "rejected",
        retryable: true,
        message: "upload_failed_network",
      };
    }
    if (upload.ok) {
      return { kind: "accepted" };
    }
    const retryable = upload.status === 503 || upload.status === 504;
    if (attempt < 2 && retryable) {
      await sleep(200 * (attempt + 1));
      continue;
    }
    return {
      kind: "rejected",
      retryable,
      message: `upload_failed_${upload.status}`,
    };
  }
  return {
    kind: "rejected",
    retryable: true,
    message: "upload_failed_network",
  };
}

export async function createUploadedEvidenceObjectBlob({
  apiBase,
  createClientTxnId,
  file,
  incidentId,
}: {
  readonly apiBase: string | undefined;
  readonly createClientTxnId: () => string;
  readonly file: File;
  readonly incidentId: string;
}): Promise<string> {
  const createBlobRequest = {
    incident_id: incidentId,
    client_txn_id: createClientTxnId(),
    byte_size: file.size,
    filename_hint: file.name || null,
    content_type_hint: file.type || null,
  } satisfies CreateObjectBlobSlotRequest;
  const createBlob = await fetchHTTPOperation<CreateObjectBlobSlotResponse>({
    apiBase,
    operationID: "createObjectBlobSlot",
    init: {
      method: "POST",
      body: JSON.stringify(createBlobRequest),
    },
  });
  if (!createBlob.ok) {
    throw new Error(evidencePublicErrorMessage(createBlob.payload));
  }
  const blobEnvelope = createBlob.payload;
  if (!createdBlobMatchesRequest(blobEnvelope, createBlobRequest)) {
    throw new Error("invalid_public_contract_response");
  }
  const upload = await uploadEvidenceObjectBlobTarget(
    apiBase,
    blobEnvelope.data.upload_target,
    file,
  );
  if (upload.kind === "rejected") {
    throw new Error(upload.message);
  }
  return blobEnvelope.data.object_blob_id;
}

function createdBlobMatchesRequest(
  envelope: CreateObjectBlobSlotResponse,
  request: CreateObjectBlobSlotRequest,
): boolean {
  const accepted = envelope.data.accepted_contract;
  return (
    envelope.data.incident_id === request.incident_id &&
    accepted.incident_id === request.incident_id &&
    accepted.byte_size === request.byte_size &&
    accepted.filename_hint === request.filename_hint &&
    accepted.content_type_hint === request.content_type_hint
  );
}

function sleep(milliseconds: number): Promise<void> {
  return new Promise((resolve) => window.setTimeout(resolve, milliseconds));
}

const rawEvidenceStorageDetailPatterns = [
  /\bhttps?:\/\//iu,
  /\bs3:\/\//iu,
  /\bobject:\/\//iu,
  /\bminio\b/iu,
  /\bseaweedfs?\b/iu,
  /\bbucket(?:\b|_)/iu,
  /\bobject[-_ ]?store(?:\b|_)/iu,
  /\bstorage[-_ ]?backend(?:\b|_)/iu,
  /\b(?:storage|object)[-_ ]?key(?:\b|_)/iu,
  /\bobject[-_ ]?blob[-_ ]?storage[-_ ]?key(?:\b|_)/iu,
  /\/(?:home|var|tmp|usr|app|workspace|data|mnt|srv)\//iu,
  /\b(?:local|filesystem|s3|gcs|azure)[-_ ]?backend\b/iu,
] as const;

function containsRawEvidenceStorageDetail(value: string): boolean {
  return rawEvidenceStorageDetailPatterns.some((pattern) =>
    pattern.test(value),
  );
}

function safeEvidencePublicText(value: unknown): string | null {
  if (
    typeof value !== "string" &&
    typeof value !== "number" &&
    typeof value !== "boolean"
  ) {
    return null;
  }
  const text = String(value).trim();
  if (
    text === "" ||
    text.length > 240 ||
    containsRawEvidenceStorageDetail(text)
  ) {
    return null;
  }
  return text;
}

function evidencePublicError(payload: unknown): ErrorEnvelope["error"] | null {
  if (!payload || typeof payload !== "object" || !("error" in payload)) {
    return null;
  }
  const error = payload.error;
  if (!error || typeof error !== "object") {
    return null;
  }
  return error as ErrorEnvelope["error"];
}

export function evidencePublicErrorMessage(
  payload: unknown,
  fallback = "Evidence request failed.",
): string {
  const error = evidencePublicError(payload);
  if (error === null) {
    return fallback;
  }
  const reason =
    error.details && typeof error.details === "object"
      ? safeEvidencePublicText(error.details.reason_code)
      : null;
  const code = safeEvidencePublicText(error.code);
  if (code !== null && reason !== null) {
    return `${code}: ${reason}`;
  }
  if (reason !== null) {
    return reason;
  }
  const message = safeEvidencePublicText(error.message);
  if (message !== null) {
    return message;
  }
  if (code !== null) {
    return code;
  }
  const statusText = safeEvidencePublicText(
    publicErrorStatusText({ status: error.status }, error.status),
  );
  return statusText ?? fallback;
}

export function evidenceAttachPublicErrorMessage(
  error: unknown,
  fallback = "Evidence attach failed.",
): string {
  if (error instanceof Error) {
    return safeEvidencePublicText(error.message) ?? fallback;
  }
  return safeEvidencePublicText(error) ?? fallback;
}

export function resolvePublicEvidenceHandleHref(href: string): string | null {
  const trimmed = href.trim();
  if (
    containsRawEvidenceStorageDetail(trimmed) ||
    !/^\/api\/v1\/evidence-handles\/[A-Za-z0-9._~!$&'()*+,;=:@%-]+(?:\?[A-Za-z0-9._~!$&'()*+,;=:@%/?-]*)?$/u.test(
      trimmed,
    )
  ) {
    return null;
  }
  return trimmed;
}
