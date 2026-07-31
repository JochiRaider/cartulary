import type { ViewContract } from "@cartulary/view-contracts";
import type { AssessmentCreateDraft } from "../models/assessmentWorkbookModel";
import type { WorkbookRow } from "../timeline/models/workbookTimelineModel";

export type WorkbookMutationCommandResult<T = unknown> = {
  readonly ok: boolean;
  readonly status: number;
  readonly payload: T | { readonly error?: { readonly message?: string } };
};

export interface TimelineMutationCommandPort {
  createLogicalActionId(): string;
  createConflictRecoveryId(): string;
  assignTag(input: {
    readonly tagName: string;
    readonly targets: readonly {
      readonly recordId: string;
      readonly baseRowVersion: number;
    }[];
  }): Promise<WorkbookMutationCommandResult>;
  fillDown(input: {
    readonly fieldKey: string;
    readonly onClientTxnId: (clientTxnId: string) => void;
    readonly value: string;
    readonly targets: readonly {
      readonly recordId: string;
      readonly baseRowVersion: number;
    }[];
  }): Promise<
    WorkbookMutationCommandResult & { readonly clientTxnId: string | null }
  >;
  createRelatedRecord(input: {
    readonly contract: ViewContract;
    readonly draft: Readonly<Record<string, string>>;
    readonly featureGroupKey: string;
  }): Promise<WorkbookMutationCommandResult>;
  linkCreatedEvidence(input: {
    readonly sourceRow: WorkbookRow;
    readonly createdRecordId: string;
  }): Promise<WorkbookMutationCommandResult>;
}

export interface GenericMutationCommandPort {
  canCreateRecord(input: {
    readonly contract: ViewContract;
    readonly draft: Readonly<Record<string, string>>;
  }): boolean;
  createRecord(input: {
    readonly contract: ViewContract;
    readonly draft: Readonly<Record<string, string>>;
    readonly linkedNoteSourceRecordId: string;
  }): Promise<WorkbookMutationCommandResult>;
  patchRecord(input: {
    readonly baseRowVersion: number;
    readonly changes: readonly Record<string, unknown>[];
    readonly purpose: string;
    readonly recordId: string;
    readonly viewSchemaId: string;
  }): Promise<WorkbookMutationCommandResult>;
  createPartyFromText(input: {
    readonly originViewSchemaId: string;
    readonly rawText: string;
  }): Promise<WorkbookMutationCommandResult>;
}

export interface EntityMutationCommandPort {
  canCreateRecord(input: {
    readonly contract: ViewContract;
    readonly draft: Readonly<Record<string, string>>;
  }): boolean;
  createRecord(input: {
    readonly contract: ViewContract;
    readonly draft: Readonly<Record<string, string>>;
  }): Promise<WorkbookMutationCommandResult>;
  patchRecord(input: {
    readonly baseRowVersion: number;
    readonly changes: readonly Record<string, unknown>[];
    readonly purpose: string;
    readonly recordId: string;
    readonly viewSchemaId: string;
  }): Promise<WorkbookMutationCommandResult>;
  pasteCreate(input: {
    readonly clipboardText: string;
    readonly columns: readonly string[];
    readonly format: string;
    readonly startFieldKey: string;
    readonly targetCount: number;
    readonly viewSchemaId: string;
  }): Promise<WorkbookMutationCommandResult>;
  merge(input: {
    readonly loserBaseRowVersion: number;
    readonly loserRecordId: string;
    readonly reason: string;
    readonly survivorBaseRowVersion: number;
    readonly survivorRecordId: string;
  }): Promise<WorkbookMutationCommandResult>;
}

export interface AssessmentMutationCommandPort {
  create(input: {
    readonly draft: AssessmentCreateDraft;
  }): Promise<WorkbookMutationCommandResult>;
}

export interface EvidenceMutationCommandPort {
  attach(input: {
    readonly baseRowVersion: number;
    readonly evidenceRecordId: string;
    readonly file: File;
  }): Promise<void>;
}

export interface CoordinationMutationCommandPort {
  updateTaskLifecycle(input: {
    readonly baseRowVersion: number;
    readonly blockedReason?: string | undefined;
    readonly recordId: string;
    readonly status: string;
  }): Promise<WorkbookMutationCommandResult>;
  supersedeDecision(input: {
    readonly baseRowVersion: number;
    readonly reason: string;
    readonly replacementRecordId: string;
    readonly targetRecordId: string;
  }): Promise<WorkbookMutationCommandResult>;
}

export type WorkbookMutationCommandPorts = {
  readonly timeline: TimelineMutationCommandPort;
  readonly generic: GenericMutationCommandPort;
  readonly entity: EntityMutationCommandPort;
  readonly assessment: AssessmentMutationCommandPort;
  readonly evidence: EvidenceMutationCommandPort;
  readonly coordination: CoordinationMutationCommandPort;
};
