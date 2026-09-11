import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type { EntityMergePlan } from "../../models/entityMergePlan";

export type EntityMergeAuthority = {
  readonly actorId: string;
  readonly sessionIdentity: string;
  readonly incidentId: string;
  readonly role: WorkbookIncidentRole;
  readonly closed: boolean;
};

export type EntityMergeReviewedRecord = {
  readonly recordId: string;
  readonly label: string;
  readonly baseRowVersion: number;
};

export type EntityMergeReview = {
  readonly survivor: EntityMergeReviewedRecord;
  readonly loser: EntityMergeReviewedRecord;
  readonly entityType: "host" | "identity";
  readonly authority: EntityMergeAuthority;
  readonly authorityGeneration: number;
  readonly originSurface: string;
  readonly presentationGeneration: number;
  readonly lifecycleKey: string;
  readonly reason: string;
  readonly plan: EntityMergePlan;
};
