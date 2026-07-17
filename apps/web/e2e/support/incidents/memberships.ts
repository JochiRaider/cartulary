import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";
import { csrfHeaders } from "../auth/browserSession";
import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { atJsonOrigin, requestPublicJson } from "../transport/publicJsonClient";

export async function createLocalUser(
  page: Page,
  options: {
    email: string;
    display_name: string;
    initial_password: string;
    mfa_required?: boolean;
    is_deployment_admin?: boolean;
  },
) {
  const response = await requestPublicJson({
    body: {
      client_txn_id: uniqueTxn("user"),
      auth_kind: "local",
      email: options.email,
      display_name: options.display_name,
      initial_password: options.initial_password,
      mfa_required: options.mfa_required ?? false,
      is_deployment_admin: options.is_deployment_admin ?? false,
    },
    headers: await csrfHeaders(page),
    method: "POST",
    path: "/api/v1/users",
    request: atJsonOrigin(page.request, apiBase),
  });
  expect(response.ok).toBeTruthy();
  return (response.body as { data: { user_id: string } }).data;
}

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
    mfa_required?: boolean;
    is_deployment_admin?: boolean;
  },
) {
  const userOptions: Parameters<typeof createLocalUser>[1] = {
    email: options.email,
    display_name: options.display_name,
    initial_password: options.initial_password,
  };
  if (options.mfa_required !== undefined) {
    userOptions.mfa_required = options.mfa_required;
  }
  if (options.is_deployment_admin !== undefined) {
    userOptions.is_deployment_admin = options.is_deployment_admin;
  }
  const user = await createLocalUser(page, userOptions);
  await createIncidentMembership(page, incidentId, options.email, options.role);
  return {
    ...user,
    email: options.email,
    initial_password: options.initial_password,
    role: options.role,
  };
}
