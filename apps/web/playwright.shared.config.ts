import path from "node:path";
import { fileURLToPath } from "node:url";

import type { PlaywrightTestConfig } from "@playwright/test";

const currentDirectory = path.dirname(fileURLToPath(import.meta.url));

export const publicOrigin =
  process.env.CARTULARY_WEB_E2E_PUBLIC_ORIGIN ?? "http://127.0.0.1:4173";

export const configuredWorkers = Number.parseInt(
  process.env.PLAYWRIGHT_WORKERS ?? "2",
  10,
);

export function webE2EBaseConfig(
  overrides: Partial<PlaywrightTestConfig> = {},
): PlaywrightTestConfig {
  const useExternalServer =
    process.env.CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER === "1";
  const { use, webServer, ...rest } = overrides;
  const defaultWebServer = {
    command: "bash ./scripts/start-web-e2e.sh",
    cwd: path.resolve(currentDirectory, "..", ".."),
    url: publicOrigin,
    reuseExistingServer: false,
    timeout: 180_000,
  };
  const resolvedWebServer = useExternalServer
    ? undefined
    : (webServer ?? defaultWebServer);

  return {
    testDir: "./e2e",
    fullyParallel: false,
    globalSetup: path.resolve(currentDirectory, "e2e", "global-setup.ts"),
    globalTeardown: path.resolve(currentDirectory, "e2e", "global-teardown.ts"),
    workers:
      Number.isNaN(configuredWorkers) || configuredWorkers < 1
        ? 1
        : configuredWorkers,
    timeout: 60_000,
    use: {
      baseURL: publicOrigin,
      trace: "retain-on-failure",
      ...use,
    },
    ...(resolvedWebServer === undefined
      ? {}
      : { webServer: resolvedWebServer }),
    ...rest,
  };
}
