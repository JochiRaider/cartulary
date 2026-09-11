import { buildHTTPOperationPath } from "@cartulary/protocol-ts/http";
import { requireViewContract } from "@cartulary/view-contracts";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import { buildPatchRecordRequest } from "../models/workbookRequestDecoders";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookPendingMutationAccepted } from "../ports/WorkbookPendingMutationPort";
import type { WorkbookOperationExecutor } from "./workbookOperationContract";
import type { WorkbookProtocolPatchRecordRequest } from "./workbookProtocolTypes";

export type CapturedRecordPatch = Readonly<{
  recordId: string;
  viewSchemaId: string;
  baseRowVersion: number;
  clientTxnId: string;
  path: string;
  body: string;
}>;
export type RecordPatchOutcome =
  | {
      readonly kind: "acknowledged";
      readonly receipt: WorkbookPendingMutationAccepted;
    }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure }
  | { readonly kind: "uncertain" };
export interface RecordPatchTransport {
  send(
    request: CapturedRecordPatch,
    signal: AbortSignal,
  ): Promise<RecordPatchOutcome>;
}
export function captureRecordPatch(
  input: Parameters<typeof buildPatchRecordRequest>[0] & {
    readonly recordId: string;
  },
): CapturedRecordPatch | null {
  const request = buildPatchRecordRequest(input);
  return request
    ? Object.freeze({
        recordId: input.recordId,
        viewSchemaId: input.viewSchemaId,
        baseRowVersion: input.baseRowVersion,
        clientTxnId: input.clientTxnId,
        path: buildHTTPOperationPath("patchRecord", {
          record_id: input.recordId,
        }),
        body: JSON.stringify(request),
      })
    : null;
}
export function acceptedRecordMutation(
  data: unknown,
  viewSchemaId: string,
  recordId?: string,
): WorkbookPendingMutationAccepted | null {
  if (
    !data ||
    typeof data !== "object" ||
    !("view_schema_id" in data) ||
    data.view_schema_id !== viewSchemaId ||
    !("change_set_id" in data) ||
    typeof data.change_set_id !== "string" ||
    !data.change_set_id.trim() ||
    !("row" in data)
  )
    return null;
  const row = normalizeRecordMutationRow(data.row, viewSchemaId, recordId);
  return row ? { changeSetId: data.change_set_id, row, viewSchemaId } : null;
}
export function normalizeRecordMutationRow(
  raw: unknown,
  viewSchemaId: string,
  recordId?: string,
): WorkbookPendingMutationAccepted["row"] | null {
  try {
    const row = normalizeWorkbookViewRows(
      requireViewContract(viewSchemaId),
      [raw],
      "ordinary mutation receipt",
    )[0];
    return row &&
      row.view_schema_id === viewSchemaId &&
      row.row_version > 0 &&
      (recordId === undefined || row.record_id === recordId)
      ? { ...row, view_schema_id: viewSchemaId }
      : null;
  } catch {
    return null;
  }
}
export function createRecordPatchTransport(
  operations: WorkbookOperationExecutor,
): RecordPatchTransport {
  return {
    async send(captured, signal) {
      let status: number | null = null;
      try {
        if (
          captured.path !==
          buildHTTPOperationPath("patchRecord", {
            record_id: captured.recordId,
          })
        )
          return { kind: "uncertain" };
        const request: WorkbookProtocolPatchRecordRequest = JSON.parse(
          captured.body,
        );
        const result = await operations.execute({
          operationID: "patchRecord",
          pathParameters: { record_id: captured.recordId },
          request,
          signal,
          observeTransport: {
            onResponseStatus: (value) => {
              status = value;
            },
          },
        });
        if (result.kind === "rejected")
          return status === null ||
            status < 400 ||
            status >= 500 ||
            result.failure.kind === "invalid_contract"
            ? { kind: "uncertain" }
            : result;
        const receipt = acceptedRecordMutation(
          result.value.data,
          captured.viewSchemaId,
          captured.recordId,
        );
        return receipt && receipt.row.row_version > captured.baseRowVersion
          ? { kind: "acknowledged", receipt }
          : { kind: "uncertain" };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}
