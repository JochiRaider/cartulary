import { createHash } from "node:crypto";
import { readFileSync } from "node:fs";
import path from "node:path";
import { createRequire } from "node:module";
import { ReviewFailure } from "./failure.mjs";

export const repoRoot = path.resolve(import.meta.dirname, "../../../..");
const sha256 = (bytes) => createHash("sha256").update(bytes).digest("hex");
export function toolProfile(root = repoRoot) {
  try {
    const pinBytes = readFileSync(path.join(root, "tools/toolchain_pins.json"));
    const pins = JSON.parse(pinBytes);
    if (process.version !== `v${pins.node_version}`) throw new Error("node pin mismatch");
    const manifest = JSON.parse(readFileSync(path.join(root, "package.json")));
    const versions = {};
    for (const [name, expected] of Object.entries(pins.ui_review)) {
      const installed = JSON.parse(readFileSync(path.join(root, "node_modules", name, "package.json")));
      if (installed.version !== expected || manifest.devDependencies[name] !== expected) throw new Error("package pin mismatch");
      versions[name] = installed.version;
    }
    if (Object.keys(versions).length !== 5 || versions.playwright !== versions["playwright-core"]) throw new Error("incomplete core profile");
    const require = createRequire(path.join(root, "package.json"));
    if (require("sharp").versions.sharp !== versions.sharp) throw new Error("native image runtime unavailable");
    const playwrightRequire = createRequire(require.resolve("playwright"));
    const axeRequire = createRequire(require.resolve("@axe-core/playwright"));
    if (playwrightRequire("playwright-core/package.json").version !== versions.playwright || axeRequire("axe-core/package.json").version !== versions["axe-core"]) throw new Error("nested engine mismatch");
    return { pins_sha256: sha256(pinBytes), lock_sha256: sha256(readFileSync(path.join(root, "pnpm-lock.yaml"))), node_version: pins.node_version, playwright_version: versions.playwright, sharp_version: versions.sharp, axe_version: versions["axe-core"] };
  } catch (cause) { throw new ReviewFailure("tool_configuration", { cause }); }
}
if (process.argv[2] === "--doctor") {
  try { toolProfile(); process.stdout.write("ok UI review core toolchain\n"); }
  catch { process.stderr.write("missing UI review core toolchain: run make bootstrap\n"); process.exitCode = 2; }
}
