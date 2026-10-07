import type { ViewContract } from "@cartulary/view-contracts";
import type { WorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookPortResult } from "../ports/WorkbookPortResult";
import type { WorkbookViewQueryAccepted } from "./WorkbookViewQueryPort";

export type WorkbookLocation =
  | { readonly outcome: "located"; readonly page: WorkbookViewQueryAccepted }
  | { readonly outcome: "outside_query" }
  | { readonly outcome: "unavailable" };

export interface WorkbookRecordLocatorPort {
  locate(input: {
    readonly contract: ViewContract;
    readonly recordId: string;
    readonly queryState: WorkbookQueryState;
    readonly signal: AbortSignal;
  }): Promise<WorkbookPortResult<WorkbookLocation>>;
}
