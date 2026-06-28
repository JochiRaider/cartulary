import type {
  ErrorEnvelope,
  EvidenceAttachBlobEnvelope,
  EvidenceAttachBlobRequest,
  EvidenceHandleEnvelope,
  EvidenceHandleIssueRequest,
  ObjectBlobCreateEnvelope,
  ObjectBlobCreateRequest,
  ObjectBlobUploadTarget,
} from "@cartulary/protocol-ts";
import { publicErrorStatusText } from "../shared/publicError";
import type { EvidenceLifecycleViewModel } from "../workbook/models/evidenceLifecycleViewModel";
import { apiPath } from "./browserApi";
import { fetchJSON, readEnvelope } from "./workbookApi";

export type EvidenceHandleKind = "preview" | "download";

export type IssuedEvidenceHandle =
  | {
      readonly ok: true;
      readonly filename: string;
      readonly href: string;
      readonly previewKind: string | null;
    }
  | {
      readonly ok: false;
      readonly message: string;
    };

async function uploadObjectBlobTarget(
  apiBase: string | undefined,
  uploadTarget: ObjectBlobUploadTarget,
  file: File,
): Promise<void> {
  const uploadHref =
    uploadTarget.href.startsWith("/") && apiBase
      ? apiPath(apiBase, uploadTarget.href)
      : uploadTarget.href;
  const headers = new Headers(uploadTarget.headers ?? undefined);
  let hasContentType = false;
  headers.forEach((_value, key) => {
    if (key.toLowerCase() === "content-type") {
      hasContentType = true;
    }
  });
  if (!hasContentType) {
    headers.set("Content-Type", file.type || "application/octet-stream");
  }
  let lastStatus = 0;
  let lastDetail = "";
  for (let attempt = 0; attempt < 3; attempt += 1) {
    const upload = await fetch(uploadHref, {
      method: uploadTarget.method ?? "PUT",
      credentials: "omit",
      headers,
      body: file,
    });
    if (upload.ok) {
      return;
    }
    lastStatus = upload.status;
    lastDetail = await readUploadFailureDetail(upload);
    if (attempt < 2 && isRetryableUploadFailure(upload.status, lastDetail)) {
      await sleep(200 * (attempt + 1));
      continue;
    }
    break;
  }
  throw new Error(
    lastDetail === ""
      ? `upload_failed_${lastStatus}`
      : `upload_failed_${lastStatus}: ${lastDetail}`,
  );
}

export async function createAndAttachEvidenceBlob({
  apiBase,
  attachClientTxnId,
  baseRowVersion,
  createClientTxnId,
  evidenceRecordId,
  file,
  incidentId,
}: {
  readonly apiBase: string | undefined;
  readonly attachClientTxnId: () => string;
  readonly baseRowVersion: number;
  readonly createClientTxnId: () => string;
  readonly evidenceRecordId: string;
  readonly file: File;
  readonly incidentId: string;
}): Promise<void> {
  const createBlobRequest = {
    incident_id: incidentId,
    client_txn_id: createClientTxnId(),
    byte_size: file.size,
    filename_hint: file.name || null,
    content_type_hint: file.type || null,
  } satisfies ObjectBlobCreateRequest;
  const createBlob = await fetchJSON<ObjectBlobCreateEnvelope>(
    apiPath(apiBase, "/api/v1/object-blobs"),
    {
      method: "POST",
      body: JSON.stringify(createBlobRequest),
    },
  );
  if (!createBlob.ok) {
    throw new Error(evidencePublicErrorMessage(createBlob.payload));
  }
  const blobEnvelope = readEnvelope<ObjectBlobCreateEnvelope>(
    createBlob.payload,
  );
  await uploadObjectBlobTarget(apiBase, blobEnvelope.data.upload_target, file);

  const attachRequest = {
    object_blob_id: blobEnvelope.data.object_blob_id,
    base_row_version: baseRowVersion,
    client_txn_id: attachClientTxnId(),
  } satisfies EvidenceAttachBlobRequest;
  const attach = await fetchJSON<EvidenceAttachBlobEnvelope>(
    apiPath(
      apiBase,
      `/api/v1/evidence-records/${evidenceRecordId}/attach-blob`,
    ),
    {
      method: "POST",
      body: JSON.stringify(attachRequest),
    },
  );
  if (!attach.ok) {
    throw new Error(evidencePublicErrorMessage(attach.payload));
  }
}

async function readUploadFailureDetail(response: Response): Promise<string> {
  try {
    return (await response.text()).replace(/\s+/g, " ").slice(0, 300);
  } catch {
    return "";
  }
}

function isRetryableUploadFailure(status: number, detail: string): boolean {
  if (status !== 503 && status !== 504) {
    return false;
  }
  return detail === "" || detail.includes('"retryable":true');
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

export async function issueEvidenceAccessHandle({
  apiBase,
  evidenceRecordId,
  kind,
}: {
  readonly apiBase: string | undefined;
  readonly evidenceRecordId: string;
  readonly kind: EvidenceHandleKind;
}): Promise<IssuedEvidenceHandle> {
  const handleRequest = {} satisfies EvidenceHandleIssueRequest;
  const result = await fetchJSON<EvidenceHandleEnvelope>(
    apiPath(
      apiBase,
      `/api/v1/evidence-records/${evidenceRecordId}/${kind}-handle`,
    ),
    { method: "POST", body: JSON.stringify(handleRequest) },
  );
  if (!result.ok) {
    return {
      ok: false,
      message: evidencePublicErrorMessage(
        result.payload,
        "Evidence access failed.",
      ),
    };
  }
  const envelope = readEnvelope<EvidenceHandleEnvelope>(result.payload);
  const href = resolvePublicEvidenceHandleHref(envelope.data.href);
  if (href === null) {
    return {
      ok: false,
      message: "Evidence handle is unavailable.",
    };
  }
  return {
    ok: true,
    filename: envelope.data.filename,
    href,
    previewKind: envelope.data.preview_kind ?? null,
  };
}

export function evidenceAccessMessageLiveRegion(
  message: string,
  evidenceAccess: EvidenceLifecycleViewModel,
): { ariaLive: "assertive" | "polite"; role: "alert" | "status" } {
  const normalized = message.toLowerCase();
  const isBlockingMessage =
    evidenceAccess.messageTone === "danger" ||
    evidenceAccess.stateKey === "blocked" ||
    evidenceAccess.stateKey === "failed" ||
    evidenceAccess.stateKey === "inconsistent" ||
    evidenceAccess.stateKey === "preview_blocked" ||
    evidenceAccess.stateKey === "public_error" ||
    /\b(?:blob_failed|evidence_access_unavailable|failed|inconsistent|quarantined|unavailable|unsupported_preview)\b/u.test(
      normalized,
    );
  return isBlockingMessage
    ? { ariaLive: "assertive", role: "alert" }
    : { ariaLive: "polite", role: "status" };
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
