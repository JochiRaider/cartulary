import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createWorkbookOperationExecutor } from "../workbook/adapters/workbookOperationExecutor";

import { createUploadedEvidenceBlob } from "../workbook/features/evidence/createUploadedEvidenceBlob";
import {
  resolvePublicEvidenceHandleHref,
  uploadEvidenceObjectBlobTarget,
} from "./workbookEvidence";

describe("workbookEvidence", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("accepts only public evidence handle routes", () => {
    expect(
      resolvePublicEvidenceHandleHref(
        "/api/v1/evidence-handles/handle-1?disposition=preview",
      ),
    ).toBe("/api/v1/evidence-handles/handle-1?disposition=preview");

    expect(resolvePublicEvidenceHandleHref("https://object-store/path")).toBe(
      null,
    );
    expect(resolvePublicEvidenceHandleHref("/api/v1/object-blobs/blob-1")).toBe(
      null,
    );
  });

  it("uses one authenticated CSRF-bound upload attempt without reading response bodies", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=evidence-csrf",
    );
    const responseText = vi.spyOn(Response.prototype, "text");
    fetchMock.mockResolvedValue(
      new Response("s3://private-bucket/private-object", { status: 503 }),
    );
    await expect(
      uploadEvidenceObjectBlobTarget(
        "/base",
        {
          expires_at: "2026-08-04T00:30:00Z",
          href: "/api/v1/object-uploads/upload-token",
          method: "PUT",
          headers: { "Content-Type": "application/cartulary-evidence" },
        },
        new File(["abc"], "evidence.txt", { type: "text/plain" }),
      ),
    ).resolves.toEqual({
      kind: "rejected",
      failure: { cause: "http", status: 503 },
    });
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const uploadInit = fetchMock.mock.calls[0]?.[1] as RequestInit;
    expect(uploadInit.credentials).toBe("include");
    const uploadHeaders = new Headers(uploadInit.headers);
    expect(uploadHeaders.get("Content-Type")).toBe(
      "application/cartulary-evidence",
    );
    expect(uploadHeaders.get("X-CSRF-Token")).toBe("evidence-csrf");
    expect(responseText).not.toHaveBeenCalled();
  });

  it("fails closed before upload when the current CSRF cookie is absent", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue("");

    await expect(
      uploadEvidenceObjectBlobTarget(
        "/base",
        {
          expires_at: "2026-08-04T00:30:00Z",
          href: "/api/v1/object-uploads/upload-token",
          method: "PUT",
          headers: { "Content-Type": "application/cartulary-evidence" },
        },
        new File(["abc"], "evidence.txt", { type: "text/plain" }),
      ),
    ).resolves.toEqual({
      kind: "rejected",
      failure: { cause: "csrf_missing" },
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("creates and uploads a private Evidence object blob without exposing its target", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=evidence-csrf",
    );
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = String(input);
      if (url === "/base/api/v1/object-blobs") {
        return Promise.resolve(jsonResponse(objectBlobEnvelope()));
      }
      if (url === "/base/api/v1/object-uploads/upload-token") {
        return Promise.resolve(new Response("", { status: 200 }));
      }
      return Promise.resolve(
        jsonResponse({ error: { code: "unexpected" } }, 500),
      );
    });

    const objectBlobId = await createUploadedEvidenceBlob({
      apiBase: "/base",
      clientTxnId: "blob-txn-1",
      operations: createWorkbookOperationExecutor({ apiBase: "/base" }),
      file: new File(["abc"], "evidence.txt", { type: "text/plain" }),
      incidentId: "00000000-0000-4000-8000-000000001001",
    });

    expect(objectBlobId).toEqual({
      kind: "accepted",
      value: { objectBlobId: "00000000-0000-4000-8000-000000003001" },
    });
    expect(fetchMock).toHaveBeenCalledTimes(2);
    const createRequest = fetchMock.mock.calls[0]?.[1] as
      | RequestInit
      | undefined;
    expect(createRequest?.method).toBe("POST");
    expect(new Headers(createRequest?.headers).get("X-CSRF-Token")).toBe(
      "evidence-csrf",
    );
    expect(JSON.parse(String(createRequest?.body))).toEqual({
      incident_id: "00000000-0000-4000-8000-000000001001",
      client_txn_id: "blob-txn-1",
      byte_size: 3,
      filename_hint: "evidence.txt",
      content_type_hint: "text/plain",
    });
    const uploadRequest = fetchMock.mock.calls[1]?.[1] as RequestInit;
    expect(uploadRequest.credentials).toBe("include");
    expect(new Headers(uploadRequest.headers).get("X-CSRF-Token")).toBe(
      "evidence-csrf",
    );
    fetchMock.mockReset();
    fetchMock
      .mockResolvedValueOnce(jsonResponse(objectBlobEnvelope()))
      .mockResolvedValueOnce(
        new Response("private upload error body", { status: 401 }),
      );
    expect(
      await createUploadedEvidenceBlob({
        apiBase: "/base",
        clientTxnId: "expired-session-upload",
        operations: createWorkbookOperationExecutor({ apiBase: "/base" }),
        file: new File(["abc"], "evidence.txt", { type: "text/plain" }),
        incidentId: "00000000-0000-4000-8000-000000001001",
      }),
    ).toMatchObject({
      kind: "rejected",
      failure: {
        kind: "authentication_required",
        presentation: { family: "authentication_required" },
        uploadFailure: { cause: "http", status: 401 },
      },
    });
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("fails closed before upload for malformed or cross-incident blob-slot responses", async () => {
    for (const responsePayload of [
      { data: { incident_id: "missing-required-fields" } },
      objectBlobEnvelope({
        incidentId: "00000000-0000-4000-8000-000000001099",
      }),
    ]) {
      fetchMock.mockReset();
      fetchMock.mockResolvedValue(jsonResponse(responsePayload));

      await expect(
        createUploadedEvidenceBlob({
          apiBase: "/base",
          clientTxnId: "blob-txn-1",
          operations: createWorkbookOperationExecutor({ apiBase: "/base" }),
          file: new File(["abc"], "evidence.txt", { type: "text/plain" }),
          incidentId: "00000000-0000-4000-8000-000000001001",
        }),
      ).resolves.toMatchObject({
        kind: "rejected",
        failure: { kind: "invalid_contract" },
      });
      expect(fetchMock).toHaveBeenCalledTimes(1);
    }
  });
});

function objectBlobEnvelope(options: { readonly incidentId?: string } = {}) {
  const incidentId =
    options.incidentId ?? "00000000-0000-4000-8000-000000001001";
  return {
    data: {
      incident_id: incidentId,
      object_blob_id: "00000000-0000-4000-8000-000000003001",
      upload_state: "pending",
      target_expires_at: "2026-07-26T12:05:00Z",
      pending_expires_at: "2026-07-26T12:10:00Z",
      upload_target: {
        href: "/api/v1/object-uploads/upload-token",
        method: "PUT",
        expires_at: "2026-07-26T12:05:00Z",
        headers: { "Content-Type": "text/plain" },
      },
      accepted_contract: {
        incident_id: incidentId,
        byte_size: 3,
        filename_hint: "evidence.txt",
        content_type_hint: "text/plain",
        sha256_hex: null,
      },
    },
    meta: { request_id: "request-create" },
  };
}

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
