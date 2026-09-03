import type { ViewContract } from "@cartulary/view-contracts";
import type { WorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookPortResult } from "../ports/WorkbookPortResult";

import type { WorkbookQueryRow } from "./WorkbookQueryRow";

export type WorkbookViewQueryAccepted = {
  readonly incidentId: string;
  readonly rows: readonly WorkbookQueryRow[];
  readonly viewSchemaId: string;
};

export type WorkbookViewQueryResult =
  WorkbookPortResult<WorkbookViewQueryAccepted>;

export interface WorkbookViewQueryPort {
  query(input: {
    readonly contract: ViewContract;
    readonly queryState: WorkbookQueryState;
    readonly signal: AbortSignal;
  }): Promise<WorkbookViewQueryResult>;
}
