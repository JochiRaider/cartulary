import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  createUploadedEvidenceObjectBlob,
  evidenceAttachPublicErrorMessage,
  evidencePublicErrorMessage,
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

  it("keeps raw storage details out of public evidence error messages", () => {
    expect(
      evidencePublicErrorMessage({
        error: {
          code: "object_store_unavailable",
          details: { reason_code: "s3://bucket/private-key" },
          message: "https://storage.example/private",
          status: 503,
        },
      }),
    ).toBe("Request failed.");

    expect(
      evidencePublicErrorMessage({
        error: {
          code: "evidence_access_unavailable",
          details: { reason_code: "unsupported_preview" },
        },
      }),
    ).toBe("evidence_access_unavailable: unsupported_preview");

    expect(
      evidenceAttachPublicErrorMessage(
        new Error(
          "upload_failed_500: https://minio.internal/cartulary-evidence-bucket/object_blob_storage_key_v1",
        ),
      ),
    ).toBe("Evidence attach failed.");

    expect(
      evidenceAttachPublicErrorMessage(new Error("upload_failed_503")),
    ).toBe("upload_failed_503");
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

  it("bounds evidence upload retries by transport outcome without reading response bodies", async () => {
    vi.useFakeTimers();
    const responseText = vi.spyOn(Response.prototype, "text");
    try {
      fetchMock.mockResolvedValue(
        new Response("s3://private-bucket/private-object", { status: 503 }),
      );
      const statusRetry = uploadEvidenceObjectBlobTarget(
        "/base",
        {
          href: "/api/v1/object-uploads/upload-token",
          method: "PUT",
          headers: {},
        },
        new File(["abc"], "evidence.txt", { type: "text/plain" }),
      );

      await vi.runAllTimersAsync();
      await expect(statusRetry).resolves.toEqual({
        kind: "rejected",
        message: "upload_failed_503",
        retryable: true,
      });
      expect(fetchMock).toHaveBeenCalledTimes(3);
      expect(responseText).not.toHaveBeenCalled();

      fetchMock.mockReset();
      fetchMock
        .mockRejectedValueOnce(new TypeError("connection unavailable"))
        .mockRejectedValueOnce(new TypeError("connection unavailable"))
        .mockResolvedValueOnce(new Response("", { status: 200 }));
      const networkRetry = uploadEvidenceObjectBlobTarget(
        "/base",
        {
          href: "/api/v1/object-uploads/upload-token",
          method: "PUT",
          headers: {},
        },
        new File(["abc"], "evidence.txt", { type: "text/plain" }),
      );

      await vi.runAllTimersAsync();
      await expect(networkRetry).resolves.toEqual({ kind: "accepted" });
      expect(fetchMock).toHaveBeenCalledTimes(3);
      expect(responseText).not.toHaveBeenCalled();
    } finally {
      vi.useRealTimers();
    }
  });

  it("creates and uploads a private Evidence object blob without exposing its target", async () => {
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = String(input);
      if (url === "/base/api/v1/object-blobs") {
        return Promise.resolve(
          jsonResponse({
            data: {
              incident_id: "00000000-0000-4000-8000-000000001001",
              object_blob_id: "00000000-0000-4000-8000-000000003001",
              upload_state: "pending",
              target_expires_at: "2026-07-26T12:05:00Z",
              pending_expires_at: "2026-07-26T12:10:00Z",
              upload_target: {
                href: "/api/v1/object-uploads/upload-token",
                method: "PUT",
                expires_at: "2026-07-26T12:05:00Z",
                headers: {},
              },
              accepted_contract: {
                incident_id: "00000000-0000-4000-8000-000000001001",
                byte_size: 3,
                filename_hint: "evidence.txt",
                content_type_hint: "text/plain",
                sha256_hex: null,
              },
            },
            meta: { request_id: "request-create" },
          }),
        );
      }
      if (url === "/base/api/v1/object-uploads/upload-token") {
        return Promise.resolve(new Response("", { status: 200 }));
      }
      return Promise.resolve(
        jsonResponse({ error: { code: "unexpected" } }, 500),
      );
    });

    const objectBlobId = await createUploadedEvidenceObjectBlob({
      apiBase: "/base",
      createClientTxnId: () => "blob-txn-1",
      file: new File(["abc"], "evidence.txt", { type: "text/plain" }),
      incidentId: "00000000-0000-4000-8000-000000001001",
    });

    expect(objectBlobId).toBe("00000000-0000-4000-8000-000000003001");
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
});

function jsonResponse(payload: unknown, status = 200) {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
