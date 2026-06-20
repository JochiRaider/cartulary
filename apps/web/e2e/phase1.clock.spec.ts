import { phase1AccountTestId, phase1AuthTestId } from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { createLocalUser } from "./authRuntime";
import { expect, test } from "./fixtures";
import { uniqueEmail } from "./helpers";
import { Phase1Page } from "./phase1Page";

// This isolated spec owns browser evidence that mutates process-global backend state.
test.beforeEach(async ({ page }) => {
  await new Phase1Page(page).resetClockOffset();
});

test.afterEach(async ({ page }) => {
  await new Phase1Page(page).resetClockOffset();
});

test("E-1-04 advances the shared clock past idle expiry and requires a fresh login afterwards", async ({
  page,
  sessionTracker,
  workerAdminRequest,
}) => {
  const phase1 = new Phase1Page(page);
  const email = uniqueEmail("phase1-e104");
  const password = "Phase1E104Pass!";
  const user = await createLocalUser(workerAdminRequest, {
    email,
    display_name: "Phase 1 E104",
    initial_password: password,
    mfa_required: false,
  });

  await clearBrowserSession(page);
  await phase1.goto();
  await phase1.login(email, password);
  await phase1.openAccountSettings("account-security");
  await expect(
    page.getByTestId(phase1AccountTestId("session-provider-type")),
  ).toHaveText("local");
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email,
    purpose: "phase1 e104 initial ordinary login",
    userId: user.user_id,
  });

  const currentSession = await phase1.currentSession();
  try {
    await phase1.setClockAfter(currentSession.session_expires_at);
    const [sessionResponse] = await Promise.all([
      page.waitForResponse((candidate) => {
        const method = candidate.request().method().toUpperCase();
        return (
          method === "GET" &&
          new URL(candidate.url()).pathname === "/api/v1/auth/session" &&
          candidate.status() === 401
        );
      }),
      phase1.refreshAccount(),
    ]);
    expect(sessionResponse.status()).toBe(401);
    await expect(
      page.getByTestId(phase1AuthTestId("login-username")),
    ).toBeVisible();
  } finally {
    await phase1.resetClockOffset();
  }

  await phase1.login(email, password);
  await phase1.openAccountSettings("account-security");
  await expect(
    page.getByTestId(phase1AccountTestId("session-provider-type")),
  ).toHaveText("local");
  await sessionTracker.captureCurrentSession(page, {
    createdBy: "phase1 ordinary shell",
    email,
    purpose: "phase1 e104 re-login after expiry",
    userId: user.user_id,
  });
});

async function clearBrowserSession(page: Page) {
  await page.context().clearCookies();
}
