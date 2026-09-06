import type { GetCurrentSessionResponse } from "@cartulary/protocol-ts/http";
import type { Page } from "@playwright/test";
import { expect } from "@playwright/test";
import { apiBase } from "../runtime/configuration";

type Kind = "profile" | "appearance";
type Fault = "conflict" | "lost" | "transaction";
export function accountResponseGate() {
  let release: () => void = () => {};
  const promise = new Promise<void>((resolve) => {
    release = resolve;
  });
  return { promise, release };
}

/** Browser-only deterministic resources and response gates; never used by production. */
export async function installAccountEditingFixture(page: Page) {
  const sessionResponse = await page.request.get(
    `${apiBase}/api/v1/auth/session`,
  );
  expect(sessionResponse.ok()).toBeTruthy();
  const session = ((await sessionResponse.json()) as GetCurrentSessionResponse)
    .data;
  const timestamp = "2026-09-05T12:00:00Z";
  let profile = {
    user_id: session.user_id,
    email: "analyst@example.test",
    display_name: "Account Analyst",
    user_version: 1,
    created_at: timestamp,
    updated_at: timestamp,
  };
  let preferences = {
    user_id: session.user_id,
    density_mode: null as "compact" | "default" | "comfortable" | null,
    preferences_version: 1,
    created_at: timestamp,
    updated_at: timestamp,
  };
  const faults = new Map<Kind, Fault>();
  const holds = new Map<
    Kind,
    {
      received: ReturnType<typeof accountResponseGate>;
      release: ReturnType<typeof accountResponseGate>;
    }
  >();
  const requests: Record<Kind, Record<string, unknown>[]> = {
    profile: [],
    appearance: [],
  };
  const committed = new Map<
    string,
    { request: Record<string, unknown>; resource: unknown }
  >();
  await page.route("**/api/v1/auth/session", (route) =>
    route.fulfill({
      json: {
        data: {
          ...session,
          display_name: profile.display_name,
          memberships: [],
          is_deployment_admin: false,
        },
        meta: { request_id: "account-fixture-session" },
      },
    }),
  );
  await page.route("**/api/v1/incidents?*", (route) =>
    route.fulfill({
      json: {
        data: { incidents: [] },
        meta: {
          request_id: "account-fixture-directory",
          paging: { limit: 100, has_more: false, next_cursor: null },
        },
      },
    }),
  );
  for (const kind of ["profile", "appearance"] as const) {
    const path = kind === "profile" ? "profile" : "preferences";
    await page.route(`**/api/v1/account/${path}`, async (route) => {
      if (route.request().method() === "GET") {
        await route.fulfill({
          json: {
            data: kind === "profile" ? profile : preferences,
            meta: { request_id: "account-fixture-read" },
          },
        });
        return;
      }
      const request = route.request().postDataJSON() as Record<string, unknown>;
      requests[kind].push(request);
      const hold = holds.get(kind);
      if (hold) {
        holds.delete(kind);
        hold.received.release();
        await hold.release.promise;
      }
      const replay = committed.get(`${kind}:${request.client_txn_id}`);
      if (replay) {
        expect(request).toEqual(replay.request);
        await route.fulfill({
          json: {
            data: replay.resource,
            meta: { request_id: "account-fixture-replay" },
          },
        });
        return;
      }
      const fault = faults.get(kind);
      faults.delete(kind);
      if (fault === "conflict" || fault === "transaction") {
        if (fault === "conflict") {
          if (kind === "profile")
            profile = {
              ...profile,
              display_name: "Saved by another session",
              user_version: profile.user_version + 1,
            };
          else
            preferences = {
              ...preferences,
              density_mode: "default",
              preferences_version: preferences.preferences_version + 1,
            };
        }
        await route.fulfill({
          status: 409,
          json: {
            error: {
              code:
                fault === "transaction"
                  ? "client_txn_conflict"
                  : kind === "profile"
                    ? "user_version_conflict"
                    : "preferences_version_conflict",
              status: 409,
              retryable: false,
            },
          },
        });
        return;
      }
      if (kind === "profile")
        profile = {
          ...profile,
          display_name: String(request.display_name),
          user_version: profile.user_version + 1,
        };
      else
        preferences = {
          ...preferences,
          density_mode: request.density_mode as typeof preferences.density_mode,
          preferences_version: preferences.preferences_version + 1,
        };
      const resource = kind === "profile" ? profile : preferences;
      committed.set(`${kind}:${request.client_txn_id}`, { request, resource });
      if (fault === "lost") {
        await route.abort("failed");
        return;
      }
      await route.fulfill({
        json: { data: resource, meta: { request_id: "account-fixture-save" } },
      });
    });
  }
  return {
    requests,
    fault: (kind: Kind, fault: Fault) => {
      faults.set(kind, fault);
    },
    hold: (kind: Kind) => {
      const received = accountResponseGate();
      const release = accountResponseGate();
      holds.set(kind, { received, release });
      return { received: received.promise, release: release.release };
    },
  };
}
