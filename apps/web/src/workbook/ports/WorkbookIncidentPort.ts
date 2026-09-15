import type { IncidentResource } from "../../shared/incidentResource";
import type { WorkbookIncidentIdentity } from "../models/workbookIncidentIdentity";
import type { WorkbookPortResult } from "./WorkbookPortResult";

export interface WorkbookIncidentPort {
  getIdentity(input: {
    readonly signal: AbortSignal;
  }): Promise<
    WorkbookPortResult<
      WorkbookIncidentIdentity & { readonly resource?: IncidentResource }
    >
  >;
}
