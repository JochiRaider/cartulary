import { type APIRequestContext, type Page, request } from "@playwright/test";

import {
  isExternalServerHarnessMode,
  usesSharedPlaywrightState,
} from "./harnessState";
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
import type { StorageState } from "./playwrightTypes";
import {
  buildWorkerAdminBlueprints,
  clearWorkerAdminCleanupMarkers,
  ensureWorkerAdminCleanupMarkerDirectory,
  loadWorkerAdminManifest,
  loadWorkerAdminManifestIfPresent,
  type TrackedSessionSnapshot,
  type WorkerAdminBlueprint,
  type WorkerAdminEntry,
  type WorkerAdminManifest,
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

export type UserResource = {
  user_id: string;
  email: string;
  display_name: string;
  user_version: number;
  is_active: boolean;
  mfa_required: boolean;
  is_deployment_admin: boolean;
};

type PatchUserBody = {
  base_user_version: number;
  display_name?: string;
  is_active?: boolean;
  mfa_required?: boolean;
  is_deployment_admin?: boolean;
};

export type WorkerAdminControlPlane = {
  canLogin: (email: string, password: string) => Promise<boolean>;
  createUser: (blueprint: WorkerAdminBlueprint) => Promise<UserResource>;
  listUsers: () => Promise<UserResource[]>;
  patchUser: (userId: string, body: PatchUserBody) => Promise<UserResource>;
  resetUserPassword: (
    userId: string,
    baseUserVersion: number,
    nextPassword: string,
    reason: string,
  ) => Promise<UserResource>;
  revokeAllSessions: (userId: string, reason: string) => Promise<void>;
};

export type DeploymentAdminMutationClient = {
  listUsers: () => Promise<UserResource[]>;
  loadUser: (userId: string) => Promise<UserResource>;
  patchUser: (userId: string, body: PatchUserBody) => Promise<UserResource>;
};

export async function prepareWorkerAdminSuite(workerCount: number) {
  ensureWorkerAdminCleanupMarkerDirectory();
  const sharedExternalHarness =
    isExternalServerHarnessMode() && usesSharedPlaywrightState();

  const controlPlane = await loginBootstrapControlPlaneContext();
  try {
    const existingManifest = loadWorkerAdminManifestIfPresent();
    const manifestWorkerCount =
      sharedExternalHarness && existingManifest !== null
        ? Math.max(
            workerCount,
            ...existingManifest.worker_admins.map(
              (entry) => entry.parallel_index + 1,
            ),
          )
        : workerCount;
    if (existingManifest !== null && !sharedExternalHarness) {
      await janitorStaleWorkerAdmins(controlPlane.request, existingManifest);
    }
    if (!sharedExternalHarness) {
      clearWorkerAdminCleanupMarkers();
      ensureWorkerAdminCleanupMarkerDirectory();
    }
    writeWorkerAdminManifest(
      await reconcileWorkerAdminManifest(
        controlPlaneClient(controlPlane.request),
        buildWorkerAdminBlueprints(manifestWorkerCount),
        existingManifest,
      ),
    );
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
    const controlPlane = await loginBootstrapControlPlaneContext();
    try {
      await janitorStaleWorkerAdmins(controlPlane.request, manifest);
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

export async function createLocalUser(
  authRequests: APIRequestContext,
  options: {
    email: string;
    display_name: string;
    initial_password: string;
    mfa_required?: boolean;
    is_deployment_admin?: boolean;
  },
) {
  const response = await authRequests.post(`${apiBase}/api/v1/users`, {
    data: {
      client_txn_id: uniqueTxn("phase1-user"),
      auth_kind: "local",
      email: options.email,
      display_name: options.display_name,
      initial_password: options.initial_password,
      mfa_required: options.mfa_required ?? true,
      is_deployment_admin: options.is_deployment_admin ?? false,
    },
  });
  if (!response.ok()) {
    throw new Error(`create local user failed: ${await response.text()}`);
  }
  return ((await response.json()) as { data: UserResource }).data;
}

export async function resetUserPassword(
  authRequests: APIRequestContext,
  userId: string,
  baseUserVersion: number,
  newPassword: string,
  reason: string,
) {
  const response = await authRequests.post(
    `${apiBase}/api/v1/users/${userId}/password/reset`,
    {
      data: {
        base_user_version: baseUserVersion,
        client_txn_id: uniqueTxn("playwright-admin-password-reset"),
        new_password: newPassword,
        reason,
      },
    },
  );
  if (!response.ok()) {
    throw new Error(
      `reset user password failed for ${userId}: ${await response.text()}`,
    );
  }
  return ((await response.json()) as { data: UserResource }).data;
}

export async function listUsers(authRequests: APIRequestContext) {
  const response = await authRequests.get(`${apiBase}/api/v1/users`);
  if (!response.ok()) {
    throw new Error(`list users failed: ${await response.text()}`);
  }
  return ((await response.json()) as { data: { users: UserResource[] } }).data
    .users;
}

export async function loadUser(
  authRequests: APIRequestContext,
  userId: string,
) {
  const response = await authRequests.get(`${apiBase}/api/v1/users/${userId}`);
  if (!response.ok()) {
    throw new Error(`load user ${userId} failed: ${await response.text()}`);
  }
  return ((await response.json()) as { data: UserResource }).data;
}

export async function patchUser(
  authRequests: APIRequestContext,
  userId: string,
  body: {
    base_user_version: number;
    display_name?: string;
    is_active?: boolean;
    mfa_required?: boolean;
    is_deployment_admin?: boolean;
  },
) {
  const response = await authRequests.patch(
    `${apiBase}/api/v1/users/${userId}`,
    {
      data: body,
    },
  );
  if (!response.ok()) {
    throw new Error(`patch user ${userId} failed: ${await response.text()}`);
  }
  return ((await response.json()) as { data: UserResource }).data;
}

export function deploymentAdminMutationClient(
  authRequests: APIRequestContext,
): DeploymentAdminMutationClient {
  return {
    listUsers: async () => listUsers(authRequests),
    loadUser: async (userId) => loadUser(authRequests, userId),
    patchUser: async (userId, body) => patchUser(authRequests, userId, body),
  };
}

export async function withOnlyActiveDeploymentAdmin<T>(
  client: DeploymentAdminMutationClient,
  retainedAdminUserId: string,
  run: () => Promise<T>,
) {
  const demotedUserIds: string[] = [];
  const activeAdmins = (await client.listUsers()).filter(
    (user) =>
      user.user_id !== retainedAdminUserId &&
      user.is_active &&
      user.is_deployment_admin,
  );

  let runError: unknown = null;
  let result: T | undefined;
  try {
    for (const user of activeAdmins) {
      await client.patchUser(user.user_id, {
        base_user_version: user.user_version,
        is_deployment_admin: false,
      });
      demotedUserIds.push(user.user_id);
    }
    result = await run();
  } catch (error) {
    runError = error;
  }

  const restoreFailures: string[] = [];
  for (const userId of demotedUserIds) {
    try {
      const reloaded = await client.loadUser(userId);
      if (reloaded.is_deployment_admin) {
        continue;
      }
      await client.patchUser(userId, {
        base_user_version: reloaded.user_version,
        is_deployment_admin: true,
      });
    } catch (error) {
      restoreFailures.push(`${userId}: ${formatUnknownError(error)}`);
    }
  }

  if (restoreFailures.length > 0) {
    const suffix =
      runError === null
        ? ""
        : `\noriginal guarded block failure: ${formatUnknownError(runError)}`;
    throw new Error(
      `failed to restore deployment-admin status after last-admin probe: ${restoreFailures.join("; ")}${suffix}`,
    );
  }
  if (runError !== null) {
    throw runError;
  }
  return result as T;
}

export async function resetUserTotp(
  authRequests: APIRequestContext,
  userId: string,
  baseUserVersion: number,
  reason: string,
) {
  const response = await authRequests.post(
    `${apiBase}/api/v1/users/${userId}/mfa/totp/reset`,
    {
      data: {
        base_user_version: baseUserVersion,
        client_txn_id: uniqueTxn("playwright-admin-totp-reset"),
        reason,
      },
    },
  );
  if (!response.ok()) {
    throw new Error(
      `reset user TOTP failed for ${userId}: ${await response.text()}`,
    );
  }
  return ((await response.json()) as { data: UserResource }).data;
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

export async function authenticatedRequestContextFromStorageState(
  storageState: StorageState,
) {
  return request.newContext({
    baseURL: apiBase,
    extraHTTPHeaders: authHeadersForStorageState(storageState),
  });
}

async function authenticatedRequestContext(storageState: StorageState) {
  return authenticatedRequestContextFromStorageState(storageState);
}

function formatUnknownError(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  return String(error);
}

export async function reconcileWorkerAdminManifest(
  controlPlane: WorkerAdminControlPlane,
  blueprints: WorkerAdminBlueprint[],
  existingManifest: WorkerAdminManifest | null,
) {
  const existingByIndex = new Map(
    (existingManifest?.worker_admins ?? []).map((entry) => [
      entry.parallel_index,
      entry,
    ]),
  );
  const usersByEmail = new Map(
    (await controlPlane.listUsers()).map((user) => [user.email, user]),
  );
  const usersByID = new Map(
    [...usersByEmail.values()].map((user) => [user.user_id, user]),
  );
  const nextEntries: WorkerAdminEntry[] = [];

  for (const blueprint of blueprints) {
    const manifestEntry = existingByIndex.get(blueprint.parallelIndex) ?? null;
    const manifestUser =
      manifestEntry !== null
        ? (usersByID.get(manifestEntry.user_id) ?? null)
        : null;
    let user =
      usersByEmail.get(blueprint.email) ??
      (manifestUser?.email === blueprint.email ? manifestUser : null) ??
      null;

    if (user === null) {
      user = await controlPlane.createUser(blueprint);
    } else {
      const patchBody = patchBodyForWorkerAdmin(blueprint, user);
      if (patchBody !== null) {
        user = await controlPlane.patchUser(user.user_id, patchBody);
      }
    }

    if (!(await controlPlane.canLogin(blueprint.email, blueprint.password))) {
      user = await controlPlane.resetUserPassword(
        user.user_id,
        user.user_version,
        blueprint.password,
        `playwright worker admin password reconcile worker=${blueprint.parallelIndex}`,
      );
      if (!(await controlPlane.canLogin(blueprint.email, blueprint.password))) {
        throw new Error(
          `worker admin ${blueprint.parallelIndex} could not authenticate after password reconciliation`,
        );
      }
    }

    usersByEmail.set(user.email, user);
    usersByID.set(user.user_id, user);
    nextEntries.push({
      parallel_index: blueprint.parallelIndex,
      user_id: user.user_id,
      email: blueprint.email,
      password: blueprint.password,
    });
  }

  return { worker_admins: nextEntries } satisfies WorkerAdminManifest;
}

function patchBodyForWorkerAdmin(
  blueprint: WorkerAdminBlueprint,
  user: UserResource,
): PatchUserBody | null {
  const patchBody: PatchUserBody = {
    base_user_version: user.user_version,
  };
  if (user.display_name !== blueprint.displayName) {
    patchBody.display_name = blueprint.displayName;
  }
  if (!user.is_active) {
    patchBody.is_active = true;
  }
  if (user.mfa_required) {
    patchBody.mfa_required = false;
  }
  if (!user.is_deployment_admin) {
    patchBody.is_deployment_admin = true;
  }
  return Object.keys(patchBody).length === 1 ? null : patchBody;
}

async function janitorStaleWorkerAdmins(
  controlPlane: APIRequestContext,
  manifest: WorkerAdminManifest,
) {
  const staleEntries = manifest.worker_admins.filter((entry) =>
    workerAdminNeedsJanitor(entry.parallel_index),
  );
  for (const entry of staleEntries) {
    await revokeAllSessions(
      controlPlane,
      entry.user_id,
      `playwright global janitor worker=${entry.parallel_index}`,
    );
  }
}

function controlPlaneClient(
  authRequests: APIRequestContext,
): WorkerAdminControlPlane {
  return {
    canLogin: async (email, password) => {
      const anonymousRequests = await request.newContext({ baseURL: apiBase });
      try {
        await waitForAPIReady(anonymousRequests);
        const response = await loginLocalAPIContext(anonymousRequests, {
          email,
          password,
        });
        return response.ok();
      } finally {
        await anonymousRequests.dispose();
      }
    },
    createUser: async (blueprint) =>
      createLocalUser(authRequests, {
        email: blueprint.email,
        display_name: blueprint.displayName,
        initial_password: blueprint.password,
        is_deployment_admin: true,
        mfa_required: false,
      }),
    listUsers: async () => listUsers(authRequests),
    patchUser: async (userId, body) => patchUser(authRequests, userId, body),
    resetUserPassword: async (userId, baseUserVersion, nextPassword, reason) =>
      resetUserPassword(
        authRequests,
        userId,
        baseUserVersion,
        nextPassword,
        reason,
      ),
    revokeAllSessions: async (userId, reason) =>
      revokeAllSessions(authRequests, userId, reason),
  };
}
