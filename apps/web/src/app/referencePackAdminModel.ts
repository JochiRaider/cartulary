import type {
  ReferencePackJobResource,
  ReferencePackPaging,
  ReferencePackQuery,
  ReferencePackVersion,
} from "../services/referencePacks";

export type {
  ReferencePackJobResource,
  ReferencePackPaging,
  ReferencePackQuery,
  ReferencePackVersion,
};

export type DeploymentAdminSession = {
  readonly is_deployment_admin: boolean;
};

export const terminalReferencePackJobStates = new Set([
  "succeeded",
  "failed",
  "canceled",
]);

export const defaultReferencePackPaging: ReferencePackPaging = {
  limit: 100,
  has_more: false,
  next_cursor: null,
};
