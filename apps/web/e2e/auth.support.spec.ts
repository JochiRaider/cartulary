import {
  accountTestId,
  appRouteTestId,
  authTestId,
  deploymentAdminTestId,
  incidentLandingTestId,
  publicErrorCodeTestId,
  publicErrorSummaryTestIds,
} from "@cartulary/ui-contracts";

import { expect, test } from "./fixtures";
import { AccountSettings } from "./pages/accountSettings";
import { DeploymentAdministration } from "./pages/deploymentAdministration";
import { IncidentDirectory } from "./pages/incidentDirectory";

test("Verify bootstrap route selectors and error-state selectors use stable test-id builders.", async ({
  page,
}) => {
  await new IncidentDirectory(page).goto();

  await expect(page.getByTestId(appRouteTestId("app-shell"))).toBeVisible();
  await expect(page.getByTestId(incidentLandingTestId("shell"))).toBeVisible();
  await expect(page.getByTestId(incidentLandingTestId("status"))).toBeVisible();
  await expect(
    page.getByTestId(publicErrorCodeTestId("landing")),
  ).toBeAttached();
  await expect(
    page.getByTestId(publicErrorSummaryTestIds("landing").container),
  ).toBeAttached();
  await new AccountSettings(page).openSecurity();
  await expect(page.getByTestId(accountTestId("refresh-state"))).toBeVisible();
  await expect(
    page.getByTestId(publicErrorSummaryTestIds("account").message),
  ).toBeAttached();
  await new DeploymentAdministration(page).open();
  await expect(page.getByTestId(deploymentAdminTestId("status"))).toBeVisible();
  expect(authTestId("bootstrap-token")).toBe("auth-bootstrap-token");

  await page.context().clearCookies();
  await page.goto("/");

  await expect(page.getByTestId(authTestId("login-username"))).toBeVisible();
  await expect(page.getByTestId(authTestId("status"))).toBeVisible();
  await expect(page.getByTestId(publicErrorCodeTestId("auth"))).toBeAttached();
  await expect(
    page.getByTestId(publicErrorSummaryTestIds("auth").container),
  ).toBeAttached();
});
