import path from "node:path";
import { fileURLToPath } from "node:url";

import type { PlaywrightTestConfig } from "@playwright/test";

import { attestedVisualRendererProfile } from "./e2e/support/runtime/visualRenderer";

const currentDirectory = path.dirname(fileURLToPath(import.meta.url));

export const publicOrigin = process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN ?? "";

export const apiOrigin = process.env.CARTULARY_WEB_E2E_API_ORIGIN ?? "";

export const configuredWorkers = 1;
if (
  process.env.CARTULARY_PLAYWRIGHT_WORKER_COUNT !== "1" ||
  process.env.CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET !== "0"
) {
  throw new Error("Playwright requires one owned worker at offset zero");
}

function requireValidatedAttachment(): void {
  if (
    process.env.CARTULARY_WEB_E2E_ATTACHMENT_VALIDATED !== "1" ||
    publicOrigin === "" ||
    apiOrigin === ""
  ) {
    throw new Error(
      "Playwright requires a Make-owned, validated cartulary.web_e2e_stack.v8 attachment",
    );
  }
}

function visualRendererUse(): NonNullable<PlaywrightTestConfig["use"]> {
  const endpoint = process.env.CARTULARY_VISUAL_RENDERER_WS_ENDPOINT ?? "";
  const profileId = process.env.CARTULARY_VISUAL_RENDERER_PROFILE_ID ?? "";
  const attested = process.env.CARTULARY_VISUAL_RENDERER_ATTESTED ?? "";
  const supplied = endpoint !== "" || profileId !== "" || attested !== "";
  if (!supplied || !process.env.CARTULARY_RENDERER_ATTACHMENT) {
    throw new Error("Playwright requires a native renderer borrow attachment");
  }
  const profile = attestedVisualRendererProfile();
  if (
    attested !== "1" ||
    !/^ws:\/\/127\.0\.0\.1:[0-9]+\/[A-Za-z0-9_-]+$/u.test(endpoint) ||
    profileId !== profile.profile_id
  ) {
    throw new Error("Playwright visual renderer attestation is invalid");
  }
  return {
    connectOptions: { wsEndpoint: endpoint, exposeNetwork: "<loopback>" },
    colorScheme: profile.color_scheme,
    deviceScaleFactor: profile.device_scale_factor,
    locale: profile.locale,
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
  if (overrides.workers !== undefined && overrides.workers !== 1)
    throw new Error("browser groups require exactly one worker");
  const { retries: _retries, use, ...rest } = overrides;

  return {
    testDir: "./e2e",
    fullyParallel: false,
    globalSetup: path.resolve(currentDirectory, "e2e", "global-setup.ts"),
    globalTeardown: path.resolve(currentDirectory, "e2e", "global-teardown.ts"),
    retries: 0,
    workers: configuredWorkers,
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
