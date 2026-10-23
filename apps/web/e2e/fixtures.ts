import { createHash, randomUUID } from "node:crypto";
import {
  type APIRequestContext,
  type Browser,
  type BrowserContext,
  test as base,
  expect,
  type Page,
  request,
} from "@playwright/test";
import { applyStorageState } from "./support/auth/browserSession";
import { createDeploymentUser } from "./support/auth/deploymentUsers";
import {
  loginBootstrapControlPlaneContext,
  loginFixtureActorContext,
  loginTrackedUserViaPage,
  logoutAndVerify,
  revokeAllSessions,
  verifyRevokedSession,
  verifySessionUnauthorized,
  type WorkerAdmin,
} from "./support/auth/sessions";
import type { StorageState } from "./support/auth/storageState";
import { authHeadersForStorageState } from "./support/auth/storageState";
import {
  loadWorkerAdminManifest,
  markWorkerAdminCleaned,
} from "./support/auth/workerAdmin";
import {
  closeOwnedBrowserContext,
  OwnedSessionTracker,
} from "./support/runtime/cleanup";
import { apiBase } from "./support/runtime/configuration";

type SessionTracker = {
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
  newTrackedContext: (
    browser: Browser,
    storageState: StorageState,
  ) => Promise<BrowserContext>;
};

type CartularyTestFixtures = {
  testActor: Omit<WorkerAdmin, "parallel_index">;
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

        const testActor = await loginFixtureActorContext(entry);
        const provisioner = await request.newContext({
          baseURL: apiBase,
          extraHTTPHeaders: authHeadersForStorageState(testActor.storageState),
        });
        try {
          await use(provisioner);
        } finally {
          await provisioner.dispose();
        }

        const controlPlane = await loginBootstrapControlPlaneContext();
        try {
          await revokeAllSessions(
            controlPlane.request,
            testActor.user_id,
            `playwright worker teardown worker=${workerAdminIndex}`,
          );
          await verifySessionUnauthorized(
            testActor.storageState,
            `worker admin cached session worker=${workerAdminIndex}`,
          );
          markWorkerAdminCleaned(workerAdminIndex);
        } finally {
          await logoutAndVerify(
            controlPlane.request,
            controlPlane.storageState,
            `bootstrap control-plane worker teardown worker=${workerAdminIndex}`,
          );
          await controlPlane.request.dispose();
        }
      },
      { scope: "worker" },
    ],

    testActor: async ({ fixtureProvisioner }, use, testInfo) => {
      // Stable presentation identity, independently persisted for each selected
      // test. Credentials are ephemeral and never enter retained diagnostics.
      const identity = createHash("sha256")
        .update(testInfo.testId)
        .digest("hex")
        .slice(0, 24);
      const email = `playwright-test-${identity}@example.test`;
      const password = `Playwright-${randomUUID()}-A1!`;
      const user = await createDeploymentUser(fixtureProvisioner, {
        email,
        initial_password: password,
        display_name: "Playwright Test Actor",
        is_deployment_admin: true,
        mfa_required: false,
      });
      let actor: Omit<WorkerAdmin, "parallel_index"> | undefined;
      try {
        actor = await loginFixtureActorContext({
          user_id: user.user_id,
          email,
          password,
        });
        await use(actor);
      } finally {
        await revokeAllSessions(
          fixtureProvisioner,
          user.user_id,
          "playwright test actor teardown",
        );
        if (actor)
          await verifySessionUnauthorized(
            actor.storageState,
            "test actor cached session",
          );
      }
    },

    context: async (
      { browser, testActor, viewport, deviceScaleFactor },
      use,
    ) => {
      const context = await browser.newContext({
        storageState: testActor.storageState,
        viewport,
        ...(deviceScaleFactor === undefined ? {} : { deviceScaleFactor }),
      });
      try {
        await use(context);
      } finally {
        await closeOwnedBrowserContext(context);
      }
    },

    page: async ({ context }, use) => {
      const page = await context.newPage();
      await use(page);
    },

    testActorPage: async ({ page }, use) => {
      await use(page);
    },

    testActorRequest: async ({ testActor }, use) => {
      const authRequests = await request.newContext({
        baseURL: apiBase,
        extraHTTPHeaders: authHeadersForStorageState(testActor.storageState),
      });
      await use(authRequests);
      await authRequests.dispose();
    },

    sessionTracker: async ({ fixtureProvisioner }, use, testInfo) => {
      const tracker = new OwnedSessionTracker({
        label: `${testInfo.file} :: ${testInfo.title}`,
        revokeAllSessions: async (userId, reason) => {
          await revokeAllSessions(fixtureProvisioner, userId, reason);
        },
        verifyRevokedSession,
      });

      const sessionTracker = {
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
          tracker.registerSession({
            ...details,
            storageState: await pageAuthStorageState(page),
          });
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
          tracker.registerSession({
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
          tracker.registerSession({
            createdBy: details.createdBy,
            email: details.email,
            purpose: details.purpose,
            storageState: await pageAuthStorageState(page),
            userId: details.userId,
          });
        },

        newTrackedContext: async (
          browser: Browser,
          storageState: StorageState,
        ) => tracker.newTrackedContext(browser, storageState),
      } satisfies SessionTracker;

      await use(sessionTracker);
      await tracker.cleanup();
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
