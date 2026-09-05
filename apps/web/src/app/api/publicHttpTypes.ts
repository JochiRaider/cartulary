export type {
  CloseIncidentRequest,
  CloseIncidentResponse,
  CreateIncidentRequest,
  CreateIncidentResponse,
  ListAdministrativeAuditEventsResponse,
  ListIncidentMembershipAuditEventsResponse,
  ListVisibleIncidentsResponse,
  ReopenIncidentRequest,
  ReopenIncidentResponse,
} from "@cartulary/protocol-ts/http";

import type { ListVisibleIncidentsResponse } from "@cartulary/protocol-ts/http";

export type IncidentDirectoryResource =
  ListVisibleIncidentsResponse["data"]["incidents"][number];
export type IncidentDirectoryPaging =
  ListVisibleIncidentsResponse["meta"]["paging"];
