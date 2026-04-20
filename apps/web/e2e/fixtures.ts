import {
  type APIRequestContext,
  type Browser,
  type BrowserContext,
  test as base,
  expect,
  type Page,
  request,
  type StorageState,
} from "@playwright/test";
import {
  loginBootstrapControlPlaneContext,
  loginTrackedUserViaPage,
  loginWorkerAdminContext,
  logoutAndVerify,
  revokeAllSessions,
  verifyRevokedSession,
  verifySessionUnauthorized,
  type WorkerAdmin,
} from "./authRuntime";
import {
  apiBase,
  applyStorageState,
  authHeadersForStorageState,
} from "./helpers";
import {
  loadWorkerAdminManifest,
  markWorkerAdminCleaned,
  OwnedSessionTracker,
} from "./sessionSupport";

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

export const test = base.extend<{
  sessionTracker: SessionTracker;
  workerAdmin: WorkerAdmin;
  workerAdminPage: Page;
  workerAdminRequest: APIRequestContext;
}>({
  workerAdmin: [
    async ({ browserName }, use, workerInfo) => {
      void browserName;
      const manifest = loadWorkerAdminManifest();
      const entry = manifest.worker_admins.find(
        (candidate) => candidate.parallel_index === workerInfo.parallelIndex,
      );
      if (!entry) {
        throw new Error(
          `missing worker admin manifest entry for parallelIndex=${workerInfo.parallelIndex}`,
        );
      }

      const workerAdmin = await loginWorkerAdminContext(entry);
      await use(workerAdmin);

      const controlPlane = await loginBootstrapControlPlaneContext();
      try {
        await revokeAllSessions(
          controlPlane.request,
          workerAdmin.user_id,
          `playwright worker teardown worker=${workerInfo.parallelIndex}`,
        );
        await verifySessionUnauthorized(
          workerAdmin.storageState,
          `${workerAdmin.user_id} (${workerAdmin.email}) worker admin cached session`,
        );
        markWorkerAdminCleaned(workerInfo.parallelIndex);
      } finally {
        await logoutAndVerify(
          controlPlane.request,
          controlPlane.storageState,
          `bootstrap control-plane worker teardown worker=${workerInfo.parallelIndex}`,
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
          secondFactorCode: details.secondFactorCode,
        });
        tracker.registerSession({
          createdBy: details.createdBy,
          email: details.email,
          purpose: details.purpose,
          storageState: await pageAuthStorageState(page),
          userId: details.userId,
        });
      },

      newTrackedContext: async (browser: Browser, storageState: StorageState) =>
        tracker.newTrackedContext(browser, storageState),
    } satisfies SessionTracker;

    await use(sessionTracker);
    await tracker.cleanup();
  },
});

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
