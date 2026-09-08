import type { IncidentResource } from "../../shared/incidentResource";
import type { WorkbookIncidentIdentity } from "../models/workbookIncidentIdentity";
import type { WorkbookPortResult } from "./WorkbookPortResult";

export type WorkbookIncidentMember = {
  readonly displayName: string;
  readonly userId: string;
};

export interface WorkbookIncidentPort {
  getIdentity(input: {
    readonly signal: AbortSignal;
  }): Promise<
    WorkbookPortResult<
      WorkbookIncidentIdentity & { readonly resource?: IncidentResource }
    >
  >;
  listMembers(input: {
    readonly signal: AbortSignal;
  }): Promise<
    WorkbookPortResult<{ readonly members: readonly WorkbookIncidentMember[] }>
  >;
}
