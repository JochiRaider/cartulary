import type { CreateIncidentMembershipRequest } from "@cartulary/protocol-ts/http";
import type { Page } from "@playwright/test";

import { csrfHeaders } from "../auth/browserSession";
import { createDeploymentUser } from "../auth/deploymentUsers";
import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";

type IncidentMembershipRole = CreateIncidentMembershipRequest["role"];

export async function createIncidentMembership(
  page: Page,
  incidentId: string,
  email: string,
  role: IncidentMembershipRole,
) {
  const body: CreateIncidentMembershipRequest = {
    client_txn_id: uniqueTxn("membership"),
    email,
    role,
  };
  const response = await publicHttpOperation({
    body,
    headers: await csrfHeaders(page),
    operationID: "createIncidentMembership",
    pathParameters: { incident_id: incidentId },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok) {
    throw new Error(
      `create incident membership failed with HTTP ${response.status}`,
    );
  }
}

export async function createIncidentMemberUser(
  page: Page,
  incidentId: string,
  options: {
    email: string;
    display_name: string;
    initial_password: string;
    role: IncidentMembershipRole;
    mfa_required: boolean;
    is_deployment_admin: boolean;
  },
) {
  const user = await createDeploymentUser(page, {
    email: options.email,
    display_name: options.display_name,
    initial_password: options.initial_password,
    is_deployment_admin: options.is_deployment_admin,
    mfa_required: options.mfa_required,
  });
  await createIncidentMembership(page, incidentId, options.email, options.role);
  return {
    ...user,
    email: options.email,
    initial_password: options.initial_password,
    role: options.role,
  };
}
