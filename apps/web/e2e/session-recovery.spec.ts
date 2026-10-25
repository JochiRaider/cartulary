import { expect, test } from "./fixtures";
import { csrfHeaders } from "./support/auth/browserSession";
import { readCurrentSession, revokeAllSessions } from "./support/auth/sessions";
import { exerciseRevokedPendingReplay } from "./support/collaboration/replay";
import { apiBase } from "./support/runtime/configuration";
import { TestClock } from "./support/runtime/testClock";

test("preserves and replays local work after owned-user revoke-all", async ({
  page,
  sessionTracker,
  testActorRequest,
}) => {
  await exerciseRevokedPendingReplay({
    createdBy: "collaboration-recovery",
    incidentKeyPrefix: "E603REVOKE",
    page,
    scenario: "revoke-all",
    sessionTracker,
    triggerRevocation: async ({ member }) => {
      await revokeAllSessions(
        testActorRequest,
        member.user_id,
        "Collaboration recovery owned-user revoke-all",
      );
    },
  });
});

test("preserves and replays local work after current-session logout", async ({
  page,
  sessionTracker,
}) => {
  await exerciseRevokedPendingReplay({
    createdBy: "collaboration-recovery",
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
});

test("preserves unsaved local work after socket idle expiry and re-authentication", async ({
  page,
  sessionTracker,
}) => {
  const clock = new TestClock(page);
  await clock.reset();
  try {
    await exerciseRevokedPendingReplay({
      createdBy: "collaboration-recovery",
      incidentKeyPrefix: "E603IDLE",
      page,
      scenario: "idle-expiry",
      sessionTracker,
      triggerRevocation: async () => {
        const session = await readCurrentSession(page);
        await clock.setAfter(session.session_expires_at);
      },
    });
  } finally {
    await clock.reset();
  }
});
