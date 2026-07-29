import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import {
  importProfileId,
  importRouteFamily,
} from "../extensions/extensionWorkspaceIdentities";
import {
  type APIResult,
  clientTxnID,
  fetchHTTPOperation,
  fetchMultipartHTTPOperation,
} from "../services/browserApi";
import type {
  ApplyImportSessionResponse,
  CancelJobResponse,
  CreateImportSessionResponse,
  CreateImportUnitRegionResponse,
  DiscoveredImportColumn,
  DiscoveredImportPreview,
  DiscoveredImportUnit,
  ExtensionMappingPreviewResource,
  GetImportSessionResponse,
  GetImportUnitPreviewResponse,
  GetJobResponse,
  ImportJobResource,
  ImportResourceRef,
  ImportSessionResource,
  ListImportUnitsResponse,
  PreviewImportUnitExtensionMappingResponse,
  PutImportUnitMappingResponse,
  SelectImportUnitResponse,
  SkipImportUnitResponse,
  WorkbookSourceColumnMapping,
} from "../services/importContractAdapter";
import { parseErrorMessage } from "../services/workbookApi";

export type {
  DiscoveredImportColumn,
  DiscoveredImportPreview,
  DiscoveredImportUnit,
  ImportJobResource,
  ImportResourceRef,
  ImportSessionResource,
  WorkbookSourceColumnMapping,
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

export type { ExtensionMappingPreviewResource };

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
  const envelope =
    await requireImportResult<PreviewImportUnitExtensionMappingResponse>(
      options.availability,
      () =>
        fetchHTTPOperation<PreviewImportUnitExtensionMappingResponse>({
          apiBase: options.apiBase,
          operationID: "previewImportUnitExtensionMapping",
          pathParameters: {
            import_session_id: discovery.sessionId,
            import_unit_id: discovery.unit.import_unit_id,
          },
          init: {
            method: "POST",
            body: JSON.stringify({
              target_kind: candidate.targetKind,
              extension_profile_id: candidate.extensionProfileId,
              owner_mapping_schema_id: candidate.ownerMappingSchemaId,
              owner_mapping: candidate.ownerMapping,
            }),
          },
        }),
    );
  return envelope.data as ExtensionMappingPreviewResource<OwnerResult>;
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
  const envelope = await requireImportResult<PutImportUnitMappingResponse>(
    options.availability,
    () =>
      fetchHTTPOperation<PutImportUnitMappingResponse>({
        apiBase: options.apiBase,
        operationID: "putImportUnitMapping",
        pathParameters: {
          import_session_id: options.sessionId,
          import_unit_id: options.discovery.unit.import_unit_id,
        },
        init: {
          method: "PUT",
          body: JSON.stringify({
            client_txn_id: clientTxnID(`${options.transactionPrefix}-mapping`),
            target_view_schema_id: options.targetViewSchemaId,
            header_row_ref: options.discovery.preview.header_row_ref,
            data_start_row_ref: options.discovery.preview.data_start_row_ref,
            unknown_column_policy: options.unknownColumnPolicy,
            source_columns: options.sourceColumns,
          }),
        },
      }),
  );
  return envelope.data;
}

export async function createWorkbookImportRegion(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly sessionId: string;
  readonly baseUnitId: string;
  readonly sourceRect: {
    readonly startRow: number;
    readonly startColumn: number;
    readonly endRow: number;
    readonly endColumn: number;
  };
  readonly transactionPrefix: string;
}): Promise<WorkbookImportUnitDiscovery> {
  const envelope = await requireImportResult<CreateImportUnitRegionResponse>(
    options.availability,
    () =>
      fetchHTTPOperation<CreateImportUnitRegionResponse>({
        apiBase: options.apiBase,
        operationID: "createImportUnitRegion",
        pathParameters: {
          import_session_id: options.sessionId,
          base_unit_id: options.baseUnitId,
        },
        init: {
          method: "POST",
          body: JSON.stringify({
            client_txn_id: clientTxnID(
              `${options.transactionPrefix}-operator-region`,
            ),
            source_rect: {
              start_row: options.sourceRect.startRow,
              start_column: options.sourceRect.startColumn,
              end_row: options.sourceRect.endRow,
              end_column: options.sourceRect.endColumn,
            },
          }),
        },
      }),
  );
  return {
    unit: envelope.data,
    preview: await fetchImportPreview(
      options.availability,
      options.apiBase,
      options.sessionId,
      envelope.data.import_unit_id,
    ),
  };
}

export async function setWorkbookImportUnitSelection(options: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase?: string | undefined;
  readonly sessionId: string;
  readonly unitId: string;
  readonly selected: boolean;
  readonly transactionPrefix: string;
}): Promise<DiscoveredImportUnit> {
  const operationID = options.selected ? "selectImportUnit" : "skipImportUnit";
  const envelope = await requireImportResult<
    SelectImportUnitResponse | SkipImportUnitResponse
  >(options.availability, () =>
    fetchHTTPOperation<SelectImportUnitResponse | SkipImportUnitResponse>({
      apiBase: options.apiBase,
      operationID,
      pathParameters: {
        import_session_id: options.sessionId,
        import_unit_id: options.unitId,
      },
      init: {
        method: "POST",
        body: JSON.stringify({
          client_txn_id: clientTxnID(`${options.transactionPrefix}-selection`),
        }),
      },
    }),
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
  const applyEnvelope = await requireImportResult<ApplyImportSessionResponse>(
    options.availability,
    () =>
      fetchHTTPOperation<ApplyImportSessionResponse>({
        apiBase: options.apiBase,
        operationID: "applyImportSession",
        pathParameters: { import_session_id: options.sessionId },
        init: {
          method: "POST",
          body: JSON.stringify({
            client_txn_id: clientTxnID(`${options.transactionPrefix}-apply`),
            selected_unit_ids: options.selectedUnitIds,
          }),
        },
      }),
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
  const envelope = await requireImportResult<CancelJobResponse>(
    options.availability,
    () =>
      fetchHTTPOperation<CancelJobResponse>({
        apiBase: options.apiBase,
        operationID: "cancelJob",
        pathParameters: { job_id: options.jobId },
        init: {
          method: "POST",
          body: JSON.stringify({
            client_txn_id: clientTxnID(`${options.transactionPrefix}-cancel`),
          }),
        },
      }),
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
  const mappingEnvelope =
    await requireImportResult<PutImportUnitMappingResponse>(
      options.availability,
      () =>
        fetchHTTPOperation<PutImportUnitMappingResponse>({
          apiBase: options.apiBase,
          operationID: "putImportUnitMapping",
          pathParameters: {
            import_session_id: discovery.sessionId,
            import_unit_id: discovery.unit.import_unit_id,
          },
          init: {
            method: "PUT",
            body: JSON.stringify({
              client_txn_id: clientTxnID(
                `${options.transactionPrefix}-mapping`,
              ),
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
            }),
          },
        }),
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
  const result = await requireImportResult<CreateImportSessionResponse>(
    options.availability,
    () =>
      fetchMultipartHTTPOperation<CreateImportSessionResponse>({
        apiBase: options.apiBase,
        operationID: "createImportSession",
        body: form,
      }),
  );
  return result.data;
}

async function fetchAllImportUnits(
  availability: ExtensionAvailabilityController,
  apiBase: string | undefined,
  sessionId: string,
): Promise<readonly DiscoveredImportUnit[]> {
  const units: DiscoveredImportUnit[] = [];
  const continuationCursors = new Set<string>();
  let cursorToken: string | undefined;
  for (;;) {
    const envelope = await requireImportResult<ListImportUnitsResponse>(
      availability,
      () =>
        fetchHTTPOperation<ListImportUnitsResponse>({
          apiBase,
          operationID: "listImportUnits",
          pathParameters: { import_session_id: sessionId },
          query: { limit: 50, cursor_token: cursorToken },
        }),
    );
    units.push(...envelope.data.import_units);

    const paging = envelope.meta.paging;
    if (paging === undefined) {
      throw new Error("invalid_import_paging_contract");
    }
    if (!paging.has_more) {
      if (paging.next_cursor !== null) {
        throw new Error("invalid_import_paging_contract");
      }
      break;
    }
    const nextCursor = paging.next_cursor;
    if (
      typeof nextCursor !== "string" ||
      nextCursor.trim() === "" ||
      continuationCursors.has(nextCursor)
    ) {
      throw new Error("invalid_import_paging_contract");
    }
    continuationCursors.add(nextCursor);
    cursorToken = nextCursor;
  }
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
  const envelope = await requireImportResult<GetImportSessionResponse>(
    availability,
    () =>
      fetchHTTPOperation<GetImportSessionResponse>({
        apiBase,
        operationID: "getImportSession",
        pathParameters: { import_session_id: sessionId },
      }),
  );
  return envelope.data;
}

async function fetchImportPreview(
  availability: ExtensionAvailabilityController,
  apiBase: string | undefined,
  sessionId: string,
  unitId: string,
): Promise<DiscoveredImportPreview> {
  const envelope = await requireImportResult<GetImportUnitPreviewResponse>(
    availability,
    () =>
      fetchHTTPOperation<GetImportUnitPreviewResponse>({
        apiBase,
        operationID: "getImportUnitPreview",
        pathParameters: {
          import_session_id: sessionId,
          import_unit_id: unitId,
        },
      }),
  );
  return envelope.data;
}

export async function pollImportJob(
  availability: ExtensionAvailabilityController,
  apiBase: string | undefined,
  jobId: string,
  onJob?: ((job: ImportJobResource) => void) | undefined,
): Promise<ImportJobResource> {
  for (let attempt = 0; attempt < 30; attempt += 1) {
    const envelope = await requireImportResult<GetJobResponse>(
      availability,
      () =>
        fetchHTTPOperation<GetJobResponse>({
          apiBase,
          operationID: "getJob",
          pathParameters: { job_id: jobId },
        }),
    );
    const job = envelope.data;
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

async function requireImportResult<T>(
  availability: ExtensionAvailabilityController,
  request: () => Promise<APIResult<T>>,
): Promise<T> {
  const result = await runImportRequest(availability, request);
  if (!result.ok) {
    throw new Error(parseErrorMessage(result.payload));
  }
  return result.payload as T;
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
