import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";

const scratch = mkdtempSync(path.join(tmpdir(), "cartulary-cleanup-test-"));
try {
  writeFileSync(path.join(scratch, "docker"), `#!/usr/bin/env bash
set -eu
printf '%s %s\\n' "$1" "$2" >>"$CLEANUP_TEST_LOG"
if [[ "$CLEANUP_TEST_MODE" == unverified && "$1" == network && "$2" == ls ]]; then exit 1; fi
if [[ "$2" == ls && ! -f "$CLEANUP_TEST_ROOT/$1.removed" ]]; then printf 'owned-resource\\n'; fi
if [[ "$2" == rm ]]; then
  if [[ "$CLEANUP_TEST_MODE" == leftover && "$1" == volume ]]; then exit 1; fi
  touch "$CLEANUP_TEST_ROOT/$1.removed"
fi
`, { mode: 0o700 });
  const helper = path.resolve(import.meta.dirname, "../package-smoke-cleanup.sh");
  for (const mode of ["success", "leftover", "unverified", "interrupt", "unsafe"]) {
    for (const kind of ["container", "volume", "network", "image"]) rmSync(path.join(scratch, `${kind}.removed`), { force: true });
    const log = path.join(scratch, "calls"); writeFileSync(log, "");
    const evidence = path.join(scratch, "cleanup.json"); rmSync(evidence, { force: true });
    const script = mode === "interrupt"
      ? 'source "$1"; trap \'status=$?; cleanup_package_resources cartularymvpsmoke123 cartulary/test:cartularymvpsmoke123 "$2" || status=1; exit "$status"\' EXIT; trap \'exit 143\' TERM; kill -TERM $$'
      : 'source "$1"; cleanup_package_resources "$3" cartulary/test:cartularymvpsmoke123 "$2"';
    const result = spawnSync("bash", ["-c", script, "cleanup-test", helper, evidence, mode === "unsafe" ? "retained-production" : "cartularymvpsmoke123"], {
      env: { ...process.env, PATH: `${scratch}:${process.env.PATH}`, CLEANUP_TEST_ROOT: scratch, CLEANUP_TEST_MODE: mode, CLEANUP_TEST_LOG: log }, encoding: "utf8", timeout: 10000,
    });
    assert.equal(result.status, mode === "success" ? 0 : mode === "interrupt" ? 143 : 1, mode);
    const calls = readFileSync(log, "utf8");
    if (mode === "unsafe") { assert.equal(calls, ""); continue; }
    const record = JSON.parse(readFileSync(evidence));
    assert.equal(record.cleanup, ["success", "interrupt"].includes(mode) ? "passed" : "failed");
    assert.ok(calls.includes("container rm") && calls.includes("volume rm") && calls.includes("image rm"), "one failure must not skip independent cleanup");
    assert.ok(calls.includes("network ls"), "network cleanup must be verified");
  }
  // Private restore roots are created by the container user. A failed ownership
  // repair must remain a failed, retained cleanup result.
  for (const failRepair of [false, true]) {
    const workspace = mkdtempSync(path.join(scratch, "cartulary-standup-recovery-smoke."));
    mkdirSync(path.join(workspace, "runtime"));
    writeFileSync(path.join(scratch, "docker"), `#!/usr/bin/env bash\nexit ${failRepair ? 1 : 0}\n`, { mode: 0o700 });
    const evidence = path.join(scratch, "workspace-cleanup.json");
    const result = spawnSync("bash", ["-c", 'source "$1"; cleanup_package_workspace cartularymvprecoverysmk123 "$2" test-image "$3"', "workspace-test", helper, workspace, evidence], {
      env: { ...process.env, PATH: `${scratch}:${process.env.PATH}` }, encoding: "utf8", timeout: 10000,
    });
    assert.equal(result.status, failRepair ? 1 : 0);
    assert.equal(JSON.parse(readFileSync(evidence)).cleanup, failRepair ? "failed" : "passed");
  }
} finally { rmSync(scratch, { recursive: true, force: true }); }
