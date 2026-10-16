import type { EvidenceAccessReasonCode } from "../../services/publicErrorIdentity";
import type {
  EvidenceRecordLifecycleState,
  ObjectBlobUploadState,
} from "../models/evidenceLifecycleViewModel";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";

export type EvidenceMetadata = Readonly<{
  recordId: string;
  rowVersion: number;
  title: string | null;
  lifecycleState: EvidenceRecordLifecycleState;
  hasBlob: boolean;
  uploadState: ObjectBlobUploadState | null;
  previewEligibility: Readonly<{
    eligible: boolean;
    reasonCode: EvidenceAccessReasonCode | null;
  }>;
}>;
export type EvidenceMetadataResult =
  | { readonly kind: "accepted"; readonly items: readonly EvidenceMetadata[] }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure }
  | { readonly kind: "aborted" };

/** Read authority only. Queries settle within 30 seconds; no access capabilities are minted. */
export interface WorkbookEvidenceMetadataPort {
  query(input: {
    readonly recordIds: readonly string[];
    readonly signal: AbortSignal;
  }): Promise<EvidenceMetadataResult>;
}
