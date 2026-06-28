import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { buildEvidenceLifecycleViewModel } from "../workbook/models/evidenceLifecycleViewModel";
import { csrfHeaderName } from "./browserApi";
import {
  createAndAttachEvidenceBlob,
  evidenceAccessMessageLiveRegion,
  evidenceAttachPublicErrorMessage,
  evidencePublicErrorMessage,
  resolvePublicEvidenceHandleHref,
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

  it("maps blocking evidence messages to assertive live regions", () => {
    const evidenceAccess = buildEvidenceLifecycleViewModel({
      evidenceLifecycleState: "available",
      objectBlobUploadState: "failed",
    });

    expect(
      evidenceAccessMessageLiveRegion("Failed: object blob upload failed.", {
        ...evidenceAccess,
        messageTone: "danger",
      }),
    ).toEqual({ ariaLive: "assertive", role: "alert" });
  });

  it("creates an object blob, uploads it, and attaches it with row-version concurrency", async () => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=test-csrf",
    );
    fetchMock.mockImplementation((input: RequestInfo | URL) => {
      const url = String(input);
      if (url === "/base/api/v1/object-blobs") {
        return Promise.resolve(
          jsonResponse({
            data: {
              object_blob_id: "blob-1",
              upload_target: {
                href: "/upload/blob-1",
                method: "PUT",
                headers: { "X-Upload": "yes" },
              },
            },
          }),
        );
      }
      if (url === "/base/upload/blob-1") {
        return Promise.resolve(new Response("", { status: 200 }));
      }
      if (url === "/base/api/v1/evidence-records/evidence-1/attach-blob") {
        return Promise.resolve(jsonResponse({ data: { ok: true } }));
      }
      return Promise.resolve(
        jsonResponse({ error: { code: "unexpected" } }, 500),
      );
    });

    await createAndAttachEvidenceBlob({
      apiBase: "/base",
      attachClientTxnId: () => "attach-txn-1",
      baseRowVersion: 9,
      createClientTxnId: () => "blob-txn-1",
      evidenceRecordId: "evidence-1",
      file: new File(["abc"], "evidence.txt", { type: "text/plain" }),
      incidentId: "incident-1",
    });

    expect(fetchMock).toHaveBeenCalledTimes(3);
    const createInit = fetchMock.mock.calls[0]?.[1] as RequestInit;
    expect(new Headers(createInit.headers).get(csrfHeaderName)).toBe(
      "test-csrf",
    );
    expect(JSON.parse(String(createInit.body))).toMatchObject({
      incident_id: "incident-1",
      client_txn_id: "blob-txn-1",
      filename_hint: "evidence.txt",
      content_type_hint: "text/plain",
      byte_size: 3,
    });

    const uploadInit = fetchMock.mock.calls[1]?.[1] as RequestInit;
    expect(fetchMock.mock.calls[1]?.[0]).toBe("/base/upload/blob-1");
    expect(uploadInit.credentials).toBe("omit");
    expect(new Headers(uploadInit.headers).get("Content-Type")).toBe(
      "text/plain",
    );
    expect(new Headers(uploadInit.headers).get("X-Upload")).toBe("yes");

    const attachInit = fetchMock.mock.calls[2]?.[1] as RequestInit;
    expect(JSON.parse(String(attachInit.body))).toEqual({
      object_blob_id: "blob-1",
      base_row_version: 9,
      client_txn_id: "attach-txn-1",
    });
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
