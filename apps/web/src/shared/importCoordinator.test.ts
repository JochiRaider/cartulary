import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { errorResponse, jsonResponse } from "../testing/fetchMockTestSupport";
import { coordinateExtensionImport } from "./importCoordinator";

describe("coordinateExtensionImport", () => {
  let fetchMock: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=import-csrf",
    );
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("preserves upload, discovery, mapping, selection, apply, and job request boundaries", async () => {
    fetchMock.mockImplementation(
      async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        const method = (init?.method ?? "GET").toUpperCase();
        if (url === "/api/v1/import-sessions" && method === "POST") {
          expect(init?.body).toBeInstanceOf(FormData);
          return jsonResponse({ data: job("upload-job", "succeeded") });
        }
        if (url === "/api/v1/jobs/upload-job" && method === "GET") {
          return jsonResponse({
            data: {
              ...job("upload-job", "succeeded"),
              result_summary: {
                code: "import_session_created",
                resource_refs: [{ kind: "import_session", id: "session-1" }],
              },
            },
          });
        }
        if (
          url === "/api/v1/import-sessions/session-1/units" &&
          method === "GET"
        ) {
          return jsonResponse({
            data: { import_units: [{ import_unit_id: "unit-1" }] },
          });
        }
        if (
          url === "/api/v1/import-sessions/session-1/units/unit-1/mapping" &&
          method === "PUT"
        ) {
          expectJSONRequest(init, {
            mapping_kind: "characterized",
          });
          expectCSRF(init);
          return jsonResponse({ data: { accepted: true } });
        }
        if (
          url === "/api/v1/import-sessions/session-1/units/unit-1/select" &&
          method === "POST"
        ) {
          expectJSONRequest(init, { client_txn_id: expect.any(String) });
          expectCSRF(init);
          return jsonResponse({ data: { selected: true } });
        }
        if (
          url === "/api/v1/import-sessions/session-1/apply" &&
          method === "POST"
        ) {
          expectJSONRequest(init, { client_txn_id: expect.any(String) });
          expectCSRF(init);
          return jsonResponse({ data: job("apply-job", "succeeded") });
        }
        if (url === "/api/v1/jobs/apply-job" && method === "GET") {
          return jsonResponse({ data: job("apply-job", "succeeded") });
        }
        throw new Error(`unexpected fetch ${method} ${url}`);
      },
    );

    const progress: string[] = [];
    await coordinateExtensionImport({
      incidentId: "incident-1",
      file: new File(["header\nvalue\n"], "flows.csv", {
        type: "text/csv",
      }),
      mappingPayload: (clientTxnId) => ({
        client_txn_id: clientTxnId,
        mapping_kind: "characterized",
      }),
      transactionPrefix: "network-flow-import",
      onProgress: (message) => progress.push(message),
    });

    expect(progress).toEqual([
      "Uploading import.",
      "Preparing mapping.",
      "Applying import.",
    ]);
    expect(fetchMock).toHaveBeenCalledTimes(7);
  });

  it("preserves public mapping errors and stops before selection or apply", async () => {
    fetchMock.mockImplementation(
      async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = String(input);
        const method = (init?.method ?? "GET").toUpperCase();
        if (url === "/api/v1/import-sessions" && method === "POST") {
          return jsonResponse({ data: job("upload-job", "succeeded") });
        }
        if (url === "/api/v1/jobs/upload-job" && method === "GET") {
          return jsonResponse({
            data: {
              ...job("upload-job", "succeeded"),
              result_summary: {
                resource_refs: [{ kind: "import_session", id: "session-1" }],
              },
            },
          });
        }
        if (
          url === "/api/v1/import-sessions/session-1/units" &&
          method === "GET"
        ) {
          return jsonResponse({
            data: { import_units: [{ import_unit_id: "unit-1" }] },
          });
        }
        if (url.endsWith("/mapping") && method === "PUT") {
          return errorResponse("invalid_import_mapping", 422, {
            reason_code: "required_field_missing",
          });
        }
        throw new Error(`unexpected fetch ${method} ${url}`);
      },
    );

    await expect(
      coordinateExtensionImport({
        incidentId: "incident-1",
        file: new File(["header\n"], "invalid.csv", { type: "text/csv" }),
        mappingPayload: (clientTxnId) => ({ client_txn_id: clientTxnId }),
        transactionPrefix: "network-flow-import",
      }),
    ).rejects.toThrow("invalid_import_mapping: required_field_missing");

    expect(
      fetchMock.mock.calls.some(([input]) => String(input).endsWith("/select")),
    ).toBe(false);
    expect(
      fetchMock.mock.calls.some(([input]) => String(input).endsWith("/apply")),
    ).toBe(false);
  });
});

function job(jobId: string, status: "succeeded") {
  return {
    job_id: jobId,
    status,
    result_summary: null,
  } as const;
}

function expectJSONRequest(
  init: RequestInit | undefined,
  expected: Record<string, unknown>,
) {
  expect(JSON.parse(String(init?.body ?? "{}"))).toMatchObject(expected);
}

function expectCSRF(init: RequestInit | undefined) {
  const headers = new Headers(init?.headers);
  expect(headers.get("X-CSRF-Token")).toBe("import-csrf");
  expect(init?.credentials).toBe("include");
}
