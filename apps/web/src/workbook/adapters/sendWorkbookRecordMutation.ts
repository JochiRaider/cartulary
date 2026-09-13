import {
  buildHTTPOperationPath,
  type CreateViewRowResponse,
} from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation } from "../../services/browserApi";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import { freezeWorkbookValue } from "../utils/freezeWorkbookValue";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";
import { acceptedRecordMutation } from "./workbookRecordPatchTransport";

type Outcome =
  | { readonly kind: "accepted"; readonly receipt: CreateViewRowResponse }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure }
  | { readonly kind: "uncertain" };

/** A complete correlated envelope is the write checkpoint, independent of any query. */
export async function sendWorkbookRecordMutation(
  input: {
    readonly operationID:
      | "createViewRow"
      | "createRecordLinkedNote"
      | "patchRecord";
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
): Promise<Outcome> {
  let status: number | null = null,
    requestId: string | null = null;
  try {
    if (
      input.path !==
        buildHTTPOperationPath(input.operationID, input.pathParameters) ||
      JSON.parse(input.body).client_txn_id !== input.clientTxnId
    )
      return { kind: "uncertain" };
    const result = await fetchHTTPOperation<CreateViewRowResponse>({
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
    });
    if (!result.ok) {
      if (
        status === null ||
        status < 400 ||
        status >= 500 ||
        !result.payload.error?.code ||
        (requestId !== null && requestId !== result.payload.error.request_id)
      )
        return { kind: "uncertain" };
      const failure = classifyWorkbookOperationFailure(
        result.status,
        result.payload,
        input.operationID,
      );
      return failure.kind === "invalid_contract"
        ? { kind: "uncertain" }
        : { kind: "rejected", failure };
    }
    const receipt = result.payload;
    if (
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
      receipt: freezeWorkbookValue(structuredClone(receipt)),
    };
  } catch {
    return { kind: "uncertain" };
  }
}
