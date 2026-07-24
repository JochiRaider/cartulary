import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import { errorResponse, jsonResponse } from "../testing/fetchMockTestSupport";
import {
  approveSelectAndApplyExtensionImport,
  ImportMappingPreviewStaleError,
  previewExtensionImportMapping,
  uploadAndDiscoverExtensionImport,
} from "./importCoordinator";

const fingerprint = "a".repeat(64);

describe("extension import coordinator stages", () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  const availability = readyExtensionAvailability();

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

  it("keeps discovery, side-effect-free preview, approval, selection, and apply as explicit stages", async () => {
    installHappyPath(fetchMock);
    const progress: string[] = [];
    const discovery = await uploadAndDiscoverExtensionImport({
      availability,
      incidentId: "incident-1",
      file: csvFile(),
      transactionPrefix: "network-flow-import",
      onProgress: (message) => progress.push(message),
    });

    const preview = await previewExtensionImportMapping<{
      readonly mapping_fingerprint: string;
    }>({ availability, discovery, candidate: mappingCandidate() });
    const refs = await approveSelectAndApplyExtensionImport({
      availability,
      discovery,
      candidate: mappingCandidate(),
      expectedMappingFingerprint: preview.owner_result.mapping_fingerprint,
      transactionPrefix: "network-flow-import",
      onProgress: (message) => progress.push(message),
    });

    expect(refs).toEqual([{ kind: "network_flow_table", id: "nft_returned" }]);
    expect(progress).toEqual([
      "Uploading workbook.",
      "Discovering workbook units.",
      "Approving mapping.",
      "Applying import.",
    ]);
    const previewCall = callFor(fetchMock, "/mapping-preview");
    expect(JSON.parse(String(previewCall?.[1]?.body))).toEqual({
      target_kind: "network_flow_table",
      extension_profile_id: "network_flow_activity",
      owner_mapping_schema_id: "cartulary.network_flow.mapping_candidate.v1",
      owner_mapping: { mapping_kind: "characterized" },
    });
    expect(String(previewCall?.[1]?.body)).not.toContain("client_txn_id");

    const mappingCall = callFor(fetchMock, "/mapping");
    expect(JSON.parse(String(mappingCall?.[1]?.body))).toMatchObject({
      client_txn_id: expect.any(String),
      header_row_ref: 4,
      data_start_row_ref: 5,
      source_columns: [
        sourceColumn(1, "Source IP"),
        sourceColumn(2, "Source IP"),
        sourceColumn(3, null),
      ],
    });
  });

  it("blocks selection and apply when durable approval returns a stale fingerprint", async () => {
    installHappyPath(fetchMock, { durableFingerprint: "b".repeat(64) });
    const discovery = await uploadAndDiscoverExtensionImport({
      availability,
      incidentId: "incident-1",
      file: csvFile(),
      transactionPrefix: "network-flow-import",
    });

    await expect(
      approveSelectAndApplyExtensionImport({
        availability,
        discovery,
        candidate: mappingCandidate(),
        expectedMappingFingerprint: fingerprint,
        transactionPrefix: "network-flow-import",
      }),
    ).rejects.toBeInstanceOf(ImportMappingPreviewStaleError);

    expect(callFor(fetchMock, "/select")).toBeUndefined();
    expect(callFor(fetchMock, "/apply")).toBeUndefined();
  });

  it("preserves public preview errors and creates no durable mapping", async () => {
    installHappyPath(fetchMock, { previewError: true });
    const discovery = await uploadAndDiscoverExtensionImport({
      availability,
      incidentId: "incident-1",
      file: csvFile(),
      transactionPrefix: "network-flow-import",
    });

    await expect(
      previewExtensionImportMapping({
        availability,
        discovery,
        candidate: mappingCandidate(),
      }),
    ).rejects.toThrow("invalid_import_mapping: required_field_missing");

    expect(callFor(fetchMock, "/mapping")).toBeUndefined();
    expect(callFor(fetchMock, "/select")).toBeUndefined();
    expect(callFor(fetchMock, "/apply")).toBeUndefined();
  });
});

function installHappyPath(
  fetchMock: ReturnType<typeof vi.fn>,
  options: {
    readonly durableFingerprint?: string;
    readonly previewError?: boolean;
  } = {},
) {
  fetchMock.mockImplementation(
    async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url === "/api/v1/import-sessions" && method === "POST") {
        expect(init?.body).toBeInstanceOf(FormData);
        return jsonResponse({ data: job("upload-job") });
      }
      if (url === "/api/v1/jobs/upload-job" && method === "GET") {
        return jsonResponse({
          data: {
            ...job("upload-job"),
            result_summary: {
              code: "import_session_created",
              resource_refs: [{ kind: "import_session", id: "session-1" }],
            },
          },
        });
      }
      if (url === "/api/v1/import-sessions/session-1" && method === "GET") {
        return jsonResponse({
          data: {
            import_session_id: "session-1",
            incident_id: "incident-1",
            original_filename: "flows.csv",
            source_file_kind: "csv",
            session_status: "discovered",
            selected_unit_ids: [],
            blocking_diagnostics: [],
            nonblocking_warning_codes: [],
          },
        });
      }
      if (
        url === "/api/v1/import-sessions/session-1/units?limit=50" &&
        method === "GET"
      ) {
        return jsonResponse({
          data: {
            import_units: [
              { import_session_id: "session-1", import_unit_id: "unit-1" },
            ],
          },
        });
      }
      if (url.endsWith("/unit-1/preview") && method === "GET") {
        return jsonResponse({ data: discoveredPreview() });
      }
      if (url.endsWith("/mapping-preview") && method === "POST") {
        expectCSRF(init);
        if (options.previewError === true) {
          return errorResponse("invalid_import_mapping", 422, {
            reason_code: "required_field_missing",
          });
        }
        return jsonResponse({
          data: {
            schema_id: "cartulary.imports.extension_mapping_preview_result.v1",
            import_session_id: "session-1",
            import_unit_id: "unit-1",
            target_kind: "network_flow_table",
            extension_profile_id: "network_flow_activity",
            owner_result_schema_id:
              "cartulary.network_flow.import_preview_result.v1",
            owner_result: { mapping_fingerprint: fingerprint },
          },
        });
      }
      if (url.endsWith("/mapping") && method === "PUT") {
        expectCSRF(init);
        return jsonResponse({
          data: {
            import_session_id: "session-1",
            import_unit_id: "unit-1",
            mapping_fingerprint: options.durableFingerprint ?? fingerprint,
          },
        });
      }
      if (url.endsWith("/select") && method === "POST") {
        expectCSRF(init);
        return jsonResponse({ data: { selected: true } });
      }
      if (url.endsWith("/session-1/apply") && method === "POST") {
        expectCSRF(init);
        return jsonResponse({ data: job("apply-job") });
      }
      if (url === "/api/v1/jobs/apply-job" && method === "GET") {
        return jsonResponse({
          data: {
            ...job("apply-job"),
            result_summary: {
              code: "import_applied",
              resource_refs: [
                { kind: "network_flow_table", id: "nft_returned" },
              ],
            },
          },
        });
      }
      throw new Error(`unexpected fetch ${method} ${url}`);
    },
  );
}

function mappingCandidate() {
  return {
    targetKind: "network_flow_table",
    extensionProfileId: "network_flow_activity",
    ownerMappingSchemaId: "cartulary.network_flow.mapping_candidate.v1",
    ownerMapping: { mapping_kind: "characterized" },
  } as const;
}

function discoveredPreview() {
  return {
    import_session_id: "session-1",
    import_unit_id: "unit-1",
    header_row_ref: 4,
    data_start_row_ref: 5,
    columns: [
      { source_column_ordinal: 1, source_header_text: "Source IP" },
      { source_column_ordinal: 2, source_header_text: "Source IP" },
      { source_column_ordinal: 3, source_header_text: null },
    ],
    preview_rows: [],
  };
}

function sourceColumn(ordinal: number, header: string | null) {
  return {
    source_column_ordinal: ordinal,
    source_header_text: header,
    field_key: null,
    entity_binding_mode: null,
    transform_id: null,
    transform_options: {},
    empty_value_policy: "omit_field",
  };
}

function csvFile() {
  return new File(["header\nvalue\n"], "flows.csv", { type: "text/csv" });
}

function job(jobId: string) {
  return { job_id: jobId, status: "succeeded", result_summary: null } as const;
}

function callFor(fetchMock: ReturnType<typeof vi.fn>, suffix: string) {
  return fetchMock.mock.calls.find(([input]) => String(input).endsWith(suffix));
}

function expectCSRF(init: RequestInit | undefined) {
  const headers = new Headers(init?.headers);
  expect(headers.get("X-CSRF-Token")).toBe("import-csrf");
  expect(init?.credentials).toBe("include");
}
