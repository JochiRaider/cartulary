import { accountTestId, authTestId } from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";
import { expect, test } from "./fixtures";
import { AccountSettings } from "./pages/accountSettings";
import { AuthGateway } from "./pages/authGateway";
import { createDeploymentUser } from "./support/auth/deploymentUsers";
import { readCurrentSession } from "./support/auth/sessions";
import { uniqueEmail } from "./support/runtime/fixtureIdentity";
import { TestClock } from "./support/runtime/testClock";

// This isolated spec owns browser evidence that mutates process-global backend state.
test.beforeEach(async ({ page }) => {
  await new TestClock(page).reset();
});

test.afterEach(async ({ page }) => {
  await new TestClock(page).reset();
});

test("advances the shared clock past idle expiry and requires a fresh login afterwards", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const testClock = new TestClock(page);
  const email = uniqueEmail("authentication-e104");
  const password = "AuthenticationE104Pass!";
  const user = await createDeploymentUser(workerAdminRequest, {
    email,
    display_name: "Authentication E104",
    initial_password: password,
    mfa_required: false,
    is_deployment_admin: false,
  });

  await clearBrowserSession(page);
  await new AuthGateway(page).goto();
  await new AuthGateway(page).login(email, password);
  await new AccountSettings(page).openSecurity();
  await expect(page.getByTestId(accountTestId("refresh-state"))).toBeVisible();
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email,
    purpose: "authentication e104 initial ordinary login",
    userId: user.user_id,
  });

  const currentSession = await readCurrentSession(page);
  try {
    await testClock.setAfter(currentSession.session_expires_at);
    const [sessionResponse] = await Promise.all([
      page.waitForResponse((candidate) => {
        const method = candidate.request().method().toUpperCase();
        return (
          method === "GET" &&
          new URL(candidate.url()).pathname === "/api/v1/auth/session" &&
          candidate.status() === 401
        );
      }),
      new AccountSettings(page).refresh(),
    ]);
    expect(sessionResponse.status()).toBe(401);
    await expect(page.getByTestId(authTestId("login-username"))).toBeVisible();
  } finally {
    await testClock.reset();
  }

  await new AuthGateway(page).login(email, password);
  await new AccountSettings(page).openSecurity();
  await expect(page.getByTestId(accountTestId("refresh-state"))).toBeVisible();
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "authentication ordinary shell",
    email,
    purpose: "authentication e104 re-login after expiry",
    userId: user.user_id,
  });
});

async function clearBrowserSession(page: Page) {
  await page.context().clearCookies();
}
