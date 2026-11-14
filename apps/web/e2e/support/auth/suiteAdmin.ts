import { createHmac } from "node:crypto";
import {
  type APIRequestContext,
  expect,
  type Page,
  request,
} from "@playwright/test";
import {
  readBrowserState,
  writeBrowserState,
} from "../../../../../tools/harness/browser/private-state.mjs";
import { browserScopeFor } from "../runtime/cleanup";
import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import { waitForAPIReady } from "../runtime/readiness";
import { loginLocalAPIContext } from "./browserSession";
import { withOwnedRequestContext } from "./requestLifetime";

export const bootstrapEmail = "dev-admin@example.test";
export const bootstrapPassword = "DevBootstrap1!";

const suiteAdminTotpStateFile = "cartulary-playwright-admin-totp.txt";

type LocalLoginResult =
  | { kind: "success" }
  | {
      kind: "error";
      status: number;
      code: string;
      details: Record<string, unknown>;
    };

type SuiteAdminAuthClient = {
  loginLocal: (secondFactorCode?: string | null) => Promise<LocalLoginResult>;
  provisionTotpFromBootstrap: (bootstrapToken: string) => Promise<string>;
};

export async function prepareSuiteAdminState() {
  await withOwnedRequestContext(
    await request.newContext({ baseURL: apiBase }),
    async (authRequests) => {
      await waitForAPIReady(authRequests);
      const secretBase32 = await reconcileSuiteAdminTotpState(
        suiteAdminAuthClient(authRequests),
        loadSuiteAdminTotpSecret(),
      );
      writeSuiteAdminTotpSecret(secretBase32);
    },
  );
}

export async function enrollTotpViaBootstrap(
  page: Page,
  email: string,
  password: string,
) {
  const authRequests = await browserScopeFor(page).ownRequestContext(() =>
    request.newContext({ baseURL: apiBase }),
  );
  await waitForAPIReady(authRequests);
  return provisionUserTotp(authRequests, email, password);
}

export function generateTotpCode(secretBase32: string, at = Date.now()) {
  const secret = decodeBase32(secretBase32);
  const counter = Math.floor(at / 1000 / 30);
  const counterBuffer = Buffer.alloc(8);
  counterBuffer.writeBigUInt64BE(BigInt(counter));
  const digest = createHmac("sha256", secret).update(counterBuffer).digest();
  const offsetSource = digest.at(-1);
  if (offsetSource === undefined) {
    throw new Error("empty TOTP digest");
  }
  const offset = offsetSource & 0x0f;
  const codeBytes = digest.subarray(offset, offset + 4);
  const [byte0, byte1, byte2, byte3] = codeBytes;
  if (
    codeBytes.length !== 4 ||
    byte0 === undefined ||
    byte1 === undefined ||
    byte2 === undefined ||
    byte3 === undefined
  ) {
    throw new Error("short TOTP digest window");
  }
  const code =
    ((byte0 & 0x7f) << 24) |
    ((byte1 & 0xff) << 16) |
    ((byte2 & 0xff) << 8) |
    (byte3 & 0xff);
  return String(code % 1_000_000).padStart(6, "0");
}

export async function reconcileSuiteAdminTotpState(
  client: SuiteAdminAuthClient,
  storedSecretBase32: string | null,
) {
  const normalizedStoredSecret = storedSecretBase32?.trim() ?? "";
  if (normalizedStoredSecret !== "") {
    const loginWithStoredSecret = await client.loginLocal(
      generateTotpCode(normalizedStoredSecret),
    );
    if (loginWithStoredSecret.kind === "success") {
      return normalizedStoredSecret;
    }
    if (loginWithStoredSecret.code === "mfa_setup_required") {
      return provisionSuiteAdminTotp(client, loginWithStoredSecret.details);
    }
    throw suiteAdminStateError(
      `stored suite admin TOTP state no longer matches the current backend state (login code ${loginWithStoredSecret.code})`,
    );
  }

  const loginWithoutSecondFactor = await client.loginLocal();
  if (loginWithoutSecondFactor.kind === "success") {
    throw new Error(
      "suite admin login unexpectedly succeeded without MFA during harness setup",
    );
  }
  if (loginWithoutSecondFactor.code === "mfa_setup_required") {
    return provisionSuiteAdminTotp(client, loginWithoutSecondFactor.details);
  }
  if (loginWithoutSecondFactor.code === "mfa_required") {
    throw suiteAdminStateError(
      "suite admin MFA is already active but harness authentication state is unavailable",
    );
  }
  throw new Error(
    `suite admin harness login failed with ${loginWithoutSecondFactor.code}`,
  );
}

export function loadSuiteAdminTotpSecret() {
  const secret = readBrowserState(suiteAdminTotpStateFile)?.trim() ?? "";
  return secret === "" ? null : secret;
}

function writeSuiteAdminTotpSecret(secretBase32: string) {
  writeBrowserState(suiteAdminTotpStateFile, `${secretBase32}\n`);
}

async function provisionUserTotp(
  authRequests: APIRequestContext,
  email: string,
  password: string,
) {
  const loginResponse = await loginLocalAPIContext(authRequests, {
    email,
    password,
  });
  const loginResult = readLocalLoginResult(
    loginResponse.status,
    loginResponse.ok ? null : loginResponse.payload,
  );
  if (
    loginResult.kind !== "error" ||
    loginResult.status !== 401 ||
    loginResult.code !== "mfa_setup_required"
  ) {
    throw new Error(
      `expected mfa_setup_required while provisioning TOTP, got ${formatLocalLoginResult(loginResult)}`,
    );
  }
  return provisionTotpFromBootstrap(
    authRequests,
    requireBootstrapToken(loginResult.details),
  );
}

function suiteAdminAuthClient(
  authRequests: APIRequestContext,
): SuiteAdminAuthClient {
  return {
    loginLocal: async (secondFactorCode) => {
      const response = await loginLocalAPIContext(authRequests, {
        email: bootstrapEmail,
        password: bootstrapPassword,
        ...(secondFactorCode === undefined ? {} : { secondFactorCode }),
      });
      return readLocalLoginResult(
        response.status,
        response.ok ? null : response.payload,
      );
    },
    provisionTotpFromBootstrap: async (bootstrapToken) =>
      provisionTotpFromBootstrap(authRequests, bootstrapToken),
  };
}

function readLocalLoginResult(
  status: number,
  payload: unknown,
): LocalLoginResult {
  if (payload === null) {
    return { kind: "success" };
  }
  const body = payload as {
    error?: { code?: string; details?: unknown };
  };
  return {
    kind: "error",
    status,
    code: body.error?.code ?? "unknown_error",
    details: toErrorDetails(body.error?.details),
  };
}

function formatLocalLoginResult(result: LocalLoginResult) {
  return result.kind === "success"
    ? "success"
    : `${result.status} ${result.code}`;
}

function toErrorDetails(value: unknown) {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : {};
}

async function provisionSuiteAdminTotp(
  client: SuiteAdminAuthClient,
  details: Record<string, unknown>,
) {
  return client.provisionTotpFromBootstrap(requireBootstrapToken(details));
}

async function provisionTotpFromBootstrap(
  authRequests: APIRequestContext,
  bootstrapToken: string,
) {
  const beginResponse = await authRequests.post("/api/v1/auth/mfa/totp/begin", {
    headers: { Authorization: `Bearer ${bootstrapToken}` },
    data: { client_txn_id: uniqueTxn("bootstrap-begin") },
  });
  expect(beginResponse.ok()).toBeTruthy();
  const beginBody = (await beginResponse.json()) as {
    data: { enrollment_id: string; totp_setup: { secret_base32: string } };
  };
  const secretBase32 = beginBody.data.totp_setup.secret_base32;
  const completeResponse = await authRequests.post(
    "/api/v1/auth/mfa/totp/complete",
    {
      headers: { Authorization: `Bearer ${bootstrapToken}` },
      data: {
        client_txn_id: uniqueTxn("bootstrap-complete"),
        enrollment_id: beginBody.data.enrollment_id,
        code: generateTotpCode(secretBase32),
      },
    },
  );
  expect(completeResponse.ok()).toBeTruthy();
  return secretBase32;
}

function requireBootstrapToken(details: Record<string, unknown>) {
  const bootstrapToken = details.bootstrap_token;
  if (typeof bootstrapToken === "string" && bootstrapToken.trim() !== "") {
    return bootstrapToken;
  }
  throw suiteAdminStateError(
    "suite admin login did not return the required enrollment credential",
  );
}

function suiteAdminStateError(message: string) {
  return new Error(
    `${message}. The acquired browser stack owns this state for its full lifetime.`,
  );
}

function decodeBase32(input: string) {
  const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";
  const normalized = input.replace(/=+$/u, "").toUpperCase();
  let bits = "";
  for (const character of normalized) {
    const index = alphabet.indexOf(character);
    if (index < 0) {
      throw new Error("invalid base32 input");
    }
    bits += index.toString(2).padStart(5, "0");
  }
  const bytes: number[] = [];
  for (let index = 0; index + 8 <= bits.length; index += 8) {
    bytes.push(Number.parseInt(bits.slice(index, index + 8), 2));
  }
  return Buffer.from(bytes);
}
