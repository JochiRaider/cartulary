import { incidentLandingTestId } from "@cartulary/ui-contracts";
import { expect, test } from "./fixtures";
import { createDeploymentUser } from "./support/auth/deploymentUsers";
import { uniqueEmail } from "./support/runtime/fixtureIdentity";

test("renders the empty incident directory for an authenticated user with no memberships", async ({
  page,
  sessionTracker,
}) => {
  const zeroMemberEmail = uniqueEmail("incident_membership-support-e205");
  const zeroMemberPassword = "IncidentMembershipSupportE205Pass!";
  const zeroMemberUser = await createDeploymentUser(page, {
    email: zeroMemberEmail,
    display_name: "Incident administration Support Zero Member",
    initial_password: zeroMemberPassword,
    is_deployment_admin: false,
    mfa_required: false,
  });

  await sessionTracker.loginTrackedUser(page, {
    createdBy: "incident_membership support zero-member landing",
    email: zeroMemberEmail,
    password: zeroMemberPassword,
    purpose: "incident_membership support zero-member login",
    userId: zeroMemberUser.user_id,
  });
  await page.goto("/");
  await expect(
    page.getByTestId(incidentLandingTestId("empty-state")),
  ).toBeVisible();
});
