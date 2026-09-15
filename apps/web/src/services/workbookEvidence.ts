import type { CreateObjectBlobSlotResponse } from "@cartulary/protocol-ts/http";
import { apiPath, csrfHeaderName, readCookie } from "./browserApi";

export type EvidenceUploadFailure =
  | { readonly cause: "csrf_missing" | "network" | "invalid_target" }
  | { readonly cause: "http"; readonly status: number };
export type EvidenceUploadOutcome =
  | { readonly kind: "accepted" }
  | { readonly kind: "rejected"; readonly failure: EvidenceUploadFailure };

export type EvidenceObjectUploadTarget =
  CreateObjectBlobSlotResponse["data"]["upload_target"];

export function validEvidenceObjectUploadTarget(
  apiBase: string | undefined,
  uploadTarget: EvidenceObjectUploadTarget,
): boolean {
  // Validate the public route without decoding or interpreting its capability.
  if (
    !/^\/api\/v1\/object-uploads\/[^/?#\\]+$/u.test(uploadTarget.href) ||
    uploadTarget.method !== "PUT" ||
    /%(?:2f|5c)/iu.test(uploadTarget.href)
  ) {
    return false;
  }
  const uploadHref =
    uploadTarget.href.startsWith("/") && apiBase
      ? apiPath(apiBase, uploadTarget.href)
      : uploadTarget.href;
  try {
    const resolved = new URL(uploadHref, window.location.href);
    const expectedPath = new URL(
      apiPath(apiBase, "/api/v1/object-uploads/"),
      window.location.href,
    ).pathname;
    if (
      resolved.origin !== window.location.origin ||
      !resolved.pathname.startsWith(expectedPath) ||
      resolved.pathname.slice(expectedPath.length).length === 0 ||
      resolved.pathname.slice(expectedPath.length).includes("/")
    )
      return false;
  } catch {
    return false;
  }
  return true;
}

export async function uploadEvidenceObjectBlobTarget(
  apiBase: string | undefined,
  uploadTarget: EvidenceObjectUploadTarget,
  file: File,
  signal?: AbortSignal,
): Promise<EvidenceUploadOutcome> {
  if (!validEvidenceObjectUploadTarget(apiBase, uploadTarget)) {
    return { kind: "rejected", failure: { cause: "invalid_target" } };
  }
  const uploadHref = apiBase
    ? apiPath(apiBase, uploadTarget.href)
    : uploadTarget.href;
  const headers = new Headers();
  for (const [key, value] of Object.entries(uploadTarget.headers)) {
    headers.set(key, value);
  }
  const csrfToken = readCookie("cartulary_csrf");
  if (!csrfToken) {
    return {
      kind: "rejected",
      failure: { cause: "csrf_missing" },
    };
  }
  headers.set(csrfHeaderName, csrfToken);
  let upload: Response;
  try {
    upload = await fetch(uploadHref, {
      method: uploadTarget.method,
      credentials: "include",
      headers,
      body: file,
      ...(signal ? { signal } : {}),
    });
  } catch {
    return {
      kind: "rejected",
      failure: { cause: "network" },
    };
  }
  if (upload.status === 204) {
    return { kind: "accepted" };
  }
  return {
    kind: "rejected",
    failure: { cause: "http", status: upload.status },
  };
}

export function resolvePublicEvidenceHandleHref(href: string): string | null {
  const trimmed = href.trim();
  if (
    !/^\/api\/v1\/evidence-handles\/[A-Za-z0-9._~!$&'()*+,;=:@%-]+(?:\?[A-Za-z0-9._~!$&'()*+,;=:@%/?-]*)?$/u.test(
      trimmed,
    )
  ) {
    return null;
  }
  try {
    const resolved = new URL(trimmed, "https://cartulary.invalid");
    if (
      resolved.origin !== "https://cartulary.invalid" ||
      !resolved.pathname.startsWith("/api/v1/evidence-handles/") ||
      /%(?:2f|5c)/iu.test(resolved.pathname)
    )
      return null;
  } catch {
    return null;
  }
  return trimmed;
}
