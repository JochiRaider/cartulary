#!/usr/bin/env node
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  buildPhaseSlicePlan,
  printablePlan,
} from "./phase-slice-plan.mjs";
import {
  buildFrontendPhaseSlicePlan,
  printableFrontendPlan,
} from "./frontend-readiness.mjs";
import { PhaseSliceSelectionError } from "./phase-row-selector.mjs";
import {
  createPhaseSliceRunnerContext,
  runPhaseSliceExecution,
  writePhaseSlicePlan,
} from "../scheduler/phase-slice-execution.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "../../..");
const phaseSliceCliPath = path.join(scriptDir, "phase-slice-cli.mjs");
const validModes = new Set(["phase", "service-backed"]);

function usage() {
  process.stderr.write(
    "usage: run-phase-slice.mjs --phase <phaseN|FE-PN> --mode <phase|service-backed> [--phase-namespace <base|frontend>] [--rows <row-id,...>] [--inside-service-wrapper]\n",
  );
  process.exit(2);
}

function parseArgs(argv) {
  const options = { phase: "", mode: "", phaseNamespace: "base", rows: "", insideServiceWrapper: false, json: false };
  for (let index = 0; index < argv.length; index += 1) {
    const arg = argv[index];
    if (arg === "--phase") {
      options.phase = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--mode") {
      options.mode = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--phase-namespace") {
      options.phaseNamespace = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--rows") {
      options.rows = argv[index + 1] ?? "";
      index += 1;
      continue;
    }
    if (arg === "--inside-service-wrapper") {
      options.insideServiceWrapper = true;
      continue;
    }
    if (arg === "--json") {
      options.json = true;
      continue;
    }
    usage();
  }
  if (!options.phase || !validModes.has(options.mode) || !["base", "frontend"].includes(options.phaseNamespace)) {
    usage();
  }
  return options;
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const context = createPhaseSliceRunnerContext();
  const mode = options.mode === "service-backed" ? "service_backed" : "phase";
  if (options.phaseNamespace === "frontend") {
    const plan = buildFrontendPhaseSlicePlan(options.phase, {
      mode,
      root: repoRoot,
      rowIDs: options.rows,
    });

    if (options.json || process.env.JSON === "1") {
      await writePhaseSlicePlan(plan, context);
      process.stdout.write(`${JSON.stringify(printableFrontendPlan(plan), null, 2)}\n`);
      return 0;
    }

    return await runPhaseSliceExecution(plan, context, options, {
      phaseSliceCliPath,
    });
  }
  if (options.phase.startsWith("FE-P")) {
    throw new PhaseSliceSelectionError(
      "frontend phases require --phase-namespace frontend",
    );
  }
  const plan = buildPhaseSlicePlan(options.phase, {
    mode,
    root: repoRoot,
    rowIDs: options.rows,
  });

  if (options.json || process.env.JSON === "1") {
    await writePhaseSlicePlan(plan, context);
    process.stdout.write(`${JSON.stringify(printablePlan(plan), null, 2)}\n`);
    return 0;
  }

  return await runPhaseSliceExecution(plan, context, options, {
    phaseSliceCliPath,
  });
}

main()
  .then((status) => {
    process.exitCode = status;
  })
  .catch((error) => {
    const message = error instanceof Error ? error.message : String(error);
    const exitCode = Number.isInteger(error?.exitCode) ? error.exitCode : 1;
    process.stderr.write(exitCode === 2 ? `${message}\n` : `phase slice failed: ${message}\n`);
    process.exitCode = exitCode;
  });
