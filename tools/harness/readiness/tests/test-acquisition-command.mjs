import assert from "node:assert/strict";
import test from "node:test";
import { spawnSync } from "node:child_process";
import { browserAcquisitionLaunchArguments } from "../../test-support/owner-fixtures.mjs";
test("browser acquisition launches its owning dispatcher and rejects invalid ownership", () => {
  const command = browserAcquisitionLaunchArguments("missing-acquisition", "invalid-launch", process.execPath, ["--version"]);
  const childEnv = { ...process.env };
  delete childEnv.CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT;
  const child = spawnSync(process.execPath, command, { encoding: "utf8", env: childEnv });
  assert.notEqual(child.status, 0, "an import-only facade must never count as a successful browser producer");
  assert.match(child.stderr, /browser acquisition launch failed/);
});
