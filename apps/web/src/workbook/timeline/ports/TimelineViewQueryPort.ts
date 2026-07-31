import type { WorkbookQueryState } from "../../models/workbookQuery";
import type { WorkbookOperationOutcome } from "../../mutations/workbookOperationOutcome";
import type { WorkbookRow } from "../models/workbookTimelineModel";

export type TimelineViewQueryAccepted = {
  readonly incidentId: string;
  readonly rows: readonly WorkbookRow[];
  readonly viewSchemaId: string;
};

export type TimelineViewQueryResult =
  | WorkbookOperationOutcome<TimelineViewQueryAccepted>
  | { readonly kind: "aborted" };

export interface TimelineViewQueryPort {
  query(input: {
    readonly queryState: WorkbookQueryState;
    readonly signal: AbortSignal;
  }): Promise<TimelineViewQueryResult>;
}
