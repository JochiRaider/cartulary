import type { WorkbookOperationResponse } from "../adapters/workbookOperationContract";
import type { AssessmentAppendTransport } from "../features/assessments/assessmentOperation";
import type { WorkbookRecordHistoryPort } from "../history/workbookHistoryOperation";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { WorkbookBatchAdmission } from "../runtime/workbookBatchOperation";
import type { WorkbookOperationOutcome } from "./workbookOperationOutcome";

export type GenericViewMutationAccepted = {
  readonly changeSetId: string;
  readonly row: WorkbookQueryRow;
  readonly viewSchemaId: string;
};

export type RecordRouteCommandPort = WorkbookRecordHistoryPort;

export type EvidenceHandleAccepted = {
  readonly filename: string;
  readonly href: string;
  readonly previewKind: string | null;
};

export type EvidenceHandleOutcome =
  WorkbookOperationOutcome<EvidenceHandleAccepted>;

export interface TimelineMutationIdentityPort {
  createLogicalActionId(): string;
  createConflictRecoveryId(): string;
}

export interface TimelineFillMutationPort {
  fillDown(
    input: {
      readonly fieldKey: string;
      readonly value: string;
      readonly targets: readonly {
        readonly recordId: string;
        readonly baseRowVersion: number;
      }[];
    },
    admission: WorkbookBatchAdmission,
  ): string | null;
}

export interface TimelineClearMutationPort {
  clearCells(
    input: {
      readonly fieldKeys: readonly string[];
      readonly targets: readonly {
        readonly recordId: string;
        readonly baseRowVersion: number;
      }[];
    },
    admission: WorkbookBatchAdmission,
  ): string | null;
}

export type TimelineMutationCommandPorts = {
  readonly clear: TimelineClearMutationPort;
  readonly fill: TimelineFillMutationPort;
  readonly identity: TimelineMutationIdentityPort;
};

export type AssessmentMutationCommandPort = AssessmentAppendTransport;

export interface EvidenceCapabilityPort {
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
  readonly assessment: AssessmentMutationCommandPort;
  readonly evidence: EvidenceCapabilityPort;
};
