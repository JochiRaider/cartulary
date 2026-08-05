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
import {
  loginBootstrapControlPlaneContext,
  loginTrackedUserViaPage,
  loginWorkerAdminContext,
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
import { OwnedSessionTracker } from "./support/runtime/cleanup";
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
  sessionTracker: SessionTracker;
  workerAdminPage: Page;
  workerAdminRequest: APIRequestContext;
};

type CartularyWorkerFixtures = {
  workerAdmin: WorkerAdmin;
};

export const test = base.extend<CartularyTestFixtures, CartularyWorkerFixtures>(
  {
    workerAdmin: [
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

        const workerAdmin = await loginWorkerAdminContext(entry);
        await use(workerAdmin);

        const controlPlane = await loginBootstrapControlPlaneContext();
        try {
          await revokeAllSessions(
            controlPlane.request,
            workerAdmin.user_id,
            `playwright worker teardown worker=${workerAdminIndex}`,
          );
          await verifySessionUnauthorized(
            workerAdmin.storageState,
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

    context: async ({ browser, workerAdmin }, use) => {
      const context = await browser.newContext({
        storageState: workerAdmin.storageState,
      });
      await use(context);
      await context.close();
    },

    page: async ({ context }, use) => {
      const page = await context.newPage();
      await use(page);
    },

    workerAdminPage: async ({ page }, use) => {
      await use(page);
    },

    workerAdminRequest: async ({ workerAdmin }, use) => {
      const authRequests = await request.newContext({
        baseURL: apiBase,
        extraHTTPHeaders: authHeadersForStorageState(workerAdmin.storageState),
      });
      await use(authRequests);
      await authRequests.dispose();
    },

    sessionTracker: async ({ workerAdminRequest }, use, testInfo) => {
      const tracker = new OwnedSessionTracker({
        label: `${testInfo.file} :: ${testInfo.title}`,
        revokeAllSessions: async (userId, reason) => {
          await revokeAllSessions(workerAdminRequest, userId, reason);
        },
        verifyRevokedSession,
      });

      const sessionTracker = {
        captureCurrentSession: async (
          page: Page,
          details: {
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
            createdBy: string;
            email: string;
            password: string;
            purpose: string;
            secondFactorCode?: string | null;
            userId: string;
          },
        ) => {
          await loginTrackedUserViaPage(page, {
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

function isScheduledBrowserGroupInvocation() {
  return Boolean(
    process.env.CARTULARY_BROWSER_SESSION_GROUP ||
      process.env.CARTULARY_BROWSER_GROUP_KIND,
  );
}

function integerEnv(name: string, options: { min: number; required: boolean }) {
  const value = process.env[name];
  if (value === undefined || value.trim() === "") {
    if (options.required) {
      throw new Error(`${name} is required for scheduled browser groups`);
    }
    return null;
  }
  const parsed = Number.parseInt(value, 10);
  if (
    !Number.isInteger(parsed) ||
    parsed < options.min ||
    String(parsed) !== value
  ) {
    const description =
      options.min === 0 ? "a non-negative integer" : "a positive integer";
    throw new Error(`${name} must be ${description}`);
  }
  return parsed;
}

function workerAdminIndexOffset(scheduled: boolean) {
  return (
    integerEnv("CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET", {
      min: 0,
      required: scheduled,
    }) ?? 0
  );
}

export function workerAdminIndexForParallelIndex(parallelIndex: number) {
  if (!Number.isInteger(parallelIndex) || parallelIndex < 0) {
    throw new Error("Playwright parallelIndex must be a non-negative integer");
  }
  const scheduled = isScheduledBrowserGroupInvocation();
  const offset = workerAdminIndexOffset(scheduled);
  if (!scheduled) {
    return parallelIndex + offset;
  }
  const workerCount = integerEnv("CARTULARY_PLAYWRIGHT_WORKER_COUNT", {
    min: 1,
    required: true,
  });
  const workerAdminIndex = parallelIndex + offset;
  if (workerCount === null || workerAdminIndex >= workerCount) {
    throw new Error(
      `scheduled browser group worker slot ${workerAdminIndex} is outside CARTULARY_PLAYWRIGHT_WORKER_COUNT=${workerCount}`,
    );
  }
  return workerAdminIndex;
}
