import type {
  DiscoveredImportPreview,
  DiscoveredImportUnit,
  ImportJobResource,
  ImportSessionResource,
} from "../services/importContractAdapter";
export const importTestIds = {
  incident: "00000000-0000-4000-8000-000000000001",
  actor: "00000000-0000-4000-8000-000000000002",
  session: "00000000-0000-4000-8000-000000000003",
  unit: "00000000-0000-4000-8000-000000000004",
  job: "00000000-0000-4000-8000-000000000005",
  secondUnit: "00000000-0000-4000-8000-000000000006",
} as const;
export const importTestScope = {
  incidentId: importTestIds.incident,
  actorId: importTestIds.actor,
  lifetime: "test-session",
};
export const importTestMapping = {
  target_view_schema_id: "cartulary.view.timeline.v2",
  unknown_column_policy: "preserve_raw_capture",
  source_columns: [
    {
      source_column_ordinal: 1,
      source_header_text: "Activity Synopsis",
      field_key: "timeline.activity_synopsis_text",
      entity_binding_mode: null,
      transform_id: null,
      transform_options: {},
      empty_value_policy: "omit_field",
    },
  ],
} as const;
export function importTestUnit(
  overrides: Partial<DiscoveredImportUnit> = {},
): DiscoveredImportUnit {
  return {
    import_session_id: importTestIds.session,
    import_unit_id: importTestIds.unit,
    locator_kind: "csv_file",
    locator: {},
    source_rect_a1: "A1:A2",
    header_row_ref: 1,
    data_start_row_ref: 2,
    inferred_row_count: 1,
    inferred_column_count: 1,
    warning_codes: [],
    unit_status: "discovered",
    ...overrides,
  };
}
export function importTestApprovedUnit(
  overrides: Partial<DiscoveredImportUnit> = {},
): DiscoveredImportUnit {
  return importTestUnit({
    unit_status: "ready",
    approved_mapping: JSON.parse(JSON.stringify(importTestMapping)),
    mapping_fingerprint: "a".repeat(64),
    ...overrides,
  });
}
export function importTestPreview(
  unit = importTestUnit(),
): DiscoveredImportPreview {
  const { approved_mapping: _approved, ...base } = unit;
  return {
    ...base,
    columns: [
      { source_column_ordinal: 1, source_header_text: "Activity Synopsis" },
    ],
    preview_rows: [
      {
        source_row_ref: 2,
        cells: [
          {
            source_column_ordinal: 1,
            display_text: "Imported observation",
            cell_kind: "string",
          },
        ],
      },
    ],
    truncated: false,
  };
}
export function importTestSession(
  overrides: Partial<ImportSessionResource> = {},
): ImportSessionResource {
  return {
    import_session_id: importTestIds.session,
    incident_id: importTestIds.incident,
    created_by_user_id: importTestIds.actor,
    created_at: "2026-09-08T10:00:00Z",
    source_file_kind: "csv",
    original_filename: "source.csv",
    source_content_sha256: "b".repeat(64),
    parser_profile_id: "tabular_default",
    parser_version: "1",
    assistant_profile: "phase2_workbook_import_v1",
    session_status: "discovered",
    selected_unit_ids: [],
    blocking_diagnostics: [],
    nonblocking_warning_codes: [],
    ...overrides,
  };
}
export function importTestJob(
  status: ImportJobResource["status"] = "queued",
  overrides: Partial<ImportJobResource> = {},
): ImportJobResource {
  const terminal = ["succeeded", "failed", "canceled"].includes(status);
  return {
    job_id: importTestIds.job,
    scope: { kind: "incident", incident_id: importTestIds.incident },
    status_route: `/api/v1/jobs/${importTestIds.job}`,
    status,
    cancelable: status === "queued" || status === "running",
    submitted_by_user_id: importTestIds.actor,
    submitted_at: "2026-09-08T10:00:00Z",
    updated_at: terminal ? "2026-09-08T10:00:02Z" : "2026-09-08T10:00:01Z",
    started_at: status === "queued" ? null : "2026-09-08T10:00:01Z",
    finished_at: terminal ? "2026-09-08T10:00:02Z" : null,
    retained_until: terminal ? "2026-09-15T10:00:02Z" : null,
    progress: { completed: status === "succeeded" ? 1 : 0, total: 1 },
    result_summary:
      status === "succeeded"
        ? {
            code: "import_session_discovered",
            message: "Discovered",
            resource_refs: [
              {
                kind: "import_session",
                id: importTestIds.session,
                route: `/api/v1/import-sessions/${importTestIds.session}`,
              },
            ],
          }
        : status === "canceled"
          ? { code: "job_canceled", message: "Canceled", resource_refs: [] }
          : null,
    error_summary:
      status === "failed"
        ? {
            code: "import_source_unsupported",
            message: "Source unsupported",
            retryable: false,
          }
        : null,
    ...overrides,
  };
}
export const importTestEnvelope = <T>(data: T) => ({
  data,
  meta: { request_id: "test-request" },
});
