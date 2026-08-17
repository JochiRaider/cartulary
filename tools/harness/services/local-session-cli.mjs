#!/usr/bin/env node

import path from "node:path";
import { pathToFileURL } from "node:url";

import {
  createLocalSession,
  assertNoCallerServiceState,
  localSessionStatus,
  resolveLocalSessionFile,
  stopLocalSession,
} from "./local-session.mjs";

const root = path.resolve(import.meta.dirname, "../../..");

function usage() {
  process.stderr.write("usage: local-session-cli.mjs up|status|down\n");
}

export function runLocalSessionCLI(args = process.argv.slice(2), environment = process.env) {
  if (args.length !== 1 || !new Set(["up", "status", "down"]).has(args[0])) {
    usage();
    return 2;
  }
  assertNoCallerServiceState(environment);
  const binary = path.resolve(
    String(environment.TEST_SERVICES_BIN || environment.CARTULARY_TEST_SERVICES_BIN || path.join(root, "tmp/toolbin/cartulary-test-services")),
  );
  const sessionFile = resolveLocalSessionFile(environment);
  const operation = args[0];
  const status = operation === "up"
    ? createLocalSession({ root, binary, sessionFile, environment })
    : operation === "down"
      ? stopLocalSession({ root, binary, sessionFile })
      : localSessionStatus({ root, binary, sessionFile });
  process.stdout.write(`${JSON.stringify(status)}\n`);
  return new Set(["expired", "invalid", "stale"]).has(status.state) ? 1 : 0;
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  try {
    process.exitCode = runLocalSessionCLI();
  } catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
    process.exitCode = 1;
  }
}
