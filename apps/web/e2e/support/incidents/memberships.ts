import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";
import { csrfHeaders } from "../auth/browserSession";
import { createDeploymentUser } from "../auth/deploymentUsers";
import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { atJsonOrigin, requestPublicJson } from "../transport/publicJsonClient";

export async function createIncidentMembership(
  page: Page,
  incidentId: string,
  email: string,
  role: string,
) {
  const response = await requestPublicJson({
    body: { client_txn_id: uniqueTxn("membership"), email, role },
    headers: await csrfHeaders(page),
    method: "POST",
    path: `/api/v1/incidents/${incidentId}/memberships`,
    request: atJsonOrigin(page.request, apiBase),
  });
  expect(response.ok).toBeTruthy();
}

export async function createIncidentMemberUser(
  page: Page,
  incidentId: string,
  options: {
    email: string;
    display_name: string;
    initial_password: string;
    role: string;
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
