import { expect, test } from "./fixtures";
import { applyStorageState, csrfHeaders } from "./support/auth/browserSession";
import { readCurrentSession, revokeAllSessions } from "./support/auth/sessions";
import { exerciseRevokedPendingReplay } from "./support/collaboration/replay";
import { apiBase } from "./support/runtime/configuration";
import { TestClock } from "./support/runtime/testClock";

test.beforeEach(async ({ page }) => {
  await new TestClock(page).reset();
});

test.afterEach(async ({ page }) => {
  await new TestClock(page).reset();
});

test("preserves unsaved local work after socket revocation and re-authentication", async ({
  page,
  sessionTracker,
  workerAdmin,
  workerAdminRequest,
}) => {
  await test.step("deployment-admin revoke-all preserves and replays local work", async () => {
    await exerciseRevokedPendingReplay({
      createdBy: "E-6-03",
      incidentKeyPrefix: "E603REVOKE",
      page,
      scenario: "revoke-all",
      sessionTracker,
      triggerRevocation: async ({ member }) => {
        await revokeAllSessions(
          workerAdminRequest,
          member.user_id,
          "Collaboration E-6-03 browser revoke-all",
        );
      },
    });
    await applyStorageState(page, workerAdmin.storageState);
  });

  await test.step("current-session logout preserves and replays local work", async () => {
    await exerciseRevokedPendingReplay({
      createdBy: "E-6-03",
      incidentKeyPrefix: "E603LOGOUT",
      page,
      scenario: "logout",
      sessionTracker,
      triggerRevocation: async () => {
        const response = await page.request.post(
          `${apiBase}/api/v1/auth/logout`,
          {
            headers: await csrfHeaders(page),
            data: {},
          },
        );
        expect(response.ok()).toBeTruthy();
      },
    });
    await applyStorageState(page, workerAdmin.storageState);
  });

  await test.step("idle expiry preserves and replays local work after re-authentication", async () => {
    const testClock = new TestClock(page);
    await exerciseRevokedPendingReplay({
      createdBy: "E-6-03",
      incidentKeyPrefix: "E603IDLE",
      page,
      scenario: "idle-expiry",
      sessionTracker,
      triggerRevocation: async () => {
        const currentSession = await readCurrentSession(page);
        await testClock.setAfter(currentSession.session_expires_at);
      },
    });
    await applyStorageState(page, workerAdmin.storageState);
  });
});
