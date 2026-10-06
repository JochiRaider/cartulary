import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { buildReceipt, goOutput, pins, verifyArchive } from "./cryptographic-module.mjs";

const policy = pins();
const go = process.env.GO || "go";
const env = { ...process.env, GOTOOLCHAIN: policy.go_toolchain, GOFIPS140: policy.cryptographic_module.selector, GODEBUG: "fips140=on", GOCACHE: process.env.GO_CACHE_DIR, GOMODCACHE: process.env.GO_MOD_CACHE_DIR, GOTMPDIR: process.env.GO_TMP_DIR };
let scratch;
try {
  for (const dir of [env.GOCACHE, env.GOMODCACHE, env.GOTMPDIR]) mkdirSync(dir, { recursive: true });
  verifyArchive(go, env);
  const buildHelper = spawnSync("bash", ["tools/harness/backend/tests/test-build-go-artifact.sh"], { env, encoding: "utf8", timeout: 30000 });
  if (buildHelper.error || buildHelper.status !== 0) throw new Error("build helper identity/receipt regression failed");
  scratch = mkdtempSync(path.join(env.GOTMPDIR, "crypto-assessment-"));
  const pinned = path.join(scratch, "pinned");
  const ordinary = path.join(scratch, "ordinary");
  goOutput(go, ["build", "-buildvcs=false", "-o", pinned, "./tools/cryptographicpolicy"], env);
  const receipt = buildReceipt(go, pinned, env);
  const applicationReceipts = [];
  const applicationRejections = [];
  const configFIFO = path.join(scratch, "configuration-must-not-be-read");
  const fifo = spawnSync("mkfifo", ["--", configFIFO], { encoding: "utf8" });
  if (fifo.error || fifo.status !== 0) throw new Error("admission probe setup failed");
  function rejectApplication(name, binary, mode, scenario) {
    const args = name === "migrate" ? ["up"] : name === "operator" ? ["object-store", "init"] : [];
    const child = spawnSync(binary, args, { env: { ...env, CARTULARY_CONFIG_FILE: configFIFO, GODEBUG: `fips140=${mode}` }, encoding: "utf8", timeout: 10000 });
    const expected = name === "server" ? 2 : 1;
    if (child.error || child.status !== expected || !`${child.stdout}${child.stderr}`.includes("cryptographic execution policy rejected")) {
      throw new Error(`application rejected at wrong boundary: ${name}/${scenario}`);
    }
    applicationRejections.push({ application: name, scenario, exit_code: child.status, configuration_unread: true });
  }
  function metadataFixture(binary, name, from, to) {
    const bytes = readFileSync(binary);
    const needle = Buffer.from(from);
    const replacement = Buffer.from(to);
    if (needle.length !== replacement.length) throw new Error("metadata fixture length mismatch");
    let found = 0;
    for (let at = bytes.indexOf(needle); at !== -1; at = bytes.indexOf(needle, at + needle.length)) {
      replacement.copy(bytes, at); found++;
    }
    if (!found) throw new Error("metadata fixture setting absent");
    const output = `${binary}.${name}`;
    writeFileSync(output, bytes, { mode: 0o700 });
    return output;
  }
  for (const name of ["server", "migrate", "operator"]) {
    const binary = path.join(scratch, name);
    goOutput(go, ["build", "-buildvcs=false", "-o", binary, `./cmd/${name}`], env);
    applicationReceipts.push({ name, receipt: buildReceipt(go, binary, env) });
    rejectApplication(name, binary, "off", "disabled_mode");
    const missing = metadataFixture(binary, "missing", "build\tGOFIPS140=", "build\tNOFIPS140=");
    rejectApplication(name, missing, "on", "missing_metadata");
    const contradictory = metadataFixture(binary, "contradictory", `build\tGOFIPS140=${policy.cryptographic_module.selector}\n`, "build\tGOFIPS140=v1.0.0-c2097c7d\n");
    rejectApplication(name, contradictory, "on", "contradictory_metadata");
    for (const [scenario, selector] of [["ordinary_build", "off"], ["wrong_module", "v1.26.0"]]) {
      const alternate = `${binary}.${scenario}`;
      goOutput(go, ["build", "-buildvcs=false", "-o", alternate, `./cmd/${name}`], { ...env, GOFIPS140: selector });
      rejectApplication(name, alternate, "on", scenario);
      let rejected = false;
      try { buildReceipt(go, alternate, env); } catch { rejected = true; }
      if (!rejected) throw new Error(`unqualified application cache artifact accepted: ${name}/${scenario}`);
    }
  }
  goOutput(go, ["build", "-buildvcs=false", "-o", ordinary, "./tools/cryptographicpolicy"], { ...env, GOFIPS140: "off" });
  let rejectedOrdinaryReceipt = false;
  try { buildReceipt(go, ordinary, env); } catch { rejectedOrdinaryReceipt = true; }
  if (!rejectedOrdinaryReceipt) throw new Error("ordinary cache artifact accepted as pinned build");
  const results = [];
  for (const [name, binary, mode, expected] of [["pinned_enabled", pinned, "on", 0], ["pinned_disabled", pinned, "off", 2], ["ordinary_enabled", ordinary, "on", 2], ["pinned_strict_diagnostic", pinned, "only", 0]]) {
    const child = spawnSync(binary, [], { env: { ...env, GODEBUG: `fips140=${mode}` }, encoding: "utf8", timeout: 30000 });
    if (child.error || child.status !== expected) throw new Error(`cryptographic assessment failed: ${name}`);
    results.push({ name, exit_code: child.status, identity: expected === 0 ? JSON.parse(child.stdout) : null });
  }
  const root = path.join(process.env.CARTULARY_TEST_RESULTS_DIR, process.env.CARTULARY_TEST_RUN_ID, "cryptographic-policy-assessment");
  mkdirSync(root, { recursive: true, mode: 0o700 });
  writeFileSync(path.join(root, "assessment.json"), `${JSON.stringify({ schema_id: "cartulary.cryptographic_policy_assessment.v1", disposition: "binary_admission_verified_state_and_package_evidence_separate", receipt, applicationReceipts, applicationRejections, rejectedOrdinaryReceipt, results }, null, 2)}\n`, { mode: 0o600 });
  process.stdout.write("cryptographic build and admission diagnostics passed\n");
} catch (error) {
  process.stderr.write(`${error.message}\n`);
  process.exitCode = 1;
} finally {
  if (scratch) rmSync(scratch, { recursive: true, force: true });
}
