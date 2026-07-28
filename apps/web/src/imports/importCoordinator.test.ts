import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import { errorResponse, jsonResponse } from "../testing/fetchMockTestSupport";
import {
  approveSelectAndApplyExtensionImport,
  cancelImportJob,
  ImportMappingPreviewStaleError,
  previewExtensionImportMapping,
  setWorkbookImportUnitSelection,
  uploadAndDiscoverExtensionImport,
  uploadAndDiscoverWorkbookImport,
} from "./importCoordinator";

const fingerprint = "a".repeat(64);
const incidentId = "00000000-0000-4000-8000-000000000001";
const userId = "00000000-0000-4000-8000-000000000002";
const sessionId = "00000000-0000-4000-8000-000000000003";
const firstUnitId = "00000000-0000-4000-8000-000000000004";
const secondUnitId = "00000000-0000-4000-8000-000000000005";
const uploadJobId = "00000000-0000-4000-8000-000000000006";
const applyJobId = "00000000-0000-4000-8000-000000000007";

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
    const workbookDiscovery = await uploadAndDiscoverWorkbookImport({
      availability,
      incidentId,
      file: csvFile(),
      transactionPrefix: "network-flow-import",
      onProgress: (message) => progress.push(message),
    });
    expect(
      workbookDiscovery.units.map((item) => item.unit.import_unit_id),
    ).toEqual([firstUnitId, secondUnitId]);
    const firstDiscovery = workbookDiscovery.units[0];
    if (firstDiscovery === undefined) {
      throw new Error("missing characterized import unit");
    }
    const discovery = {
      sessionId: workbookDiscovery.session.import_session_id,
      unit: firstDiscovery.unit,
      preview: firstDiscovery.preview,
    };

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
    const skipped = await setWorkbookImportUnitSelection({
      availability,
      sessionId,
      unitId: secondUnitId,
      selected: false,
      transactionPrefix: "network-flow-import",
    });
    const canceled = await cancelImportJob({
      availability,
      jobId: applyJobId,
      transactionPrefix: "network-flow-import",
    });

    expect(refs).toEqual([
      {
        kind: "network_flow_table",
        id: "nft_returned",
        route: `/api/v1/incidents/${incidentId}/network-flow/tables/nft_returned`,
      },
    ]);
    expect(skipped).toEqual(
      expect.objectContaining({
        import_unit_id: secondUnitId,
        unit_status: "skipped",
      }),
    );
    expect(canceled.job_id).toBe(applyJobId);
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
    expect(fetchMock).toHaveBeenCalledWith(
      `/api/v1/import-sessions/${sessionId}/units?cursor_token=opaque%20%2F%2B%20cursor&limit=50`,
      expect.any(Object),
    );

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

    for (const [pagingScenario, expectedError] of [
      ["empty", "import_unit_not_returned"],
      ["missing_cursor", "invalid_import_paging_contract"],
      ["repeated_cursor", "invalid_import_paging_contract"],
      ["terminal_cursor", "invalid_import_paging_contract"],
    ] as const) {
      fetchMock.mockReset();
      installHappyPath(fetchMock, { pagingScenario });
      await expect(
        uploadAndDiscoverWorkbookImport({
          availability,
          incidentId,
          file: csvFile(),
          transactionPrefix: `paging-${pagingScenario}`,
        }),
      ).rejects.toThrow(expectedError);
    }
  });

  it("blocks selection and apply when durable approval returns a stale fingerprint", async () => {
    installHappyPath(fetchMock, { durableFingerprint: "b".repeat(64) });
    const discovery = await uploadAndDiscoverExtensionImport({
      availability,
      incidentId,
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
      incidentId,
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
    readonly pagingScenario?:
      | "empty"
      | "missing_cursor"
      | "repeated_cursor"
      | "terminal_cursor";
    readonly previewError?: boolean;
  } = {},
) {
  fetchMock.mockImplementation(
    async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url === "/api/v1/import-sessions" && method === "POST") {
        expect(init?.body).toBeInstanceOf(FormData);
        return jsonResponse(envelope(job(uploadJobId, null, "queued")));
      }
      if (url === `/api/v1/jobs/${uploadJobId}` && method === "GET") {
        return jsonResponse(
          envelope(
            job(uploadJobId, {
              code: "import_session_discovered",
              message: "Import session discovery completed.",
              resource_refs: [
                {
                  kind: "import_session",
                  id: sessionId,
                  route: `/api/v1/import-sessions/${sessionId}`,
                },
              ],
            }),
          ),
        );
      }
      if (url === `/api/v1/import-sessions/${sessionId}` && method === "GET") {
        return jsonResponse(envelope(importSession()));
      }
      if (
        url === `/api/v1/import-sessions/${sessionId}/units?limit=50` &&
        method === "GET"
      ) {
        if (options.pagingScenario === "empty") {
          return jsonResponse(pagedUnits([], false, null));
        }
        if (options.pagingScenario === "missing_cursor") {
          return jsonResponse(
            pagedUnits([importUnit(firstUnitId)], true, null),
          );
        }
        if (options.pagingScenario === "terminal_cursor") {
          return jsonResponse(
            pagedUnits([importUnit(firstUnitId)], false, "unexpected"),
          );
        }
        return jsonResponse(
          pagedUnits(
            [importUnit(firstUnitId)],
            true,
            options.pagingScenario === "repeated_cursor"
              ? "repeat"
              : "opaque /+ cursor",
          ),
        );
      }
      if (
        url ===
          `/api/v1/import-sessions/${sessionId}/units?cursor_token=opaque%20%2F%2B%20cursor&limit=50` &&
        method === "GET"
      ) {
        return jsonResponse(
          pagedUnits([importUnit(secondUnitId)], false, null),
        );
      }
      if (
        url ===
          `/api/v1/import-sessions/${sessionId}/units?cursor_token=repeat&limit=50` &&
        method === "GET"
      ) {
        return jsonResponse(
          pagedUnits([importUnit(secondUnitId)], true, "repeat"),
        );
      }
      if (url.endsWith(`/${firstUnitId}/preview`) && method === "GET") {
        return jsonResponse(envelope(discoveredPreview(firstUnitId)));
      }
      if (url.endsWith(`/${secondUnitId}/preview`) && method === "GET") {
        return jsonResponse(
          envelope(discoveredPreview(secondUnitId, "Destination IP")),
        );
      }
      if (url.endsWith("/mapping-preview") && method === "POST") {
        expectCSRF(init);
        if (options.previewError === true) {
          return errorResponse("invalid_import_mapping", 422, {
            reason_code: "required_field_missing",
          });
        }
        return jsonResponse(
          envelope({
            schema_id: "cartulary.imports.extension_mapping_preview_result.v1",
            import_session_id: sessionId,
            import_unit_id: firstUnitId,
            target_kind: "network_flow_table",
            extension_profile_id: "network_flow_activity",
            owner_result_schema_id:
              "cartulary.network_flow.import_preview_result.v1",
            owner_result: { mapping_fingerprint: fingerprint },
          }),
        );
      }
      if (url.endsWith("/mapping") && method === "PUT") {
        expectCSRF(init);
        return jsonResponse(
          envelope({
            ...importUnit(firstUnitId, "mapped"),
            mapping_fingerprint: options.durableFingerprint ?? fingerprint,
          }),
        );
      }
      if (url.endsWith("/select") && method === "POST") {
        expectCSRF(init);
        return jsonResponse(
          envelope({
            import_session_id: sessionId,
            selected_unit_ids: [firstUnitId],
            session_status: "ready_to_apply",
            unit: importUnit(firstUnitId, "ready"),
          }),
        );
      }
      if (url.endsWith("/skip") && method === "POST") {
        expectCSRF(init);
        return jsonResponse(
          envelope({
            import_session_id: sessionId,
            selected_unit_ids: [firstUnitId],
            session_status: "ready_to_apply",
            unit: importUnit(secondUnitId, "skipped"),
          }),
        );
      }
      if (
        url === `/api/v1/import-sessions/${sessionId}/apply` &&
        method === "POST"
      ) {
        expectCSRF(init);
        return jsonResponse(envelope(job(applyJobId, null, "queued")));
      }
      if (url === `/api/v1/jobs/${applyJobId}` && method === "GET") {
        return jsonResponse(
          envelope(
            job(applyJobId, {
              code: "import_applied",
              message: "Import session applied.",
              resource_refs: [
                {
                  kind: "network_flow_table",
                  id: "nft_returned",
                  route: `/api/v1/incidents/${incidentId}/network-flow/tables/nft_returned`,
                },
              ],
            }),
          ),
        );
      }
      if (url === `/api/v1/jobs/${applyJobId}/cancel` && method === "POST") {
        expectCSRF(init);
        return jsonResponse(
          envelope(
            job(
              applyJobId,
              {
                code: "job_canceled",
                message: "Job canceled.",
                resource_refs: [],
              },
              "canceled",
            ),
          ),
        );
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

function importSession() {
  return {
    import_session_id: sessionId,
    incident_id: incidentId,
    created_by_user_id: userId,
    created_at: "2026-07-28T12:00:00Z",
    original_filename: "flows.csv",
    source_file_kind: "csv",
    source_content_sha256: "b".repeat(64),
    parser_profile_id: "tabular_default",
    parser_version: "1.0.0",
    assistant_profile: "workbook_import",
    session_status: "discovered",
    selected_unit_ids: [],
    blocking_diagnostics: [],
    nonblocking_warning_codes: [],
  } as const;
}

function importUnit(
  unitId: string,
  unitStatus:
    | "discovered"
    | "mapped"
    | "ready"
    | "skipped"
    | "applying"
    | "applied"
    | "failed" = "discovered",
) {
  return {
    import_session_id: sessionId,
    import_unit_id: unitId,
    unit_status: unitStatus,
    locator_kind: "csv_file",
    locator: { file: "source" },
    source_rect_a1: "A1:C2",
    header_row_ref: 4,
    data_start_row_ref: 5,
    inferred_row_count: 1,
    inferred_column_count: 3,
    warning_codes: [],
  } as const;
}

function discoveredPreview(unitId = firstUnitId, secondHeader = "Source IP") {
  return {
    import_session_id: sessionId,
    import_unit_id: unitId,
    locator_kind: "csv_file",
    locator: { file: "source" },
    source_rect_a1: "A1:C2",
    header_row_ref: 4,
    data_start_row_ref: 5,
    inferred_row_count: 1,
    inferred_column_count: 3,
    warning_codes: [],
    unit_status: "discovered",
    columns: [
      { source_column_ordinal: 1, source_header_text: "Source IP" },
      { source_column_ordinal: 2, source_header_text: secondHeader },
      { source_column_ordinal: 3, source_header_text: null },
    ],
    preview_rows: [],
    truncated: false,
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

function job(
  jobId: string,
  resultSummary: {
    readonly code: string;
    readonly message: string;
    readonly resource_refs: readonly {
      readonly kind: string;
      readonly id: string;
      readonly route: string;
    }[];
  } | null = null,
  status:
    | "queued"
    | "running"
    | "cancel_requested"
    | "succeeded"
    | "failed"
    | "canceled" = "succeeded",
) {
  const terminal =
    status === "succeeded" || status === "failed" || status === "canceled";
  return {
    job_id: jobId,
    scope: { kind: "incident", incident_id: incidentId },
    status_route: `/api/v1/jobs/${jobId}`,
    submitted_by_user_id: userId,
    status,
    cancelable: status === "queued" || status === "running",
    progress: { completed: terminal ? 1 : 0, total: 1 },
    result_summary: resultSummary,
    error_summary: null,
    submitted_at: "2026-07-28T12:00:00Z",
    started_at: status === "queued" ? null : "2026-07-28T12:00:01Z",
    finished_at: terminal ? "2026-07-28T12:00:02Z" : null,
    retained_until: terminal ? "2026-08-04T12:00:02Z" : null,
    updated_at: "2026-07-28T12:00:02Z",
  } as const;
}

function envelope<T>(data: T) {
  return {
    data,
    meta: { request_id: "request-import" },
  };
}

function pagedUnits(
  importUnits: readonly ReturnType<typeof importUnit>[],
  hasMore: boolean,
  nextCursor: string | null,
) {
  return {
    data: { import_units: importUnits },
    meta: {
      request_id: "request-import",
      paging: { limit: 50, has_more: hasMore, next_cursor: nextCursor },
    },
  };
}

function callFor(fetchMock: ReturnType<typeof vi.fn>, suffix: string) {
  return fetchMock.mock.calls.find(([input]) => String(input).endsWith(suffix));
}

function expectCSRF(init: RequestInit | undefined) {
  const headers = new Headers(init?.headers);
  expect(headers.get("X-CSRF-Token")).toBe("import-csrf");
  expect(init?.credentials).toBe("include");
}
