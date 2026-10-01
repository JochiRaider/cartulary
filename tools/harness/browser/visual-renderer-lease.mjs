import { randomUUID } from "node:crypto";
import { createServer } from "node:net";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { pathToFileURL } from "node:url";

import { resolvePlaywrightPackages } from "../readiness/playwright-packages.mjs";
import { loadVisualRendererProfile, verifyRendererFonts, validateRendererAttestation, sha256 } from "./visual-renderer-profile.mjs";
export { loadVisualRendererProfile, visualRendererProfilePath } from "./visual-renderer-profile.mjs";
export const visualRendererEnvironmentKeys = Object.freeze([
  "CARTULARY_VISUAL_RENDERER_ATTESTED",
  "CARTULARY_VISUAL_RENDERER_PROFILE_ID",
  "CARTULARY_VISUAL_RENDERER_PROFILE_JSON",
  "CARTULARY_VISUAL_RENDERER_WS_ENDPOINT",
]);

function docker(args, options = {}) {
  const result = spawnSync("docker", args, {
    encoding: "utf8",
    maxBuffer: 1024 * 1024,
    timeout: 30000,
    ...options,
  });
  if (result.status !== 0) {
    throw new Error(`pinned visual renderer lifecycle failed at docker ${args[0]}`);
  }
  return String(result.stdout ?? "").trim();
}

export function assertVisualRendererEnvironmentIsPrivate(environment) {
  for (const key of visualRendererEnvironmentKeys) {
    if ((environment[key] ?? "") !== "") {
      throw new Error(`${key} is harness-private and cannot be inherited`);
    }
  }
}

export function verifyRendererImage(profile, run = docker) {
  const [observed] = JSON.parse(run(["image", "inspect", profile.container_image]));
  if (`${observed.Os}/${observed.Architecture}` !== profile.platform || !observed.RepoDigests?.includes(profile.container_image)) {
    throw new Error("visual renderer image identity mismatch");
  }
  return { image_id: observed.Id, container_image: observed.RepoDigests.find((entry) => entry === profile.container_image), platform: `${observed.Os}/${observed.Architecture}` };
}

export function rendererRelease(containerID, run = docker) {
  let closed = false;
  return () => {
    if (closed) return;
    let removalError;
    try { run(["rm", "--force", containerID]); } catch (error) { removalError = error; }
    const remaining = run(["ps", "--all", "--no-trunc", "--filter", `id=${containerID}`, "--format", "{{.ID}}"]);
    if (remaining.trim()) throw removalError ?? new Error("visual renderer removal not confirmed");
    closed = true;
  };
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

async function waitForEndpoint(containerID, port, run = docker) {
  const deadline = Date.now() + 30_000;
  const endpointPattern = new RegExp(
    `ws://0\\.0\\.0\\.0:${port}/[A-Za-z0-9_-]+`,
    "u",
  );
  while (Date.now() < deadline) {
    const state = run([
      "inspect",
      containerID,
      "--format",
      "{{.State.Running}} {{.State.ExitCode}}",
    ]);
    if (!state.startsWith("true ")) {
      throw new Error("pinned visual renderer exited before readiness");
    }
    const logs = run(["logs", containerID]);
    const match = logs.match(endpointPattern);
    if (match) return match[0].replace("ws://0.0.0.0", "ws://127.0.0.1");
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
  throw new Error("pinned visual renderer did not become ready");
}

async function verifyRemoteBrowser(packages, endpoint, profile) {
  const { chromium } = await import(pathToFileURL(path.join(packages.playwrightPath, "index.mjs")).href);
  const browser = await chromium.connect(endpoint, { timeout: 10000 });
  try {
    if (browser.version() !== profile.chromium_version) {
      throw new Error(
        `visual renderer Chromium mismatch: expected ${profile.chromium_version}, got ${browser.version()}`,
      );
    }
    return browser.version();
  } finally {
    await browser.close();
  }
}

export async function startVisualRendererLease({ root, environment, run = docker, connect = verifyRemoteBrowser, endpointReady = waitForEndpoint }) {
  assertVisualRendererEnvironmentIsPrivate(environment);
  const profile = loadVisualRendererProfile(root);
  const image = verifyRendererImage(profile, run);
  const packages = resolvePlaywrightPackages(root);
  if (packages.version !== profile.playwright_version || packages.chromium.revision !== profile.chromium_revision || packages.chromium.browserVersion !== profile.chromium_version) throw new Error("renderer package descriptor mismatch");
  const fonts = verifyRendererFonts(root, profile);
  const port = await allocateLoopbackPort();
  const endpointToken = randomUUID().replaceAll("-", "");
  const containerName = `cartulary-visual-${process.pid}-${endpointToken.slice(0, 12)}`;
  let containerID = "";
  let release, creationAttempted = false;
  const cleanup = () => {
    if (!release && creationAttempted) {
      const acquired = run(["ps", "--all", "--no-trunc", "--filter", `name=^/${containerName}$`, "--filter", `label=cartulary.visual-owner=${endpointToken}`, "--format", "{{.ID}}"]);
      if (acquired) {
        if (!/^[a-f0-9]{64}$/u.test(acquired)) throw new Error("ambiguous renderer acquisition");
        release = rendererRelease(acquired, run);
      }
    }
    release?.();
  };
  const interrupt = () => { try { cleanup(); } catch { process.stderr.write("visual renderer cleanup failed (cleanup_error)\n"); } process.exit(130); };
  const terminate = () => { try { cleanup(); } catch { process.stderr.write("visual renderer cleanup failed (cleanup_error)\n"); } process.exit(143); };
  process.once("SIGINT", interrupt); process.once("SIGTERM", terminate);
  try {
    creationAttempted = true;
    containerID = run([
      "create",
      "--name",
      containerName,
      "--label",
      `cartulary.visual-owner=${endpointToken}`,
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
    release = rendererRelease(containerID, run);
    run(["cp", packages.playwrightPath, `${containerID}:/home/pwuser/playwright`]);
    run(["cp", packages.corePath, `${containerID}:/home/pwuser/playwright-core`]);
    run(["start", containerID]);
    const observed = JSON.parse(run(["exec", containerID, "node", "-e", `const p=require('/home/pwuser/playwright/package.json'); const c=require('/home/pwuser/playwright-core/package.json'); const b=require('/home/pwuser/playwright-core/browsers.json').browsers.find(x=>x.name==='chromium'); process.stdout.write(JSON.stringify({playwright:p.version,core:c.version,revision:b.revision,version:b.browserVersion}));`]));
    if (observed.playwright !== profile.playwright_version || observed.core !== profile.playwright_version || observed.revision !== profile.chromium_revision || observed.version !== profile.chromium_version) throw new Error("container package identity mismatch");
    if (run(["inspect", containerID, "--format", "{{.Image}}"]) !== image.image_id) throw new Error("container image mismatch");
    const endpoint = await endpointReady(containerID, port, run);
    const browserVersion = await connect(packages, endpoint, profile);
    const attestation = validateRendererAttestation({ schema_id: "cartulary.frontend_visual_renderer_attestation.v1", profile, observed: {
      ...image, playwright_version: observed.playwright, core_version: observed.core, chromium_revision: observed.revision, chromium_version: browserVersion,
      font_manifest_sha256: sha256(fonts.manifestBytes), font_files: fonts.files,
    } }, profile, fonts.files);
    return {
      profile,
      attestation,
      environment: {
        CARTULARY_VISUAL_RENDERER_ATTESTED: "1",
        CARTULARY_VISUAL_RENDERER_PROFILE_ID: profile.profile_id,
        CARTULARY_VISUAL_RENDERER_PROFILE_JSON: JSON.stringify(profile),
        CARTULARY_VISUAL_RENDERER_WS_ENDPOINT: endpoint,
      },
      cleanup,
    };
  } catch (error) {
    try { cleanup(); } catch (cleanupError) { error.cleanupFailures = [cleanupError]; }
    throw error;
  } finally { process.removeListener("SIGINT", interrupt); process.removeListener("SIGTERM", terminate); }
}
