import { buildHTTPOperationPath } from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation } from "../../services/browserApi";
import {
  type NoteAssociationTransport,
  noteAssociationView,
} from "../features/notes/noteAssociationOperation";
import { acceptWorkbookRowObservation } from "../query/acceptWorkbookRowObservation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { freezeWorkbookValue } from "../utils/freezeWorkbookValue";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";
import type {
  WorkbookProtocolNoteAssociationsPage,
  WorkbookProtocolNoteAssociationsReceipt,
  WorkbookProtocolNoteAssociationsRequest,
} from "./workbookProtocolTypes";
import { normalizeRecordMutationRow } from "./workbookRecordPatchTransport";

export function createNoteAssociationTransport(
  apiBase: string | undefined,
  readScope?: WorkbookReadScopeSource,
): NoteAssociationTransport {
  return {
    capture(review, clientTxnId) {
      const [first, ...rest] = review.actions;
      if (!first || review.actions.length > 64)
        throw new Error("Select between one and 64 association changes.");
      const request: WorkbookProtocolNoteAssociationsRequest = {
        kind: review.kind,
        base_row_version: review.row.row_version,
        client_txn_id: clientTxnId,
        actions: [first, ...rest],
      };
      return freezeWorkbookValue({
        review: structuredClone(review),
        clientTxnId,
        apiBase,
        request,
        body: JSON.stringify(request),
        path: buildHTTPOperationPath("mutateNoteAssociations", {
          note_record_id: review.row.record_id,
        }),
      });
    },
    async send(attempt, signal) {
      const scope = readScope?.() ?? null;
      let status: number | null = null,
        responseId: string | null = null;
      try {
        const pathParameters = { note_record_id: attempt.review.row.record_id };
        if (
          attempt.body !== JSON.stringify(attempt.request) ||
          attempt.clientTxnId !== attempt.request.client_txn_id ||
          attempt.path !==
            buildHTTPOperationPath("mutateNoteAssociations", pathParameters)
        )
          return { kind: "uncertain" };
        const result =
          await fetchHTTPOperation<WorkbookProtocolNoteAssociationsReceipt>({
            apiBase: attempt.apiBase,
            operationID: "mutateNoteAssociations",
            pathParameters,
            init: { method: "POST", body: attempt.body, signal },
            onResponse: (response) => {
              status = response.status;
              responseId = response.headers.get("X-Request-ID");
            },
          });
        if (!result.ok) {
          if (
            status === null ||
            status < 400 ||
            status >= 500 ||
            !result.payload.error?.code ||
            (responseId !== null &&
              responseId !== result.payload.error.request_id)
          )
            return { kind: "uncertain" };
          const failure = classifyWorkbookOperationFailure(
            result.status,
            result.payload,
            "mutateNoteAssociations",
          );
          return failure.kind === "invalid_contract"
            ? { kind: "uncertain" }
            : { kind: "rejected", failure };
        }
        const receipt = result.payload,
          row = normalizeRecordMutationRow(
            receipt.data.row,
            noteAssociationView,
            attempt.review.row.record_id,
          );
        // Duplicate additions have a valid retained no-op receipt at the base
        // version. A changed receipt must advance it and carry its change set.
        if (
          status !== 200 ||
          !receipt.meta?.request_id?.trim() ||
          (responseId !== null && responseId !== receipt.meta.request_id) ||
          receipt.data.view_schema_id !== noteAssociationView ||
          !row ||
          (receipt.data.change_set_id
            ? row.row_version <= attempt.request.base_row_version
            : row.row_version !== attempt.request.base_row_version)
        )
          return { kind: "uncertain" };
        return {
          kind: "accepted",
          receipt: freezeWorkbookValue(
            structuredClone({
              ...receipt,
              data: {
                ...receipt.data,
                row: acceptWorkbookRowObservation(row, scope),
              },
            }),
          ),
        };
      } catch {
        return { kind: "uncertain" };
      }
    },
    async list(noteId, kind, cursor, signal) {
      try {
        const result =
          await fetchHTTPOperation<WorkbookProtocolNoteAssociationsPage>({
            apiBase,
            operationID: "listNoteAssociations",
            pathParameters: { note_record_id: noteId },
            query: {
              kind,
              limit: 100,
              ...(cursor ? { cursor_token: cursor } : {}),
            },
            init: { method: "GET", signal },
          });
        if (!result.ok)
          return {
            kind: "rejected",
            failure: classifyWorkbookOperationFailure(
              result.status,
              result.payload,
              "listNoteAssociations",
            ),
          };
        if (
          result.payload.data.note_record_id !== noteId ||
          result.payload.data.kind !== kind ||
          (cursor !== null &&
            result.payload.data.next_cursor_token === cursor) ||
          result.payload.meta.paging?.limit !== 100 ||
          result.payload.meta.paging?.has_more !==
            (result.payload.data.next_cursor_token !== null) ||
          result.payload.meta.paging?.next_cursor !==
            result.payload.data.next_cursor_token
        )
          return {
            kind: "rejected",
            failure: {
              kind: "invalid_contract",
              message: "The association response did not match this Note.",
            },
          };
        return {
          kind: "accepted",
          page: freezeWorkbookValue(result.payload.data),
        };
      } catch {
        return {
          kind: "rejected",
          failure: {
            kind: "retryable",
            message: "Could not load Note associations. Retry the read.",
          },
        };
      }
    },
  };
}
