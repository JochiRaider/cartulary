import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import { apiPath, clientTxnID, extractError } from "../services/browserApi";
import { requestMultipartJSON } from "../services/httpTransport";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
} from "../services/workbookApi";

type JobResource = {
  readonly job_id: string;
  readonly status:
    | "queued"
    | "running"
    | "cancel_requested"
    | "succeeded"
    | "failed"
    | "canceled";
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
  readonly header_row_ref: number;
  readonly data_start_row_ref: number;
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
  readonly mapping_fingerprint?: string | null | undefined;
  readonly approved_mapping?: Record<string, unknown> | null | undefined;
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

export class ImportMappingPreviewStaleError extends Error {
  constructor() {
    super("mapping_preview_stale");
    this.name = "ImportMappingPreviewStaleError";
  }
}

export async function uploadAndDiscoverExtensionImport(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly file: File;
  readonly transactionPrefix: string;
  readonly onProgress?: ((message: string) => void) | undefined;
}): Promise<ExtensionImportDiscovery> {
  options.onProgress?.("Uploading import.");
  const uploadJob = await uploadImportSession(options);
  const discovered = await pollJob(
    options.availability,
    options.apiBase,
    uploadJob.job_id,
  );
  const sessionId = importSessionIdFromJob(discovered);
  if (sessionId === null) {
    throw new Error("import_session_not_returned");
  }
  options.onProgress?.("Discovering source columns.");
  const unit = await firstImportUnit(
    options.availability,
    options.apiBase,
    sessionId,
  );
  const preview = await fetchImportPreview(
    options.availability,
    options.apiBase,
    sessionId,
    unit.import_unit_id,
  );
  return { sessionId, unit, preview };
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

  await postImportJSON(
    options.availability,
    options.apiBase,
    `/api/v1/import-sessions/${discovery.sessionId}/units/${discovery.unit.import_unit_id}/select`,
    { client_txn_id: clientTxnID(`${options.transactionPrefix}-select`) },
  );

  options.onProgress?.("Applying import.");
  const applyEnvelope = await postImportJSON<{ data: JobResource }>(
    options.availability,
    options.apiBase,
    `/api/v1/import-sessions/${discovery.sessionId}/apply`,
    { client_txn_id: clientTxnID(`${options.transactionPrefix}-apply`) },
  );
  const applied = await pollJob(
    options.availability,
    options.apiBase,
    applyEnvelope.data.job_id,
  );
  return applied.result_summary?.resource_refs ?? [];
}

async function uploadImportSession(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly file: File;
  readonly transactionPrefix: string;
}): Promise<JobResource> {
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
  const result = await fetchUploadJSON<{ data: JobResource }>(
    options.availability,
    options.apiBase,
    "/api/v1/import-sessions",
    form,
  );
  return result.data;
}

async function firstImportUnit(
  availability: ExtensionAvailabilityController,
  apiBase: string | undefined,
  sessionId: string,
): Promise<DiscoveredImportUnit> {
  const result = await availability.runRequest(() =>
    fetchWorkbookJSON<{
      readonly data: { readonly import_units: readonly DiscoveredImportUnit[] };
    }>(apiPath(apiBase, `/api/v1/import-sessions/${sessionId}/units`)),
  );
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  const unit = readEnvelope<{
    data: { import_units: DiscoveredImportUnit[] };
  }>(result.payload).data.import_units[0];
  if (unit === undefined || unit.import_unit_id.trim() === "") {
    throw new Error("import_unit_not_returned");
  }
  return unit;
}

async function fetchImportPreview(
  availability: ExtensionAvailabilityController,
  apiBase: string | undefined,
  sessionId: string,
  unitId: string,
): Promise<DiscoveredImportPreview> {
  const result = await availability.runRequest(() =>
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

async function pollJob(
  availability: ExtensionAvailabilityController,
  apiBase: string | undefined,
  jobId: string,
): Promise<JobResource> {
  for (let attempt = 0; attempt < 30; attempt += 1) {
    const result = await availability.runRequest(() =>
      fetchWorkbookJSON<{ data: JobResource }>(
        apiPath(apiBase, `/api/v1/jobs/${jobId}`),
      ),
    );
    if (!result.ok) {
      throw new Error(parseErrorMessage(result.payload));
    }
    const job = readEnvelope<{ data: JobResource }>(result.payload).data;
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
  const result = await availability.runRequest(() =>
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
  const response = await availability.runRequest(() =>
    requestMultipartJSON<T>(apiPath(apiBase, path), body),
  );
  const payload = response.payload;
  if (!response.ok) {
    const error = extractError(payload);
    throw new Error(error?.code ?? "upload_failed");
  }
  return payload as T;
}

function importSessionIdFromJob(job: JobResource): string | null {
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
