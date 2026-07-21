import { spawn, spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import {
  existsSync,
  mkdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";

import {
  createSecureWriteStream,
  redactString,
  secureMkdir,
  secureWriteFile,
} from "../../contract/index.mjs";
import {
  goChildEnv,
  renderGoTestCommand,
} from "./command.mjs";
import { prepareSharedArtifactDir } from "./context.mjs";
import {
  isCrossTargetSharedReport,
  resolveExecutionFamilyPostgresFixturePolicy,
  resolveExecutionFamilySpec,
  runtimeRowsForExecution,
} from "./planning.mjs";
import {
  runtimeBinaryIDsForRows,
  validateRuntimeBinaries,
} from "./runtime-binaries.mjs";
import { runHelper } from "./summary-emission.mjs";
import {
  captureFinish,
  captureStart,
  nowUTC,
  monotonicMs,
  sleep,
} from "./util.mjs";

function hashGoTestDependencyInputs(ctx) {
  const hash = createHash("sha256");
  const version = spawnSync(ctx.goBin, ["version"], {
    cwd: ctx.repoRoot,
    encoding: "utf8",
  });
  hash.update(version.stdout || "");
  hash.update("\n-- go.mod --\n");
  hash.update(readFileSync(path.join(ctx.repoRoot, "go.mod")));
  hash.update("\n-- go.sum --\n");
  hash.update(readFileSync(path.join(ctx.repoRoot, "go.sum")));
  return hash.digest("hex");
}

async function acquireDirectoryLock(lockDir, label, timeoutSeconds, metadata) {
  const started = monotonicMs();
  while (true) {
    try {
      mkdirSync(lockDir, { recursive: false, mode: 0o700 });
      for (const [name, value] of Object.entries(metadata)) {
        secureWriteFile(path.join(lockDir, name), `${value}\n`);
      }
      return;
    } catch (error) {
      if (error?.code !== "EEXIST") {
        throw error;
      }
    }

    const ownerPid = Number.parseInt(
      existsSync(path.join(lockDir, "pid"))
        ? readFileSync(path.join(lockDir, "pid"), "utf8")
        : "",
      10,
    );
    if (Number.isInteger(ownerPid)) {
      try {
        process.kill(ownerPid, 0);
      } catch (error) {
        if (error?.code === "ESRCH") {
          rmSync(lockDir, { recursive: true, force: true });
          continue;
        }
      }
    }

    if (monotonicMs() - started >= timeoutSeconds * 1000) {
      throw new Error(`${label}_timeout lock=${lockDir}`);
    }
    await sleep(100);
  }
}

async function warmGoTestDependencies(ctx) {
  const warmRoot = path.join(ctx.goModCacheDir, ".cartulary-go-test-warm");
  const lockDir = path.join(warmRoot, "lock");
  const timeoutSeconds = Number.parseInt(
    ctx.env.CARTULARY_GO_TEST_WARM_LOCK_TIMEOUT_SECONDS || "300",
    10,
  );
  if (!Number.isInteger(timeoutSeconds) || timeoutSeconds < 1) {
    throw new Error(
      `invalid CARTULARY_GO_TEST_WARM_LOCK_TIMEOUT_SECONDS=${ctx.env.CARTULARY_GO_TEST_WARM_LOCK_TIMEOUT_SECONDS}`,
    );
  }
  mkdirSync(warmRoot, { recursive: true });
  const warmKey = hashGoTestDependencyInputs(ctx);
  const stampFile = path.join(warmRoot, `${warmKey}.stamp`);
  if (existsSync(stampFile)) {
    return;
  }
  await acquireDirectoryLock(
    lockDir,
    "go_test_dependency_warm_lock",
    timeoutSeconds,
    {
      pid: process.pid,
      acquired_at: nowUTC(),
    },
  );
  try {
    if (existsSync(stampFile)) {
      return;
    }
    for (const args of [
      ["mod", "download"],
      ["list", "-deps", "-test", "./..."],
    ]) {
      const result = spawnSync(ctx.goBin, args, {
        cwd: ctx.repoRoot,
        env: goChildEnv(ctx),
        stdio: "inherit",
      });
      if ((result.status ?? 1) !== 0) {
        process.exitCode = result.status ?? 1;
        throw new Error(
          `${ctx.goBin} ${args.join(" ")} exited ${result.status ?? 1}`,
        );
      }
    }
    writeFileSync(
      `${stampFile}.${process.pid}`,
      `warmed_at=${nowUTC()}\ngo=${spawnSync(ctx.goBin, ["version"], { encoding: "utf8" }).stdout || ""}`,
    );
    rmSync(stampFile, { force: true });
    writeFileSync(stampFile, readFileSync(`${stampFile}.${process.pid}`));
    rmSync(`${stampFile}.${process.pid}`, { force: true });
  } finally {
    rmSync(lockDir, { recursive: true, force: true });
  }
}

async function runGoTestCapture(ctx, regex, args, reportDir, policy = {}) {
  await warmGoTestDependencies(ctx);
  const runnerLog = path.join(reportDir, "runner.jsonl");
  const stderrLog = path.join(reportDir, "stderr.log");
  const stdoutStream = createSecureWriteStream(runnerLog);
  const stderrStream = createSecureWriteStream(stderrLog);
  let stdoutBuffer = "";

  const child = spawn(ctx.goBin, ["test", "-json", "-run", regex, ...args], {
    cwd: ctx.repoRoot,
    env: goChildEnv(ctx, policy),
    stdio: ["ignore", "pipe", "pipe"],
  });
  child.stdout.on("data", (chunk) => {
    const redactedChunk = redactString(chunk.toString("utf8"));
    stdoutStream.write(redactedChunk);
    if (ctx.outputMode === "quiet") {
      return;
    }
    stdoutBuffer += redactedChunk;
    const lines = stdoutBuffer.split(/\r?\n/u);
    stdoutBuffer = lines.pop() ?? "";
    for (const line of lines) {
      if (!line.trim()) {
        continue;
      }
      try {
        const entry = JSON.parse(line);
        if (typeof entry.Output === "string") {
          process.stderr.write(entry.Output);
        }
      } catch {
        // Preserve raw JSON in the log; malformed streaming lines are not lifecycle output.
      }
    }
  });
  child.stderr.on("data", (chunk) => {
    const redactedChunk = redactString(chunk.toString("utf8"));
    stderrStream.write(redactedChunk);
    if (ctx.outputMode !== "quiet") {
      process.stderr.write(redactedChunk);
    }
  });
  return await new Promise((resolve, reject) => {
    child.on("error", reject);
    child.on("close", (status) => {
      stdoutStream.end();
      stderrStream.end();
      resolve(status ?? 1);
    });
  });
}

export async function acquireSharedReportLock(ctx, sharedDir, sharedName) {
  const timeoutSeconds = Number.parseInt(
    ctx.env.CARTULARY_SHARED_REPORT_LOCK_TIMEOUT_SECONDS || "300",
    10,
  );
  if (!Number.isInteger(timeoutSeconds) || timeoutSeconds < 1) {
    throw new Error(
      `invalid CARTULARY_SHARED_REPORT_LOCK_TIMEOUT_SECONDS=${ctx.env.CARTULARY_SHARED_REPORT_LOCK_TIMEOUT_SECONDS}`,
    );
  }
  try {
    await acquireDirectoryLock(
      path.join(sharedDir, "capture.lock"),
      "shared_go_report_lock",
      timeoutSeconds,
      {
        pid: process.pid,
        shared_report: sharedName,
        acquired_at: nowUTC(),
      },
    );
  } catch (error) {
    if (String(error.message).startsWith("shared_go_report_lock_timeout")) {
      process.stderr.write(
        `shared_go_report_lock_timeout report=${sharedName} lock=${path.join(sharedDir, "capture.lock")}\n`,
      );
    }
    throw error;
  }
}

export function releaseSharedReportLock(sharedDir) {
  rmSync(path.join(sharedDir, "capture.lock"), {
    recursive: true,
    force: true,
  });
}

async function writeCrossTargetSharedExecutionMetadata(
  ctx,
  sharedDir,
  sharedName,
  window,
  status,
) {
  if (
    !isCrossTargetSharedReport(ctx, "backend-integration", sharedName) &&
    !isCrossTargetSharedReport(ctx, "backend-integration-support", sharedName)
  ) {
    return;
  }
  await runHelper(ctx, [
    "shared-execution",
    "backend-integration-shards",
    sharedName,
    status === 0 ? "pass" : "fail",
    window.startTime,
    window.endTime,
    String(window.durationMs),
    String(status),
    path.join(sharedDir, "shared-execution.json"),
  ]);
}

async function captureGoReportLocked(
  ctx,
  sharedDir,
  sharedName,
  regex,
  args,
  policy = {},
) {
  const commandText = renderGoTestCommand(ctx, regex, args, policy);
  const completeFile = path.join(sharedDir, "complete");
  if (existsSync(completeFile)) {
    const existingCommand = existsSync(path.join(sharedDir, "command.txt"))
      ? readFileSync(path.join(sharedDir, "command.txt"), "utf8").trimEnd()
      : "";
    if (existingCommand !== commandText) {
      throw new Error(
        [
          `shared_go_report_command_mismatch report=${sharedName}`,
          `shared go report ${sharedName} was created with a different command`,
          `existing: ${existingCommand}`,
          `current:  ${commandText}`,
        ].join("\n"),
      );
    }
    return { reportDir: sharedDir, usage: "reused" };
  }

  const started = captureStart();
  const status = await runGoTestCapture(ctx, regex, args, sharedDir, policy);
  const window = captureFinish(started);
  secureWriteFile(path.join(sharedDir, "command.txt"), `${commandText}\n`);
  secureWriteFile(
    path.join(sharedDir, "start_time.txt"),
    `${window.startTime}\n`,
  );
  secureWriteFile(path.join(sharedDir, "end_time.txt"), `${window.endTime}\n`);
  secureWriteFile(
    path.join(sharedDir, "duration_ms.txt"),
    `${window.durationMs}\n`,
  );
  secureWriteFile(path.join(sharedDir, "exit_status.txt"), `${status}\n`);
  await writeCrossTargetSharedExecutionMetadata(
    ctx,
    sharedDir,
    sharedName,
    window,
    status,
  );
  secureWriteFile(completeFile, "");
  return { reportDir: sharedDir, usage: "actual" };
}

export async function captureGoReport(
  ctx,
  sharedName,
  regex,
  args,
  policy = {},
) {
  const sharedDir = prepareSharedArtifactDir(ctx, sharedName);
  await acquireSharedReportLock(ctx, sharedDir, sharedName);
  try {
    return await captureGoReportLocked(
      ctx,
      sharedDir,
      sharedName,
      regex,
      args,
      policy,
    );
  } finally {
    releaseSharedReportLock(sharedDir);
  }
}

export async function assignExecutionFamily(ctx, target, familyOrShard) {
  const spec = resolveExecutionFamilySpec(ctx, target, familyOrShard);
  const policy = resolveExecutionFamilyPostgresFixturePolicy(
    ctx,
    target,
    familyOrShard,
  );
  const runtimeRows = runtimeRowsForExecution(ctx, target, familyOrShard);
  if (runtimeBinaryIDsForRows(runtimeRows).length > 0) {
    validateRuntimeBinaries(
      ctx,
      runtimeRows,
      prepareSharedArtifactDir(ctx, familyOrShard),
    );
  }
  const captured = await captureGoReport(
    ctx,
    familyOrShard,
    spec.regex,
    spec.args,
    policy,
  );
  if (
    captured.usage === "actual" &&
    isCrossTargetSharedReport(ctx, target, familyOrShard)
  ) {
    return { ...captured, usage: "reused" };
  }
  return captured;
}

export async function captureUnshardedGroup(ctx, group) {
  if (runtimeBinaryIDsForRows(group.rows).length > 0) {
    validateRuntimeBinaries(
      ctx,
      group.rows,
      prepareSharedArtifactDir(ctx, group.name),
    );
  }
  return await captureGoReport(
    ctx,
    group.name,
    group.regex,
    group.args,
    group.fixture_policy,
  );
}

function writeShardMetadata(metadataDir, sharedName, captured) {
  secureMkdir(metadataDir);
  const file = path.join(metadataDir, `${sharedName}.meta`);
  secureWriteFile(
    `${file}.${process.pid}`,
    `${captured.reportDir}\n${captured.usage}\n`,
  );
  rmSync(file, { force: true });
  secureWriteFile(file, readFileSync(`${file}.${process.pid}`));
  rmSync(`${file}.${process.pid}`, { force: true });
}

export async function captureScheduledShard(
  ctx,
  target,
  sharedName,
  metadataDir,
) {
  const captured = await assignExecutionFamily(ctx, target, sharedName);
  writeShardMetadata(metadataDir, sharedName, captured);
}

export async function captureNamedSharedReportsParallel(
  ctx,
  target,
  jobs,
  metadataDir,
  shardNames,
) {
  if (!Number.isInteger(jobs) || jobs < 1) {
    throw new Error(`invalid shard job count: ${jobs}`);
  }
  secureMkdir(metadataDir);
  let status = 0;
  let next = 0;
  const workerCount = Math.min(jobs, shardNames.length);
  await Promise.all(
    Array.from({ length: workerCount }, async () => {
      while (next < shardNames.length) {
        const shardName = shardNames[next];
        next += 1;
        try {
          await captureScheduledShard(ctx, target, shardName, metadataDir);
        } catch (error) {
          status = 1;
          process.stderr.write(`${error.message}\n`);
        }
      }
    }),
  );
  return status;
}
