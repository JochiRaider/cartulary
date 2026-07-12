import { existsSync } from "node:fs";
import { mkdir, writeFile } from "node:fs/promises";
import path from "node:path";

import { runLifecycle } from "../scheduler-runner.mjs";
import {
  readStringEnvFile,
  schedulerChildEnv,
} from "./runtime-command-helpers.mjs";

export function serviceSessionTarget(unit) {
  return typeof unit?.serviceSession?.target === "string" &&
    unit.serviceSession.target.trim() !== ""
    ? unit.serviceSession.target.trim()
    : "";
}

async function readServiceSessionEnv(envFile) {
  return readStringEnvFile(
    envFile,
    `service session env file ${envFile} must contain an object`,
  );
}

export function createServiceSessionRuntime({
  repoRoot,
  workUnits,
  tempDir,
  testServicesBin,
  resultsDir,
  runId,
}) {
  const targets = Array.from(
    new Set(workUnits.map(serviceSessionTarget).filter(Boolean)),
  ).sort((left, right) => left.localeCompare(right));
  const files = new Map(
    targets.map((target) => [
      target,
      {
        envFile: path.join(tempDir, `${target}-env.json`),
        leaseFile: path.join(tempDir, `${target}-lease.json`),
        metadataDir: path.join(tempDir, `${target}-go-shard-metadata`),
      },
    ]),
  );
  const cleanupStatus = new Map(
    targets.map((target) => [target, "not_started"]),
  );
  const cleanupDurationMs = new Map(
    targets.map((target) => [target, null]),
  );

  const requireTestServicesBin = () => {
    if (!testServicesBin) {
      throw new Error("TEST_SERVICES_BIN is required for scheduler service sessions");
    }
  };

  const serviceEnvFor = async (_unit, target) => {
    const sessionFiles = files.get(target);
    if (!sessionFiles) {
      return process.env;
    }
    return {
      ...process.env,
      ...(await readServiceSessionEnv(sessionFiles.envFile)),
    };
  };

  const writeDiagnosticEnv = async (unit) => {
    const sessionFiles = files.get(serviceSessionTarget(unit));
    if (!sessionFiles?.envFile) {
      return;
    }
    await mkdir(path.dirname(sessionFiles.envFile), { recursive: true });
    await writeFile(
      sessionFiles.envFile,
      `${JSON.stringify(
        {
          CARTULARY_TEST_RESULTS_DIR: resultsDir,
          CARTULARY_TEST_RUN_ID: runId,
          CARTULARY_TEST_TARGET: unit.target,
          CARTULARY_TEST_SERVICES_LIFECYCLE_MODE: "owned",
          CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
        },
        null,
        2,
      )}\n`,
      { mode: 0o600 },
    );
  };

  const recordChildLifecycle = async (unit, event) => {
    if (!unit?.serviceSession?.target) {
      return;
    }
    if (["service_session", "service_complete"].includes(unit.kind)) {
      return;
    }
    const sessionFiles = files.get(serviceSessionTarget(unit));
    if (!sessionFiles?.envFile || !existsSync(sessionFiles.envFile)) {
      return;
    }
    requireTestServicesBin();
    await runLifecycle(repoRoot, testServicesBin, [
      "record-lifecycle",
      "--env-file",
      sessionFiles.envFile,
      "--event",
      event,
      "--child-key",
      unit.id,
    ]);
  };

  const attachCommands = () => {
    for (const unit of workUnits) {
      if (unit.kind === "service_session") {
        const sessionFiles = files.get(serviceSessionTarget(unit));
        unit.command = () => {
          requireTestServicesBin();
          return {
            command: testServicesBin,
            args: [
              "start-suite",
              "--env-file",
              sessionFiles.envFile,
              "--lease-file",
              sessionFiles.leaseFile,
            ],
            env: schedulerChildEnv({
              ...process.env,
              ...unit.env,
              CARTULARY_TEST_RESULTS_DIR: resultsDir,
              CARTULARY_TEST_RUN_ID: runId,
              CARTULARY_TEST_TARGET: unit.target,
              CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
            }),
          };
        };
      } else if (unit.kind === "service_complete") {
        unit.command = () => ({
          command: process.execPath,
          args: ["-e", ""],
          env: process.env,
        });
      }
    }
  };

  const beforeUnitStart = async (unit) => {
    if (unit.kind === "service_session") {
      await writeDiagnosticEnv(unit);
    }
    await recordChildLifecycle(unit, "child_started");
  };

  const afterUnitFinish = async (unit) => {
    await recordChildLifecycle(unit, "child_finished");
  };

  const cleanup = async () => {
    let cleanupFailure = null;
    for (const target of targets) {
      const sessionFiles = files.get(target);
      if (!sessionFiles?.leaseFile) {
        continue;
      }
      if (!existsSync(sessionFiles.leaseFile)) {
        cleanupStatus.set(target, "skipped_no_lease");
        continue;
      }
      cleanupStatus.set(target, "running");
      const startedAt = Date.now();
      const status = await runLifecycle(repoRoot, testServicesBin, [
        "terminate-suite",
        "--lease",
        sessionFiles.leaseFile,
      ]).then(
        () => 0,
        () => 1,
      );
      cleanupDurationMs.set(target, Math.max(0, Date.now() - startedAt));
      if (status !== 0 && !cleanupFailure) {
        cleanupStatus.set(target, "failed");
        cleanupFailure = {
          status,
          label: `${target}:terminate-suite`,
        };
      } else if (status === 0) {
        cleanupStatus.set(target, "pass");
      }
    }
    return cleanupFailure;
  };

  const summary = (reporter, relPath = (value) => value) =>
    targets.map((target) => {
      const sessionFiles = files.get(target);
      const setupRecord = reporter.completedWork.find(
        (record) =>
          record.service_session_target === target &&
          record.work_unit_type === "service_session",
      );
      const childWork = reporter.completedWork.filter(
        (record) =>
          record.service_session_target === target &&
          !["service_session", "service_complete"].includes(
            record.work_unit_type,
          ),
      );
      const childWorkStartedAt =
        childWork.length > 0
          ? Math.min(...childWork.map((record) => record.started_monotonic_ms))
          : null;
      return {
        target,
        env_file: relPath(sessionFiles.envFile),
        lease_file: relPath(sessionFiles.leaseFile),
        metadata_dir: relPath(sessionFiles.metadataDir),
        cleanup_status: cleanupStatus.get(target) ?? "unknown",
        setup_duration_ms: setupRecord?.duration_ms ?? null,
        ready_at_monotonic_ms:
          setupRecord?.status === 0 ? setupRecord.finished_monotonic_ms : null,
        child_work_started_at_monotonic_ms: childWorkStartedAt,
        cleanup_duration_ms: cleanupDurationMs.get(target) ?? null,
      };
    });

  return {
    targets,
    files,
    attachCommands,
    serviceEnvFor,
    metadataDirForUnit: (unit) =>
      files.get(serviceSessionTarget(unit))?.metadataDir ?? tempDir,
    aggregateMetadataDirForUnit: (unit) =>
      files.get(serviceSessionTarget(unit) || targets[0])?.metadataDir ?? tempDir,
    beforeUnitStart,
    afterUnitFinish,
    cleanup,
    summary,
  };
}
