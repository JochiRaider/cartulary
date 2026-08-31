import path from "node:path";
import { fileURLToPath } from "node:url";

import type { PlaywrightTestConfig } from "@playwright/test";

const currentDirectory = path.dirname(fileURLToPath(import.meta.url));

export const publicOrigin = process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN ?? "";

export const apiOrigin = process.env.CARTULARY_WEB_E2E_API_ORIGIN ?? "";

export const configuredWorkers = Number.parseInt(
  process.env.PLAYWRIGHT_WORKERS ?? "2",
  10,
);

function requireValidatedAttachment(): void {
  if (
    process.env.CARTULARY_WEB_E2E_ATTACHMENT_VALIDATED !== "1" ||
    publicOrigin === "" ||
    apiOrigin === ""
  ) {
    throw new Error(
      "Playwright requires a Make-owned, validated cartulary.web_e2e_stack.v6 attachment",
    );
  }
}

function visualRendererUse(): NonNullable<PlaywrightTestConfig["use"]> {
  const endpoint = process.env.CARTULARY_VISUAL_RENDERER_WS_ENDPOINT ?? "";
  const profileId = process.env.CARTULARY_VISUAL_RENDERER_PROFILE_ID ?? "";
  const attested = process.env.CARTULARY_VISUAL_RENDERER_ATTESTED ?? "";
  const supplied = endpoint !== "" || profileId !== "" || attested !== "";
  if (!supplied) return {};
  if (
    attested !== "1" ||
    !/^ws:\/\/127\.0\.0\.1:[0-9]+\/[A-Za-z0-9_-]+$/u.test(endpoint) ||
    profileId !== "visual.renderer.playwright_1_59_1_chromium_1217_linux_amd64"
  ) {
    throw new Error("Playwright visual renderer attestation is invalid");
  }
  return {
    connectOptions: { wsEndpoint: endpoint, exposeNetwork: "<loopback>" },
    colorScheme: "light",
    deviceScaleFactor: 1,
    locale: "en-US",
  };
}

export function webE2EBaseConfig(
  overrides: Partial<PlaywrightTestConfig> = {},
): PlaywrightTestConfig {
  requireValidatedAttachment();
  if ("webServer" in overrides) {
    throw new Error(
      "Cartulary browser configuration does not permit Playwright webServer",
    );
  }
  if (overrides.retries !== undefined && overrides.retries !== 0) {
    throw new Error(
      "Cartulary browser evidence does not permit Playwright retries",
    );
  }
  const { retries: _retries, use, ...rest } = overrides;

  return {
    testDir: "./e2e",
    fullyParallel: false,
    globalSetup: path.resolve(currentDirectory, "e2e", "global-setup.ts"),
    globalTeardown: path.resolve(currentDirectory, "e2e", "global-teardown.ts"),
    retries: 0,
    workers:
      Number.isNaN(configuredWorkers) || configuredWorkers < 1
        ? 1
        : configuredWorkers,
    timeout: 60_000,
    updateSnapshots: "none",
    use: {
      baseURL: publicOrigin,
      trace: "retain-on-failure",
      ...visualRendererUse(),
      ...use,
    },
    ...rest,
  };
}
