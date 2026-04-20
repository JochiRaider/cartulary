import {
  type APIRequestContext,
  type Page,
  request,
  type StorageState,
} from "@playwright/test";

import {
  apiBase,
  applyCookies,
  authHeadersForStorageState,
  bootstrapEmail,
  bootstrapPassword,
  csrfCookieName,
  generateTotpCode,
  loadSuiteAdminTotpSecret,
  loginLocalAPIContext,
  requireCookie,
  sessionCookieName,
  storageStateFromCookieValues,
  uniqueTxn,
  waitForAPIReady,
} from "./helpers";
import {
  buildWorkerAdminBlueprints,
  clearWorkerAdminSuiteState,
  ensureWorkerAdminCleanupMarkerDirectory,
  loadWorkerAdminManifest,
  type TrackedSessionSnapshot,
  type WorkerAdminEntry,
  workerAdminNeedsJanitor,
  writeWorkerAdminManifest,
} from "./sessionSupport";

export type WorkerAdmin = WorkerAdminEntry & {
  storageState: StorageState;
};

type AuthenticatedRequestContext = {
  request: APIRequestContext;
  storageState: StorageState;
};

export async function prepareWorkerAdminSuite(workerCount: number) {
  clearWorkerAdminSuiteState();
  ensureWorkerAdminCleanupMarkerDirectory();

  const controlPlane = await loginBootstrapControlPlaneContext();
  try {
    const workerAdmins: WorkerAdminEntry[] = [];
    for (const blueprint of buildWorkerAdminBlueprints(workerCount)) {
      const createResponse = await controlPlane.request.post("/api/v1/users", {
        data: {
          client_txn_id: uniqueTxn(
            `playwright-worker-admin-${blueprint.parallelIndex}`,
          ),
          auth_kind: "local",
          email: blueprint.email,
          display_name: blueprint.displayName,
          initial_password: blueprint.password,
          mfa_required: false,
          is_deployment_admin: true,
        },
      });
      if (!createResponse.ok()) {
        throw new Error(
          `create worker admin ${blueprint.parallelIndex} failed: ${await createResponse.text()}`,
        );
      }
      const body = (await createResponse.json()) as {
        data: { user_id: string };
      };
      workerAdmins.push({
        parallel_index: blueprint.parallelIndex,
        user_id: body.data.user_id,
        email: blueprint.email,
        password: blueprint.password,
      });
    }
    writeWorkerAdminManifest({ worker_admins: workerAdmins });
  } finally {
    await logoutAndVerify(
      controlPlane.request,
      controlPlane.storageState,
      "bootstrap control-plane provisioning",
    );
    await controlPlane.request.dispose();
  }
}

export async function cleanupWorkerAdminSuite() {
  try {
    const manifest = loadWorkerAdminManifest();
    const staleEntries = manifest.worker_admins.filter((entry) =>
      workerAdminNeedsJanitor(entry.parallel_index),
    );
    if (staleEntries.length === 0) {
      return;
    }

    const controlPlane = await loginBootstrapControlPlaneContext();
    try {
      for (const entry of staleEntries) {
        await revokeAllSessions(
          controlPlane.request,
          entry.user_id,
          `playwright global janitor worker=${entry.parallel_index}`,
        );
      }
    } finally {
      await logoutAndVerify(
        controlPlane.request,
        controlPlane.storageState,
        "bootstrap control-plane janitor",
      );
      await controlPlane.request.dispose();
    }
  } catch (error) {
    if (
      error instanceof Error &&
      error.message.startsWith("missing worker admin manifest")
    ) {
      return;
    }
    throw error;
  } finally {
    clearWorkerAdminSuiteState();
  }
}

export async function loginBootstrapControlPlaneContext() {
  const secretBase32 = loadSuiteAdminTotpSecret();
  if (!secretBase32) {
    throw new Error("missing suite admin TOTP state; global setup did not run");
  }

  const anonymousRequests = await request.newContext({ baseURL: apiBase });
  try {
    await waitForAPIReady(anonymousRequests);
    const loginResponse = await loginLocalAPIContext(anonymousRequests, {
      email: bootstrapEmail,
      password: bootstrapPassword,
      secondFactorCode: generateTotpCode(secretBase32),
    });
    if (!loginResponse.ok()) {
      throw new Error(
        `bootstrap control-plane login failed: ${await loginResponse.text()}`,
      );
    }
    const storageState = storageStateFromCookieValues(
      requireCookie(loginResponse, sessionCookieName),
      requireCookie(loginResponse, csrfCookieName),
    );
    return {
      request: await authenticatedRequestContext(storageState),
      storageState,
    } satisfies AuthenticatedRequestContext;
  } finally {
    await anonymousRequests.dispose();
  }
}

export async function loginWorkerAdminContext(entry: WorkerAdminEntry) {
  const anonymousRequests = await request.newContext({ baseURL: apiBase });
  try {
    await waitForAPIReady(anonymousRequests);
    const loginResponse = await loginLocalAPIContext(anonymousRequests, {
      email: entry.email,
      password: entry.password,
    });
    if (!loginResponse.ok()) {
      throw new Error(
        `worker admin login failed for ${entry.email}: ${await loginResponse.text()}`,
      );
    }
    const storageState = storageStateFromCookieValues(
      requireCookie(loginResponse, sessionCookieName),
      requireCookie(loginResponse, csrfCookieName),
    );
    return {
      ...entry,
      storageState,
    } satisfies WorkerAdmin;
  } finally {
    await anonymousRequests.dispose();
  }
}

export async function loginTrackedUserViaPage(
  page: Page,
  options: {
    email: string;
    password: string;
    secondFactorCode?: string | null;
  },
) {
  await page.context().clearCookies();
  const loginResponse = await page.request.post(
    `${apiBase}/api/v1/auth/login`,
    {
      data: {
        username: options.email,
        password: options.password,
        ...(options.secondFactorCode?.trim()
          ? {
              second_factor: {
                kind: "totp",
                assertion: {
                  code: options.secondFactorCode.trim(),
                },
              },
            }
          : {}),
      },
    },
  );
  if (!loginResponse.ok()) {
    throw new Error(
      `tracked user login failed for ${options.email}: ${await loginResponse.text()}`,
    );
  }
  await applyCookies(
    page,
    requireCookie(loginResponse, sessionCookieName),
    requireCookie(loginResponse, csrfCookieName),
  );
}

export async function revokeAllSessions(
  controlPlane: APIRequestContext,
  userId: string,
  reason: string,
) {
  const response = await controlPlane.post(
    `/api/v1/users/${userId}/sessions/revoke-all`,
    {
      data: {
        client_txn_id: uniqueTxn("playwright-revoke-all"),
        reason,
      },
    },
  );
  if (!response.ok()) {
    throw new Error(
      `revoke-all failed for ${userId}: ${await response.text()}`,
    );
  }
}

export async function verifyRevokedSession(snapshot: TrackedSessionSnapshot) {
  await verifySessionUnauthorized(
    snapshot.storageState,
    `${snapshot.userId} (${snapshot.email}) [${snapshot.purpose}]`,
  );
}

export async function verifySessionUnauthorized(
  storageState: StorageState,
  label: string,
) {
  const authRequests = await authenticatedRequestContext(storageState);
  try {
    const sessionResponse = await authRequests.get("/api/v1/auth/session");
    if (sessionResponse.status() !== 401) {
      throw new Error(
        `expected 401 after revocation for ${label}, got ${sessionResponse.status()}`,
      );
    }
    const body = (await sessionResponse.json()) as {
      error?: { code?: string };
    };
    if (body.error?.code !== "session_required") {
      throw new Error(
        `expected session_required after revocation for ${label}, got ${JSON.stringify(body)}`,
      );
    }
  } finally {
    await authRequests.dispose();
  }
}

export async function logoutAndVerify(
  authRequests: APIRequestContext,
  storageState: StorageState,
  label: string,
) {
  const response = await authRequests.post("/api/v1/auth/logout");
  if (!response.ok()) {
    throw new Error(`logout failed for ${label}: ${await response.text()}`);
  }
  await verifySessionUnauthorized(storageState, label);
}

async function authenticatedRequestContext(storageState: StorageState) {
  return request.newContext({
    baseURL: apiBase,
    extraHTTPHeaders: authHeadersForStorageState(storageState),
  });
}
