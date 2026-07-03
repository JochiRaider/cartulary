#!/usr/bin/env node

import { readFileSync } from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";

import {
  HarnessConfigError,
  compactJSONString,
  generateTestRouteToken,
  preflightPublicTarget,
  redactString,
  redactValue,
  resolveArtifactIdentityForTarget,
  runCleanup,
  validateSchema,
} from "./harness-contract.mjs";

function usage(programName) {
  process.stderr.write(
    `usage: ${programName} preflight <target> | retained-artifact-env <target> | cleanup <clean|distclean> <path...> | validate-schema <schema-id> <json-file> | redact | generate-test-route-token\n`,
  );
}

async function main(argv, programName) {
  const [command, ...args] = argv;
  if (command === "preflight") {
    const [target, ...extra] = args;
    if (!target || extra.length > 0) {
      usage(programName);
      return 2;
    }
    preflightPublicTarget(target);
    return 0;
  }
  if (command === "retained-artifact-env") {
    const [target, ...extra] = args;
    if (!target || extra.length > 0) {
      usage(programName);
      return 2;
    }
    const identity = resolveArtifactIdentityForTarget(target);
    process.stdout.write(`${identity.result_root}\n${identity.run_id}\n`);
    return 0;
  }
  if (command === "cleanup") {
    const [scope, ...paths] = args;
    if (!["clean", "distclean"].includes(scope)) {
      usage(programName);
      return 2;
    }
    runCleanup({ scope, candidates: paths });
    return 0;
  }
  if (command === "validate-schema") {
    const [schemaID, file, ...extra] = args;
    if (!schemaID || !file || extra.length > 0) {
      usage(programName);
      return 2;
    }
    const value = JSON.parse(readFileSync(file, "utf8"));
    await validateSchema(schemaID, value);
    return 0;
  }
  if (command === "redact") {
    process.stdin.setEncoding("utf8");
    let input = "";
    for await (const chunk of process.stdin) {
      input += chunk;
    }
    try {
      process.stdout.write(compactJSONString(redactValue(JSON.parse(input))));
    } catch {
      process.stdout.write(redactString(input));
    }
    return 0;
  }
  if (command === "generate-test-route-token") {
    process.stdout.write(`${generateTestRouteToken()}\n`);
    return 0;
  }
  usage(programName);
  return 2;
}

export async function runHarnessContractCLI(
  argv = process.argv.slice(2),
  { programName = path.basename(process.argv[1] ?? "harness-contract-cli.mjs") } = {},
) {
  try {
    process.exitCode = await main(argv, programName);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    if (error instanceof HarnessConfigError) {
      process.stderr.write(
        `[FAIL] failure_class=${error.failure_class} failure_reason=${error.failure_reason} exit_code=${error.exit_code} ${message}\n`,
      );
      process.exitCode = error.exit_code;
    } else {
      process.stderr.write(`${message}\n`);
      process.exitCode = 1;
    }
  }
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
  await runHarnessContractCLI();
}
