import { randomUUID } from "node:crypto";
import { createServer } from "node:net";
import { existsSync, readFileSync } from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { pathToFileURL } from "node:url";

export const visualRendererProfilePath =
  "tools/frontend_visual_renderer_profile.json";
export const visualRendererEnvironmentKeys = Object.freeze([
  "CARTULARY_VISUAL_RENDERER_ATTESTED",
  "CARTULARY_VISUAL_RENDERER_PROFILE_ID",
  "CARTULARY_VISUAL_RENDERER_WS_ENDPOINT",
]);

function docker(args, options = {}) {
  const result = spawnSync("docker", args, {
    encoding: "utf8",
    maxBuffer: 1024 * 1024,
    ...options,
  });
  if (result.status !== 0) {
    throw new Error(`pinned visual renderer lifecycle failed at docker ${args[0]}`);
  }
  return String(result.stdout ?? "").trim();
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

export function loadVisualRendererProfile(root) {
  const profile = readJSON(path.join(root, visualRendererProfilePath));
  const required = {
    schema_id: "cartulary.frontend_visual_renderer_profile.v1",
    container_image:
      "mcr.microsoft.com/playwright@sha256:eac9b0a5312cdab40ee8c2429df5bf19bffdccf8f3bf3c42268e173f97541645",
    platform: "linux/amd64",
    playwright_version: "1.59.1",
    chromium_revision: "1217",
    chromium_version: "147.0.7727.15",
    font_manifest_sha256:
      "c21f8663e6c8fe72681b2be644aa8398538afc59a0f0cda06b94d46d5fbba5fe",
    locale: "en-US",
    device_scale_factor: 1,
    color_scheme: "light",
  };
  for (const [key, value] of Object.entries(required)) {
    if (profile[key] !== value) {
      throw new Error(`visual renderer profile has unexpected ${key}`);
    }
  }
  if (!/^visual\.renderer\.[a-z0-9_]+$/u.test(profile.profile_id ?? "")) {
    throw new Error("visual renderer profile has invalid profile_id");
  }
  return profile;
}

export function assertVisualRendererEnvironmentIsPrivate(environment) {
  for (const key of visualRendererEnvironmentKeys) {
    if ((environment[key] ?? "") !== "") {
      throw new Error(`${key} is harness-private and cannot be inherited`);
    }
  }
}

function verifyLocalPlaywrightPackage(root, profile) {
  const packagePath = path.join(
    root,
    "node_modules/.pnpm/playwright@1.59.1/node_modules/playwright/package.json",
  );
  const corePath = path.join(
    root,
    "node_modules/.pnpm/playwright-core@1.59.1/node_modules/playwright-core",
  );
  const playwrightPath = path.dirname(packagePath);
  if (!existsSync(packagePath) || !existsSync(corePath)) {
    throw new Error("pinned local Playwright server package is unavailable");
  }
  const packageValue = readJSON(packagePath);
  if (packageValue.version !== profile.playwright_version) {
    throw new Error("local Playwright server package does not match renderer profile");
  }
  return { playwrightPath, corePath };
}

function verifyImage(profile) {
  const observed = docker([
    "image",
    "inspect",
    profile.container_image,
    "--format",
    "{{.Os}}/{{.Architecture}}",
  ]);
  if (observed !== profile.platform) {
    throw new Error(
      `visual renderer image platform mismatch: expected ${profile.platform}, got ${observed}`,
    );
  }
}

async function allocateLoopbackPort() {
  return await new Promise((resolve, reject) => {
    const server = createServer();
    server.unref();
    server.once("error", reject);
    server.listen(0, "127.0.0.1", () => {
      const address = server.address();
      const port = typeof address === "object" && address ? address.port : 0;
      server.close((error) => (error ? reject(error) : resolve(port)));
    });
  });
}

async function waitForEndpoint(containerID, port) {
  const deadline = Date.now() + 30_000;
  const endpointPattern = new RegExp(
    `ws://0\\.0\\.0\\.0:${port}/[A-Za-z0-9_-]+`,
    "u",
  );
  while (Date.now() < deadline) {
    const state = docker([
      "inspect",
      containerID,
      "--format",
      "{{.State.Running}} {{.State.ExitCode}}",
    ]);
    if (!state.startsWith("true ")) {
      throw new Error("pinned visual renderer exited before readiness");
    }
    const logs = docker(["logs", containerID]);
    const match = logs.match(endpointPattern);
    if (match) return match[0].replace("ws://0.0.0.0", "ws://127.0.0.1");
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  throw new Error("pinned visual renderer did not become ready");
}

async function verifyRemoteBrowser(root, endpoint, profile) {
  const modulePath = path.join(
    root,
    "apps/web/node_modules/@playwright/test/index.mjs",
  );
  const { chromium } = await import(pathToFileURL(modulePath).href);
  const browser = await chromium.connect(endpoint);
  try {
    if (browser.version() !== profile.chromium_version) {
      throw new Error(
        `visual renderer Chromium mismatch: expected ${profile.chromium_version}, got ${browser.version()}`,
      );
    }
  } finally {
    await browser.close();
  }
}

export async function startVisualRendererLease({ root, environment }) {
  assertVisualRendererEnvironmentIsPrivate(environment);
  const profile = loadVisualRendererProfile(root);
  verifyImage(profile);
  const packages = verifyLocalPlaywrightPackage(root, profile);
  const port = await allocateLoopbackPort();
  const endpointToken = randomUUID().replaceAll("-", "");
  const containerName = `cartulary-visual-${process.pid}-${endpointToken.slice(0, 12)}`;
  let containerID = "";
  let closed = false;
  const cleanup = () => {
    if (closed) return;
    closed = true;
    if (containerID) {
      spawnSync("docker", ["rm", "--force", containerID], {
        encoding: "utf8",
        maxBuffer: 1024 * 1024,
      });
    }
  };
  try {
    containerID = docker([
      "create",
      "--name",
      containerName,
      "--platform",
      profile.platform,
      "--publish",
      `127.0.0.1:${port}:${port}`,
      "--ipc",
      "host",
      "--init",
      "--user",
      "pwuser",
      "--env",
      "LANG=en_US.UTF-8",
      "--env",
      "NODE_PATH=/home/pwuser",
      profile.container_image,
      "node",
      "/home/pwuser/playwright/cli.js",
      "run-server",
      "--host",
      "0.0.0.0",
      "--port",
      String(port),
      "--path",
      `/${endpointToken}`,
    ]);
    docker(["cp", packages.playwrightPath, `${containerID}:/home/pwuser/playwright`]);
    docker(["cp", packages.corePath, `${containerID}:/home/pwuser/playwright-core`]);
    docker(["start", containerID]);
    const endpoint = await waitForEndpoint(containerID, port);
    await verifyRemoteBrowser(root, endpoint, profile);
    return {
      profile,
      environment: {
        CARTULARY_VISUAL_RENDERER_ATTESTED: "1",
        CARTULARY_VISUAL_RENDERER_PROFILE_ID: profile.profile_id,
        CARTULARY_VISUAL_RENDERER_WS_ENDPOINT: endpoint,
      },
      cleanup,
    };
  } catch (error) {
    cleanup();
    throw error;
  }
}
