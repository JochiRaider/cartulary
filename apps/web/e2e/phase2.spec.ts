import type { APIResponse, Page } from "@playwright/test";

import { contractArtifactIndex } from "../../../packages/protocol-ts/src/generated/contracts";
import { expect, test } from "./fixtures";
import {
  apiBase,
  csrfHeaders,
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "./helpers";

type ErrorContract = {
  code: string;
  http_status: number;
};

type ExtensionProfileContract = {
  profile_id: string;
  route_families: string[];
};

test("E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface", async ({
  page,
}) => {
  await page.goto("/");

  const incidentKey = uniqueIncidentKey("E201");
  await page.getByTestId("landing-incident-key").fill(incidentKey);
  await page.getByTestId("landing-incident-title").fill("Phase 2 E-2-01");
  await page.getByTestId("landing-create-button").click();

  await expect(page).toHaveURL(/incident_id=/);
  await expect(page.getByTestId("surface-tab-timeline")).toBeVisible();
  await expect(page.getByText("Current incident role: admin")).toBeVisible();

  const incidentId = currentIncidentId(page.url());
  const sessionResponse = await page.request.get(
    `${apiBase}/api/v1/auth/session`,
  );
  const sessionBody = (await sessionResponse.json()) as {
    data: { memberships: Array<{ incident_id: string; role: string }> };
  };
  expect(
    sessionBody.data.memberships.some(
      (membership) =>
        membership.incident_id === incidentId && membership.role === "admin",
    ),
  ).toBeTruthy();

  const defaultPrefsResponse = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/workbook-preferences/default`,
  );
  const defaultPrefsBody = (await defaultPrefsResponse.json()) as {
    data: { default_sheet_ref: string | null };
  };
  expect(defaultPrefsBody.data.default_sheet_ref).toBeNull();

  const userPrefsResponse = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/workbook-preferences/me`,
  );
  const userPrefsBody = (await userPrefsResponse.json()) as {
    data: { home_sheet_ref: string | null };
  };
  expect(userPrefsBody.data.home_sheet_ref).toBeNull();
});

test("E-2-02 shows incident discovery, direct retrieval, and promoted-field-only patching", async ({
  page,
}) => {
  const incidentKey = uniqueIncidentKey("E202");
  const incidentId = await createIncident(page, incidentKey, "Phase 2 E-2-02");

  await openIncidentFromLanding(page, incidentId);
  await expect(page.getByTestId("surface-tab-timeline")).toBeVisible();

  const patchResponse = await page.request.patch(
    `${apiBase}/api/v1/incidents/${incidentId}`,
    {
      headers: await csrfHeaders(page),
      data: {
        base_incident_version: 1,
        tlp: "amber",
        current_phase: "containment",
        primary_external_case_ref: "CASE-E202",
      },
    },
  );
  expect(patchResponse.ok()).toBeTruthy();

  const incidentResponse = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}`,
  );
  expect(incidentResponse.ok()).toBeTruthy();
  const incidentBody = (await incidentResponse.json()) as {
    data: {
      incident_key: string;
      title: string;
      tlp: string | null;
      current_phase: string | null;
      primary_external_case_ref: string | null;
      incident_version: number;
    };
  };
  expect(incidentBody.data.incident_key).toBe(incidentKey);
  expect(incidentBody.data.title).toBe("Phase 2 E-2-02");
  expect(incidentBody.data.tlp).toBe("amber");
  expect(incidentBody.data.current_phase).toBe("containment");
  expect(incidentBody.data.primary_external_case_ref).toBe("CASE-E202");
  expect(incidentBody.data.incident_version).toBe(2);
});

test("E-2-03 lets admins manage memberships and denies the same actions to non-admin members", async ({
  browser,
  page,
  sessionTracker,
}) => {
  const targetEmail = uniqueEmail("phase2-e203-member");
  const targetPassword = "Phase2E203Pass!";
  const incidentKey = uniqueIncidentKey("E203");
  const targetUser = await createLocalUser(page, {
    email: targetEmail,
    display_name: "Phase 2 E203 Member",
    initial_password: targetPassword,
  });
  const incidentId = await createIncident(page, incidentKey, "Phase 2 E-2-03");

  await openIncidentFromLanding(page, incidentId);
  const createResponse = await page.request.post(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships`,
    {
      headers: await csrfHeaders(page),
      data: {
        client_txn_id: uniqueTxn("phase2-membership"),
        email: targetEmail,
        role: "viewer",
      },
    },
  );
  expect(createResponse.status()).toBe(201);

  let memberships = await listMemberships(page, incidentId);
  expect(
    memberships.find((membership) => membership.user_id === targetUser.user_id),
  ).toMatchObject({
    role: "viewer",
    user_id: targetUser.user_id,
  });

  const patchResponse = await page.request.patch(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships/${targetUser.user_id}`,
    {
      headers: await csrfHeaders(page),
      data: {
        base_membership_version: 1,
        role: "reviewer",
      },
    },
  );
  expect(patchResponse.ok()).toBeTruthy();

  memberships = await listMemberships(page, incidentId);
  expect(
    memberships.find((membership) => membership.user_id === targetUser.user_id),
  ).toMatchObject({
    role: "reviewer",
    membership_version: 2,
    user_id: targetUser.user_id,
  });

  const memberContext = await browser.newContext();
  const memberPage = await memberContext.newPage();
  await sessionTracker.loginTrackedUser(memberPage, {
    createdBy: "phase2 member denied context",
    email: targetEmail,
    password: targetPassword,
    purpose: "phase2 e203 member login",
    userId: targetUser.user_id,
  });
  await openIncidentFromLanding(memberPage, incidentId);

  await expectAPIError(
    await memberPage.request.post(
      `${apiBase}/api/v1/incidents/${incidentId}/memberships`,
      {
        headers: await csrfHeaders(memberPage),
        data: {
          client_txn_id: uniqueTxn("phase2-member-denied"),
          email: uniqueEmail("phase2-e203-denied"),
          role: "viewer",
        },
      },
    ),
    "authorization_denied",
  );

  await expectAPIError(
    await memberPage.request.patch(
      `${apiBase}/api/v1/incidents/${incidentId}/memberships/${targetUser.user_id}`,
      {
        headers: await csrfHeaders(memberPage),
        data: {
          base_membership_version: 2,
          role: "admin",
        },
      },
    ),
    "authorization_denied",
  );

  await expectAPIError(
    await memberPage.request.delete(
      `${apiBase}/api/v1/incidents/${incidentId}/memberships/${targetUser.user_id}`,
      {
        headers: await csrfHeaders(memberPage),
        data: {
          base_membership_version: 2,
        },
      },
    ),
    "authorization_denied",
  );
  await memberContext.close();

  const deleteResponse = await page.request.delete(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships/${targetUser.user_id}`,
    {
      headers: await csrfHeaders(page),
      data: {
        base_membership_version: 2,
      },
    },
  );
  expect(deleteResponse.status()).toBe(204);

  memberships = await listMemberships(page, incidentId);
  expect(
    memberships.some((membership) => membership.user_id === targetUser.user_id),
  ).toBeFalsy();
});

test("E-2-04 rejects unknown or forbidden top-level members with route-owned errors", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("E204"),
    "Phase 2 E-2-04",
  );

  await expectAPIError(
    await page.request.post(`${apiBase}/api/v1/incidents`, {
      headers: await csrfHeaders(page),
      data: {
        client_txn_id: uniqueTxn("phase2-invalid-create-memberships"),
        incident_key: uniqueIncidentKey("E204INVALID"),
        title: "Invalid create",
        initial_memberships: [],
      },
    }),
    "invalid_incident_create",
    {
      field: "initial_memberships",
      reason_code: "initial_memberships_not_supported",
    },
  );

  await expectAPIError(
    await page.request.post(`${apiBase}/api/v1/incidents`, {
      headers: await csrfHeaders(page),
      data: {
        client_txn_id: uniqueTxn("phase2-invalid-create-unknown"),
        incident_key: uniqueIncidentKey("E204UNKNOWN"),
        title: "Invalid create",
        unexpected: true,
      },
    }),
    "invalid_incident_create",
    {
      field: "unexpected",
      reason_code: "unknown_top_level_member",
    },
  );

  await expectAPIError(
    await page.request.patch(`${apiBase}/api/v1/incidents/${incidentId}`, {
      headers: await csrfHeaders(page),
      data: {
        base_incident_version: 1,
        title: "forbidden",
      },
    }),
    "invalid_incident_patch",
    {
      field: "title",
      reason_code: "forbidden_field",
    },
  );

  await expectAPIError(
    await page.request.patch(`${apiBase}/api/v1/incidents/${incidentId}`, {
      headers: await csrfHeaders(page),
      data: {
        base_incident_version: 1,
        unknown: "field",
      },
    }),
    "invalid_incident_patch",
    {
      field: "unknown",
      reason_code: "unknown_top_level_member",
    },
  );
});

test("E-2-05 allows zero-membership extension discovery and rejects singleton pagination semantics", async ({
  page,
  sessionTracker,
}) => {
  const zeroMemberEmail = uniqueEmail("phase2-e205");
  const zeroMemberPassword = "Phase2E205Pass!";
  const zeroMemberUser = await createLocalUser(page, {
    email: zeroMemberEmail,
    display_name: "Phase 2 E205 User",
    initial_password: zeroMemberPassword,
  });

  await sessionTracker.loginTrackedUser(page, {
    createdBy: "phase2 zero-member landing",
    email: zeroMemberEmail,
    password: zeroMemberPassword,
    purpose: "phase2 e205 zero-member login",
    userId: zeroMemberUser.user_id,
  });
  await page.goto("/");
  await expect(page.getByTestId("landing-empty-state")).toBeVisible();

  const extensionsResponse = await page.request.get(
    `${apiBase}/api/v1/extensions`,
  );
  expect(extensionsResponse.ok()).toBeTruthy();
  const extensionsBody = (await extensionsResponse.json()) as {
    data: {
      extensions: Array<{
        profile_id: string;
        claimed: boolean;
        route_families: string[];
      }>;
    };
  };
  expect(extensionsBody.data.extensions).toEqual(
    extensionRegistry().map((profile) => ({
      profile_id: profile.profile_id,
      claimed: false,
      route_families: profile.route_families,
    })),
  );

  await expectAPIError(
    await page.request.get(`${apiBase}/api/v1/extensions?cursor_token=opaque`),
    "invalid_pagination_request",
    {
      reason_code: "pagination_not_supported",
    },
  );
});

test("E-2-06 shows reserved-family 404 precedence while base and outside-reserved paths keep their ordinary dispatch", async ({
  page,
  sessionTracker,
}) => {
  const zeroMemberEmail = uniqueEmail("phase2-e206");
  const zeroMemberPassword = "Phase2E206Pass!";
  const zeroMemberUser = await createLocalUser(page, {
    email: zeroMemberEmail,
    display_name: "Phase 2 E206 User",
    initial_password: zeroMemberPassword,
  });

  await sessionTracker.loginTrackedUser(page, {
    createdBy: "phase2 reserved-family zero-member",
    email: zeroMemberEmail,
    password: zeroMemberPassword,
    purpose: "phase2 e206 zero-member login",
    userId: zeroMemberUser.user_id,
  });
  await page.goto("/");
  await expect(page.getByTestId("landing-empty-state")).toBeVisible();

  const readyResponse = await page.request.get(`${apiBase}/readyz`);
  expect(readyResponse.ok()).toBeTruthy();
  expect(await readyResponse.text()).toContain("ready");

  const importProfile = extensionProfile("import");
  await expectAPIError(
    await page.request.get(`${apiBase}${importProfile.route_families[0]}`),
    "extension_profile_not_claimed",
    {
      profile_id: importProfile.profile_id,
      route_family: importProfile.route_families[0],
    },
  );

  const enterpriseProfile = extensionProfile("enterprise_authentication");
  await expectAPIError(
    await page.request.get(
      `${apiBase}${enterpriseProfile.route_families[0].replace("{user_id}", "00000000-0000-0000-0000-000000000001")}/provider`,
    ),
    "extension_profile_not_claimed",
    {
      profile_id: enterpriseProfile.profile_id,
      route_family: enterpriseProfile.route_families[0],
    },
  );

  const outsideReserved = await page.request.get(
    `${apiBase}/api/v1/outside-reserved-families`,
  );
  expect(outsideReserved.status()).toBe(404);
  expect(await outsideReserved.text()).not.toContain(
    "extension_profile_not_claimed",
  );
});

async function createIncident(page: Page, incidentKey: string, title: string) {
  const response = await page.request.post(`${apiBase}/api/v1/incidents`, {
    headers: await csrfHeaders(page),
    data: {
      client_txn_id: uniqueTxn("phase2-incident"),
      incident_key: incidentKey,
      title,
    },
  });
  expect(response.ok()).toBeTruthy();
  const body = (await response.json()) as { data: { incident_id: string } };
  return body.data.incident_id;
}

async function openIncidentFromLanding(page: Page, incidentId: string) {
  await expect
    .poll(async () => {
      const response = await page.request.get(`${apiBase}/api/v1/incidents`);
      expect(response.ok()).toBeTruthy();
      const body = (await response.json()) as {
        data: { incidents: Array<{ incident_id: string }> };
      };
      return body.data.incidents.some(
        (incident) => incident.incident_id === incidentId,
      );
    })
    .toBe(true);

  await page.goto("/");
  await expect(
    page.getByTestId(`landing-incident-${incidentId}`),
  ).toBeVisible();
  await page.getByTestId(`landing-open-${incidentId}`).click();
  await expect(page).toHaveURL(new RegExp(`incident_id=${incidentId}`));
}

async function createLocalUser(
  page: Page,
  options: {
    email: string;
    display_name: string;
    initial_password: string;
    is_deployment_admin?: boolean;
  },
) {
  const response = await page.request.post(`${apiBase}/api/v1/users`, {
    headers: await csrfHeaders(page),
    data: {
      client_txn_id: uniqueTxn("phase2-user"),
      auth_kind: "local",
      email: options.email,
      display_name: options.display_name,
      initial_password: options.initial_password,
      mfa_required: false,
      is_deployment_admin: options.is_deployment_admin ?? false,
    },
  });
  expect(response.ok()).toBeTruthy();
  return ((await response.json()) as { data: { user_id: string } }).data;
}

async function listMemberships(page: Page, incidentId: string) {
  const response = await page.request.get(
    `${apiBase}/api/v1/incidents/${incidentId}/memberships`,
  );
  expect(response.ok()).toBeTruthy();
  return (
    (await response.json()) as {
      data: {
        memberships: Array<{
          user_id: string;
          role: string;
          membership_version: number;
        }>;
      };
    }
  ).data.memberships;
}

async function expectAPIError(
  response: APIResponse,
  code: string,
  details?: Record<string, unknown>,
) {
  const contract = errorContract(code);
  expect(response.status()).toBe(contract.http_status);
  const body = (await response.json()) as {
    error: { code: string; details: Record<string, unknown> };
  };
  expect(body.error.code).toBe(code);
  if (details) {
    expect(body.error.details).toMatchObject(details);
  }
  return body.error;
}

function currentIncidentId(urlText: string): string {
  const url = new URL(urlText);
  const incidentId = url.searchParams.get("incident_id");
  if (!incidentId) {
    throw new Error(`missing incident_id in ${urlText}`);
  }
  return incidentId;
}

function errorContract(code: string): ErrorContract {
  const registry = JSON.parse(
    contractArtifactIndex["contracts/errors/index.json"].json,
  ) as {
    errors: ErrorContract[];
  };
  const match = registry.errors.find((candidate) => candidate.code === code);
  if (!match) {
    throw new Error(`missing error contract for ${code}`);
  }
  return match;
}

function extensionRegistry(): ExtensionProfileContract[] {
  return (
    JSON.parse(
      contractArtifactIndex["contracts/extensions/index.json"].json,
    ) as {
      profiles: ExtensionProfileContract[];
    }
  ).profiles;
}

function extensionProfile(profileID: string): ExtensionProfileContract {
  const match = extensionRegistry().find(
    (profile) => profile.profile_id === profileID,
  );
  if (!match) {
    throw new Error(`missing extension profile contract for ${profileID}`);
  }
  return match;
}
