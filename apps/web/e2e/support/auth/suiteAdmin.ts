import { createHmac } from "node:crypto";
import { existsSync, readFileSync, unlinkSync } from "node:fs";
import { type APIRequestContext, expect, request } from "@playwright/test";
import { apiBase } from "../runtime/configuration";
import { uniqueTxn } from "../runtime/fixtureIdentity";
import {
  isExternalServerHarnessMode,
  resolvePlaywrightStateFile,
  sharedPlaywrightStateDir,
} from "../runtime/harnessState";
import { atomicWritePrivateFile } from "../runtime/privateState";
import { waitForAPIReady } from "../runtime/readiness";
import { loginLocalAPIContext } from "./browserSession";

export const bootstrapEmail = "dev-admin@example.test";
export const bootstrapPassword = "DevBootstrap1!";

const suiteAdminTotpStatePath = resolvePlaywrightStateFile(
  "cartulary-playwright-admin-totp.txt",
);

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

type SuiteAdminStateContext = {
  externalServerMode: boolean;
  sharedStateDir: string | null;
  stateFilePath: string;
};

export async function prepareSuiteAdminState() {
  const authRequests = await request.newContext({ baseURL: apiBase });
  try {
    await waitForAPIReady(authRequests);
    const secretBase32 = await reconcileSuiteAdminTotpState(
      suiteAdminAuthClient(authRequests),
      loadSuiteAdminTotpSecret(),
    );
    writeSuiteAdminTotpSecret(secretBase32);
  } finally {
    await authRequests.dispose();
  }
}

export async function enrollTotpViaBootstrap(email: string, password: string) {
  const authRequests = await request.newContext({ baseURL: apiBase });
  try {
    await waitForAPIReady(authRequests);
    return await provisionUserTotp(authRequests, email, password);
  } finally {
    await authRequests.dispose();
  }
}

export function generateTotpCode(secretBase32: string) {
  const secret = decodeBase32(secretBase32);
  const counter = Math.floor(Date.now() / 1000 / 30);
  const counterBuffer = Buffer.alloc(8);
  counterBuffer.writeBigUInt64BE(BigInt(counter));
  const digest = createHmac("sha1", secret).update(counterBuffer).digest();
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

function currentSuiteAdminStateContext(): SuiteAdminStateContext {
  return {
    externalServerMode: isExternalServerHarnessMode(),
    sharedStateDir: sharedPlaywrightStateDir(),
    stateFilePath: suiteAdminTotpStatePath,
  };
}

export async function reconcileSuiteAdminTotpState(
  client: SuiteAdminAuthClient,
  storedSecretBase32: string | null,
  context: SuiteAdminStateContext = currentSuiteAdminStateContext(),
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
      return provisionSuiteAdminTotp(
        client,
        loginWithStoredSecret.details,
        context,
      );
    }
    throw suiteAdminStateError(
      `stored suite admin TOTP state no longer matches the current backend state (login code ${loginWithStoredSecret.code})`,
      context,
    );
  }

  const loginWithoutSecondFactor = await client.loginLocal();
  if (loginWithoutSecondFactor.kind === "success") {
    throw new Error(
      "suite admin login unexpectedly succeeded without MFA during harness setup",
    );
  }
  if (loginWithoutSecondFactor.code === "mfa_setup_required") {
    return provisionSuiteAdminTotp(
      client,
      loginWithoutSecondFactor.details,
      context,
    );
  }
  if (loginWithoutSecondFactor.code === "mfa_required") {
    throw suiteAdminStateError(
      "suite admin MFA is already active but harness authentication state is unavailable",
      context,
    );
  }
  throw new Error(
    `suite admin harness login failed with ${loginWithoutSecondFactor.code}`,
  );
}

export function loadSuiteAdminTotpSecret() {
  if (!existsSync(suiteAdminTotpStatePath)) {
    return null;
  }
  const secret = readFileSync(suiteAdminTotpStatePath, "utf8").trim();
  return secret === "" ? null : secret;
}

function writeSuiteAdminTotpSecret(secretBase32: string) {
  atomicWritePrivateFile(suiteAdminTotpStatePath, `${secretBase32}\n`);
}

export function clearSuiteAdminTotpSecret() {
  if (existsSync(suiteAdminTotpStatePath)) {
    unlinkSync(suiteAdminTotpStatePath);
  }
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
    requireBootstrapToken(loginResult.details, currentSuiteAdminStateContext()),
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
  context: SuiteAdminStateContext,
) {
  return client.provisionTotpFromBootstrap(
    requireBootstrapToken(details, context),
  );
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

function requireBootstrapToken(
  details: Record<string, unknown>,
  context: SuiteAdminStateContext,
) {
  const bootstrapToken = details.bootstrap_token;
  if (typeof bootstrapToken === "string" && bootstrapToken.trim() !== "") {
    return bootstrapToken;
  }
  throw suiteAdminStateError(
    "suite admin login did not return the required enrollment credential",
    context,
  );
}

function suiteAdminStateError(
  message: string,
  context: SuiteAdminStateContext,
) {
  const ownership = context.externalServerMode
    ? "The reused external-server stack owns this state for its full lifetime."
    : "The harness expected to provision this state during Playwright global setup.";
  return new Error(`${message}. ${ownership}`);
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
