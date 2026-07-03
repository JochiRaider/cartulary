import {
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import os from "node:os";
import path from "node:path";
import { spawnSync } from "node:child_process";

import { validateManifest } from "../../../tools/harness/planning/phase-manifest.mjs";
import { phaseLedgerOutputs, renderPhaseLedger } from "./render-phase-ledger.mjs";

const diffLineLimit = 200;

function renderTempLedger(root, tempRoot, phase, outputPath) {
  if (!phase.startsWith("FE-P")) {
    validateManifest(root, phase, { allowPlanned: true });
  }
  const rendered = renderPhaseLedger(root, phase);
  const tempOutputPath = path.join(tempRoot, outputPath);
  mkdirSync(path.dirname(tempOutputPath), { recursive: true });
  writeFileSync(tempOutputPath, rendered, "utf8");
  return tempOutputPath;
}

function diffExcerpt(committedPath, renderedPath, outputPath) {
  const result = spawnSync(
    "diff",
    [
      "-u",
      "--label",
      outputPath,
      "--label",
      `rendered ${outputPath}`,
      committedPath,
      renderedPath,
    ],
    { encoding: "utf8" },
  );
  const output = `${result.stdout ?? ""}${result.stderr ?? ""}`.trimEnd();
  if (output === "") {
    return "(diff unavailable)";
  }
  return output.split("\n").slice(0, diffLineLimit).join("\n");
}

function main() {
  const root = process.cwd();
  const tempRoot = mkdtempSync(path.join(os.tmpdir(), "cartulary-phase-ledgers-"));
  const drifts = [];

  try {
    for (const { phase, outputPath } of phaseLedgerOutputs(root)) {
      const committedPath = path.join(root, outputPath);
      const renderedPath = renderTempLedger(root, tempRoot, phase, outputPath);
      const committed = readFileSync(committedPath, "utf8");
      const rendered = readFileSync(renderedPath, "utf8");

      if (committed !== rendered) {
        drifts.push({
          phase,
          outputPath,
          excerpt: diffExcerpt(committedPath, renderedPath, outputPath),
        });
      }
    }
  } finally {
    rmSync(tempRoot, { recursive: true, force: true });
  }

  if (drifts.length > 0) {
    for (const drift of drifts) {
      const source =
        drift.phase.startsWith("FE-P")
          ? `tools/frontend_phase_maps/fe_p${drift.phase.slice("FE-P".length).toLowerCase()}_test_map.json`
          : `tools/${drift.phase}_test_map.json`;
      process.stderr.write(
        `${drift.phase} coverage ledger drift: regenerate ${drift.outputPath} from ${source}\n`,
      );
      process.stderr.write(`${drift.excerpt}\n`);
    }
    process.exit(1);
  }

  console.log("phase coverage ledgers verified");
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
