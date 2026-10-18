import { randomUUID } from "node:crypto";
import path from "node:path";
import { spawnSync } from "../workspace/child-process.mjs";
import { pathToFileURL } from "node:url";

import { redactString } from "../contract/index.mjs";
import { stageBrowserTrust } from "./browser-trust.mjs";
import { validateRendererProof, removeOwnedRenderer } from "../runtime/owned-renderer.mjs";
import { CommandFailure } from "../runtime/command-failure.mjs";

import { resolvePlaywrightPackages } from "../readiness/playwright-packages.mjs";
import { loadVisualRendererProfile, verifyRendererFonts, validateRendererAttestation, sha256 } from "./visual-renderer-profile.mjs";
export { loadVisualRendererProfile, visualRendererProfilePath } from "./visual-renderer-profile.mjs";
export const visualRendererEnvironmentKeys = Object.freeze([
  "CARTULARY_VISUAL_RENDERER_ATTESTED",
  "CARTULARY_VISUAL_RENDERER_PROFILE_ID",
  "CARTULARY_VISUAL_RENDERER_PROFILE_JSON",
  "CARTULARY_VISUAL_RENDERER_WS_ENDPOINT",
]);

const rendererContainerPort = 3000;
const diagnosticContainerPort = 3001;

export function rendererLifecycleFailureReason(result) {
  if (result.error?.code === "ETIMEDOUT") return "command_timeout";
  if (result.error?.code === "ENOENT") return "command_unavailable";
  if (result.signal) return "command_signalled";
  const diagnostic = String(result.stderr ?? "").slice(-8192).toLowerCase();
  if (/port is already allocated|address already in use|failed to bind host port/u.test(diagnostic)) return "port_unavailable";
  if (/cannot connect to the docker daemon|daemon is not running/u.test(diagnostic)) return "daemon_unavailable";
  if (/permission denied|access is denied/u.test(diagnostic)) return "permission_denied";
  if (diagnostic.includes("no space left on device")) return "storage_exhausted";
  return "command_failed";
}

function docker(args, options = {}) {
  const result = spawnSync("docker", args, {
    encoding: "utf8",
    maxBuffer: 1024 * 1024,
    timeout: 30000,
    ...options,
  });
  if (result.status !== 0) {
    const operation = args[0] === "exec" ? `exec ${path.basename(args[2] ?? "")}` : args[0];
    const diagnostic = args[0] === "exec" && ["dpkg-deb", "certutil", "mkdir"].includes(path.basename(args[2] ?? "")) ? `: ${redactString(String(result.stderr ?? "").slice(-2048))}` : "";
    throw new Error(`pinned browser renderer lifecycle failed at docker ${operation} (exit ${result.status}, reason ${rendererLifecycleFailureReason(result)})${diagnostic}`);
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

export function rendererPublishedPorts(value, diagnostics) {
  const expected = [rendererContainerPort, ...(diagnostics ? [diagnosticContainerPort] : [])];
  if (!value || typeof value !== "object" || Array.isArray(value)) throw new Error("invalid renderer published ports");
  const published = Object.entries(value).filter(([, bindings]) => bindings !== null);
  if (published.length !== expected.length) throw new Error("unexpected renderer published ports");
  const ports = new Map();
  for (const [key, bindings] of published) {
    const containerPort = expected.find((port) => key === `${port}/tcp`);
    if (!containerPort || !Array.isArray(bindings) || bindings.length !== 1) throw new Error("invalid renderer port binding");
    const binding = bindings[0];
    if (!binding || Object.keys(binding).sort().join(",") !== "HostIp,HostPort" || binding.HostIp !== "127.0.0.1" || typeof binding.HostPort !== "string" || !/^[1-9][0-9]{0,4}$/u.test(binding.HostPort) || Number(binding.HostPort) > 65535) throw new Error("renderer port binding is not exact loopback");
    const hostPort = Number(binding.HostPort);
    if ([...ports.values()].includes(hostPort)) throw new Error("duplicate renderer host port");
    ports.set(containerPort, hostPort);
  }
  return { renderer: { containerPort: rendererContainerPort, hostPort: ports.get(rendererContainerPort) }, diagnostic: diagnostics ? { containerPort: diagnosticContainerPort, hostPort: ports.get(diagnosticContainerPort) } : null };
}

async function waitForEndpoint(containerID, { containerPort, hostPort }, run = docker) {
  const deadline = Date.now() + 30_000;
  const endpointPattern = new RegExp(
    `ws://0\\.0\\.0\\.0:${containerPort}/[A-Za-z0-9_-]+`,
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
    if (match) return match[0].replace(`ws://0.0.0.0:${containerPort}`, `ws://127.0.0.1:${hostPort}`);
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

// Read only the container owned by this acquisition, before cleanup removes it.
// The daemon's error text, endpoint, process identity and logs stay private.
function rendererStateBeforeCleanup(containerID, run) {
  try {
    const state = JSON.parse(run(["inspect", containerID, "--format", "{{json .State}}"], { timeout: 3000 }));
    if (!["created", "running", "paused", "restarting", "removing", "exited", "dead"].includes(state?.Status) ||
        typeof state.Running !== "boolean" || typeof state.OOMKilled !== "boolean" ||
        !Number.isInteger(state.ExitCode) || state.ExitCode < 0 || state.ExitCode > 255) throw new Error("invalid renderer state");
    return Object.freeze({ status: state.Status, running: state.Running, oom_killed: state.OOMKilled, exit_code: state.ExitCode });
  } catch { return Object.freeze({ status: "unavailable" }); }
}

export function recoverVisualRenderer(proof, run = docker) { removeOwnedRenderer(proof, run); }

export async function startVisualRendererLease({ root, environment, diagnostics = false, onOwnedResource = () => {}, run = docker, connect = verifyRemoteBrowser, endpointReady = waitForEndpoint, trust = stageBrowserTrust }) {
  assertVisualRendererEnvironmentIsPrivate(environment);
  const profile = loadVisualRendererProfile(root);
  const image = verifyRendererImage(profile, run);
  const packages = resolvePlaywrightPackages(root);
  if (packages.version !== profile.playwright_version || packages.chromium.revision !== profile.chromium_revision || packages.chromium.browserVersion !== profile.chromium_version) throw new Error("renderer package descriptor mismatch");
  const fonts = verifyRendererFonts(root, profile);
  const endpointToken = randomUUID().replaceAll("-", "");
  const containerName = `cartulary-visual-${process.pid}-${endpointToken.slice(0, 12)}`;
  const proof = validateRendererProof({ name: containerName, token: endpointToken, image_id: image.image_id, daemon_id: run(["info", "--format", "{{.ID}}"]) });
  let containerID = "";
  let release, creationAttempted = false, published = false;
  const cleanup = () => {
    if (!release && creationAttempted) {
      const acquired = run(["ps", "--all", "--no-trunc", "--filter", `name=^/${containerName}$`, "--filter", `label=cartulary.visual-owner=${endpointToken}`, "--format", "{{.ID}}"]);
      if (acquired) {
        if (!/^[a-f0-9]{64}$/u.test(acquired)) throw new Error("ambiguous renderer acquisition");
        release = rendererRelease(acquired, run);
      }
    }
    release?.();
    if (published) { onOwnedResource({ kind: "browser_renderer", target: proof, state: "released" }); published = false; }
  };
  const interrupt = () => { try { cleanup(); } catch { process.stderr.write("visual renderer cleanup failed (cleanup_error)\n"); } process.exit(130); };
  const terminate = () => { try { cleanup(); } catch { process.stderr.write("visual renderer cleanup failed (cleanup_error)\n"); } process.exit(143); };
  process.once("SIGINT", interrupt); process.once("SIGTERM", terminate);
  try {
    published = true;
    onOwnedResource({ kind: "browser_renderer", target: proof, state: "pending" });
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
      `127.0.0.1::${rendererContainerPort}`,
      ...(diagnostics ? ["--publish", `127.0.0.1::${diagnosticContainerPort}`] : []),
      "--ipc",
      "host",
      "--init",
      "--user",
      "pwuser",
      "--env",
      "LANG=en_US.UTF-8",
      "--env",
      "NODE_PATH=/home/pwuser",
      "--env",
      "NODE_EXTRA_CA_CERTS=/home/pwuser/trust.pem",
      profile.container_image,
      "node",
      "/home/pwuser/playwright/cli.js",
      "run-server",
      "--host",
      "0.0.0.0",
      "--port",
      String(rendererContainerPort),
      "--path",
      `/${endpointToken}`,
    ]);
    release = rendererRelease(containerID, run);
    onOwnedResource({ kind: "browser_renderer", target: proof, state: "acquired" });
    run(["cp", packages.playwrightPath, `${containerID}:/home/pwuser/playwright`]);
    run(["cp", packages.corePath, `${containerID}:/home/pwuser/playwright-core`]);
    const stagedTrust = trust({ root, environment, containerID, run });
    run(["start", containerID]);
    const ports = rendererPublishedPorts(JSON.parse(run(["inspect", containerID, "--format", "{{json .NetworkSettings.Ports}}"])), diagnostics);
    stagedTrust.initialize();
    const observed = JSON.parse(run(["exec", containerID, "node", "-e", `const p=require('/home/pwuser/playwright/package.json'); const c=require('/home/pwuser/playwright-core/package.json'); const b=require('/home/pwuser/playwright-core/browsers.json').browsers.find(x=>x.name==='chromium'); process.stdout.write(JSON.stringify({playwright:p.version,core:c.version,revision:b.revision,version:b.browserVersion}));`]));
    if (observed.playwright !== profile.playwright_version || observed.core !== profile.playwright_version || observed.revision !== profile.chromium_revision || observed.version !== profile.chromium_version) throw new Error("container package identity mismatch");
    if (run(["inspect", containerID, "--format", "{{.Image}}"]) !== image.image_id) throw new Error("container image mismatch");
    const trustAttestation = stagedTrust.attestation;
    const endpoint = await endpointReady(containerID, ports.renderer, run);
    let browserVersion;
    try { browserVersion = await connect(packages, endpoint, profile); }
    catch (cause) {
      const state = rendererStateBeforeCleanup(containerID, run);
      const failure = new CommandFailure(`pinned browser renderer connection failed; container_state=${JSON.stringify(state)}`,
        { failure_class: "infra", failure_reason: "service_start_error" }, { cause });
      failure.rendererState = state;
      throw failure;
    }
    const attestation = validateRendererAttestation({ schema_id: "cartulary.frontend_visual_renderer_attestation.v1", profile, observed: {
      ...image, playwright_version: observed.playwright, core_version: observed.core, chromium_revision: observed.revision, chromium_version: browserVersion,
      font_manifest_sha256: sha256(fonts.manifestBytes), font_files: fonts.files,
    } }, profile, fonts.files);
    return {
      profile,
      attestation,
      trustAttestation,
      diagnosticBinding: ports.diagnostic,
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
