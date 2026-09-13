import type { ViewContract } from "@cartulary/view-contracts";
import type { WorkbookOperationResponse } from "../adapters/workbookOperationContract";
import type { WorkbookProtocolPatchRecordRequest } from "../adapters/workbookProtocolTypes";
import type { AssessmentAppendTransport } from "../features/assessments/assessmentOperation";
import type { WorkbookRecordHistoryPort } from "../history/workbookHistoryOperation";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { WorkbookOperationOutcome } from "./workbookOperationOutcome";

export type GenericViewMutationAccepted = {
  readonly changeSetId: string;
  readonly row: WorkbookQueryRow;
  readonly viewSchemaId: string;
};

type RecordPatchChange = WorkbookProtocolPatchRecordRequest["changes"][number];

export type GenericMutationOutcome =
  WorkbookOperationOutcome<GenericViewMutationAccepted>;

export type RecordRouteCommandPort = WorkbookRecordHistoryPort;

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

export type EntityCreateOutcome =
  WorkbookOperationOutcome<EntityCreateAccepted>;
export type EntityPatchOutcome = WorkbookOperationOutcome<EntityPatchAccepted>;

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

export interface TimelineMutationIdentityPort {
  createLogicalActionId(): string;
  createConflictRecoveryId(): string;
}

export type TimelineFillAccepted = {
  readonly affectedRowCount: number;
  readonly changeSetId: string | null;
  readonly conflictCount: number;
};

export type TimelineFillOutcome =
  WorkbookOperationOutcome<TimelineFillAccepted>;

export interface TimelineFillMutationPort {
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
    readonly outcome: TimelineFillOutcome;
  }>;
}

export type TimelineRelatedRecordCreated = {
  readonly changeSetId: string;
  readonly recordId: string;
  readonly viewSchemaId: string;
};

export interface TimelineRelatedRecordPort {
  createRelatedRecord(input: {
    readonly contract: ViewContract;
    readonly draft: Readonly<Record<string, string>>;
    readonly featureGroupKey: string;
  }): Promise<WorkbookOperationOutcome<TimelineRelatedRecordCreated>>;
}

export type TimelineMutationCommandPorts = {
  readonly fill: TimelineFillMutationPort;
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
  }): Promise<GenericMutationOutcome>;
  patchRecord(input: {
    readonly baseRowVersion: number;
    readonly changes: readonly RecordPatchChange[];
    readonly purpose: string;
    readonly recordId: string;
    readonly viewSchemaId: string;
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
    readonly changes: readonly RecordPatchChange[];
    readonly purpose: string;
    readonly recordId: string;
    readonly viewSchemaId: string;
  }): Promise<EntityPatchOutcome>;
}

export type AssessmentMutationCommandPort = AssessmentAppendTransport;

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

type IndicatorObservationListResponse =
  WorkbookOperationResponse<"listIndicatorObservations">;
export type IndicatorObservation =
  IndicatorObservationListResponse["data"]["observations"][number];
export type WorkbookMutationCommandPorts = {
  readonly records: RecordRouteCommandPort;
  readonly timeline: TimelineMutationCommandPorts;
  readonly generic: GenericMutationCommandPort;
  readonly entity: EntityMutationCommandPort;
  readonly assessment: AssessmentMutationCommandPort;
  readonly evidence: EvidenceCapabilityPort;
};
