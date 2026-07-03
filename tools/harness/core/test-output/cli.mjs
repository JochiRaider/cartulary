#!/usr/bin/env node

import { handleGoJSONStream } from "./go-json-stream.mjs";
import { handleRunStart, handleStepStart, handleTargetStart } from "./lifecycle.mjs";
import { handleRunSummary } from "./run-summary.mjs";
import { handleSharedExecution } from "./shared-execution.mjs";
import { handleTargetSummary } from "./target-summary.mjs";
import { handleTimingSpan } from "./timing.mjs";
import { handleGoPhase } from "./runners/go.mjs";
import { handlePlaywrightPhase } from "./runners/playwright.mjs";
import { handleShellPhase } from "./runners/shell.mjs";
import { handleVitestPhase } from "./runners/vitest.mjs";

function main() {
  const [command, ...rest] = process.argv.slice(2);
  switch (command) {
    case "shell-phase":
      process.exit(handleShellPhase());
      break;
    case "go-phase":
      process.exit(handleGoPhase({ manifestAware: false }));
      break;
    case "go-manifest-phase":
      process.exit(handleGoPhase({ manifestAware: true }));
      break;
    case "vitest-phase":
      process.exit(handleVitestPhase({ manifestAware: false }));
      break;
    case "vitest-manifest-phase":
      process.exit(handleVitestPhase({ manifestAware: true }));
      break;
    case "playwright-phase":
      process.exit(handlePlaywrightPhase({ manifestAware: false }));
      break;
    case "playwright-manifest-phase":
      process.exit(handlePlaywrightPhase({ manifestAware: true }));
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
  process.exit(1);
}
