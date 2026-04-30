import type { APIResponse } from "@playwright/test";

import { contractArtifactIndex } from "../../../packages/protocol-ts/src/generated/contracts";
import { expect, test } from "./fixtures";
import {
  apiBase,
  createIncident,
  createLocalUser,
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

test("supports route-owned incident validation errors through browser-authenticated request probes", async ({
  page,
}) => {
  const incidentId = await createIncident(
    page,
    uniqueIncidentKey("S204"),
    "Phase 2 Support Incident Validation",
  );

  await expectAPIError(
    await page.request.post(`${apiBase}/api/v1/incidents`, {
      headers: await csrfHeaders(page),
      data: {
        client_txn_id: uniqueTxn("phase2-support-invalid-create-memberships"),
        incident_key: uniqueIncidentKey("S204INVALID"),
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
        client_txn_id: uniqueTxn("phase2-support-invalid-create-unknown"),
        incident_key: uniqueIncidentKey("S204UNKNOWN"),
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

test("supports zero-membership extension discovery and singleton pagination rejection through browser-authenticated request probes", async ({
  page,
  sessionTracker,
}) => {
  const zeroMemberEmail = uniqueEmail("phase2-support-e205");
  const zeroMemberPassword = "Phase2SupportE205Pass!";
  const zeroMemberUser = await createLocalUser(page, {
    email: zeroMemberEmail,
    display_name: "Phase 2 Support Zero Member",
    initial_password: zeroMemberPassword,
  });

  await sessionTracker.loginTrackedUser(page, {
    createdBy: "phase2 support zero-member landing",
    email: zeroMemberEmail,
    password: zeroMemberPassword,
    purpose: "phase2 support zero-member login",
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

test("supports reserved-family dispatch precedence probes while ordinary base and outside paths keep their dispatch", async ({
  page,
  sessionTracker,
}) => {
  const zeroMemberEmail = uniqueEmail("phase2-support-e206");
  const zeroMemberPassword = "Phase2SupportE206Pass!";
  const zeroMemberUser = await createLocalUser(page, {
    email: zeroMemberEmail,
    display_name: "Phase 2 Support Reserved Family User",
    initial_password: zeroMemberPassword,
  });

  await sessionTracker.loginTrackedUser(page, {
    createdBy: "phase2 support reserved-family zero-member",
    email: zeroMemberEmail,
    password: zeroMemberPassword,
    purpose: "phase2 support reserved-family login",
    userId: zeroMemberUser.user_id,
  });
  await page.goto("/");
  await expect(page.getByTestId("landing-empty-state")).toBeVisible();

  const readyResponse = await page.request.get(`${apiBase}/readyz`);
  expect(readyResponse.ok()).toBeTruthy();
  expect(await readyResponse.text()).toContain("ready");

  const importProfile = extensionProfile("import");
  const importRouteFamily = routeFamily(importProfile);
  await expectAPIError(
    await page.request.get(`${apiBase}${importRouteFamily}`),
    "extension_profile_not_claimed",
    {
      profile_id: importProfile.profile_id,
      route_family: importRouteFamily,
    },
  );

  const enterpriseProfile = extensionProfile("enterprise_authentication");
  const enterpriseRouteFamily = routeFamily(enterpriseProfile);
  await expectAPIError(
    await page.request.get(
      `${apiBase}${enterpriseRouteFamily.replace("{user_id}", "00000000-0000-0000-0000-000000000001")}/provider`,
    ),
    "extension_profile_not_claimed",
    {
      profile_id: enterpriseProfile.profile_id,
      route_family: enterpriseRouteFamily,
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

function errorContract(code: string): ErrorContract {
  const registry = JSON.parse(
    contractArtifactJSON("contracts/errors/index.json"),
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
    JSON.parse(contractArtifactJSON("contracts/extensions/index.json")) as {
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

function routeFamily(profile: ExtensionProfileContract) {
  const route = profile.route_families[0];
  if (!route) {
    throw new Error(
      `missing route family for extension profile ${profile.profile_id}`,
    );
  }
  return route;
}

function contractArtifactJSON(path: string) {
  const artifact = contractArtifactIndex[path];
  if (!artifact) {
    throw new Error(`missing contract artifact ${path}`);
  }
  return artifact.json;
}
