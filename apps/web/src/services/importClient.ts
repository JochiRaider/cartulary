import type { HTTPOperationID } from "@cartulary/protocol-ts/http";
import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import {
  importProfileId,
  importRouteFamily,
} from "../extensions/extensionWorkspaceIdentities";
import {
  type ImportWriteAttempt,
  immutableImportValue,
} from "../imports/importRequests";
import {
  fetchHTTPOperation,
  fetchMultipartHTTPOperation,
  type HTTPOperationResult,
  publicErrorView,
} from "./browserApi";
import { equalJSONResource } from "./commonJobContract";
import type {
  DiscoveredImportPreview,
  DiscoveredImportUnit,
  ExtensionMappingPreviewRequest,
  ExtensionMappingPreviewResource,
  ImportJobResource,
  ImportSelectionReceipt,
  ImportSessionResource,
  ListImportUnitsResponse,
} from "./importContractAdapter";
import { validWorkbookImportJob } from "./importJobContract";

export type ImportFailure = {
  readonly kind: "public" | "transport" | "contract" | "authority";
  readonly status: number;
  readonly code: string;
  readonly reason: string | null;
  readonly field: string | null;
  readonly retryable: boolean;
};
export type ImportReadResult<T> =
  | { readonly kind: "received"; readonly value: T }
  | { readonly kind: "failed"; readonly failure: ImportFailure };
export type ImportWriteReceipt =
  | { readonly kind: "job"; readonly job: ImportJobResource }
  | { readonly kind: "unit"; readonly unit: DiscoveredImportUnit }
  | { readonly kind: "selection"; readonly selection: ImportSelectionReceipt };
export type ImportWriteResult =
  | { readonly kind: "accepted"; readonly receipt: ImportWriteReceipt }
  | {
      readonly kind: "rejected" | "uncertain";
      readonly failure: ImportFailure;
    };

export const importContractFailure = (): ImportFailure => ({
  kind: "contract",
  status: 0,
  code: "invalid_public_contract_response",
  reason: null,
  field: null,
  retryable: true,
});
export const importInterruptedFailure = (): ImportFailure => ({
  kind: "transport",
  status: 0,
  code: "observation_interrupted",
  reason: null,
  field: null,
  retryable: true,
});
const failed = <T>(failure: ImportFailure): ImportReadResult<T> => ({
  kind: "failed",
  failure,
});
const received = <T>(value: T): ImportReadResult<T> => ({
  kind: "received",
  value: immutableImportValue(value),
});

export function importFailureMessage(failure: ImportFailure): string {
  if (failure.status === 401)
    return "Your session needs to be confirmed. Sign in again if requested.";
  if (failure.status === 403)
    return "This action is not currently authorized. Refresh access before continuing.";
  if (failure.code === "incident_closed")
    return "Closed, read-only. Your local mapping is retained; no source changes can be submitted.";
  if (failure.kind === "authority")
    return "Import access changed. Refresh access before continuing.";
  if (failure.kind === "contract")
    return "The server response could not be validated. Refresh the known resource or retry the exact unresolved request.";
  if (failure.kind === "transport")
    return "The response could not be confirmed. Server work may continue; use the recovery action below.";
  const reasons: Record<string, string> = {
    duplicate_target_field: "Each target field can be mapped only once.",
    overlapping_units:
      "Selected units overlap. Skip an overlapping unit before applying.",
    duplicate_apply_blocked:
      "This mapping was already applied. Intentional re-import requires a new import session.",
    unit_not_ready:
      "A selected unit is not ready. Review its mapping and diagnostics.",
    session_applying:
      "This session is already applying. Refresh its known job and outcomes.",
    session_terminal:
      "This session has finished. Review its outcomes or start a new import.",
    unit_terminal: "This unit has finished and cannot be changed.",
    already_terminal:
      "The job finished before cancellation. Refresh to review its outcome.",
    already_cancel_requested:
      "Cancellation has already been requested. Refresh to observe the outcome.",
    not_cancelable: "The current job phase does not accept cancellation.",
    source_changed:
      "The source or approved mapping changed. Refresh the session before continuing.",
    field_not_import_writable:
      "The selected field does not permit import values.",
    invalid_source_columns:
      "Review the source-column mapping and unmapped columns.",
    invalid_source_rect:
      "Enter a whole-number rectangle inside the worksheet preview.",
    formula_cached_value_missing:
      "A formula has no cached value. Exclude or remap the affected column.",
  };
  if (failure.reason !== null && reasons[failure.reason])
    return reasons[failure.reason] as string;
  if (failure.status === 404)
    return "This resource is unavailable or no longer visible. Committed work is not necessarily removed.";
  if (failure.code === "client_txn_conflict")
    return "This request identity conflicts with a previous request. Refresh before starting a new action.";
  if (failure.code === "import_source_rejected")
    return "The source exceeds the permitted import or archive limits.";
  if (failure.code === "import_source_unsupported")
    return "This source cannot be imported. Review workbook format, encryption, and formula warnings.";
  return "The request was rejected. Review the affected mapping or refresh its current state.";
}

/** Rejects late resolution locally even if the fetch implementation ignores abort. */
function abortable<T>(promise: Promise<T>, signal: AbortSignal): Promise<T> {
  return new Promise((resolve, reject) => {
    const stop = () =>
      reject(new DOMException("Request interrupted", "AbortError"));
    if (signal.aborted) {
      stop();
      return;
    }
    signal.addEventListener("abort", stop, { once: true });
    promise
      .then(resolve, reject)
      .finally(() => signal.removeEventListener("abort", stop));
  });
}

export class ImportClient {
  constructor(
    private readonly options: {
      readonly availability: ExtensionAvailabilityController;
      readonly apiBase?: string | undefined;
      readonly incidentId: string;
      /** Revalidate consumer authority inside the extension dispatch queue. */
      readonly canDispatch?: () => boolean;
    },
  ) {}

  private async request<T>(
    operationID: HTTPOperationID,
    signal: AbortSignal,
    request: {
      readonly paths?: Record<string, string>;
      readonly query?: Record<string, string | number | undefined>;
      readonly method?: string;
      readonly body?: unknown;
      readonly form?: FormData;
    } = {},
  ): Promise<ImportReadResult<T>> {
    let status = 0;
    try {
      if (signal.aborted || this.options.canDispatch?.() === false)
        return failed(importInterruptedFailure());
      const result = await abortable(
        this.options.availability.runProfileRequest(
          importProfileId,
          importRouteFamily,
          async () => {
            if (signal.aborted || this.options.canDispatch?.() === false)
              throw new DOMException("Request interrupted", "AbortError");
            const onResponse = (response: Response) => {
              status = response.status;
            };
            const operation =
              request.form === undefined
                ? fetchHTTPOperation<T>({
                    operationID,
                    apiBase: this.options.apiBase,
                    pathParameters: request.paths,
                    query: request.query,
                    init: {
                      signal,
                      cache: "no-store",
                      ...(request.method ? { method: request.method } : {}),
                      ...(request.body === undefined
                        ? {}
                        : { body: JSON.stringify(request.body) }),
                    },
                    onResponse,
                  })
                : fetchMultipartHTTPOperation<T>({
                    operationID,
                    apiBase: this.options.apiBase,
                    body: request.form,
                    init: { signal },
                    onResponse,
                  });
            return abortable(operation, signal);
          },
        ),
        signal,
      );
      if (signal.aborted || this.options.canDispatch?.() === false)
        return failed(importInterruptedFailure());
      if (result.ok) {
        const statuses =
          operationID === "createImportSession" ||
          operationID === "applyImportSession"
            ? [202]
            : operationID === "createImportUnitRegion"
              ? [200, 201]
              : [200];
        return statuses.includes(status || result.status)
          ? received(result.payload)
          : failed(importContractFailure());
      }
      return failed(this.failure(result, status));
    } catch {
      if (status >= 400 && status < 500 && status !== 408)
        return failed({
          kind: "public",
          status,
          code: "request_rejected",
          reason: null,
          field: null,
          retryable: false,
        });
      if (
        !this.options.availability.isRouteAvailable(
          importProfileId,
          importRouteFamily,
        )
      )
        return failed({
          kind: "authority",
          status,
          code: "extension_profile_not_claimed",
          reason: null,
          field: null,
          retryable: false,
        });
      return failed(importInterruptedFailure());
    }
  }

  private failure(
    result: Extract<HTTPOperationResult<unknown>, { ok: false }>,
    status: number,
  ): ImportFailure {
    const view = publicErrorView(result.payload.error, status || result.status);
    const contract =
      result.payload.error?.code === "invalid_public_contract_response";
    return {
      kind: contract ? "contract" : "public",
      status: status || result.status,
      code: view?.code ?? "request_rejected",
      reason:
        view?.details.find((item) => item.key === "reason_code")?.value ?? null,
      field: view?.details.find((item) => item.key === "field")?.value ?? null,
      retryable: result.payload.error?.retryable === true,
    };
  }

  async send(
    attempt: ImportWriteAttempt,
    signal: AbortSignal,
    replay = false,
  ): Promise<ImportWriteResult> {
    if (attempt.scope.incidentId !== this.options.incidentId)
      return { kind: "rejected", failure: importContractFailure() };
    let operationID: HTTPOperationID;
    let method: string;
    let paths: Record<string, string> = {};
    let form: FormData | undefined;
    if (attempt.kind === "upload") {
      operationID = "createImportSession";
      method = "POST";
      form = new FormData();
      form.append(
        "metadata",
        new Blob([JSON.stringify(attempt.metadata)], {
          type: "application/json",
        }),
      );
      form.append("file", attempt.file, attempt.filename);
    } else if (attempt.kind === "cancel") {
      operationID = "cancelJob";
      method = "POST";
      paths = { job_id: attempt.jobId };
    } else {
      paths = { import_session_id: attempt.sessionId };
      if (attempt.kind === "apply") {
        operationID = "applyImportSession";
        method = "POST";
      } else if (attempt.kind === "region") {
        operationID = "createImportUnitRegion";
        method = "POST";
        paths.base_unit_id = attempt.unitId;
      } else {
        paths.import_unit_id = attempt.unitId;
        operationID =
          attempt.kind === "mapping"
            ? "putImportUnitMapping"
            : attempt.kind === "select"
              ? "selectImportUnit"
              : "skipImportUnit";
        method = attempt.kind === "mapping" ? "PUT" : "POST";
      }
    }
    const result = await this.request<{
      data: ImportJobResource | DiscoveredImportUnit | ImportSelectionReceipt;
    }>(operationID, signal, {
      paths,
      method,
      ...(form
        ? { form }
        : { body: "body" in attempt ? attempt.body : undefined }),
    });
    if (result.kind === "failed")
      return {
        kind:
          result.failure.kind === "public" &&
          result.failure.status >= 400 &&
          result.failure.status < 500 &&
          result.failure.status !== 408
            ? "rejected"
            : "uncertain",
        failure: result.failure,
      };
    const data = result.value.data;
    if (
      attempt.kind === "upload" ||
      attempt.kind === "apply" ||
      attempt.kind === "cancel"
    ) {
      const job = data as ImportJobResource;
      if (
        !validWorkbookImportJob(
          job,
          attempt.scope.incidentId,
          attempt.kind === "cancel" ? attempt.jobId : undefined,
          attempt.kind === "apply" ? attempt.sessionId : undefined,
          attempt.kind === "cancel"
            ? undefined
            : attempt.kind === "upload"
              ? "discovery"
              : "apply",
        ) ||
        (attempt.kind !== "cancel" &&
          !replay &&
          !["queued", "running"].includes(job.status)) ||
        (attempt.kind === "cancel" &&
          !["cancel_requested", "canceled", "failed", "succeeded"].includes(
            job.status,
          )) ||
        (attempt.kind === "upload" &&
          job.submitted_by_user_id !== attempt.scope.actorId)
      )
        return { kind: "uncertain", failure: importContractFailure() };
      return { kind: "accepted", receipt: { kind: "job", job } };
    }
    if (attempt.kind === "select" || attempt.kind === "skip") {
      const selection = data as ImportSelectionReceipt;
      if (
        selection.import_session_id !== attempt.sessionId ||
        !validImportUnit(selection.unit, attempt.sessionId, attempt.unitId) ||
        new Set(selection.selected_unit_ids).size !==
          selection.selected_unit_ids.length ||
        selection.selected_unit_ids.includes(attempt.unitId) !==
          (attempt.kind === "select")
      )
        return { kind: "uncertain", failure: importContractFailure() };
      return { kind: "accepted", receipt: { kind: "selection", selection } };
    }
    const unit = data as DiscoveredImportUnit;
    if (
      !validImportUnit(
        unit,
        attempt.sessionId,
        attempt.kind === "region" ? undefined : attempt.unitId,
      ) ||
      (attempt.kind === "region" &&
        (unit.locator_kind !== "operator_region" ||
          unit.approved_mapping !== undefined ||
          unit.unit_status !== "discovered")) ||
      (attempt.kind === "mapping" &&
        (unit.approved_mapping === undefined ||
          !mappingTargetMatches(attempt.body, unit)))
    )
      return { kind: "uncertain", failure: importContractFailure() };
    return { kind: "accepted", receipt: { kind: "unit", unit } };
  }

  async readJob(
    jobId: string,
    signal: AbortSignal,
    sessionId?: string,
    purpose?: "discovery" | "apply",
  ): Promise<ImportReadResult<ImportJobResource>> {
    const r = await this.request<{ data: ImportJobResource }>(
      "getJob",
      signal,
      { paths: { job_id: jobId } },
    );
    if (r.kind === "failed") return r;
    return validWorkbookImportJob(
      r.value.data,
      this.options.incidentId,
      jobId,
      sessionId,
      purpose,
    )
      ? received(r.value.data)
      : failed(importContractFailure());
  }
  async readSession(
    sessionId: string,
    signal: AbortSignal,
  ): Promise<ImportReadResult<ImportSessionResource>> {
    const r = await this.request<{ data: ImportSessionResource }>(
      "getImportSession",
      signal,
      { paths: { import_session_id: sessionId } },
    );
    if (r.kind === "failed") return r;
    const s = r.value.data;
    return s.import_session_id === sessionId &&
      s.incident_id === this.options.incidentId &&
      new Set(s.selected_unit_ids).size === s.selected_unit_ids.length
      ? received(s)
      : failed(importContractFailure());
  }
  async listUnits(
    sessionId: string,
    signal: AbortSignal,
  ): Promise<ImportReadResult<readonly DiscoveredImportUnit[]>> {
    const units: DiscoveredImportUnit[] = [],
      cursors = new Set<string>(),
      ids = new Set<string>();
    let cursor: string | undefined;
    for (;;) {
      const r = await this.request<ListImportUnitsResponse>(
        "listImportUnits",
        signal,
        {
          paths: { import_session_id: sessionId },
          query: { limit: 50, cursor_token: cursor },
        },
      );
      if (r.kind === "failed") return r;
      for (const unit of r.value.data.import_units) {
        if (!validImportUnit(unit, sessionId) || ids.has(unit.import_unit_id))
          return failed(importContractFailure());
        ids.add(unit.import_unit_id);
        units.push(unit);
      }
      const paging = r.value.meta.paging;
      if (paging === undefined) return failed(importContractFailure());
      if (!paging.has_more)
        return paging.next_cursor === null
          ? received(units)
          : failed(importContractFailure());
      if (
        typeof paging.next_cursor !== "string" ||
        paging.next_cursor.trim() === "" ||
        cursors.has(paging.next_cursor) ||
        r.value.data.import_units.length === 0
      )
        return failed(importContractFailure());
      cursor = paging.next_cursor;
      cursors.add(cursor);
    }
  }
  async readUnit(
    sessionId: string,
    unitId: string,
    signal: AbortSignal,
  ): Promise<ImportReadResult<DiscoveredImportUnit>> {
    const r = await this.request<{ data: DiscoveredImportUnit }>(
      "getImportUnit",
      signal,
      { paths: { import_session_id: sessionId, import_unit_id: unitId } },
    );
    if (r.kind === "failed") return r;
    return validImportUnit(r.value.data, sessionId, unitId)
      ? received(r.value.data)
      : failed(importContractFailure());
  }
  async preview(
    unit: DiscoveredImportUnit,
    signal: AbortSignal,
  ): Promise<ImportReadResult<DiscoveredImportPreview>> {
    const r = await this.request<{ data: DiscoveredImportPreview }>(
      "getImportUnitPreview",
      signal,
      {
        paths: {
          import_session_id: unit.import_session_id,
          import_unit_id: unit.import_unit_id,
        },
      },
    );
    if (r.kind === "failed") return r;
    const p = r.value.data;
    if (
      p.import_session_id !== unit.import_session_id ||
      p.import_unit_id !== unit.import_unit_id ||
      p.source_rect_a1 !== unit.source_rect_a1 ||
      p.locator_kind !== unit.locator_kind ||
      !equalJSONResource(p.locator, unit.locator) ||
      p.header_row_ref !== unit.header_row_ref ||
      p.data_start_row_ref !== unit.data_start_row_ref ||
      p.inferred_column_count !== unit.inferred_column_count ||
      p.columns.length !== unit.inferred_column_count ||
      p.preview_rows.length > 50 ||
      p.columns.some((c, i) => c.source_column_ordinal !== i + 1) ||
      p.preview_rows.some(
        (row, i, rows) =>
          row.source_row_ref < p.data_start_row_ref ||
          (i > 0 && row.source_row_ref <= (rows[i - 1]?.source_row_ref ?? 0)) ||
          row.cells.length !== p.columns.length ||
          row.cells.some((cell, j) => cell.source_column_ordinal !== j + 1),
      )
    )
      return failed(importContractFailure());
    return received(p);
  }
  async previewMapping(
    unit: DiscoveredImportUnit,
    body: ExtensionMappingPreviewRequest,
    signal: AbortSignal,
  ): Promise<ImportReadResult<ExtensionMappingPreviewResource<unknown>>> {
    const r = await this.request<{
      data: ExtensionMappingPreviewResource<unknown>;
    }>("previewImportUnitExtensionMapping", signal, {
      method: "POST",
      paths: {
        import_session_id: unit.import_session_id,
        import_unit_id: unit.import_unit_id,
      },
      body,
    });
    if (r.kind === "failed") return r;
    const p = r.value.data;
    return p.import_session_id === unit.import_session_id &&
      p.import_unit_id === unit.import_unit_id &&
      p.target_kind === body.target_kind &&
      p.extension_profile_id === body.extension_profile_id
      ? received(p)
      : failed(importContractFailure());
  }
}

export function validImportUnit(
  unit: DiscoveredImportUnit,
  sessionId: string,
  unitId?: string,
): boolean {
  return (
    (!["mapped", "ready", "applying", "applied"].includes(unit.unit_status) ||
      unit.approved_mapping !== undefined) &&
    (!["discovered", "selected"].includes(unit.unit_status) ||
      unit.approved_mapping === undefined) &&
    unit.import_session_id === sessionId &&
    (unitId === undefined || unit.import_unit_id === unitId) &&
    (unit.approved_mapping === undefined) ===
      (unit.mapping_fingerprint === undefined) &&
    (unit.mapping_fingerprint === undefined ||
      /^[a-f0-9]{64}$/.test(unit.mapping_fingerprint))
  );
}
function mappingTargetMatches(
  body: Extract<ImportWriteAttempt, { kind: "mapping" }>["body"],
  unit: DiscoveredImportUnit,
): boolean {
  const approved = unit.approved_mapping;
  if (!approved) return false;
  if ("target_view_schema_id" in body)
    return (
      "target_view_schema_id" in approved &&
      approved.target_view_schema_id === body.target_view_schema_id
    );
  return (
    "target_kind" in approved &&
    approved.target_kind === body.target_kind &&
    approved.extension_profile_id === body.extension_profile_id
  );
}

/** These source identities cannot change while reloading the same durable resource. */
export function sameImportSessionSource(
  previous: ImportSessionResource,
  next: ImportSessionResource,
): boolean {
  return (
    [
      "import_session_id",
      "incident_id",
      "created_by_user_id",
      "created_at",
      "source_file_kind",
      "original_filename",
      "source_content_sha256",
      "parser_profile_id",
      "parser_version",
      "assistant_profile",
    ] as const
  ).every((key) => previous[key] === next[key]);
}
export function sameImportUnitSource(
  previous: DiscoveredImportUnit,
  next: DiscoveredImportUnit,
): boolean {
  return (
    previous.import_session_id === next.import_session_id &&
    previous.import_unit_id === next.import_unit_id &&
    previous.locator_kind === next.locator_kind &&
    previous.source_rect_a1 === next.source_rect_a1 &&
    equalJSONResource(previous.locator, next.locator)
  );
}
