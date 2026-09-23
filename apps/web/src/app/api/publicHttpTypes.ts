export type {
  CloseIncidentRequest,
  CreateIncidentMembershipRequest,
  CreateIncidentRequest,
  CreateIncidentResponse,
  DeleteIncidentMembershipRequest,
  ListAdministrativeAuditEventsResponse,
  ListIncidentMembershipAuditEventsResponse,
  ListIncidentMembershipsResponse,
  ListVisibleIncidentsResponse,
  PatchIncidentMembershipRequest,
} from "@cartulary/protocol-ts/http";

import type { ListVisibleIncidentsResponse } from "@cartulary/protocol-ts/http";

export type IncidentDirectoryResource =
  ListVisibleIncidentsResponse["data"]["incidents"][number];
export type IncidentDirectoryPaging =
  ListVisibleIncidentsResponse["meta"]["paging"];

import type {
  GetCredentialStateResponse,
  GetCurrentAccountPreferencesResponse,
  GetCurrentAccountProfileResponse,
  GetCurrentSessionResponse,
  GetDeploymentUserResponse,
  ListDeploymentExtensionsResponse,
  ListEnterpriseAuthProvidersResponse,
} from "@cartulary/protocol-ts/http";

export type SessionData = GetCurrentSessionResponse["data"];
export type CredentialState = GetCredentialStateResponse["data"];
export type UserResource = GetDeploymentUserResponse["data"];
export type AccountProfileResource = GetCurrentAccountProfileResponse["data"];
export type AccountPreferencesResource =
  GetCurrentAccountPreferencesResponse["data"];
export type DensityMode = Exclude<
  AccountPreferencesResource["density_mode"],
  null
>;
export type ExtensionProfileResource =
  ListDeploymentExtensionsResponse["data"]["extensions"][number];
export type EnterpriseAuthProvider =
  ListEnterpriseAuthProvidersResponse["data"]["providers"][number];
