#!/usr/bin/env node
import { workspaceLayout } from "../../workspace_layout.generated.mjs";

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

export async function runLocalSessionCLI(args = process.argv.slice(2), environment = process.env, signal) {
  if (args.length !== 1 || !new Set(["up", "status", "down"]).has(args[0])) {
    usage();
    return 2;
  }
  assertNoCallerServiceState(environment);
  const binary = path.resolve(
    String(environment.TEST_SERVICES_BIN || environment.CARTULARY_TEST_SERVICES_BIN || path.join(root, `${workspaceLayout.toolbin}/cartulary-test-services`)),
  );
  const sessionFile = resolveLocalSessionFile(environment);
  const operation = args[0];
  const status = operation === "up"
    ? await createLocalSession({ root, binary, sessionFile, environment, signal })
    : operation === "down"
      ? await stopLocalSession({ root, binary, sessionFile, signal })
      : localSessionStatus({ root, binary, sessionFile });
  process.stdout.write(`${JSON.stringify(status)}\n`);
  return new Set(["expired", "invalid", "stale"]).has(status.state) ? 1 : 0;
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  const cancellation = new AbortController();
  let interrupted;
  const interrupt = (name) => { interrupted = name === "SIGINT" ? 130 : 143; cancellation.abort(new Error("local service lifecycle interrupted")); };
  process.on("SIGINT", interrupt);
  process.on("SIGTERM", interrupt);
  try {
    const code = await runLocalSessionCLI(process.argv.slice(2), process.env, cancellation.signal);
    process.exitCode = interrupted ?? code;
  } catch (error) {
    process.stderr.write(`${error instanceof Error ? error.message : String(error)}\n`);
    process.exitCode = interrupted ?? 1;
  } finally { process.off("SIGINT", interrupt); process.off("SIGTERM", interrupt); }
}
