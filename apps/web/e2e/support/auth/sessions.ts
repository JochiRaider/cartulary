import type {
  PatchDeploymentUserRequest,
  ResetDeploymentUserPasswordRequest,
  ResetDeploymentUserTOTPRequest,
  RevokeAllDeploymentUserSessionsRequest,
  SafeUserResource,
  SessionResource,
} from "@cartulary/protocol-ts/http";
import { authTestId } from "@cartulary/ui-contracts";
import {
  type APIRequestContext,
  expect,
  type Page,
  request,
} from "@playwright/test";
import type { TrackedSessionSnapshot } from "../runtime/cleanup";
import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import {
  isExternalServerHarnessMode,
  usesSharedPlaywrightState,
} from "../runtime/harnessState";
import { waitForAPIReady } from "../runtime/readiness";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import {
  atJsonOrigin,
  type JsonRequestContextLike,
} from "../transport/publicJsonClient";
import {
  applyCookies,
  loginLocalAPIContext,
  requireCookie,
} from "./browserSession";
import { createDeploymentUser } from "./deploymentUsers";
import type { StorageState } from "./storageState";
import {
  authHeadersForStorageState,
  csrfCookieName,
  sessionCookieName,
  storageStateFromCookieValues,
} from "./storageState";
import {
  bootstrapEmail,
  bootstrapPassword,
  generateTotpCode,
  loadSuiteAdminTotpSecret,
} from "./suiteAdmin";
import {
  buildWorkerAdminBlueprints,
  clearWorkerAdminCleanupMarkers,
  ensureWorkerAdminCleanupMarkerDirectory,
  loadWorkerAdminManifest,
  loadWorkerAdminManifestIfPresent,
  type WorkerAdminBlueprint,
  type WorkerAdminEntry,
  type WorkerAdminManifest,
  workerAdminManifestSchemaID,
  workerAdminNeedsJanitor,
  writeWorkerAdminManifest,
} from "./workerAdmin";

export type WorkerAdmin = WorkerAdminEntry & {
  storageState: StorageState;
};

type AuthenticatedRequestContext = {
  request: APIRequestContext;
  storageState: StorageState;
};

export async function readCurrentSession(page: Page): Promise<SessionResource> {
  const storageState = await page.context().storageState();
  const response = await publicHttpOperation({
    headers: authHeadersForStorageState(storageState),
    operationID: "getCurrentSession",
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok) {
    throw new Error(`current session request failed with ${response.status}`);
  }
  return response.payload.data;
}

type ReconcileWorkerAdminManifestOptions = {
  trustExistingManifestPasswords?: boolean;
};

type WorkerAdminControlPlane = {
  canLogin: (email: string, password: string) => Promise<boolean>;
  createUser: (blueprint: WorkerAdminBlueprint) => Promise<SafeUserResource>;
  listUsers: () => Promise<SafeUserResource[]>;
  patchUser: (
    userId: string,
    body: PatchDeploymentUserRequest,
  ) => Promise<SafeUserResource>;
  resetUserPassword: (
    userId: string,
    baseUserVersion: number,
    nextPassword: string,
    reason: string,
  ) => Promise<SafeUserResource>;
  revokeAllSessions: (userId: string, reason: string) => Promise<void>;
};

export type DeploymentAdminMutationClient = {
  listUsers: () => Promise<SafeUserResource[]>;
  loadUser: (userId: string) => Promise<SafeUserResource>;
  patchUser: (
    userId: string,
    body: PatchDeploymentUserRequest,
  ) => Promise<SafeUserResource>;
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
        {
          trustExistingManifestPasswords:
            sharedExternalHarness && existingManifest !== null,
        },
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
    if (!loginResponse.ok) {
      throw new Error(
        `bootstrap control-plane login failed with HTTP ${loginResponse.status}`,
      );
    }
    const storageState = storageStateFromCookieValues(
      requireCookie(loginResponse.response, sessionCookieName),
      requireCookie(loginResponse.response, csrfCookieName),
    );
    return {
      request: await request.newContext({
        baseURL: apiBase,
        extraHTTPHeaders: authHeadersForStorageState(storageState),
      }),
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
    if (!loginResponse.ok) {
      throw new Error(
        `worker admin login failed with HTTP ${loginResponse.status}`,
      );
    }
    const storageState = storageStateFromCookieValues(
      requireCookie(loginResponse.response, sessionCookieName),
      requireCookie(loginResponse.response, csrfCookieName),
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
  const authShell = page.getByTestId(authTestId("shell"));
  if (await authShell.isVisible()) {
    const previousSessionCookie = (await page.context().cookies()).find(
      (cookie) => cookie.name === sessionCookieName,
    )?.value;
    const username = page.getByTestId(authTestId("login-username"));
    const password = page.getByTestId(authTestId("login-password"));
    const submit = page.getByTestId(authTestId("login-submit"));
    await username.fill(options.email);
    await password.fill(options.password);
    const secondFactorCode = options.secondFactorCode?.trim() ?? "";
    if (secondFactorCode !== "") {
      const totpCode = page.getByTestId(authTestId("login-totp-code"));
      if (!(await totpCode.isVisible())) {
        await submit.click();
        await totpCode.waitFor({ state: "visible" });
      }
      await totpCode.fill(secondFactorCode);
    }
    const successfulLoginResponse = page.waitForResponse((response) => {
      if (response.request().method() !== "POST" || !response.ok()) {
        return false;
      }
      return new URL(response.url()).pathname === "/api/v1/auth/login";
    });
    await submit.click();
    await successfulLoginResponse;
    await expect
      .poll(async () => {
        const cookies = await page.context().cookies();
        const sessionCookie = cookies.find(
          (cookie) => cookie.name === sessionCookieName,
        );
        const csrfCookie = cookies.find(
          (cookie) => cookie.name === csrfCookieName,
        );
        return (
          sessionCookie !== undefined &&
          sessionCookie.value !== previousSessionCookie &&
          csrfCookie !== undefined
        );
      })
      .toBe(true);
    const authenticatedCookies = await page.context().cookies();
    const sessionCookie = authenticatedCookies.find(
      (cookie) => cookie.name === sessionCookieName,
    );
    const csrfCookie = authenticatedCookies.find(
      (cookie) => cookie.name === csrfCookieName,
    );
    if (sessionCookie === undefined || csrfCookie === undefined) {
      throw new Error("in-page login did not commit authenticated cookies");
    }
    await applyCookies(page, sessionCookie.value, csrfCookie.value);
    await authShell.waitFor({ state: "hidden" });
    return;
  }

  await page.context().clearCookies();
  const loginResponse = await loginLocalAPIContext(
    atJsonOrigin(page.request, apiBase),
    options,
  );
  if (!loginResponse.ok) {
    throw new Error(
      `tracked user login failed with HTTP ${loginResponse.status}`,
    );
  }
  await applyCookies(
    page,
    requireCookie(loginResponse.response, sessionCookieName),
    requireCookie(loginResponse.response, csrfCookieName),
  );
}

export async function revokeAllSessions(
  controlPlane: JsonRequestContextLike,
  userId: string,
  reason: string,
) {
  const body: RevokeAllDeploymentUserSessionsRequest = {
    client_txn_id: uniqueTxn("playwright-revoke-all"),
    reason,
  };
  const response = await publicHttpOperation({
    body,
    operationID: "revokeAllDeploymentUserSessions",
    pathParameters: { user_id: userId },
    request: controlPlane,
  });
  if (!response.ok) {
    throw new Error(`revoke-all failed with HTTP ${response.status}`);
  }
}

async function resetUserPassword(
  authRequests: JsonRequestContextLike,
  userId: string,
  baseUserVersion: number,
  newPassword: string,
  reason: string,
) {
  const body: ResetDeploymentUserPasswordRequest = {
    base_user_version: baseUserVersion,
    client_txn_id: uniqueTxn("playwright-admin-password-reset"),
    new_password: newPassword,
    reason,
  };
  const response = await publicHttpOperation({
    body,
    operationID: "resetDeploymentUserPassword",
    pathParameters: { user_id: userId },
    request: authRequests,
  });
  if (!response.ok) {
    throw new Error(`reset user password failed with HTTP ${response.status}`);
  }
  return response.payload.data;
}

async function listUsers(authRequests: JsonRequestContextLike) {
  const response = await publicHttpOperation({
    operationID: "listDeploymentUsers",
    request: authRequests,
  });
  if (!response.ok) {
    throw new Error(`list users failed with HTTP ${response.status}`);
  }
  return response.payload.data.users;
}

export async function loadUser(
  authRequests: JsonRequestContextLike,
  userId: string,
) {
  const response = await publicHttpOperation({
    operationID: "getDeploymentUser",
    pathParameters: { user_id: userId },
    request: authRequests,
  });
  if (!response.ok) {
    throw new Error(`load user failed with HTTP ${response.status}`);
  }
  return response.payload.data;
}

export async function patchUser(
  authRequests: JsonRequestContextLike,
  userId: string,
  body: PatchDeploymentUserRequest,
) {
  const response = await publicHttpOperation({
    body,
    operationID: "patchDeploymentUser",
    pathParameters: { user_id: userId },
    request: authRequests,
  });
  if (!response.ok) {
    throw new Error(`patch user failed with HTTP ${response.status}`);
  }
  return response.payload.data;
}

export function deploymentAdminMutationClient(
  authRequests: JsonRequestContextLike,
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
  authRequests: JsonRequestContextLike,
  userId: string,
  baseUserVersion: number,
  reason: string,
) {
  const body: ResetDeploymentUserTOTPRequest = {
    base_user_version: baseUserVersion,
    client_txn_id: uniqueTxn("playwright-admin-totp-reset"),
    reason,
  };
  const response = await publicHttpOperation({
    body,
    operationID: "resetDeploymentUserTOTP",
    pathParameters: { user_id: userId },
    request: authRequests,
  });
  if (!response.ok) {
    throw new Error(`reset user TOTP failed with HTTP ${response.status}`);
  }
  return response.payload.data;
}

export async function verifyRevokedSession(snapshot: TrackedSessionSnapshot) {
  await verifySessionUnauthorized(
    snapshot.storageState,
    "owned session revocation",
  );
}

export async function verifySessionUnauthorized(
  storageState: StorageState,
  label: string,
) {
  const authRequests = await request.newContext({
    baseURL: apiBase,
    extraHTTPHeaders: authHeadersForStorageState(storageState),
  });
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
        `expected session_required after revocation for ${label}, got a different error code`,
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
  const response = await publicHttpOperation({
    operationID: "logoutCurrentSession",
    request: authRequests,
  });
  if (!response.ok) {
    throw new Error(`logout failed for ${label} with HTTP ${response.status}`);
  }
  await verifySessionUnauthorized(storageState, label);
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
  options: ReconcileWorkerAdminManifestOptions = {},
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
    let patched = false;

    if (user === null) {
      user = await controlPlane.createUser(blueprint);
    } else {
      const patchBody = patchBodyForWorkerAdmin(blueprint, user);
      if (patchBody !== null) {
        user = await controlPlane.patchUser(user.user_id, patchBody);
        patched = true;
      }
    }

    const canTrustExistingPassword =
      options.trustExistingManifestPasswords === true &&
      !patched &&
      manifestEntry !== null &&
      manifestEntry.user_id === user.user_id &&
      manifestEntry.email === blueprint.email &&
      manifestEntry.password === blueprint.password;

    if (
      !canTrustExistingPassword &&
      !(await controlPlane.canLogin(blueprint.email, blueprint.password))
    ) {
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

  return {
    schema_id: workerAdminManifestSchemaID,
    worker_admins: nextEntries,
  } satisfies WorkerAdminManifest;
}

function patchBodyForWorkerAdmin(
  blueprint: WorkerAdminBlueprint,
  user: SafeUserResource,
): PatchDeploymentUserRequest | null {
  const patchBody: PatchDeploymentUserRequest = {
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
        return response.ok;
      } finally {
        await anonymousRequests.dispose();
      }
    },
    createUser: async (blueprint) =>
      createDeploymentUser(authRequests, {
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
