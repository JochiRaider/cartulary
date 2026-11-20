import type {
  HTTPOperationID,
  HTTPOperationResponse,
} from "@cartulary/protocol-ts/http";
import { importTargetRegistry } from "@cartulary/protocol-ts/import-targets";
import { requireViewContract } from "@cartulary/view-contracts";
import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { authHeadersForStorageState } from "../auth/storageState";
import { ownedJobsFor } from "../jobs/ownership";
import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import {
  publicHttpMultipartOperation,
  publicHttpOperation,
  publicHttpOperationFailure,
} from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";

type Job = HTTPOperationResponse<"getJob">["data"];

type OperationResult = {
  readonly ok: boolean;
  readonly payload: unknown;
  readonly status: number;
};

function requireSuccess<Result extends OperationResult>(
  operationID: HTTPOperationID,
  result: Result,
): Extract<Result, { readonly ok: true }>["payload"] {
  if (!result.ok)
    throw new Error(
      publicHttpOperationFailure(operationID, result.status, result.payload),
    );
  return (result as Extract<Result, { readonly ok: true }>).payload;
}

export type SeededWorkbookImport = {
  readonly importSessionId: string;
  readonly importUnitId: string;
};

/**
 * Applies one single-unit workbook source through the public import API, as
 * the Import Assistant would. The mapping is derived from the import-target
 * registry and the target view contract; `fieldKeys[i]` maps source column
 * `i + 1`. Each admitted job is owned by the page's context for teardown.
 */
export async function seedWorkbookImport(
  page: Page,
  options: {
    readonly incidentId: string;
    readonly targetViewSchemaId: string;
    readonly fieldKeys: readonly string[];
    readonly file: {
      readonly buffer: Buffer;
      readonly mimeType: string;
      readonly name: string;
    };
    readonly settlementTimeoutMs: number;
  },
): Promise<SeededWorkbookImport> {
  const headers = authHeadersForStorageState(
    await page.context().storageState(),
  );
  const request = atJsonOrigin(page.request, apiBase);
  const admitJob = async (dispatch: () => Promise<Job>) =>
    (
      await ownedJobsFor(page).admit(dispatch, async (job) => ({
        jobId: job.job_id,
        headers,
      }))
    ).job_id;
  const settleJob = async (jobId: string): Promise<Job> => {
    const observed: { job?: Job } = {};
    await expect
      .poll(
        async () => {
          observed.job = requireSuccess(
            "getJob",
            await publicHttpOperation({
              operationID: "getJob",
              headers,
              pathParameters: { job_id: jobId },
              request,
            }),
          ).data;
          return observed.job.status;
        },
        { timeout: options.settlementTimeoutMs },
      )
      .toMatch(/^(succeeded|failed|canceled)$/u);
    if (observed.job?.status !== "succeeded")
      throw new Error(`Workbook import job ${observed.job?.status}`);
    return observed.job;
  };

  const discovered = await settleJob(
    await admitJob(
      async () =>
        requireSuccess(
          "createImportSession",
          await publicHttpMultipartOperation({
            operationID: "createImportSession",
            file: options.file,
            headers,
            metadata: {
              assistant_profile: "workbook_import_v1",
              client_txn_id: uniqueTxn("ir-seed-upload"),
              incident_id: options.incidentId,
            },
            request,
          }),
        ).data,
    ),
  );
  const importSessionId = discovered.result_summary?.resource_refs?.find(
    (ref) => ref.kind === "import_session",
  )?.id;
  if (!importSessionId) throw new Error("Missing discovered import session");

  const units = requireSuccess(
    "listImportUnits",
    await publicHttpOperation({
      operationID: "listImportUnits",
      headers,
      pathParameters: { import_session_id: importSessionId },
      request,
    }),
  ).data.import_units;
  expect(units).toHaveLength(1);
  const importUnitId = units[0]?.import_unit_id;
  if (!importUnitId) throw new Error("Missing discovered import unit");
  const unit = {
    import_session_id: importSessionId,
    import_unit_id: importUnitId,
  };
  const preview = requireSuccess(
    "getImportUnitPreview",
    await publicHttpOperation({
      operationID: "getImportUnitPreview",
      headers,
      pathParameters: unit,
      request,
    }),
  ).data;

  const targets = importTargetRegistry.targets.filter(
    (target) =>
      target.target_view_schema_id === options.targetViewSchemaId &&
      target.public_projection_disposition === "selectable",
  );
  const target = targets[0];
  if (targets.length !== 1 || target === undefined)
    throw new Error("Missing selectable workbook import target");
  const fields = requireViewContract(options.targetViewSchemaId).fields;
  expect(preview.columns).toHaveLength(options.fieldKeys.length);
  const sourceColumns = preview.columns.map((column) => {
    const fieldKey = options.fieldKeys[column.source_column_ordinal - 1];
    const field = fields.find(
      (candidate) =>
        candidate.fieldKey === fieldKey && candidate.createWritable,
    );
    if (!field) throw new Error(`Unwritable import field ${fieldKey}`);
    return {
      source_column_ordinal: column.source_column_ordinal,
      source_header_text: column.source_header_text,
      field_key: field.fieldKey,
      entity_binding_mode: field.entityBindingMode,
      transform_id: null,
      transform_options: {},
      empty_value_policy: "omit_field",
    };
  });
  requireSuccess(
    "putImportUnitMapping",
    await publicHttpOperation({
      operationID: "putImportUnitMapping",
      body: {
        client_txn_id: uniqueTxn("ir-seed-mapping"),
        target_view_schema_id: options.targetViewSchemaId,
        unknown_column_policy: target.default_unknown_column_policy,
        header_row_ref: preview.header_row_ref,
        data_start_row_ref: preview.data_start_row_ref,
        source_columns: sourceColumns,
      },
      headers,
      pathParameters: unit,
      request,
    }),
  );
  requireSuccess(
    "selectImportUnit",
    await publicHttpOperation({
      operationID: "selectImportUnit",
      body: { client_txn_id: uniqueTxn("ir-seed-selection") },
      headers,
      pathParameters: unit,
      request,
    }),
  );
  await settleJob(
    await admitJob(
      async () =>
        requireSuccess(
          "applyImportSession",
          await publicHttpOperation({
            operationID: "applyImportSession",
            body: {
              client_txn_id: uniqueTxn("ir-seed-apply"),
              selected_unit_ids: [importUnitId],
            },
            headers,
            pathParameters: { import_session_id: importSessionId },
            request,
          }),
        ).data,
    ),
  );
  return { importSessionId, importUnitId };
}
