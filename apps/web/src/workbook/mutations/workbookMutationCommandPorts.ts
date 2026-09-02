import type { ViewContract } from "@cartulary/view-contracts";
import type { WorkbookOperationResponse } from "../adapters/workbookOperationExecutor";
import type {
  RecordHistoryData,
  RecordHistoryRollbackTarget,
} from "../inspector/workbookRecordHistoryModel";
import type { AssessmentCreateDraft } from "../models/assessmentWorkbookModel";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type {
  TimelineApiRow,
  WorkbookRow,
} from "../timeline/models/workbookTimelineModel";
import type { WorkbookOperationOutcome } from "./workbookOperationOutcome";

export type GenericViewMutationAccepted = {
  readonly changeSetId: string;
  readonly row: WorkbookQueryRow;
  readonly viewSchemaId: string;
};

export type GenericMutationOutcome =
  WorkbookOperationOutcome<GenericViewMutationAccepted>;

export type RecordLifecycleAccepted = {
  readonly recordId: string;
  readonly rowVersion: number;
};

export interface RecordRouteCommandPort {
  execute(input: {
    readonly action: "delete" | "restore";
    readonly baseRowVersion: number;
    readonly reason: string;
    readonly recordId: string;
  }): Promise<WorkbookOperationOutcome<RecordLifecycleAccepted>>;
  loadHistory(input: {
    readonly recordId: string;
  }): Promise<WorkbookOperationOutcome<RecordHistoryData>>;
  rollback(input: {
    readonly baseRowVersion: number;
    readonly reason: string;
    readonly recordId: string;
    readonly target: RecordHistoryRollbackTarget;
  }): Promise<WorkbookOperationOutcome<RecordLifecycleAccepted>>;
}

export type EntityCreateAccepted = {
  readonly changeSetId: string;
  readonly row: WorkbookQueryRow;
  readonly viewSchemaId: string;
};

export type EntityPatchAccepted = {
  readonly changeSetId: string;
  readonly row: WorkbookQueryRow;
  readonly viewSchemaId: string;
};

export type EntityPasteAccepted = {
  readonly changeSetId: string | null;
  readonly rows: readonly WorkbookQueryRow[];
  readonly viewSchemaId: string;
};

export type EntityCreateOutcome =
  WorkbookOperationOutcome<EntityCreateAccepted>;
export type EntityPatchOutcome = WorkbookOperationOutcome<EntityPatchAccepted>;
export type EntityPasteOutcome = WorkbookOperationOutcome<EntityPasteAccepted>;

export type EntityMergeAccepted = {
  readonly changeSetId: string;
  readonly loserRecordId: string;
  readonly loserRowVersion: number;
  readonly mergedIntoRecordId: string;
  readonly recordType: "host" | "identity";
  readonly survivorRecordId: string;
  readonly survivorRowVersion: number;
};

export type EntityMergeOutcome = WorkbookOperationOutcome<EntityMergeAccepted>;

export type AssessmentCreateAccepted = {
  readonly changeSetId: string;
  readonly row: WorkbookQueryRow;
  readonly viewSchemaId: string;
};

export type AssessmentCreateOutcome =
  WorkbookOperationOutcome<AssessmentCreateAccepted>;

export type EvidenceAttachAccepted = {
  readonly evidenceRecordId: string;
};

export type EvidenceHandleAccepted = {
  readonly filename: string;
  readonly href: string;
  readonly previewKind: string | null;
};

export type EvidenceAttachOutcome =
  WorkbookOperationOutcome<EvidenceAttachAccepted>;
export type EvidenceHandleOutcome =
  WorkbookOperationOutcome<EvidenceHandleAccepted>;

export type TaskLifecycleStatus =
  | "open"
  | "in_progress"
  | "blocked"
  | "done"
  | "canceled";

export type TaskLifecycleAccepted = {
  readonly changeSetId: string;
  readonly row: WorkbookQueryRow;
  readonly status: TaskLifecycleStatus;
  readonly viewSchemaId: string;
};

export type DecisionSupersedeAccepted = {
  readonly changeSetId: string;
  readonly replacementRecordId: string;
  readonly replacementRowVersion: number;
  readonly targetRecordId: string;
  readonly targetRowVersion: number;
  readonly targetStatus: string;
  readonly viewSchemaId: string;
};

export type TaskLifecycleOutcome =
  WorkbookOperationOutcome<TaskLifecycleAccepted>;
export type DecisionSupersedeOutcome =
  WorkbookOperationOutcome<DecisionSupersedeAccepted>;

export interface TimelineMutationIdentityPort {
  createLogicalActionId(): string;
  createConflictRecoveryId(): string;
}

export type TimelineBulkMutationAccepted = {
  readonly affectedRowCount: number;
  readonly changeSetId: string | null;
  readonly conflictCount: number;
};

export type TimelineBulkMutationOutcome =
  WorkbookOperationOutcome<TimelineBulkMutationAccepted>;

export interface TimelineBulkMutationPort {
  assignTag(input: {
    readonly tagName: string;
    readonly targets: readonly {
      readonly recordId: string;
      readonly baseRowVersion: number;
    }[];
  }): Promise<TimelineBulkMutationOutcome>;
  fillDown(input: {
    readonly fieldKey: string;
    readonly onClientTxnId: (clientTxnId: string) => void;
    readonly value: string;
    readonly targets: readonly {
      readonly recordId: string;
      readonly baseRowVersion: number;
    }[];
  }): Promise<{
    readonly clientTxnId: string | null;
    readonly outcome: TimelineBulkMutationOutcome;
  }>;
}

export type TimelineRelatedRecordCreated = {
  readonly changeSetId: string;
  readonly recordId: string;
  readonly viewSchemaId: string;
};

export type TimelineRelatedEvidenceLinked = {
  readonly changeSetId: string;
  readonly row: TimelineApiRow;
  readonly viewSchemaId: string;
};

export interface TimelineRelatedRecordPort {
  createRelatedRecord(input: {
    readonly contract: ViewContract;
    readonly draft: Readonly<Record<string, string>>;
    readonly featureGroupKey: string;
  }): Promise<WorkbookOperationOutcome<TimelineRelatedRecordCreated>>;
  linkCreatedEvidence(input: {
    readonly sourceRow: WorkbookRow;
    readonly createdRecordId: string;
  }): Promise<WorkbookOperationOutcome<TimelineRelatedEvidenceLinked>>;
}

export type TimelineMutationCommandPorts = {
  readonly bulk: TimelineBulkMutationPort;
  readonly identity: TimelineMutationIdentityPort;
  readonly related: TimelineRelatedRecordPort;
};

export interface GenericMutationCommandPort {
  canCreateRecord(input: {
    readonly contract: ViewContract;
    readonly draft: Readonly<Record<string, string>>;
  }): boolean;
  createRecord(input: {
    readonly contract: ViewContract;
    readonly draft: Readonly<Record<string, string>>;
    readonly linkedNoteSourceRecordId: string;
  }): Promise<GenericMutationOutcome>;
  patchRecord(input: {
    readonly baseRowVersion: number;
    readonly changes: readonly Record<string, unknown>[];
    readonly purpose: string;
    readonly recordId: string;
    readonly viewSchemaId: string;
  }): Promise<GenericMutationOutcome>;
  createPartyFromText(input: {
    readonly originViewSchemaId: string;
    readonly rawText: string;
  }): Promise<GenericMutationOutcome>;
}

export interface EntityMutationCommandPort {
  canCreateRecord(input: {
    readonly contract: ViewContract;
    readonly draft: Readonly<Record<string, string>>;
  }): boolean;
  createRecord(input: {
    readonly contract: ViewContract;
    readonly draft: Readonly<Record<string, string>>;
  }): Promise<EntityCreateOutcome>;
  patchRecord(input: {
    readonly baseRowVersion: number;
    readonly changes: readonly Record<string, unknown>[];
    readonly purpose: string;
    readonly recordId: string;
    readonly viewSchemaId: string;
  }): Promise<EntityPatchOutcome>;
  pasteCreate(input: {
    readonly clipboardText: string;
    readonly columns: readonly string[];
    readonly format: string;
    readonly startFieldKey: string;
    readonly targetCount: number;
    readonly viewSchemaId: string;
  }): Promise<EntityPasteOutcome>;
  merge(input: {
    readonly loserBaseRowVersion: number;
    readonly loserRecordId: string;
    readonly reason: string;
    readonly survivorBaseRowVersion: number;
    readonly survivorRecordId: string;
  }): Promise<EntityMergeOutcome>;
}

export interface AssessmentMutationCommandPort {
  canCreate(input: { readonly draft: AssessmentCreateDraft }): boolean;
  create(input: {
    readonly draft: AssessmentCreateDraft;
  }): Promise<AssessmentCreateOutcome>;
}

export interface EvidenceCapabilityPort {
  attach(input: {
    readonly baseRowVersion: number;
    readonly evidenceRecordId: string;
    readonly file: File;
  }): Promise<EvidenceAttachOutcome>;
  issueHandle(input: {
    readonly evidenceRecordId: string;
    readonly kind: "download" | "preview";
  }): Promise<EvidenceHandleOutcome>;
}

export interface CoordinationMutationCommandPort {
  updateTaskLifecycle(input: {
    readonly baseRowVersion: number;
    readonly blockedReason?: string | undefined;
    readonly recordId: string;
    readonly status: TaskLifecycleStatus;
  }): Promise<TaskLifecycleOutcome>;
  supersedeDecision(input: {
    readonly baseRowVersion: number;
    readonly reason: string;
    readonly replacementRecordId: string;
    readonly targetRecordId: string;
  }): Promise<DecisionSupersedeOutcome>;
}

type IndicatorObservationListResponse =
  WorkbookOperationResponse<"listIndicatorObservations">;
type IndicatorLifecycleListResponse =
  WorkbookOperationResponse<"listIndicatorStateIntervals">;

export type IndicatorObservation =
  IndicatorObservationListResponse["data"]["observations"][number];
export type IndicatorStateInterval =
  IndicatorLifecycleListResponse["data"]["intervals"][number];
export type IndicatorLifecycleState = IndicatorStateInterval["lifecycle_state"];
export type IndicatorAffectedRecord =
  WorkbookOperationResponse<"createManualIndicatorObservation">["data"]["affected_records"][number];
export type IndicatorPaging = NonNullable<
  IndicatorObservationListResponse["meta"]["paging"]
>;

export type IndicatorPage<Resource> = {
  readonly items: readonly Resource[];
  readonly paging: IndicatorPaging | null;
};

export type IndicatorMutationAccepted<Resource> = {
  readonly affectedRecords: readonly IndicatorAffectedRecord[];
  readonly changeSetId: string;
  readonly replayed: boolean;
  readonly resource: Resource;
};

export interface IndicatorWorkflowPort {
  listSourceObservations(input: {
    readonly cursorToken?: string | undefined;
    readonly limit?: number | undefined;
    readonly sourceRecordId: string;
  }): Promise<WorkbookOperationOutcome<IndicatorPage<IndicatorObservation>>>;
  listObservations(input: {
    readonly cursorToken?: string | undefined;
    readonly indicatorRecordId: string;
    readonly limit?: number | undefined;
  }): Promise<WorkbookOperationOutcome<IndicatorPage<IndicatorObservation>>>;
  listStateIntervals(input: {
    readonly cursorToken?: string | undefined;
    readonly indicatorRecordId: string;
    readonly limit?: number | undefined;
  }): Promise<WorkbookOperationOutcome<IndicatorPage<IndicatorStateInterval>>>;
  createManualObservation(input: {
    readonly baseRowVersion: number;
    readonly parsedIndicatorType?:
      | Exclude<IndicatorObservation["parsed_indicator_type"], null>
      | undefined;
    readonly resolvedIndicatorRecordId?: string | undefined;
    readonly sourceFieldKey: string;
    readonly sourceRecordId: string;
    readonly spanEndByte: number;
    readonly spanStartByte: number;
  }): Promise<
    WorkbookOperationOutcome<IndicatorMutationAccepted<IndicatorObservation>>
  >;
  transitionObservation(input: {
    readonly action: "dismiss" | "resolve" | "restore";
    readonly baseRowVersion: number;
    readonly observationId: string;
    readonly resolvedIndicatorRecordId?: string | undefined;
  }): Promise<
    WorkbookOperationOutcome<IndicatorMutationAccepted<IndicatorObservation>>
  >;
  appendStateInterval(input: {
    readonly assessor: string | null;
    readonly baseRowVersion: number;
    readonly confidence: number | null;
    readonly indicatorRecordId: string;
    readonly lifecycleState: IndicatorLifecycleState;
    readonly rationale: string | null;
    readonly supportRefs: readonly string[];
    readonly validFrom: string;
    readonly validTo: string | null;
  }): Promise<
    WorkbookOperationOutcome<IndicatorMutationAccepted<IndicatorStateInterval>>
  >;
}

export type WorkbookMutationCommandPorts = {
  readonly records: RecordRouteCommandPort;
  readonly timeline: TimelineMutationCommandPorts;
  readonly generic: GenericMutationCommandPort;
  readonly entity: EntityMutationCommandPort;
  readonly assessment: AssessmentMutationCommandPort;
  readonly evidence: EvidenceCapabilityPort;
  readonly coordination: CoordinationMutationCommandPort;
  readonly indicators: IndicatorWorkflowPort;
};
