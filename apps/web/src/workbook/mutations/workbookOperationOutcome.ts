import type { CartularyErrorPresentation } from "@cartulary/ui-contracts";
import type {
  HTTPOperationErrorCode,
  HTTPOperationID,
} from "../../services/browserApi";
import type { PublicErrorReasonCode } from "../../services/publicErrorIdentity";
import type { EvidenceUploadFailure } from "../../services/workbookEvidence";
import type { WorkbookSameFieldConflictPayload } from "../runtime/workbookConflictModel";

export type WorkbookOperationFieldFailure = {
  readonly field: string;
  readonly message: string;
};

/**
 * A classified Workbook operation failure. `publicCode` is the decoded server
 * code, typed from the generated bindings of the operations that produced
 * it; client-originated and contract failures carry none.
 */
export type WorkbookOperationFailure<
  OperationID extends HTTPOperationID = HTTPOperationID,
> = (
  | {
      readonly kind: "validation";
      readonly message: string;
      readonly fields?: readonly WorkbookOperationFieldFailure[];
    }
  | {
      readonly kind: "same_field_conflict";
      readonly message: string;
      readonly conflict: WorkbookSameFieldConflictPayload;
    }
  | { readonly kind: "client_txn_conflict"; readonly message: string }
  | { readonly kind: "authentication_required"; readonly message: string }
  | { readonly kind: "authorization_lost"; readonly message: string }
  | { readonly kind: "stale_target"; readonly message: string }
  | { readonly kind: "retryable"; readonly message: string }
  | { readonly kind: "invalid_contract"; readonly message: string }
  | { readonly kind: "terminal"; readonly message: string }
) & {
  readonly presentation?: CartularyErrorPresentation | undefined;
  readonly publicCode?: HTTPOperationErrorCode<OperationID> | undefined;
  readonly publicReason?: PublicErrorReasonCode | undefined;
  readonly uploadFailure?: EvidenceUploadFailure | undefined;
};

/** A client-originated Workbook failure: no request decided it, so it carries no public code. */
export type WorkbookClientFailure = WorkbookOperationFailure<never>;

export type WorkbookOperationOutcome<
  Accepted,
  OperationID extends HTTPOperationID = HTTPOperationID,
> =
  | { readonly kind: "accepted"; readonly value: Accepted }
  | {
      readonly kind: "rejected";
      readonly failure: WorkbookOperationFailure<OperationID>;
    };
