import { workspaceLayout } from "../../workspace_layout.generated.mjs";
import { spawnSync } from "../workspace/child-process.mjs";
import { createHash } from "node:crypto";
import { accessSync, constants, readFileSync, readdirSync } from "node:fs";
import { createRequire } from "node:module";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { CommandFailure } from "../runtime/command-failure.mjs";
import { atomicLocalFile, privateDirectory, readLocalFile, removePrivateFile } from "../runtime/secure-local-files.mjs";

import { resolvePlaywrightPackages } from "./playwright-packages.mjs";

const repoRoot = path.resolve(import.meta.dirname, "../../..");
const hash = (value) => createHash("sha256").update(value).digest("hex");
const json = (file) => JSON.parse(readFileSync(file, "utf8"));
const receiptFile = (root) => path.join(root, `${workspaceLayout.frontend_install}/proof/installed.v1.json`);
export function readinessFailure(subject_id, condition, recovery_id, cause) {
  return new CommandFailure("installed prerequisite unavailable", { failure_class: "config", failure_reason: "configuration_error", phase: "prerequisites", subject_id, condition, recovery_id }, { cause });
}
function probe(subject, recovery, action) {
  try { return action(); }
  catch (cause) { if (cause instanceof CommandFailure) throw cause; throw readinessFailure(subject, cause.code === "ENOENT" ? "missing" : "incompatible", recovery, cause); }
}
function execute(command, args, environment, runner = spawnSync) {
  const result = runner(command, args, { env: environment, encoding: "utf8", timeout: 30000, maxBuffer: 1024 * 1024, stdio: ["ignore", "pipe", "pipe"] });
  if (result.error) throw result.error;
  if (result.status !== 0) throw new Error("readiness probe failed");
  return result.stdout.trim();
}
export function nodeReadiness(root = repoRoot) {
  return probe("node", "bootstrap_node", () => {
    const pins = json(path.join(root, "tools/toolchain_pins.json"));
    if (process.version !== `v${pins.node_version}`) throw new Error("node pin mismatch");
    return pins;
  });
}
export function frontendInstallationIdentity(root = repoRoot) {
  const pins = json(path.join(root, "tools/toolchain_pins.json"));
  const files = ["package.json", "pnpm-lock.yaml", "pnpm-workspace.yaml", ".npmrc", "apps/web/package.json",
    ...readdirSync(path.join(root, "packages"), { withFileTypes: true }).filter((entry) => entry.isDirectory()).map((entry) => `packages/${entry.name}/package.json`)];
  const inputs = files.sort().map((file) => {
    try { return [file, hash(readFileSync(path.join(root, file)))]; }
    catch (error) { if (file === ".npmrc" && error.code === "ENOENT") return [file, null]; throw error; }
  });
  return { schema_id: "cartulary.frontend_installation.v1", node_version: pins.node_version, pnpm_version: pins.pnpm_version,
    inputs_sha256: hash(JSON.stringify(inputs)), installed_lock_sha256: hash(readFileSync(path.join(root, "node_modules/.pnpm/lock.yaml"))) };
}
export function publishFrontendInstallation(root = repoRoot) {
  const value = frontendInstallationIdentity(root);
  validateSchemaSync("cartulary.frontend_installation.v1", value);
  privateDirectory(path.dirname(receiptFile(root)));
  atomicLocalFile(receiptFile(root), `${JSON.stringify(value)}\n`, { replace: true });
}
export function validateFrontendInstallation(root = repoRoot) {
  return probe("frontend_dependencies", "frontend_install", () => {
    const actual = parseStrictJSON(readLocalFile(receiptFile(root), { maximum: 4096 }).toString());
    validateSchemaSync("cartulary.frontend_installation.v1", actual);
    const expected = frontendInstallationIdentity(root);
    if (Object.keys(actual).sort().join(",") !== Object.keys(expected).sort().join(",") || Object.entries(expected).some(([key, value]) => actual[key] !== value)) {
      throw readinessFailure("frontend_dependencies", "stale_installation", "frontend_install");
    }
    return actual;
  });
}
export function coreReadiness(root = repoRoot, { runner = spawnSync, environment = process.env } = {}) {
  const pins = nodeReadiness(root);
  probe("pnpm", "frontend_toolchain", () => {
    const version = execute(path.join(root, `${workspaceLayout.node_runtime}/bin/pnpm`), ["--version"], { ...environment,
      PATH: `${path.join(root, `${workspaceLayout.node_runtime}/bin`)}:${environment.PATH ?? "/usr/bin:/bin"}`,
      COREPACK_HOME: path.join(root, `${workspaceLayout.node_runtime}/corepack`), COREPACK_ENABLE_NETWORK: "0" }, runner);
    if (version !== pins.pnpm_version) throw new Error("pnpm pin mismatch");
  });
  validateFrontendInstallation(root);
  return probe("frontend_dependencies", "frontend_install", () => {
    resolvePlaywrightPackages(root);
    const manifest = json(path.join(root, "package.json"));
    const versions = {};
    for (const [name, expected] of Object.entries(pins.ui_review)) {
      const installed = json(path.join(root, "node_modules", name, "package.json"));
      if (installed.version !== expected || manifest.devDependencies[name] !== expected) throw new Error("package pin mismatch");
      versions[name] = installed.version;
    }
    const require = createRequire(path.join(root, "package.json"));
    if (require("sharp").versions.sharp !== versions.sharp) throw new Error("native image runtime unavailable");
    const playwrightRequire = createRequire(require.resolve("playwright"));
    const axeRequire = createRequire(require.resolve("@axe-core/playwright"));
    if (Object.keys(versions).length !== 5 || versions.playwright !== versions["playwright-core"] || playwrightRequire("playwright-core/package.json").version !== versions.playwright || axeRequire("axe-core/package.json").version !== versions["axe-core"]) throw new Error("nested engine mismatch");
    return { pins_sha256: hash(readFileSync(path.join(root, "tools/toolchain_pins.json"))), lock_sha256: hash(readFileSync(path.join(root, "pnpm-lock.yaml"))), node_version: pins.node_version,
      playwright_version: versions.playwright, sharp_version: versions.sharp, axe_version: versions["axe-core"] };
  });
}
export function frontendBuildReadiness(root = repoRoot, { runner = spawnSync, environment = process.env } = {}) {
  return probe("frontend_dependencies", "frontend_install", () => {
    const webRequire = createRequire(path.join(root, "apps/web/package.json"));
    // Loading the actual bundler also probes its platform-specific native binding.
    execute(process.execPath, ["--input-type=module", "--eval", `await import(${JSON.stringify(webRequire.resolve("vite"))}); await import(${JSON.stringify(webRequire.resolve("@vitejs/plugin-react"))});`], environment, runner);
  });
}
export function browserReadiness(root = repoRoot) {
  probe("chromium", "playwright_install", () => {
    const require = createRequire(path.join(root, "package.json"));
    const executable = require("playwright").chromium.executablePath();
    accessSync(executable, constants.X_OK);
    const descriptor = json(path.join(path.dirname(require.resolve("playwright-core/package.json")), "browsers.json")).browsers.find((entry) => entry.name === "chromium");
    const version = execute(executable, ["--version"], process.env);
    if (!descriptor || !version.endsWith(` ${descriptor.browserVersion}`)) throw new Error("browser pin mismatch");
  });
  probe("host_lock", "doctor", () => accessSync("/usr/bin/flock", constants.X_OK));
}
export function goReadiness(environment, root = repoRoot, runner = spawnSync) {
  return probe("go", "bootstrap", () => {
    const executable = execute("bash", [path.join(root, "tools/harness/readiness/go-toolchain-readiness.sh"), "resolve"], environment, runner);
    const directories = execute(executable, ["list", "-deps", "-e", "-mod=readonly", "-tags=cartulary_harness", "-f", "{{if not .Dir}}MISSING{{else}}{{if .Module}}{{if .Module.Error}}MISSING{{end}}{{end}}{{end}}", "./cmd/server", "./cmd/migrate", "./tools/testservices", "./tools/embedwebassets"], {
      ...environment, GOTOOLCHAIN: "local", GOPROXY: "off", GONOPROXY: "none", GOSUMDB: "off", GOTELEMETRY: "off",
      GOCACHE: environment.GO_CACHE_DIR, GOMODCACHE: environment.GO_MOD_CACHE_DIR, GOTMPDIR: environment.GO_TMP_DIR,
    }, runner);
    if (directories.split("\n").includes("MISSING")) throw readinessFailure("go", "missing", "bootstrap");
    return executable;
  });
}
export function dockerReadiness(environment, runner = spawnSync) {
  try { execute("docker", ["info", "--format", "{{.ServerVersion}}"], environment, runner); }
  catch (cause) { throw new CommandFailure("Docker preflight failed", { failure_class: "infra", failure_reason: "preflight_error", phase: "prerequisites", subject_id: "docker", condition: "preflight_failed", recovery_id: "doctor" }, { cause }); }
}
export function serviceImageReadiness(environment, root = repoRoot, runner = spawnSync) {
  dockerReadiness(environment, runner);
  const images = probe("test_service_images", "test_service_images", () => execute(path.join(root, `${workspaceLayout.toolbin}/cartulary-test-services`), ["images"], environment, runner).split("\n").filter(Boolean));
  if (!images.length) throw readinessFailure("test_service_images", "incompatible", "test_service_images");
  for (const image of images) {
    try { execute("docker", ["image", "inspect", image], environment, runner); }
    catch (cause) { throw readinessFailure("test_service_images", "missing", "test_service_images", cause); }
  }
}
if (process.argv[1] && fileURLToPath(import.meta.url) === path.resolve(process.argv[1])) {
  try {
    if (process.argv[2] === "invalidate") removePrivateFile(receiptFile(repoRoot));
    else if (process.argv[2] === "publish") publishFrontendInstallation();
    else if (process.argv[2] === "frontend") { coreReadiness(); process.stdout.write("ok installed frontend prerequisites\n"); }
    else if (process.argv[2] === "doctor") { coreReadiness(); frontendBuildReadiness(); browserReadiness(); goReadiness(process.env); dockerReadiness(process.env); process.stdout.write("ok installed UI review prerequisites\n"); }
    else throw new Error("unsupported installed readiness operation");
  } catch (error) {
    process.stderr.write(`installed readiness: subject=${error.subject_id ?? "frontend_dependencies"} condition=${error.condition ?? "incompatible"} recovery=${error.recovery_id ?? "frontend_install"}\n`);
    process.exitCode = 2;
  }
}
