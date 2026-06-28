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
import { Phase1Page } from "./phase1Page";

test("FE-S-P1-01 Verify bootstrap route selectors and error-state selectors use stable test-id builders.", async ({
  page,
}) => {
  const phase1 = new Phase1Page(page);
  await phase1.gotoIncidentDirectory();

  await expect(page.getByTestId(phase1RouteTestId("app-shell"))).toBeVisible();
  await expect(page.getByTestId(phase1LandingTestId("shell"))).toBeVisible();
  await expect(page.getByTestId(phase1LandingTestId("status"))).toBeVisible();
  await expect(
    page.getByTestId(phase1ErrorCodeTestId("landing")),
  ).toBeAttached();
  await expect(
    page.getByTestId(phase1ErrorSummaryTestIds("landing").container),
  ).toBeAttached();
  await phase1.openAccountSettings("account-security");
  await expect(
    page.getByTestId(phase1AccountTestId("refresh-state")),
  ).toBeVisible();
  await expect(
    page.getByTestId(phase1ErrorSummaryTestIds("account").message),
  ).toBeAttached();
  await phase1.openDeploymentAdministration();
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
