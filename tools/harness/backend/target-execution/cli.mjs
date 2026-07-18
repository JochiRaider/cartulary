import { captureScheduledShard } from "./capture.mjs";
import { createGoTargetContext } from "./context.mjs";
import { inspectAggregateCommand } from "./planning.mjs";
import {
  finalizeScheduledShards,
  runShardedTarget,
  runUnshardedTarget,
} from "./targets.mjs";
import { captureStart } from "./util.mjs";

function usage() {
  return [
    "usage: run-go-target.mjs <backend-unit|backend-store|backend-integration|backend-integration-support|backend-process>",
    "       run-go-target.mjs inspect-aggregate-command <target> <execution-family-or-shard>",
    "       run-go-target.mjs capture-shard <target> <shard-name> <metadata-dir>",
    "       run-go-target.mjs finalize-shards <target> <metadata-dir> [shard-name...]",
  ].join("\n");
}

export async function runGoTargetCLI(argv, options = {}) {
  if (argv.length === 0) {
    process.stderr.write(`${usage()}\n`);
    return 2;
  }
  const ctx = createGoTargetContext(options);
  const [command, ...rest] = argv;
  try {
    switch (command) {
      case "inspect-aggregate-command":
        if (rest.length !== 2) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        process.stdout.write(
          `${inspectAggregateCommand(ctx, rest[0], rest[1])}\n`,
        );
        return 0;
      case "capture-shard":
        if (rest.length !== 3) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        await captureScheduledShard(ctx, rest[0], rest[1], rest[2]);
        return 0;
      case "finalize-shards":
        if (rest.length < 2) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await finalizeScheduledShards(ctx, rest[0], rest[1], rest.slice(2));
      case "backend-unit":
        if (rest.length !== 0) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await runUnshardedTarget(ctx, "backend-unit");
      case "backend-store":
        if (rest.length !== 0) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await runShardedTarget(ctx, "backend-store");
      case "backend-integration":
        if (rest.length !== 0) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await runShardedTarget(ctx, "backend-integration");
      case "backend-integration-support":
        if (rest.length !== 0) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await runShardedTarget(ctx, "backend-integration-support");
      case "backend-process":
        if (rest.length !== 0) {
          process.stderr.write(`${usage()}\n`);
          return 2;
        }
        ctx.invocation = captureStart();
        return await runShardedTarget(ctx, "backend-process");
      default:
        process.stderr.write(`${usage()}\n`);
        return 2;
    }
  } catch (error) {
    const detail =
      error instanceof Error && error.stack ? error.stack : String(error);
    process.stderr.write(
      `${detail}\n`,
    );
    return process.exitCode && process.exitCode !== 0 ? process.exitCode : 1;
  }
}
