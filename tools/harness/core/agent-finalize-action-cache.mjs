import { spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";

export const actionCacheRecordSchemaID =
  "cartulary.agent_finalize_action_cache_record.v1";
export const actionCacheCommandID =
  "cartulary.harness.command.agent_finalize.v1";

const cacheSchemaID = actionCacheRecordSchemaID;

const commonSuffixes = [
  ".css",
  ".go",
  ".json",
  ".mjs",
  ".js",
  ".md",
  ".sh",
  ".sql",
  ".ts",
  ".tsx",
  ".yaml",
  ".yml",
];

const implementationFiles = [
  "Makefile",
  "tools/harness/core/agent-finalize-cli.mjs",
  "tools/harness/core/agent-finalize-action-cache.mjs",
  "tools/harness/core/harness-contract.mjs",
  "tools/schemas/cartulary.agent_finalize_action_cache_record.v1.schema.json",
  "tools/schemas/cartulary.agent_finalize_summary.v3.schema.json",
  "tools/task_surface.generated.mk",
  "tools/task_surface_manifest.json",
];

const profileDefinitions = {
  "agent_finalize.structure_ledger_refresh.v1": {
    prefixes: [
      "docs/spec/",
      "docs/testing/",
      "tools/",
      "packages/",
      "apps/web/",
    ],
    suffixes: commonSuffixes,
    files: [
      "Makefile",
      "docs/testing-harness-nlspec.md",
      "docs/domain.md",
      "package.json",
      "pnpm-lock.yaml",
      "pnpm-workspace.yaml",
    ],
    env: [
      "NODE_BIN",
      "MAKE",
      "CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT",
      "CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT",
    ],
    outputPaths: [
      "tools/browser_e2e_batch_manifest.json",
      "tools/execution_topology_render_index.json",
      "tools/scheduler_manifest.json",
      "tools/task_surface.generated.mk",
      "tools/task_surface_manifest.json",
    ],
    outputPrefixes: ["docs/testing/"],
    outputSuffixes: ["_coverage_ledger.md"],
  },
  "agent_finalize.schema_shape_validation.v1": {
    prefixes: ["docs/spec/", "docs/testing/", "tools/"],
    suffixes: commonSuffixes,
    files: [
      "Makefile",
      "docs/testing-harness-nlspec.md",
      "package.json",
      "pnpm-lock.yaml",
      "pnpm-workspace.yaml",
    ],
    env: [
      "NODE_BIN",
      "MAKE",
      "CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT",
      "CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT",
    ],
    outputPaths: [],
  },
  "agent_finalize.duration_baseline_refresh.v1": {
    prefixes: ["docs/testing/", "tools/"],
    suffixes: commonSuffixes,
    files: [
      "Makefile",
      "docs/testing-harness-nlspec.md",
      "package.json",
      "pnpm-lock.yaml",
      "pnpm-workspace.yaml",
    ],
    env: [
      "NODE_BIN",
      "MAKE",
      "PRUNE_OBSERVED_PACKAGES",
      "ALLOW_COMMAND_OVERHEAD_DECREASE",
      "CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT",
      "CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT",
    ],
    outputPaths: [
      "tools/browser_e2e_batch_manifest.json",
      "tools/browser_e2e_duration_baselines.json",
      "tools/execution_topology_render_index.json",
      "tools/go_test_duration_baselines.json",
      "tools/harness_smoke_duration_baselines.json",
      "tools/scheduler_manifest.json",
      "tools/service_backed_make_target_duration_baselines.json",
      "tools/task_surface.generated.mk",
      "tools/task_surface_manifest.json",
    ],
  },
  "agent_finalize.duration_baseline_coverage.v1": {
    prefixes: ["tools/"],
    suffixes: commonSuffixes,
    files: ["Makefile", "docs/testing-harness-nlspec.md"],
    env: [
      "NODE_BIN",
      "MAKE",
      "GO_TEST_DURATION_BASELINE",
      "CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT",
      "CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT",
    ],
    outputPaths: [],
  },
  "agent_finalize.duration_baseline_drift_validation.v1": {
    prefixes: ["tools/"],
    suffixes: commonSuffixes,
    files: ["Makefile", "docs/testing-harness-nlspec.md"],
    env: [
      "NODE_BIN",
      "MAKE",
      "GO_TEST_DURATION_BASELINE",
      "BROWSER_E2E_DURATION_BASELINE",
      "SERVICE_BACKED_MAKE_TARGET_DURATION_BASELINE",
      "HARNESS_SMOKE_DURATION_BASELINE",
      "CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT",
      "CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT",
    ],
    outputPaths: [],
  },
  "agent_finalize.scheduler_drift_validation.v1": {
    prefixes: ["tools/"],
    suffixes: commonSuffixes,
    files: ["Makefile", "docs/testing-harness-nlspec.md"],
    env: [
      "NODE_BIN",
      "MAKE",
      "SCHEDULER_WARM_CHECK_BUDGET_MS",
      "SCHEDULER_WARM_CHECK_BALANCE_RATIO",
      "CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT",
      "CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT",
    ],
    outputPaths: [],
  },
};

function stableJSONString(value) {
  if (Array.isArray(value)) {
    return `[${value.map((entry) => stableJSONString(entry)).join(",")}]`;
  }
  if (value && typeof value === "object") {
    return `{${Object.keys(value)
      .sort((left, right) => left.localeCompare(right))
      .map((key) => `${JSON.stringify(key)}:${stableJSONString(value[key])}`)
      .join(",")}}`;
  }
  return JSON.stringify(value);
}

function sha256Hex(value) {
  return createHash("sha256").update(value).digest("hex");
}

function sha256Digest(value) {
  return `sha256:${sha256Hex(value)}`;
}

function repoRelative(repoRoot, file) {
  const relative = path.relative(repoRoot, file).replaceAll("\\", "/");
  return relative.startsWith("../") || path.isAbsolute(relative)
    ? file.replaceAll("\\", "/")
    : relative;
}

function sanitize(value) {
  return value.replace(/[^A-Za-z0-9._-]+/g, "-");
}

function readJSON(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch {
    return null;
  }
}

function gitFiles(repoRoot) {
  const result = spawnSync("git", ["ls-files", "-z"], {
    cwd: repoRoot,
    encoding: "utf8",
  });
  if (result.status !== 0) {
    throw new Error(`git ls-files failed: ${result.stderr || result.status}`);
  }
  return result.stdout
    .split("\0")
    .filter(Boolean)
    .sort((left, right) => left.localeCompare(right));
}

function matchesProfile(file, profile) {
  if ((profile.files ?? []).includes(file)) {
    return true;
  }
  const prefixMatch = (profile.prefixes ?? []).some((prefix) =>
    file.startsWith(prefix),
  );
  const suffixMatch = (profile.suffixes ?? []).some((suffix) =>
    file.endsWith(suffix),
  );
  return prefixMatch && suffixMatch;
}

function hashPathSet(repoRoot, files) {
  const hash = createHash("sha256");
  for (const file of [...new Set(files)].sort((left, right) =>
    left.localeCompare(right),
  )) {
    const absolute = path.join(repoRoot, file);
    hash.update(`file\0${file}\0`);
    if (!existsSync(absolute)) {
      hash.update("missing\0");
      continue;
    }
    hash.update(readFileSync(absolute));
    hash.update("\0");
  }
  return `sha256:${hash.digest("hex")}`;
}

function walkFiles(root) {
  const files = [];
  const stack = [root];
  while (stack.length > 0) {
    const current = stack.pop();
    let entries = [];
    try {
      entries = readdirSync(current, { withFileTypes: true });
    } catch {
      continue;
    }
    for (const entry of entries) {
      const next = path.join(current, entry.name);
      if (entry.isDirectory()) {
        stack.push(next);
      } else if (entry.isFile()) {
        files.push(next);
      }
    }
  }
  return files.sort((left, right) => left.localeCompare(right));
}

function hashDirectory(root) {
  if (!root) {
    return null;
  }
  const resolved = path.resolve(root);
  const hash = createHash("sha256");
  hash.update(`directory\0${resolved}\0`);
  for (const file of walkFiles(resolved)) {
    const relative = path.relative(resolved, file).replaceAll("\\", "/");
    hash.update(`file\0${relative}\0`);
    hash.update(readFileSync(file));
    hash.update("\0");
  }
  return `sha256:${hash.digest("hex")}`;
}

function outputPathsForProfile(repoRoot, profile, trackedFiles) {
  const paths = new Set(profile.outputPaths ?? []);
  for (const file of trackedFiles) {
    const prefixMatch = (profile.outputPrefixes ?? []).some((prefix) =>
      file.startsWith(prefix),
    );
    const suffixMatch = (profile.outputSuffixes ?? []).some((suffix) =>
      file.endsWith(suffix),
    );
    if (prefixMatch && suffixMatch) {
      paths.add(file);
    }
  }
  const testOutput = (process.env.CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT || "").trim();
  if (testOutput) {
    paths.add(repoRelative(repoRoot, path.resolve(testOutput)));
  }
  return [...paths].sort((left, right) => left.localeCompare(right));
}

function outputDigest(repoRoot, outputPaths) {
  const hash = createHash("sha256");
  const missing = [];
  for (const file of outputPaths) {
    const absolute = path.isAbsolute(file) ? file : path.join(repoRoot, file);
    const display = repoRelative(repoRoot, absolute);
    hash.update(`output\0${display}\0`);
    if (!existsSync(absolute)) {
      missing.push(display);
      hash.update("missing\0");
      continue;
    }
    hash.update(readFileSync(absolute));
    hash.update("\0");
  }
  return {
    digest: `sha256:${hash.digest("hex")}`,
    missing,
    paths: outputPaths.map((file) =>
      repoRelative(repoRoot, path.isAbsolute(file) ? file : path.join(repoRoot, file)),
    ),
  };
}

function envDigest(names) {
  const values = Object.fromEntries(
    [...new Set(names)].sort((left, right) => left.localeCompare(right)).map((name) => [
      name,
      process.env[name] ?? "",
    ]),
  );
  return sha256Digest(stableJSONString(values));
}

function toolchainDigest({ makeBin }) {
  return sha256Digest(
    stableJSONString({
      node_exec_path: process.execPath,
      node_version: process.version,
      make: makeBin || "make",
      node_bin_env: process.env.NODE_BIN ?? "",
    }),
  );
}

function cacheRoot(repoRoot) {
  const configured = (
    process.env.CARTULARY_AGENT_FINALIZE_ACTION_CACHE_DIR || ""
  ).trim();
  return configured
    ? path.resolve(configured)
    : path.join(repoRoot, ".cache", "cartulary", "agent-finalize-action-cache");
}

function recordPath(repoRoot, actionID, keyHash) {
  return path.join(cacheRoot(repoRoot), sanitize(actionID), `${keyHash}.json`);
}

function matchingRecordWithSameInput(repoRoot, actionID, inputDigestValue) {
  const actionDir = path.join(cacheRoot(repoRoot), sanitize(actionID));
  if (!existsSync(actionDir)) {
    return null;
  }
  for (const entry of readdirSync(actionDir, { withFileTypes: true })) {
    if (!entry.isFile() || !entry.name.endsWith(".json")) {
      continue;
    }
    const file = path.join(actionDir, entry.name);
    const record = readJSON(file);
    if (record?.digests?.input_digest_sha256 === inputDigestValue) {
      return { file, record };
    }
  }
  return null;
}

function cacheSummary({
  enabled,
  state,
  keyHash = null,
  profileID = null,
  actionContractVersion = null,
  inputDigestValue = null,
  outputDigestValue = null,
  record = null,
  reasonCode,
}) {
  return {
    enabled,
    state,
    cache_schema_id: cacheSchemaID,
    action_contract_version: actionContractVersion,
    key_sha256: keyHash ? `sha256:${keyHash}` : null,
    input_profile_id: profileID,
    input_digest_sha256: inputDigestValue,
    output_digest_sha256: outputDigestValue,
    record_path: record,
    reason_code: reasonCode,
  };
}

function disabledSummary(cache, reasonCode) {
  return {
    reusable: false,
    writable: false,
    summary: cacheSummary({
      enabled: false,
      state: "disabled",
      profileID: cache?.inputProfileID ?? null,
      actionContractVersion: cache?.actionContractVersion ?? null,
      reasonCode,
    }),
  };
}

export function evaluateActionCache({
  actionDefinition,
  repoRoot,
  makeBin,
  retainedRunRoot = null,
}) {
  const cache = actionDefinition.cache ?? null;
  if (!cache?.eligible) {
    return {
      reusable: false,
      writable: false,
      summary: cacheSummary({
        enabled: false,
        state: "ineligible",
        reasonCode: "action_ineligible",
      }),
    };
  }
  if (process.env.CI === "1") {
    return disabledSummary(cache, "ci_disabled");
  }
  if (process.env.CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE === "1") {
    return disabledSummary(cache, "env_disabled");
  }

  const profile = profileDefinitions[cache.inputProfileID];
  if (!profile) {
    return {
      reusable: false,
      writable: false,
      summary: cacheSummary({
        enabled: false,
        state: "ineligible",
        profileID: cache.inputProfileID,
        actionContractVersion: cache.actionContractVersion,
        reasonCode: "input_profile_unknown",
      }),
    };
  }

  const trackedFiles = gitFiles(repoRoot);
  const profileFiles = trackedFiles.filter((file) => matchesProfile(file, profile));
  const repoInputDigest = hashPathSet(repoRoot, [
    ...profileFiles,
    ...(profile.files ?? []),
  ]);
  const implementationDigest = hashPathSet(repoRoot, implementationFiles);
  const output = outputDigest(
    repoRoot,
    outputPathsForProfile(repoRoot, profile, trackedFiles),
  );
  const retainedRunDigest = retainedRunRoot ? hashDirectory(retainedRunRoot) : null;
  const substepDigest = sha256Digest(
    stableJSONString(actionDefinition.substeps ?? []),
  );
  const envDigestValue = envDigest(profile.env ?? []);
  const toolchainDigestValue = toolchainDigest({ makeBin });
  const inputDigestValue = sha256Digest(
    stableJSONString({
      schema_id: cacheSchemaID,
      action_id: actionDefinition.actionID,
      action_contract_version: cache.actionContractVersion,
      command_id: actionCacheCommandID,
      input_profile_id: cache.inputProfileID,
      repo_input_digest: repoInputDigest,
      implementation_digest: implementationDigest,
      toolchain_digest: toolchainDigestValue,
      environment_digest: envDigestValue,
      retained_run_digest: retainedRunDigest,
      substep_digest: substepDigest,
    }),
  );
  const keyMaterial = stableJSONString({
    schema_id: cacheSchemaID,
    action_id: actionDefinition.actionID,
    action_contract_version: cache.actionContractVersion,
    command_id: actionCacheCommandID,
    input_profile_id: cache.inputProfileID,
    input_digest_sha256: inputDigestValue,
    output_digest_sha256: output.digest,
  });
  const keyHash = sha256Hex(keyMaterial);
  const file = recordPath(repoRoot, actionDefinition.actionID, keyHash);
  const recordRel = repoRelative(repoRoot, file);

  if (output.missing.length > 0) {
    return {
      reusable: false,
      writable: false,
      summary: cacheSummary({
        enabled: true,
        state: "miss",
        keyHash,
        profileID: cache.inputProfileID,
        actionContractVersion: cache.actionContractVersion,
        inputDigestValue,
        outputDigestValue: output.digest,
        record: recordRel,
        reasonCode: "output_missing",
      }),
      record: {
        repoRoot,
        file,
        keyHash,
        inputDigestValue,
        output,
        digests: {
          repo_input_digest: repoInputDigest,
          implementation_digest: implementationDigest,
          toolchain_digest: toolchainDigestValue,
          environment_digest: envDigestValue,
          retained_run_digest: retainedRunDigest,
          substep_digest: substepDigest,
        },
      },
    };
  }

  const existing = readJSON(file);
  if (!existing && existsSync(file)) {
    return {
      reusable: false,
      writable: true,
      summary: cacheSummary({
        enabled: true,
        state: "corrupt",
        keyHash,
        profileID: cache.inputProfileID,
        actionContractVersion: cache.actionContractVersion,
        inputDigestValue,
        outputDigestValue: output.digest,
        record: recordRel,
        reasonCode: "cache_record_invalid",
      }),
      record: {
        repoRoot,
        file,
        keyHash,
        inputDigestValue,
        output,
        digests: {
          repo_input_digest: repoInputDigest,
          implementation_digest: implementationDigest,
          toolchain_digest: toolchainDigestValue,
          environment_digest: envDigestValue,
          retained_run_digest: retainedRunDigest,
          substep_digest: substepDigest,
        },
      },
    };
  }
  if (existing) {
    const valid =
      existing.schema_id === cacheSchemaID &&
      existing.action_id === actionDefinition.actionID &&
      existing.command_id === actionCacheCommandID &&
      existing.key_sha256 === `sha256:${keyHash}` &&
      existing.input_profile_id === cache.inputProfileID &&
      existing.action_contract_version === cache.actionContractVersion &&
      existing.digests?.input_digest_sha256 === inputDigestValue &&
      existing.digests?.output_digest_sha256 === output.digest &&
      Array.isArray(existing.output_paths) &&
      existing.output_paths.every((entry) => output.paths.includes(entry));
    if (valid) {
      return {
        reusable: true,
        writable: false,
        summary: cacheSummary({
          enabled: true,
          state: "hit",
          keyHash,
          profileID: cache.inputProfileID,
          actionContractVersion: cache.actionContractVersion,
          inputDigestValue,
          outputDigestValue: output.digest,
          record: recordRel,
          reasonCode: "cache_hit",
        }),
      };
    }
    return {
      reusable: false,
      writable: true,
      summary: cacheSummary({
        enabled: true,
        state: "corrupt",
        keyHash,
        profileID: cache.inputProfileID,
        actionContractVersion: cache.actionContractVersion,
        inputDigestValue,
        outputDigestValue: output.digest,
        record: recordRel,
        reasonCode: "cache_record_invalid",
      }),
      record: {
        repoRoot,
        file,
        keyHash,
        inputDigestValue,
        output,
        digests: {
          repo_input_digest: repoInputDigest,
          implementation_digest: implementationDigest,
          toolchain_digest: toolchainDigestValue,
          environment_digest: envDigestValue,
          retained_run_digest: retainedRunDigest,
          substep_digest: substepDigest,
        },
      },
    };
  }

  const sameInputRecord = matchingRecordWithSameInput(
    repoRoot,
    actionDefinition.actionID,
    inputDigestValue,
  );
  const reasonCode =
    sameInputRecord &&
    sameInputRecord.record?.digests?.output_digest_sha256 !== output.digest
      ? "output_changed"
      : "cache_record_missing";

  return {
    reusable: false,
    writable: true,
    summary: cacheSummary({
      enabled: true,
      state: "miss",
      keyHash,
      profileID: cache.inputProfileID,
      actionContractVersion: cache.actionContractVersion,
      inputDigestValue,
      outputDigestValue: output.digest,
      record: recordRel,
      reasonCode,
    }),
    record: {
      repoRoot,
      file,
      keyHash,
      inputDigestValue,
      output,
      digests: {
        repo_input_digest: repoInputDigest,
        implementation_digest: implementationDigest,
        toolchain_digest: toolchainDigestValue,
        environment_digest: envDigestValue,
        retained_run_digest: retainedRunDigest,
        substep_digest: substepDigest,
      },
    },
  };
}

export function writeActionCacheRecord({ actionDefinition, evaluation }) {
  if (!evaluation?.writable || !evaluation.record) {
    return;
  }
  const refreshedOutput = outputDigest(
    evaluation.record.repoRoot,
    evaluation.record.output.paths,
  );
  if (refreshedOutput.missing.length > 0) {
    return;
  }
  const record = {
    schema_id: cacheSchemaID,
    action_id: actionDefinition.actionID,
    command_id: actionCacheCommandID,
    action_contract_version: actionDefinition.cache.actionContractVersion,
    input_profile_id: actionDefinition.cache.inputProfileID,
    key_sha256: `sha256:${evaluation.record.keyHash}`,
    cache_schema_id: cacheSchemaID,
    digests: {
      ...evaluation.record.digests,
      input_digest_sha256: evaluation.record.inputDigestValue,
      output_digest_sha256: refreshedOutput.digest,
    },
    output_paths: refreshedOutput.paths,
    updated_at: new Date().toISOString(),
  };
  mkdirSync(path.dirname(evaluation.record.file), { recursive: true });
  writeFileSync(evaluation.record.file, `${stableJSONString(record)}\n`);
}
