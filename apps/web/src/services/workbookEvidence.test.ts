import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { csrfHeaderName } from "./browserApi";
import {
  createAndAttachEvidenceBlob,
  createEvidenceWithInitialBlob,
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
              incident_id: "00000000-0000-4000-8000-000000001001",
              object_blob_id: "00000000-0000-4000-8000-000000003001",
              upload_state: "pending",
              target_expires_at: "2026-07-26T12:05:00Z",
              pending_expires_at: "2026-07-26T12:10:00Z",
              upload_target: {
                href: "/api/v1/object-uploads/upload-token",
                method: "PUT",
                expires_at: "2026-07-26T12:05:00Z",
                headers: { "X-Upload": "yes" },
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
      if (
        url ===
        "/base/api/v1/evidence-records/00000000-0000-4000-8000-000000004001/attach-blob"
      ) {
        return Promise.resolve(
          jsonResponse({
            data: {
              view_schema_id: "cartulary.view.evidence.v1",
              change_set_id: "00000000-0000-4000-8000-000000005001",
              row: {
                record_id: "00000000-0000-4000-8000-000000004001",
                row_version: 10,
                cells: {},
              },
              object_blob_id: "00000000-0000-4000-8000-000000003001",
            },
            meta: { request_id: "request-attach" },
          }),
        );
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
      evidenceRecordId: "00000000-0000-4000-8000-000000004001",
      file: new File(["abc"], "evidence.txt", { type: "text/plain" }),
      incidentId: "00000000-0000-4000-8000-000000001001",
    });

    expect(fetchMock).toHaveBeenCalledTimes(3);
    const createInit = fetchMock.mock.calls[0]?.[1] as RequestInit;
    expect(new Headers(createInit.headers).get(csrfHeaderName)).toBe(
      "test-csrf",
    );
    expect(JSON.parse(String(createInit.body))).toMatchObject({
      incident_id: "00000000-0000-4000-8000-000000001001",
      client_txn_id: "blob-txn-1",
      filename_hint: "evidence.txt",
      content_type_hint: "text/plain",
      byte_size: 3,
    });

    const uploadInit = fetchMock.mock.calls[1]?.[1] as RequestInit;
    expect(fetchMock.mock.calls[1]?.[0]).toBe(
      "/base/api/v1/object-uploads/upload-token",
    );
    expect(uploadInit.credentials).toBe("omit");
    expect(new Headers(uploadInit.headers).get("Content-Type")).toBe(
      "text/plain",
    );
    expect(new Headers(uploadInit.headers).get("X-Upload")).toBe("yes");

    const attachInit = fetchMock.mock.calls[2]?.[1] as RequestInit;
    expect(JSON.parse(String(attachInit.body))).toEqual({
      object_blob_id: "00000000-0000-4000-8000-000000003001",
      base_row_version: 9,
      client_txn_id: "attach-txn-1",
    });
  });

  it("creates a blob-backed Evidence row atomically and reuses the row transaction ID after response uncertainty", async () => {
    let rowAttempts = 0;
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
      if (
        url ===
        "/base/api/v1/incidents/00000000-0000-4000-8000-000000001001/views/cartulary.view.evidence.v1/rows"
      ) {
        rowAttempts += 1;
        if (rowAttempts === 1) {
          return Promise.reject(new TypeError("response lost"));
        }
        return Promise.resolve(
          jsonResponse({
            data: {
              view_schema_id: "cartulary.view.evidence.v1",
              change_set_id: "00000000-0000-4000-8000-000000005001",
              row: {
                record_id: "00000000-0000-4000-8000-000000004001",
                row_version: 1,
                cells: {},
              },
            },
            meta: { request_id: "request-row" },
          }),
        );
      }
      return Promise.resolve(
        jsonResponse({ error: { code: "unexpected" } }, 500),
      );
    });
    const rowTxn = vi.fn(() => "row-txn-1");

    const created = await createEvidenceWithInitialBlob({
      apiBase: "/base",
      createBlobClientTxnId: () => "blob-txn-1",
      createRowClientTxnId: rowTxn,
      file: new File(["abc"], "evidence.txt", { type: "text/plain" }),
      incidentId: "00000000-0000-4000-8000-000000001001",
      values: {
        "evidence.title": "evidence.txt",
        "evidence.collector_party_text": "Workbook upload",
      },
      viewSchemaId: "cartulary.view.evidence.v1",
    });

    expect(created).toEqual({
      recordId: "00000000-0000-4000-8000-000000004001",
      rowVersion: 1,
    });
    expect(rowTxn).toHaveBeenCalledTimes(1);
    const firstRowBody = String(
      (fetchMock.mock.calls[2]?.[1] as RequestInit).body,
    );
    const replayRowBody = String(
      (fetchMock.mock.calls[3]?.[1] as RequestInit).body,
    );
    expect(replayRowBody).toBe(firstRowBody);
    expect(JSON.parse(firstRowBody)).toEqual({
      client_txn_id: "row-txn-1",
      "evidence.collector_party_text": "Workbook upload",
      "evidence.initial_object_blob_id": "00000000-0000-4000-8000-000000003001",
      "evidence.title": "evidence.txt",
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
