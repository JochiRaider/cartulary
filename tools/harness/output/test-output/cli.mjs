#!/usr/bin/env node

import { handleGoJSONStream } from "./go-json-stream.mjs";
import { handleRunStart, handleStepStart, handleTargetStart } from "./lifecycle.mjs";
import { handleRunSummary } from "./run-summary.mjs";
import { handleSharedExecution } from "./shared-execution.mjs";
import { handleTargetSummary } from "./target-summary.mjs";
import { handleTimingSpan } from "./timing.mjs";
import { handleGoStep } from "./runners/go.mjs";
import { handlePlaywrightStep } from "./runners/playwright.mjs";
import { handleShellStep } from "./runners/shell.mjs";
import { handleVitestStep } from "./runners/vitest.mjs";

function main() {
  const [command, ...rest] = process.argv.slice(2);
  switch (command) {
    case "shell-step":
      process.exit(handleShellStep());
      break;
    case "go-step":
      process.exit(handleGoStep({ catalogAware: false }));
      break;
    case "go-catalog-step":
      process.exit(handleGoStep({ catalogAware: true }));
      break;
    case "vitest-step":
      process.exit(handleVitestStep({ catalogAware: false }));
      break;
    case "vitest-catalog-step":
      process.exit(handleVitestStep({ catalogAware: true }));
      break;
    case "playwright-step":
      process.exit(handlePlaywrightStep({ catalogAware: false }));
      break;
    case "playwright-catalog-step":
      process.exit(handlePlaywrightStep({ catalogAware: true }));
      break;
    case "go-json-stream":
      process.exit(handleGoJSONStream());
      break;
    case "target-summary":
      process.exit(handleTargetSummary(rest));
      break;
    case "timing-span":
      process.exit(handleTimingSpan());
      break;
    case "shared-execution":
      process.exit(handleSharedExecution(rest));
      break;
    case "run-summary":
      process.exit(handleRunSummary(rest));
      break;
    case "run-start":
      process.exit(handleRunStart(rest));
      break;
    case "step-start":
      process.exit(handleStepStart(rest));
      break;
    case "target-start":
      process.exit(handleTargetStart(rest));
      break;
    default:
      throw new Error(`unknown test-output command ${command}`);
  }
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${message}\n`);
  const publicExitCode = Number(error?.publicExitCode);
  process.exit(Number.isInteger(publicExitCode) && publicExitCode > 0 ? publicExitCode : 1);
}
