#!/usr/bin/env node

import { readFileSync } from "node:fs";

import {
  HarnessConfigError,
  compactJSONString,
  generateTestRouteToken,
  preflightPublicTarget,
  redactString,
  redactValue,
  runCleanup,
  validateSchema,
} from "./lib/harness-contract.mjs";

function usage() {
  process.stderr.write(
    "usage: harness-contract.mjs preflight <target> | cleanup <clean|distclean> <path...> | validate-schema <schema-id> <json-file> | redact | generate-test-route-token\n",
  );
}

async function main(argv) {
  const [command, ...args] = argv;
  if (command === "preflight") {
    const [target, ...extra] = args;
    if (!target || extra.length > 0) {
      usage();
      return 2;
    }
    preflightPublicTarget(target);
    return 0;
  }
  if (command === "cleanup") {
    const [scope, ...paths] = args;
    if (!["clean", "distclean"].includes(scope)) {
      usage();
      return 2;
    }
    runCleanup({ scope, candidates: paths });
    return 0;
  }
  if (command === "validate-schema") {
    const [schemaID, file, ...extra] = args;
    if (!schemaID || !file || extra.length > 0) {
      usage();
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
  usage();
  return 2;
}

try {
  process.exitCode = await main(process.argv.slice(2));
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
