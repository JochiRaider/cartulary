import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";

export type WorkbookMutationAuthority = Readonly<{
  actorId: string;
  sessionIdentity: string;
  incidentId: string;
  role: WorkbookIncidentRole;
  closed: boolean;
}>;
