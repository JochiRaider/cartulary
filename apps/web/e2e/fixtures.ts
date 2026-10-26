import { createHash, randomUUID } from "node:crypto";
import {
  type APIRequestContext,
  test as base,
  expect,
  type Page,
  request,
} from "@playwright/test";
import { applyStorageState } from "./support/auth/browserSession";
import { createDeploymentUser } from "./support/auth/deploymentUsers";
import {
  findProvisionedUser,
  loginBootstrapControlPlaneContext,
  loginFixtureActorContext,
  loginTrackedUserViaPage,
  logoutAndVerify,
  revokeAllSessions,
  verifySessionUnauthorized,
  type WorkerAdmin,
} from "./support/auth/sessions";
import type { StorageState } from "./support/auth/storageState";
import { authHeadersForStorageState } from "./support/auth/storageState";
import {
  loadWorkerAdminManifest,
  markWorkerAdminCleaned,
} from "./support/auth/workerAdmin";
import { installOwnedJobs } from "./support/jobs/ownership";
import { OwnedBrowserScope } from "./support/runtime/cleanup";
import { apiBase } from "./support/runtime/configuration";

export type SessionTracker = {
  loginProvisionedUser: (
    page: Page,
    details: {
      email: string;
      password: string;
      createdBy: string;
      purpose: string;
    },
  ) => Promise<void>;
  captureCurrentSession: (
    page: Page,
    details: {
      createdBy: string;
      email: string;
      purpose: string;
      userId: string;
    },
  ) => Promise<void>;
  captureStorageState: (
    storageState: StorageState,
    details: {
      createdBy: string;
      email: string;
      purpose: string;
      userId: string;
    },
  ) => Promise<void>;
  loginTrackedUser: (
    page: Page,
    details: {
      recovery?: boolean;
      createdBy: string;
      email: string;
      password: string;
      purpose: string;
      secondFactorCode?: string | null;
      userId: string;
    },
  ) => Promise<void>;
  newContext: OwnedBrowserScope["newContext"];
  registerActor: OwnedBrowserScope["registerActor"];
};

type CartularyTestFixtures = {
  browserScope: OwnedBrowserScope;
  jobSettlementTimeout: number;
  testActor: Omit<WorkerAdmin, "parallel_index">;
  testActorEmail: string | undefined;
  sessionTracker: SessionTracker;
  testActorPage: Page;
  testActorRequest: APIRequestContext;
};

type CartularyWorkerFixtures = {
  fixtureProvisioner: APIRequestContext;
};

export const test = base.extend<CartularyTestFixtures, CartularyWorkerFixtures>(
  {
    fixtureProvisioner: [
      async ({ browserName }, use, workerInfo) => {
        void browserName;
        const workerAdminIndex = workerAdminIndexForParallelIndex(
          workerInfo.parallelIndex,
        );
        const manifest = loadWorkerAdminManifest();
        const entry = manifest.worker_admins.find(
          (candidate) => candidate.parallel_index === workerAdminIndex,
        );
        if (!entry) {
          throw new Error(
            `missing worker admin manifest entry for parallelIndex=${workerAdminIndex}`,
          );
        }

        let actor:
          | Awaited<ReturnType<typeof loginFixtureActorContext>>
          | undefined;
        let provisioner: APIRequestContext | undefined;
        let primary: unknown;
        let failed = false;
        try {
          actor = await loginFixtureActorContext(entry);
          provisioner = await request.newContext({
            baseURL: apiBase,
            extraHTTPHeaders: authHeadersForStorageState(actor.storageState),
          });
          await use(provisioner);
        } catch (error) {
          failed = true;
          primary = error;
        }

        // Worker control-plane ownership is outside every test scope. Attempt
        // each retirement stage even after setup/use or another stage fails.
        const failures: string[] = [];
        const attempt = async (
          stage: string,
          action: () => Promise<unknown>,
        ) => {
          try {
            await action();
            return true;
          } catch {
            failures.push(stage);
            return false;
          }
        };
        let controlPlane:
          | Awaited<ReturnType<typeof loginBootstrapControlPlaneContext>>
          | undefined;
        await attempt("control_plane.acquire", async () => {
          controlPlane = await loginBootstrapControlPlaneContext();
        });
        if (controlPlane) {
          const control = controlPlane;
          const revoked = await attempt("worker.revoke", () =>
            revokeAllSessions(
              control.request,
              entry.user_id,
              "playwright worker cleanup",
            ),
          );
          const acquiredActor = actor;
          const verified =
            acquiredActor === undefined ||
            (await attempt("worker.verify", () =>
              verifySessionUnauthorized(
                acquiredActor.storageState,
                "worker session",
                control.request,
              ),
            ));
          if (revoked && verified) markWorkerAdminCleaned(workerAdminIndex);
          await attempt("control_plane.logout", () =>
            logoutAndVerify(
              control.request,
              control.storageState,
              "worker control-plane cleanup",
            ),
          );
          await attempt("control_plane.dispose", () =>
            control.request.dispose(),
          );
        }
        if (provisioner) {
          const owned = provisioner;
          await attempt("provisioner.dispose", () => owned.dispose());
        }
        if (failures.length) {
          const cleanup = new Error(
            `Worker control-plane cleanup failed: ${failures.join(", ")}`,
          );
          if (failed)
            throw new AggregateError(
              [primary, cleanup],
              "Worker setup/use and cleanup failed",
            );
          throw cleanup;
        }
        if (failed) throw primary;
      },
      { scope: "worker" },
    ],

    jobSettlementTimeout: [30_000, { option: true }],

    browserScope: [
      async ({ browser, fixtureProvisioner, jobSettlementTimeout }, use) => {
        const scope: OwnedBrowserScope = new OwnedBrowserScope({
          browser,
          initializeContext: (context) =>
            installOwnedJobs(context, scope, jobSettlementTimeout),
          revokeAllSessions: (userId, reason) =>
            revokeAllSessions(fixtureProvisioner, userId, reason),
          verifyRevokedSession: (snapshot) =>
            verifySessionUnauthorized(
              snapshot.storageState,
              "owned session revocation",
              fixtureProvisioner,
            ),
        });
        try {
          await use(scope);
        } finally {
          await scope.dispose();
        }
      },
      { auto: true },
    ],

    // A presentation fixture may own literal visible identity independently of
    // Playwright's file/title-derived test ID. Every actor remains freshly owned.
    testActorEmail: [undefined, { option: true }],
    testActor: async (
      { fixtureProvisioner, browserScope, testActorEmail },
      use,
      testInfo,
    ) => {
      // Stable presentation identity, independently persisted for each selected
      // test. Credentials are ephemeral and never enter retained diagnostics.
      const identity = createHash("sha256")
        .update(testInfo.testId)
        .digest("hex")
        .slice(0, 24);
      const email =
        testActorEmail ?? `playwright-test-${identity}@example.test`;
      const password = `Playwright-${randomUUID()}-A1!`;
      const user = await browserScope.acquireActor(() =>
        createDeploymentUser(fixtureProvisioner, {
          email,
          initial_password: password,
          display_name: "Playwright Test Actor",
          is_deployment_admin: true,
          mfa_required: false,
        }),
      );
      const actor = await loginFixtureActorContext(
        {
          user_id: user.user_id,
          email,
          password,
        },
        browserScope,
      );
      await use(actor);
    },

    context: async ({ browserScope, testActor }, use) => {
      const context = await browserScope.newContext({
        storageState: testActor.storageState,
      });
      await use(context);
    },

    page: async ({ context }, use) => {
      const page = await context.newPage();
      await use(page);
    },

    testActorPage: async ({ page }, use) => {
      await use(page);
    },

    testActorRequest: async ({ testActor, browserScope }, use) => {
      const authRequests = await browserScope.ownRequestContext(() =>
        request.newContext({
          baseURL: apiBase,
          extraHTTPHeaders: authHeadersForStorageState(testActor.storageState),
        }),
      );
      await use(authRequests);
    },

    sessionTracker: async ({ browserScope, fixtureProvisioner }, use) => {
      const sessionTracker: SessionTracker = {
        loginProvisionedUser: async (page, details) => {
          const userId = await findProvisionedUser(
            fixtureProvisioner,
            details.email,
          );
          await sessionTracker.loginTrackedUser(page, { ...details, userId });
        },
        captureCurrentSession: async (
          page: Page,
          details: {
            recovery?: boolean;
            createdBy: string;
            email: string;
            purpose: string;
            userId: string;
          },
        ) => {
          await browserScope.acquireSession(details, async () => ({
            storageState: await pageAuthStorageState(page),
          }));
        },

        captureStorageState: async (
          storageState: StorageState,
          details: {
            recovery?: boolean;
            createdBy: string;
            email: string;
            purpose: string;
            userId: string;
          },
        ) => {
          browserScope.registerSession({
            ...details,
            storageState,
          });
        },

        loginTrackedUser: async (
          page: Page,
          details: {
            recovery?: boolean;
            createdBy: string;
            email: string;
            password: string;
            purpose: string;
            secondFactorCode?: string | null;
            userId: string;
          },
        ) => {
          await browserScope.acquireSession(details, async () => {
            await loginTrackedUserViaPage(page, {
              ...(details.recovery === undefined
                ? {}
                : { recovery: details.recovery }),
              email: details.email,
              password: details.password,
              ...(details.secondFactorCode === undefined
                ? {}
                : { secondFactorCode: details.secondFactorCode }),
            });
            return { storageState: await pageAuthStorageState(page) };
          });
        },

        newContext: (options) => browserScope.newContext(options),
        registerActor: (userId) => browserScope.registerActor(userId),
      } satisfies SessionTracker;

      await use(sessionTracker);
    },
  },
);

export async function restoreTrackedStorageState(
  page: Page,
  storageState: StorageState,
) {
  await applyStorageState(page, storageState);
}

export { expect };

async function pageAuthStorageState(page: Page): Promise<StorageState> {
  return page.context().storageState();
}

export function workerAdminIndexForParallelIndex(parallelIndex: number) {
  if (process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET !== "0") {
    throw new Error("owned browser worker offset must be exactly zero");
  }
  if (process.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT !== "1") {
    throw new Error("owned browser worker count must be exactly one");
  }
  if (parallelIndex !== 0) {
    throw new Error("owned browser parallel index must be exactly zero");
  }
  return 0;
}
