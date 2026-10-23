#!/usr/bin/env node
import assert from "node:assert/strict";
import { mkdirSync, mkdtempSync, rmSync, symlinkSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import { execFileSync } from "node:child_process";
import { fileURLToPath } from "node:url";

import {
  HarnessConfigError,
  resolveHarnessConfig,
  resolveMachineStatePaths,
} from "../harness-contract.mjs";

const scratch = mkdtempSync(path.join(os.tmpdir(), "cartulary-machine-state-"));
const repo = path.join(scratch, "repo");
mkdirSync(repo);

function values(resolved) {
  return Object.fromEntries(
    Object.entries(resolved).map(([name, entry]) => [name, entry.value]),
  );
}

function assertConfigError(fn, pattern) {
  assert.throws(fn, (error) => {
    assert.ok(error instanceof HarnessConfigError);
    assert.match(error.message, pattern);
    return true;
  });
}

try {
  const xdg = path.join(scratch, "xdg");
  const home = path.join(scratch, "home");
  assert.deepEqual(
    values(resolveMachineStatePaths({ XDG_CACHE_HOME: xdg, HOME: home }, { root: repo })),
    {
      CARTULARY_MACHINE_CACHE_DIR: path.join(xdg, "cartulary"),
      GO_CACHE_DIR: path.join(xdg, "cartulary", "go", "build"),
      GO_MOD_CACHE_DIR: path.join(xdg, "cartulary", "go", "mod"),
      GO_TMP_DIR: path.join(xdg, "cartulary", "go", "tmp"),
      CARTULARY_HARNESS_SCRATCH_ROOT: path.join(xdg, "cartulary", "harness-scratch"),
    },
  );
  assert.equal(
    resolveMachineStatePaths(
      { XDG_CACHE_HOME: "relative-cache", HOME: home },
      { root: repo },
    ).CARTULARY_MACHINE_CACHE_DIR.value,
    path.join(home, ".cache", "cartulary"),
  );

  const custom = path.join(scratch, "custom");
  const override = resolveMachineStatePaths(
    {
      HOME: home,
      CARTULARY_MACHINE_CACHE_DIR: custom,
      GO_CACHE_DIR: path.join(scratch, "build-override"),
      GO_MOD_CACHE_DIR: path.join(scratch, "mod-override"),
      GO_TMP_DIR: path.join(scratch, "tmp-override"),
      CARTULARY_HARNESS_SCRATCH_ROOT: path.join(scratch, "scratch-override"),
    },
    { root: repo },
  );
  assert.equal(override.GO_CACHE_DIR.value, path.join(scratch, "build-override"));
  assert.equal(override.CARTULARY_HARNESS_SCRATCH_ROOT.value, path.join(scratch, "scratch-override"));
  for (const [value, pattern] of [["", /must/u], ["relative", /absolute/u], [repo, /outside/u]]) {
    assertConfigError(() => resolveMachineStatePaths({ HOME: home, CARTULARY_MACHINE_CACHE_DIR: custom, CARTULARY_HARNESS_SCRATCH_ROOT: value }, { root: repo }), pattern);
  }

  const helper = fileURLToPath(new URL("../../test-support/harness-scratch.sh", import.meta.url));
  for (const [environment, expected] of [
    [{ CARTULARY_MACHINE_CACHE_DIR: custom, TMPDIR: path.join(scratch, "small-tmpfs") }, path.join(custom, "harness-scratch")],
    [{ XDG_CACHE_HOME: xdg }, path.join(xdg, "cartulary", "harness-scratch")],
    [{ XDG_CACHE_HOME: "relative" }, path.join(home, ".cache", "cartulary", "harness-scratch")],
    [{ CARTULARY_HARNESS_SCRATCH_ROOT: path.join(scratch, "explicit") }, path.join(scratch, "explicit")],
  ]) {
    const directory = execFileSync("bash", ["-c", 'source "$1"; cartulary_harness_mktemp_dir placement.XXXXXX', "scratch-test", helper], { encoding: "utf8", env: { PATH: process.env.PATH, HOME: home, CARTULARY_HARNESS_REPO_ROOT: repo, ...environment } }).trim();
    assert.equal(path.dirname(directory), expected);
    rmSync(directory, { recursive: true });
  }

  assertConfigError(
    () =>
      resolveMachineStatePaths(
        { HOME: home, CARTULARY_MACHINE_CACHE_DIR: "relative" },
        { root: repo },
      ),
    /must be absolute/u,
  );
  assertConfigError(
    () =>
      resolveMachineStatePaths(
        { HOME: home, CARTULARY_MACHINE_CACHE_DIR: path.join(repo, ".cache") },
        { root: repo },
      ),
    /outside the repository/u,
  );
  assertConfigError(
    () =>
      resolveMachineStatePaths(
        {
          HOME: home,
          CARTULARY_MACHINE_CACHE_DIR: custom,
          GO_TMP_DIR: "relative-tmp",
        },
        { root: repo },
      ),
    /GO_TMP_DIR must be absolute/u,
  );
  assertConfigError(
    () =>
      resolveMachineStatePaths(
        {
          HOME: home,
          CARTULARY_MACHINE_CACHE_DIR: custom,
          GO_CACHE_DIR: path.join(repo, "go-build"),
        },
        { root: repo },
      ),
    /GO_CACHE_DIR must be outside the repository/u,
  );
  assertConfigError(
    () =>
      resolveMachineStatePaths(
        {
          HOME: home,
          CARTULARY_MACHINE_CACHE_DIR: custom,
          GO_CACHE_DIR: path.join(scratch, "overlap"),
          GO_MOD_CACHE_DIR: path.join(scratch, "overlap", "mod"),
          GO_TMP_DIR: path.join(scratch, "tmp-distinct"),
        },
        { root: repo },
      ),
    /distinct and non-overlapping/u,
  );

  const symlinkRoot = path.join(scratch, "symlink-root");
  symlinkSync(repo, symlinkRoot, "dir");
  assertConfigError(
    () =>
      resolveMachineStatePaths(
        { HOME: home, CARTULARY_MACHINE_CACHE_DIR: path.join(symlinkRoot, "cache") },
        { root: repo },
      ),
    /outside the repository/u,
  );

  const globalEnv = {
    ...process.env,
    HOME: home,
    CARTULARY_MACHINE_CACHE_DIR: custom,
    GO_CACHE_DIR: path.join(custom, "go", "build"),
    GO_MOD_CACHE_DIR: path.join(custom, "go", "mod"),
    GO_TMP_DIR: path.join(custom, "go", "tmp"),
    CARTULARY_MAKE_INPUT_SOURCES:
      "CARTULARY_MACHINE_CACHE_DIR=cli GO_CACHE_DIR=cli GO_MOD_CACHE_DIR=cli GO_TMP_DIR=cli",
  };
  delete globalEnv.CARTULARY_HARNESS_IDENTITY_PREPARED;
  delete globalEnv.CARTULARY_TEST_TARGET;
  const resolved = resolveHarnessConfig("help", globalEnv, { root: repo });
  assert.equal(
    resolved.variables.global_inputs.GO_TMP_DIR.source,
    "make_command_line",
  );
  for (const nativeAlias of ["GOCACHE", "GOMODCACHE", "GOTMPDIR", "GOROOT"]) {
    assertConfigError(
      () =>
        resolveHarnessConfig(
          "help",
          {
            ...globalEnv,
            [nativeAlias]: path.join(scratch, "native-alias"),
            CARTULARY_MAKE_INPUT_SOURCES:
              `${globalEnv.CARTULARY_MAKE_INPUT_SOURCES} ${nativeAlias}=cli`,
          },
          { root: repo },
        ),
      new RegExp(`${nativeAlias} is not declared`, "u"),
    );
  }
} finally {
  rmSync(scratch, { recursive: true, force: true });
}

process.stdout.write("machine-state configuration checks passed\n");
