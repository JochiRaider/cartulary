import {
  buildHTTPOperationPath,
  type CreateViewRowResponse,
} from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation } from "../../services/browserApi";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import { acceptWorkbookRowObservation } from "../query/acceptWorkbookRowObservation";
import type { WorkbookReadScope } from "../query/WorkbookQueryRow";
import { freezeWorkbookValue } from "../utils/freezeWorkbookValue";
import { explicitWorkbookFailureOutcome } from "./workbookOperationErrorPolicy";
import { acceptedRecordMutation } from "./workbookRecordPatchTransport";

type RecordMutationOperation =
  | "createViewRow"
  | "createRecordLinkedNote"
  | "patchRecord";
type Outcome<OperationID extends RecordMutationOperation> =
  | {
      readonly kind: "accepted";
      readonly receipt: CreateViewRowResponse;
      readonly status: number;
    }
  | {
      readonly kind: "rejected";
      readonly failure: WorkbookOperationFailure<OperationID>;
    }
  | { readonly kind: "uncertain" };

/** A complete correlated envelope is the write checkpoint, independent of any query. */
export async function sendWorkbookRecordMutation<
  OperationID extends RecordMutationOperation,
>(
  input: {
    readonly operationID: OperationID;
    readonly apiBase: string | undefined;
    readonly path: string;
    readonly pathParameters: Readonly<Record<string, string>>;
    readonly body: string;
    readonly clientTxnId: string;
    readonly viewSchemaId: string;
    readonly recordId?: string;
    readonly baseRowVersion?: number;
  },
  signal: AbortSignal,
  scope: WorkbookReadScope | null = null,
): Promise<Outcome<OperationID>> {
  let status: number | null = null,
    requestId: string | null = null;
  try {
    if (
      input.path !==
        buildHTTPOperationPath(input.operationID, input.pathParameters) ||
      JSON.parse(input.body).client_txn_id !== input.clientTxnId
    )
      return { kind: "uncertain" };
    const result = await fetchHTTPOperation<CreateViewRowResponse, OperationID>(
      {
        apiBase: input.apiBase,
        operationID: input.operationID,
        pathParameters: input.pathParameters,
        init: {
          method: input.operationID === "patchRecord" ? "PATCH" : "POST",
          body: input.body,
          signal,
        },
        onResponse: (response) => {
          status = response.status;
          requestId = response.headers.get("X-Request-ID");
        },
      },
    );
    if (!result.ok)
      return explicitWorkbookFailureOutcome(
        status,
        result.payload.error,
        input.operationID,
        requestId,
      );
    const receipt = result.payload;
    if (
      (status !== 200 && status !== 201) ||
      typeof receipt.meta?.request_id !== "string" ||
      !receipt.meta.request_id.trim() ||
      (requestId !== null && requestId !== receipt.meta.request_id) ||
      !acceptedRecordMutation(
        receipt.data,
        input.viewSchemaId,
        input.recordId,
      ) ||
      (input.baseRowVersion !== undefined &&
        receipt.data.row.row_version <= input.baseRowVersion)
    )
      return { kind: "uncertain" };
    return {
      kind: "accepted",
      status,
      receipt: freezeWorkbookValue(
        structuredClone({
          ...receipt,
          data: {
            ...receipt.data,
            row: acceptWorkbookRowObservation(receipt.data.row, scope),
          },
        }),
      ),
    };
  } catch {
    return { kind: "uncertain" };
  }
}
