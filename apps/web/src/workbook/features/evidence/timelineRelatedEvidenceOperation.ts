import type {
  WorkbookProtocolConflictResolutionReceipt,
  WorkbookProtocolCreateViewRowReceipt,
} from "../../adapters/workbookProtocolTypes";
import type { RecordChangedMessage } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { TimelineRelatedEvidenceDraft } from "./timelineRelatedEvidenceModel";

export type RelatedEvidenceStage = "create" | "link";
export type RelatedEvidenceReview = Readonly<{
  authority: WorkbookMutationAuthority;
  draft: TimelineRelatedEvidenceDraft;
  source: WorkbookQueryRow;
  presentationRevision: number;
}>;
export type RelatedEvidenceAttempt = Readonly<{
  stage: RelatedEvidenceStage;
  review: RelatedEvidenceReview;
  evidenceRecordId: string | null;
  clientTxnId: string;
  path: string;
  body: string;
  apiBase: string | undefined;
}>;
export type RelatedEvidenceReceipt =
  Readonly<WorkbookProtocolCreateViewRowReceipt>;
export type RelatedEvidenceOutcome =
  | { readonly kind: "accepted"; readonly receipt: RelatedEvidenceReceipt }
  | { readonly kind: "rejected"; readonly failure: WorkbookOperationFailure }
  | { readonly kind: "uncertain" };
export interface RelatedEvidenceTransport {
  capture(
    stage: RelatedEvidenceStage,
    review: RelatedEvidenceReview,
    clientTxnId: string,
    evidenceRecordId: string | null,
  ): RelatedEvidenceAttempt;
  send(
    attempt: RelatedEvidenceAttempt,
    signal: AbortSignal,
  ): Promise<RelatedEvidenceOutcome>;
}
export type RelatedEvidenceStageEntry = Readonly<{
  attempt: RelatedEvidenceAttempt;
  phase: "submitting" | "uncertain" | "rejected" | "accepted";
  transportPending: boolean;
  receipt: RelatedEvidenceReceipt | null;
  failure: WorkbookOperationFailure | null;
  uncertain: boolean;
  refresh: "none" | "required" | "refreshing" | "complete";
  observations: readonly RecordChangedMessage[];
}>;
export type RelatedEvidenceCheckpoint = Readonly<{
  id: string;
  observationRevision: number;
  create: RelatedEvidenceStageEntry;
  links: readonly RelatedEvidenceStageEntry[];
  review: RelatedEvidenceReview | null;
  message: string | null;
  sourceUnavailable: boolean;
  evidenceUnavailable: boolean;
  associationPresent: boolean;
  resolution?: Readonly<{
    kind: string;
    receipt: WorkbookProtocolConflictResolutionReceipt;
  }>;
}>;
