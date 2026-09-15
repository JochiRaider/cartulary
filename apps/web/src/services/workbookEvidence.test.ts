import {
  evidenceViewSchemaId,
  requireViewContract,
} from "@cartulary/view-contracts";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../testing/timelineWorkbookTestSupport";
import { createEvidenceFileTransport } from "../workbook/adapters/createEvidenceFileTransport";
import { EvidenceUploadSession } from "../workbook/features/evidence/EvidenceUploadSession";

const authority = {
  actorId: "actor",
  sessionIdentity: "session",
  incidentId: "00000000-0000-4000-8000-000000001001",
  role: "editor" as const,
  closed: false,
};
function prepareUpload(file: File, clientTxnId = "blob-txn-1") {
  return new EvidenceUploadSession(
    file,
    createEvidenceFileTransport("/base"),
    { create: () => clientTxnId },
    () => {},
    () => {},
  );
}

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

  it("attaches an empty file without changing Evidence custody", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=evidence-csrf",
    );
    const slot = objectBlobEnvelope();
    slot.data.accepted_contract.byte_size = 0;
    const recordId = "00000000-0000-4000-8000-000000002001";
    const row = fullWorkbookViewRow(
      requireViewContract(evidenceViewSchemaId),
      recordId,
      2,
      {
        "evidence.title": "empty",
        "evidence.lifecycle_state": "requested",
        "evidence.storage_ref": `object://${slot.data.object_blob_id}`,
      },
    );
    fetchMock
      .mockResolvedValueOnce(jsonResponse(slot))
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
      .mockResolvedValueOnce(
        jsonResponse({
          data: {
            view_schema_id: evidenceViewSchemaId,
            object_blob_id: slot.data.object_blob_id,
            row,
            change_set_id: "00000000-0000-4000-8000-000000004001",
          },
          meta: { request_id: "attach-request" },
        }),
      );
    const upload = prepareUpload(
      new File([], "evidence.txt", { type: "text/plain" }),
    );
    expect(await upload.prepare(authority)).toBe(true);
    const transport = createEvidenceFileTransport("/base");
    const attempt = transport.capture({
      stage: "attach",
      authority,
      clientTxnId: "attach-empty",
      recordId,
      baseRowVersion: 1,
      objectBlobId: slot.data.object_blob_id,
    });
    expect(
      await transport.finalize(attempt, new AbortController().signal),
    ).toMatchObject({ kind: "accepted", receipt: { data: { row } } });
    expect(fetchMock.mock.calls.map(([, init]) => init.method)).toEqual([
      "POST",
      "PUT",
      "POST",
    ]);
    expect(JSON.parse(fetchMock.mock.calls[0]?.[1].body).byte_size).toBe(0);
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
        return Promise.resolve(new Response(null, { status: 204 }));
      }
      return Promise.resolve(
        jsonResponse({ error: { code: "unexpected" } }, 500),
      );
    });

    const upload = prepareUpload(
      new File(["abc"], "evidence.txt", { type: "text/plain" }),
    );
    expect(await upload.prepare(authority)).toBe(true);
    expect(upload.blob?.object_blob_id).toBe(
      "00000000-0000-4000-8000-000000003001",
    );
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
    const expired = prepareUpload(
      new File(["abc"], "evidence.txt", { type: "text/plain" }),
      "expired-session-upload",
    );
    expect(await expired.prepare(authority)).toBe(false);
    expect(expired.status.phase).toBe("transfer_uncertain");
    expect(fetchMock).toHaveBeenCalledTimes(2);

    // Filename hints are normalized by the slot owner, not echoed raw.
    fetchMock.mockReset();
    fetchMock
      .mockResolvedValueOnce(jsonResponse(objectBlobEnvelope()))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    const normalized = prepareUpload(
      new File(["abc"], "  evidence.txt  ", { type: "text/plain" }),
      "normalized-hint",
    );
    expect(await normalized.prepare(authority)).toBe(true);
    expect(fetchMock).toHaveBeenCalledTimes(2);
    fetchMock.mockReset();
    fetchMock
      .mockResolvedValueOnce(jsonResponse(objectBlobEnvelope()))
      .mockResolvedValueOnce(new Response(null, { status: 204 }));
    const unicodeWhitespace = prepareUpload(
      new File(["abc"], "\u0085evidence.txt\u0085", { type: "text/plain" }),
      "normalized-unicode-hint",
    );
    expect(await unicodeWhitespace.prepare(authority)).toBe(true);
    expect(unicodeWhitespace.blob?.accepted_contract.filename_hint).toBe(
      "evidence.txt",
    );
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("fails closed before upload for malformed or cross-incident blob-slot responses", async () => {
    const foreign = objectBlobEnvelope();
    foreign.data.upload_target.href =
      "https://foreign.invalid/api/v1/object-uploads/private";
    const traversal = objectBlobEnvelope();
    traversal.data.upload_target.href = "/api/v1/object-uploads/%2e%2e";
    for (const responsePayload of [
      foreign,
      traversal,
      { data: { incident_id: "missing-required-fields" } },
      objectBlobEnvelope({
        incidentId: "00000000-0000-4000-8000-000000001099",
      }),
    ]) {
      fetchMock.mockReset();
      fetchMock.mockResolvedValue(jsonResponse(responsePayload));

      const upload = prepareUpload(
        new File(["abc"], "evidence.txt", { type: "text/plain" }),
      );
      expect(await upload.prepare(authority)).toBe(false);
      expect(upload.status.phase).toBe("slot_uncertain");
      expect(fetchMock).toHaveBeenCalledTimes(1);
    }
  });
});

function objectBlobEnvelope(options: { readonly incidentId?: string } = {}) {
  const incidentId =
    options.incidentId ?? "00000000-0000-4000-8000-000000001001";
  const targetExpiry = new Date(Date.now() + 3_600_000).toISOString();
  return {
    data: {
      incident_id: incidentId,
      object_blob_id: "00000000-0000-4000-8000-000000003001",
      upload_state: "pending",
      target_expires_at: targetExpiry,
      pending_expires_at: new Date(Date.now() + 86_400_000).toISOString(),
      upload_target: {
        href: "/api/v1/object-uploads/upload-token",
        method: "PUT",
        expires_at: targetExpiry,
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
