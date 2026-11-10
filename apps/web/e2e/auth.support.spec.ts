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

test("renders stable selectors across landing, account, administration and login states", async ({
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

  await page.context().clearCookies();
  await page.goto("/");

  await expect(page.getByTestId(authTestId("login-username"))).toBeVisible();
  await expect(page.getByTestId(authTestId("feedback"))).toBeVisible();
});
