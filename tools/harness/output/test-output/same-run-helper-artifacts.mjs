import { createHash } from "node:crypto";
import {
  existsSync,
  readFileSync,
  statSync,
} from "node:fs";
import path from "node:path";

import {
  prettyJSONString,
  secureMkdir,
  secureWriteFile,
  validateSchemaSync,
} from "../../contract/harness-contract.mjs";

export const sameRunHelperArtifactRefSchemaID =
  "cartulary.same_run_helper_artifact_ref.v2";

function normalizePath(value) {
  return String(value ?? "").replaceAll("\\", "/");
}

function relToRepo(value, repoRoot) {
  if (!value) {
    return "";
  }
  const normalized = normalizePath(value);
  if (!path.isAbsolute(value)) {
    return normalized;
  }
  const relative = normalizePath(path.relative(repoRoot, value));
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return normalized;
}

function resolveArtifactPath(value, repoRoot) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

function assertUnderRunRoot(file, runDir, label) {
  const relative = path.relative(runDir, file);
  if (
    relative === "" ||
    (!relative.startsWith("../") && relative !== ".." && !path.isAbsolute(relative))
  ) {
    return;
  }
  throw new Error(`${label} must be produced under the current run root`);
}

function sha256Buffer(buffer) {
  return `sha256:${createHash("sha256").update(buffer).digest("hex")}`;
}

function sha256File(file) {
  return sha256Buffer(readFileSync(file));
}

function stableDigest(value) {
  return sha256Buffer(Buffer.from(`${JSON.stringify(value)}\n`, "utf8"));
}

function sanitizePathSegment(value) {
  return String(value ?? "helper").replace(/[^A-Za-z0-9._-]+/gu, "-");
}

function artifactSortKey(artifact) {
  return `${artifact.role}\0${artifact.path_kind}\0${artifact.format}\0${artifact.path}`;
}

function appendArtifact(target, seen, entry, { repoRoot, runDir, label }) {
  if (!entry.path) {
    return;
  }
  const file = resolveArtifactPath(entry.path, repoRoot);
  if (!existsSync(file) || !statSync(file).isFile()) {
    throw new Error(`${label} missing same-run helper artifact ${entry.path}`);
  }
  assertUnderRunRoot(file, runDir, `${label} artifact ${entry.path}`);
  const artifact = {
    role: entry.role,
    path_kind: "file",
    format: entry.format,
    path: normalizePath(path.relative(runDir, file)),
    sha256: sha256File(file),
  };
  const key = artifactSortKey(artifact);
  if (seen.has(key)) {
    return;
  }
  seen.add(key);
  target.push(artifact);
}

function helperProducerArtifacts(helper, { repoRoot, runDir }) {
  const artifacts = [];
  const seen = new Set();
  for (const [index, phase] of (helper.phase_summaries ?? []).entries()) {
    const label = `helper ${helper.target} phase ${index + 1}`;
    appendArtifact(
      artifacts,
      seen,
      {
        role: "phase_summary",
        format: "json",
        path: phase.artifact,
      },
      { repoRoot, runDir, label },
    );
    appendArtifact(
      artifacts,
      seen,
      {
        role: "runner_json",
        format: "json",
        path: phase.runner_json,
      },
      { repoRoot, runDir, label },
    );
    appendArtifact(
      artifacts,
      seen,
      {
        role: "stdout_log",
        format: "log",
        path: phase.stdout_log,
      },
      { repoRoot, runDir, label },
    );
    appendArtifact(
      artifacts,
      seen,
      {
        role: "stderr_log",
        format: "log",
        path: phase.stderr_log,
      },
      { repoRoot, runDir, label },
    );
  }
  return artifacts.sort((left, right) =>
    artifactSortKey(left).localeCompare(artifactSortKey(right)),
  );
}

function writeValidatedJson(file, schemaID, value) {
  validateSchemaSync(schemaID, value);
  secureMkdir(path.dirname(file));
  secureWriteFile(file, prettyJSONString(value));
}

export function writeSameRunHelperArtifactRefs(
  helperArtifacts,
  {
    repoRoot,
    resultsRoot,
    runId,
    consumerTarget,
  },
) {
  if (!Array.isArray(helperArtifacts) || helperArtifacts.length === 0) {
    return [];
  }
  const runDir = path.join(resultsRoot, runId);
  const outputDir = path.join(runDir, "_shared", "same-run-helper-artifacts");
  const refs = [];
  for (const helper of helperArtifacts) {
    const producerArtifacts = helperProducerArtifacts(helper, { repoRoot, runDir });
    if (producerArtifacts.length === 0) {
      continue;
    }
    const declaredInputs = producerArtifacts.filter(
      (artifact) => artifact.role === "phase_summary",
    );
    const inputDigest = stableDigest({
      run_id: runId,
      helper_target: helper.target,
      declared_inputs: declaredInputs,
    });
    const outputDigest = stableDigest({
      run_id: runId,
      helper_target: helper.target,
      producer_artifacts: producerArtifacts,
    });
    const suffix = outputDigest.slice("sha256:".length, "sha256:".length + 12);
    const file = path.join(
      outputDir,
      `${sanitizePathSegment(helper.target)}-${suffix}.json`,
    );
    const artifactPath = relToRepo(file, repoRoot);
    const record = {
      schema_id: sameRunHelperArtifactRefSchemaID,
      run_id: runId,
      run_root: relToRepo(runDir, repoRoot),
      helper_target: helper.target,
      producer_work_unit_id: helper.target,
      reuse_scope: "same_run_only",
      accounting_mode: "helper_reused",
      scheduler_reused: false,
      declared_inputs: declaredInputs,
      producer_artifacts: producerArtifacts,
      consumer_refs: [
        {
          consumer_target: consumerTarget,
          consumer_work_unit_id: consumerTarget,
          accounting_mode: "helper_reused",
        },
      ],
      input_digest_sha256: inputDigest,
      output_digest_sha256: outputDigest,
      failure_behavior: "fail_closed_on_missing_or_digest_mismatch",
      created_at: new Date().toISOString(),
    };
    writeValidatedJson(file, sameRunHelperArtifactRefSchemaID, record);
    refs.push({
      target: helper.target,
      artifact: artifactPath,
      output_digest_sha256: outputDigest,
    });
  }
  return refs.sort((left, right) =>
    `${left.target}\0${left.artifact}`.localeCompare(`${right.target}\0${right.artifact}`),
  );
}
