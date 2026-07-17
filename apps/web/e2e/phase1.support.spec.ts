import {
  phase1AccountTestId,
  phase1AdminTestId,
  phase1AuthTestId,
  phase1ErrorCodeTestId,
  phase1ErrorSummaryTestIds,
  phase1LandingTestId,
  phase1RouteTestId,
} from "@cartulary/ui-contracts";

import { expect, test } from "./fixtures";
import { AccountSettings } from "./pages/accountSettings";
import { DeploymentAdministration } from "./pages/deploymentAdministration";
import { IncidentDirectory } from "./pages/incidentDirectory";

test("FE-S-P1-01 Verify bootstrap route selectors and error-state selectors use stable test-id builders.", async ({
  page,
}) => {
  await new IncidentDirectory(page).goto();

  await expect(page.getByTestId(phase1RouteTestId("app-shell"))).toBeVisible();
  await expect(page.getByTestId(phase1LandingTestId("shell"))).toBeVisible();
  await expect(page.getByTestId(phase1LandingTestId("status"))).toBeVisible();
  await expect(
    page.getByTestId(phase1ErrorCodeTestId("landing")),
  ).toBeAttached();
  await expect(
    page.getByTestId(phase1ErrorSummaryTestIds("landing").container),
  ).toBeAttached();
  await new AccountSettings(page).open("account-security");
  await expect(
    page.getByTestId(phase1AccountTestId("refresh-state")),
  ).toBeVisible();
  await expect(
    page.getByTestId(phase1ErrorSummaryTestIds("account").message),
  ).toBeAttached();
  await new DeploymentAdministration(page).open();
  await expect(page.getByTestId(phase1AdminTestId("status"))).toBeVisible();
  expect(phase1AuthTestId("bootstrap-token")).toBe("auth-bootstrap-token");

  await page.context().clearCookies();
  await page.goto("/");

  await expect(
    page.getByTestId(phase1AuthTestId("login-username")),
  ).toBeVisible();
  await expect(page.getByTestId(phase1AuthTestId("status"))).toBeVisible();
  await expect(page.getByTestId(phase1ErrorCodeTestId("auth"))).toBeAttached();
  await expect(
    page.getByTestId(phase1ErrorSummaryTestIds("auth").container),
  ).toBeAttached();
});
