import { revokeAllSessions } from "./authRuntime";
import { expect, test } from "./fixtures";
import { apiBase, applyStorageState, csrfHeaders } from "./helpers";
import { Phase1Page } from "./phase1Page";
import {
  createUntrackedLoginSessions,
  exerciseRevokedPendingReplay,
} from "./phase6Harness";

test.beforeEach(async ({ page }) => {
  await new Phase1Page(page).resetClockOffset();
});

test.afterEach(async ({ page }) => {
  await new Phase1Page(page).resetClockOffset();
});

test("E-6-03 preserves unsaved local work after socket revocation and re-authentication", async ({
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
          "Phase 6 E-6-03 browser revoke-all",
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

  await test.step("concurrency-limit revocation preserves and replays local work", async () => {
    await exerciseRevokedPendingReplay({
      createdBy: "E-6-03",
      incidentKeyPrefix: "E603CONCURRENCY",
      page,
      scenario: "concurrency",
      sessionTracker,
      triggerRevocation: async ({ member }) => {
        await createUntrackedLoginSessions(
          member.email,
          member.initial_password,
          5,
        );
      },
    });
    await applyStorageState(page, workerAdmin.storageState);
  });

  await test.step("idle expiry preserves and replays local work after re-authentication", async () => {
    const phase1 = new Phase1Page(page);
    await exerciseRevokedPendingReplay({
      createdBy: "E-6-03",
      incidentKeyPrefix: "E603IDLE",
      page,
      scenario: "idle-expiry",
      sessionTracker,
      triggerRevocation: async () => {
        const currentSession = await phase1.currentSession();
        await phase1.setClockAfter(currentSession.session_expires_at);
      },
    });
    await applyStorageState(page, workerAdmin.storageState);
  });
});
