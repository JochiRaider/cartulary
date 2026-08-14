import type { MentionResolutionAction } from "../../collaboration/workbookCollaborationMessages";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";

export type TimelineMentionEntityCreated = {
  readonly recordId: string;
};

type TimelineMentionActionAccepted = {
  readonly entityMention: {
    readonly entityType: "host" | "identity" | null;
    readonly rawText: string | null;
    readonly resolutionMethod: string | null;
    readonly rowVersion: number;
    readonly sourceFieldKey: string | null;
  };
  readonly sourceRecord: {
    readonly recordId: string;
    readonly rowVersion: number;
  };
};

export interface TimelineMentionPort {
  createEntity(input: {
    readonly clientTxnId: string;
    readonly entityType: "host" | "identity";
    readonly rawText: string;
  }): Promise<WorkbookOperationOutcome<TimelineMentionEntityCreated>>;
  resolve(input: {
    readonly action: MentionResolutionAction;
    readonly baseMentionRowVersion: number;
    readonly clientTxnId: string;
    readonly expectedSourceRecordId: string;
    readonly mentionId: string;
    readonly resolvedRecordId?: string | undefined;
  }): Promise<WorkbookOperationOutcome<TimelineMentionActionAccepted>>;
}
