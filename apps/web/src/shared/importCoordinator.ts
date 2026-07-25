import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import {
  importProfileId,
  importRouteFamily,
} from "../extensions/extensionWorkspaceIdentities";
import { apiPath, clientTxnID, extractError } from "../services/browserApi";
import { requestMultipartJSON } from "../services/httpTransport";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
} from "../services/workbookApi";

export type ImportJobResource = {
  readonly job_id: string;
  readonly status:
    | "queued"
    | "running"
    | "cancel_requested"
    | "succeeded"
    | "failed"
    | "canceled";
  readonly progress?: {
    readonly completed: number;
    readonly total?: number | null;
  };
  readonly result_summary?: {
    readonly code?: string;
    readonly resource_refs?: readonly ImportResourceRef[];
  } | null;
};

export type ImportResourceRef = {
  readonly kind: string;
  readonly id: string;
  readonly route?: string | undefined;
};

export type DiscoveredImportColumn = {
  readonly source_column_ordinal: number;
  readonly source_header_text: string | null;
};

export type DiscoveredImportPreview = {
  readonly import_session_id: string;
  readonly import_unit_id: string;
  readonly locator_kind: string;
  readonly locator: unknown;
  readonly source_rect_a1: string;
  readonly header_row_ref: number;
  readonly data_start_row_ref: number;
  readonly inferred_row_count: number;
  readonly inferred_column_count: number;
  readonly warning_codes: readonly string[];
  readonly columns: readonly DiscoveredImportColumn[];
  readonly preview_rows: readonly {
    readonly source_row_ref: number;
    readonly cells: readonly {
      readonly source_column_ordinal: number;
      readonly display_text: string;
      readonly cell_kind: string;
    }[];
  }[];
};

export type DiscoveredImportUnit = {
  readonly import_session_id: string;
  readonly import_unit_id: string;
  readonly unit_status: string;
  readonly locator_kind: string;
  readonly locator: unknown;
  readonly source_rect_a1: string;
  readonly inferred_row_count: number;
  readonly inferred_column_count: number;
  readonly warning_codes: readonly string[];
  readonly mapping_fingerprint?: string | null | undefined;
  readonly approved_mapping?: Record<string, unknown> | null | undefined;
};

export type ImportSessionResource = {
  readonly import_session_id: string;
  readonly incident_id: string;
  readonly original_filename: string;
  readonly source_file_kind: "csv" | "xlsx";
  readonly session_status: string;
  readonly selected_unit_ids: readonly string[];
  readonly blocking_diagnostics: readonly Record<string, unknown>[];
  readonly nonblocking_warning_codes: readonly string[];
};

export type WorkbookImportUnitDiscovery = {
  readonly unit: DiscoveredImportUnit;
  readonly preview: DiscoveredImportPreview;
};

export type WorkbookImportDiscovery = {
  readonly session: ImportSessionResource;
  readonly units: readonly WorkbookImportUnitDiscovery[];
};

export type ExtensionImportDiscovery = {
  readonly sessionId: string;
  readonly unit: DiscoveredImportUnit;
  readonly preview: DiscoveredImportPreview;
};

export type ExtensionMappingCandidate = {
  readonly targetKind: string;
  readonly extensionProfileId: string;
  readonly ownerMappingSchemaId: string;
  readonly ownerMapping: Record<string, unknown>;
};

export type ExtensionMappingPreviewResource<OwnerResult> = {
  readonly schema_id: "cartulary.imports.extension_mapping_preview_result.v1";
  readonly import_session_id: string;
  readonly import_unit_id: string;
  readonly target_kind: string;
  readonly extension_profile_id: string;
  readonly owner_result_schema_id: string;
  readonly owner_result: OwnerResult;
};

export type WorkbookSourceColumnMapping = {
  readonly source_column_ordinal: number;
  readonly source_header_text: string | null;
  readonly field_key: string | null;
  readonly entity_binding_mode: string | null;
  readonly transform_id: string | null;
  readonly transform_options: Readonly<Record<string, unknown>>;
  readonly empty_value_policy: "omit_field" | "write_null";
};

export class ImportMappingPreviewStaleError extends Error {
  constructor() {
    super("mapping_preview_stale");
    this.name = "ImportMappingPreviewStaleError";
  }
}

export async function uploadAndDiscoverWorkbookImport(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly file: File;
  readonly transactionPrefix: string;
  readonly onProgress?: ((message: string) => void) | undefined;
}): Promise<WorkbookImportDiscovery> {
  options.onProgress?.("Uploading workbook.");
  const uploadJob = await uploadImportSession(options);
  const discovered = await pollImportJob(
    options.availability,
    options.apiBase,
    uploadJob.job_id,
  );
  const sessionId = importSessionIdFromJob(discovered);
  if (sessionId === null) {
    throw new Error("import_session_not_returned");
  }
  options.onProgress?.("Discovering workbook units.");
  const [session, units] = await Promise.all([
    fetchImportSession(options.availability, options.apiBase, sessionId),
    fetchAllImportUnits(options.availability, options.apiBase, sessionId),
  ]);
  const discoveries = await Promise.all(
    units.map(async (unit) => ({
      unit,
      preview: await fetchImportPreview(
        options.availability,
        options.apiBase,
        sessionId,
        unit.import_unit_id,
      ),
    })),
  );
  return { session, units: discoveries };
}

export async function uploadAndDiscoverExtensionImport(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly file: File;
  readonly transactionPrefix: string;
  readonly onProgress?: ((message: string) => void) | undefined;
}): Promise<ExtensionImportDiscovery> {
  const discovery = await uploadAndDiscoverWorkbookImport(options);
  const first = discovery.units[0];
  if (first === undefined) {
    throw new Error("import_unit_not_returned");
  }
  return {
    sessionId: discovery.session.import_session_id,
    unit: first.unit,
    preview: first.preview,
  };
}

export async function previewExtensionImportMapping<OwnerResult>(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly discovery: ExtensionImportDiscovery;
  readonly candidate: ExtensionMappingCandidate;
}): Promise<ExtensionMappingPreviewResource<OwnerResult>> {
  const { candidate, discovery } = options;
  const envelope = await postImportJSON<{
    readonly data: ExtensionMappingPreviewResource<OwnerResult>;
  }>(
    options.availability,
    options.apiBase,
    `/api/v1/import-sessions/${discovery.sessionId}/units/${discovery.unit.import_unit_id}/mapping-preview`,
    {
      target_kind: candidate.targetKind,
      extension_profile_id: candidate.extensionProfileId,
      owner_mapping_schema_id: candidate.ownerMappingSchemaId,
      owner_mapping: candidate.ownerMapping,
    },
  );
  return envelope.data;
}

export async function approveWorkbookImportMapping(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly sessionId: string;
  readonly discovery: WorkbookImportUnitDiscovery;
  readonly targetViewSchemaId: string;
  readonly unknownColumnPolicy:
    | "preserve_raw_capture"
    | "preserve_custom_attrs"
    | "reject_if_unmapped";
  readonly sourceColumns: readonly WorkbookSourceColumnMapping[];
  readonly transactionPrefix: string;
}): Promise<DiscoveredImportUnit> {
  const envelope = await postImportJSON<{
    readonly data: DiscoveredImportUnit;
  }>(
    options.availability,
    options.apiBase,
    `/api/v1/import-sessions/${options.sessionId}/units/${options.discovery.unit.import_unit_id}/mapping`,
    {
      client_txn_id: clientTxnID(`${options.transactionPrefix}-mapping`),
      target_view_schema_id: options.targetViewSchemaId,
      header_row_ref: options.discovery.preview.header_row_ref,
      data_start_row_ref: options.discovery.preview.data_start_row_ref,
      unknown_column_policy: options.unknownColumnPolicy,
      source_columns: options.sourceColumns,
    },
    "PUT",
  );
  return envelope.data;
}

export async function setWorkbookImportUnitSelection(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly sessionId: string;
  readonly unitId: string;
  readonly selected: boolean;
  readonly transactionPrefix: string;
}): Promise<DiscoveredImportUnit> {
  const envelope = await postImportJSON<{
    readonly data: {
      readonly unit: DiscoveredImportUnit;
    };
  }>(
    options.availability,
    options.apiBase,
    `/api/v1/import-sessions/${options.sessionId}/units/${options.unitId}/${options.selected ? "select" : "skip"}`,
    { client_txn_id: clientTxnID(`${options.transactionPrefix}-selection`) },
  );
  return envelope.data.unit;
}

export async function applyWorkbookImport(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly sessionId: string;
  readonly selectedUnitIds: readonly string[];
  readonly transactionPrefix: string;
  readonly onJob?: ((job: ImportJobResource) => void) | undefined;
}): Promise<ImportJobResource> {
  const applyEnvelope = await postImportJSON<{
    readonly data: ImportJobResource;
  }>(
    options.availability,
    options.apiBase,
    `/api/v1/import-sessions/${options.sessionId}/apply`,
    {
      client_txn_id: clientTxnID(`${options.transactionPrefix}-apply`),
      selected_unit_ids: options.selectedUnitIds,
    },
  );
  options.onJob?.(applyEnvelope.data);
  return pollImportJob(
    options.availability,
    options.apiBase,
    applyEnvelope.data.job_id,
    options.onJob,
  );
}

export async function cancelImportJob(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly jobId: string;
  readonly transactionPrefix: string;
}): Promise<ImportJobResource> {
  const envelope = await postImportJSON<{ readonly data: ImportJobResource }>(
    options.availability,
    options.apiBase,
    `/api/v1/jobs/${options.jobId}/cancel`,
    { client_txn_id: clientTxnID(`${options.transactionPrefix}-cancel`) },
  );
  return envelope.data;
}

export async function approveSelectAndApplyExtensionImport(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly discovery: ExtensionImportDiscovery;
  readonly candidate: ExtensionMappingCandidate;
  readonly expectedMappingFingerprint: string;
  readonly transactionPrefix: string;
  readonly onProgress?: ((message: string) => void) | undefined;
}): Promise<readonly ImportResourceRef[]> {
  const { candidate, discovery } = options;
  options.onProgress?.("Approving mapping.");
  const mappingEnvelope = await postImportJSON<{
    readonly data: DiscoveredImportUnit;
  }>(
    options.availability,
    options.apiBase,
    `/api/v1/import-sessions/${discovery.sessionId}/units/${discovery.unit.import_unit_id}/mapping`,
    {
      client_txn_id: clientTxnID(`${options.transactionPrefix}-mapping`),
      target_kind: candidate.targetKind,
      extension_profile_id: candidate.extensionProfileId,
      owner_mapping_schema_id: candidate.ownerMappingSchemaId,
      owner_mapping: candidate.ownerMapping,
      header_row_ref: discovery.preview.header_row_ref,
      data_start_row_ref: discovery.preview.data_start_row_ref,
      source_columns: discovery.preview.columns.map((column) => ({
        source_column_ordinal: column.source_column_ordinal,
        source_header_text: column.source_header_text,
        field_key: null,
        entity_binding_mode: null,
        transform_id: null,
        transform_options: {},
        empty_value_policy: "omit_field",
      })),
    },
    "PUT",
  );
  if (
    mappingEnvelope.data.mapping_fingerprint !==
    options.expectedMappingFingerprint
  ) {
    throw new ImportMappingPreviewStaleError();
  }

  await setWorkbookImportUnitSelection({
    availability: options.availability,
    apiBase: options.apiBase,
    sessionId: discovery.sessionId,
    unitId: discovery.unit.import_unit_id,
    selected: true,
    transactionPrefix: options.transactionPrefix,
  });

  options.onProgress?.("Applying import.");
  const applied = await applyWorkbookImport({
    availability: options.availability,
    apiBase: options.apiBase,
    sessionId: discovery.sessionId,
    selectedUnitIds: [discovery.unit.import_unit_id],
    transactionPrefix: options.transactionPrefix,
  });
  return applied.result_summary?.resource_refs ?? [];
}

async function uploadImportSession(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly file: File;
  readonly transactionPrefix: string;
}): Promise<ImportJobResource> {
  const form = new FormData();
  form.append(
    "metadata",
    new Blob(
      [
        JSON.stringify({
          incident_id: options.incidentId,
          client_txn_id: clientTxnID(`${options.transactionPrefix}-upload`),
        }),
      ],
      { type: "application/json" },
    ),
  );
  form.append("file", options.file, options.file.name);
  const result = await fetchUploadJSON<{ data: ImportJobResource }>(
    options.availability,
    options.apiBase,
    "/api/v1/import-sessions",
    form,
  );
  return result.data;
}

async function fetchAllImportUnits(
  availability: ExtensionAvailabilityController,
  apiBase: string | undefined,
  sessionId: string,
): Promise<readonly DiscoveredImportUnit[]> {
  const result = await runImportRequest(availability, () =>
    fetchWorkbookJSON<{
      readonly data: { readonly import_units: readonly DiscoveredImportUnit[] };
    }>(apiPath(apiBase, `/api/v1/import-sessions/${sessionId}/units?limit=50`)),
  );
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  const units = readEnvelope<{
    data: { import_units: DiscoveredImportUnit[] };
  }>(result.payload).data.import_units;
  if (units.length === 0) {
    throw new Error("import_unit_not_returned");
  }
  return units;
}

async function fetchImportSession(
  availability: ExtensionAvailabilityController,
  apiBase: string | undefined,
  sessionId: string,
): Promise<ImportSessionResource> {
  const result = await runImportRequest(availability, () =>
    fetchWorkbookJSON<{ readonly data: ImportSessionResource }>(
      apiPath(apiBase, `/api/v1/import-sessions/${sessionId}`),
    ),
  );
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  return readEnvelope<{ data: ImportSessionResource }>(result.payload).data;
}

async function fetchImportPreview(
  availability: ExtensionAvailabilityController,
  apiBase: string | undefined,
  sessionId: string,
  unitId: string,
): Promise<DiscoveredImportPreview> {
  const result = await runImportRequest(availability, () =>
    fetchWorkbookJSON<{ data: DiscoveredImportPreview }>(
      apiPath(
        apiBase,
        `/api/v1/import-sessions/${sessionId}/units/${unitId}/preview`,
      ),
    ),
  );
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  return readEnvelope<{ data: DiscoveredImportPreview }>(result.payload).data;
}

export async function pollImportJob(
  availability: ExtensionAvailabilityController,
  apiBase: string | undefined,
  jobId: string,
  onJob?: ((job: ImportJobResource) => void) | undefined,
): Promise<ImportJobResource> {
  for (let attempt = 0; attempt < 30; attempt += 1) {
    const result = await runImportRequest(availability, () =>
      fetchWorkbookJSON<{ data: ImportJobResource }>(
        apiPath(apiBase, `/api/v1/jobs/${jobId}`),
      ),
    );
    if (!result.ok) {
      throw new Error(parseErrorMessage(result.payload));
    }
    const job = readEnvelope<{ data: ImportJobResource }>(result.payload).data;
    onJob?.(job);
    if (job.status === "succeeded") {
      return job;
    }
    if (job.status === "failed" || job.status === "canceled") {
      throw new Error(`job_${job.status}`);
    }
    await new Promise((resolve) => window.setTimeout(resolve, 750));
  }
  throw new Error("job_poll_timeout");
}

async function postImportJSON<T extends object = Record<string, unknown>>(
  availability: ExtensionAvailabilityController,
  apiBase: string | undefined,
  path: string,
  body: Record<string, unknown>,
  method = "POST",
): Promise<T> {
  const result = await runImportRequest(availability, () =>
    fetchWorkbookJSON<T>(apiPath(apiBase, path), {
      method,
      body: JSON.stringify(body),
    }),
  );
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  return readEnvelope<T>(result.payload);
}

async function fetchUploadJSON<T>(
  availability: ExtensionAvailabilityController,
  apiBase: string | undefined,
  path: string,
  body: FormData,
): Promise<T> {
  const response = await runImportRequest(availability, () =>
    requestMultipartJSON<T>(apiPath(apiBase, path), body),
  );
  const payload = response.payload;
  if (!response.ok) {
    const error = extractError(payload);
    throw new Error(error?.code ?? "upload_failed");
  }
  return payload as T;
}

function runImportRequest<T>(
  availability: ExtensionAvailabilityController,
  request: () => Promise<T>,
): Promise<T> {
  return availability.runProfileRequest(
    importProfileId,
    importRouteFamily,
    request,
  );
}

function importSessionIdFromJob(job: ImportJobResource): string | null {
  const refs = job.result_summary?.resource_refs;
  if (!Array.isArray(refs)) {
    return null;
  }
  for (const ref of refs) {
    if (
      ref.kind === "import_session" &&
      typeof ref.id === "string" &&
      ref.id.trim() !== ""
    ) {
      return ref.id;
    }
  }
  return null;
}
