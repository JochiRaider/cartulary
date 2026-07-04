import { createHash } from "node:crypto";
import {
  accessSync,
  constants,
  existsSync,
  lstatSync,
  readFileSync,
  statSync,
} from "node:fs";
import path from "node:path";

import { secureWriteFile } from "../../contract/index.mjs";
import { compareStrings, relToRepo, resolvePath } from "./util.mjs";

const runtimeBinaryRegistry = Object.freeze({
  operator: Object.freeze({
    id: "operator",
    producerTarget: "build-operator",
    consumerEnv: "CARTULARY_OPERATOR_BIN",
  }),
});

class RuntimeBinaryError extends Error {
  constructor(message, { exitCode, reason }) {
    super(message);
    this.name = "RuntimeBinaryError";
    this.exitCode = exitCode;
    this.reason = reason;
  }
}

export function runtimeBinaryIDsForRows(rows) {
  return Array.from(new Set(rows.flatMap((row) => row.runtime_binaries ?? []))).sort(compareStrings);
}

function fileSha256(file) {
  return `sha256:${createHash("sha256").update(readFileSync(file)).digest("hex")}`;
}

function buildArtifactOutputDigest(ctx, file) {
  const display = relToRepo(ctx, file);
  const material = `output\t${display}\t${fileSha256(file)}\n`;
  return `sha256:${createHash("sha256").update(material).digest("hex")}`;
}

function runtimeBinaryPath(ctx, record) {
  const raw = ctx.env[record.consumerEnv] ?? "";
  if (String(raw).trim() === "") {
    throw new RuntimeBinaryError(
      `${record.consumerEnv} is required for runtime binary ${record.id}`,
      { exitCode: 2, reason: "configuration_error" },
    );
  }
  if (String(raw).includes("\0")) {
    throw new RuntimeBinaryError(`${record.consumerEnv} must not contain NUL`, {
      exitCode: 2,
      reason: "configuration_error",
    });
  }
  return resolvePath(ctx.repoRoot, String(raw).trim());
}

function readBuildArtifact(ctx, record) {
  const file = path.join(
    ctx.resultsRoot,
    ctx.runId,
    record.producerTarget,
    `build-artifact-cache-${record.producerTarget}.json`,
  );
  if (!existsSync(file)) {
    throw new RuntimeBinaryError(
      `missing ${record.producerTarget} build-artifact cache reference for runtime binary ${record.id}`,
      { exitCode: 11, reason: "artifact_error" },
    );
  }
  try {
    return { file, artifact: JSON.parse(readFileSync(file, "utf8")) };
  } catch (error) {
    throw new RuntimeBinaryError(
      `invalid ${record.producerTarget} build-artifact cache reference: ${error.message}`,
      { exitCode: 11, reason: "artifact_error" },
    );
  }
}

function validateRuntimeBinary(ctx, id) {
  const record = runtimeBinaryRegistry[id];
  if (!record) {
    throw new RuntimeBinaryError(`unknown runtime binary ${id}`, {
      exitCode: 2,
      reason: "configuration_error",
    });
  }
  const binaryPath = runtimeBinaryPath(ctx, record);
  let info;
  try {
    info = lstatSync(binaryPath);
  } catch {
    throw new RuntimeBinaryError(`${record.consumerEnv} does not exist: ${relToRepo(ctx, binaryPath)}`, {
      exitCode: 2,
      reason: "configuration_error",
    });
  }
  if (info.isSymbolicLink() || !info.isFile()) {
    throw new RuntimeBinaryError(`${record.consumerEnv} must name a regular executable file`, {
      exitCode: 2,
      reason: "configuration_error",
    });
  }
  try {
    accessSync(binaryPath, constants.X_OK);
  } catch {
    throw new RuntimeBinaryError(`${record.consumerEnv} is not executable: ${relToRepo(ctx, binaryPath)}`, {
      exitCode: 2,
      reason: "configuration_error",
    });
  }
  const stat = statSync(binaryPath);
  if (!stat.isFile()) {
    throw new RuntimeBinaryError(`${record.consumerEnv} must name a regular executable file`, {
      exitCode: 2,
      reason: "configuration_error",
    });
  }
  const { file: artifactFile, artifact } = readBuildArtifact(ctx, record);
  const expectedDigest = buildArtifactOutputDigest(ctx, binaryPath);
  if (artifact.output_digest_sha256 !== expectedDigest) {
    throw new RuntimeBinaryError(
      `${record.producerTarget} artifact digest does not match ${record.consumerEnv}`,
      { exitCode: 11, reason: "artifact_error" },
    );
  }
  return {
    id,
    producer_target: record.producerTarget,
    consumer_env: record.consumerEnv,
    source: "scheduler-produced",
    path: relToRepo(ctx, binaryPath),
    sha256: fileSha256(binaryPath),
    build_artifact_ref: relToRepo(ctx, artifactFile),
    build_artifact_output_digest: artifact.output_digest_sha256,
  };
}

export function validateRuntimeBinaries(ctx, rows, reportDir) {
  const ids = runtimeBinaryIDsForRows(rows);
  if (ids.length === 0) {
    return [];
  }
  const records = ids.map((id) => validateRuntimeBinary(ctx, id));
  secureWriteFile(
    path.join(reportDir, "runtime-binaries.json"),
    `${JSON.stringify({ runtime_binaries: records }, null, 2)}\n`,
  );
  return records;
}
