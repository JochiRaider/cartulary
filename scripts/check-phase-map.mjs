import { readFileSync } from "node:fs";
import path from "node:path";

import { validateManifest } from "./lib/phase-manifest.mjs";
import { renderPhaseLedger } from "./render-phase-ledger.mjs";

const phase = process.argv[2];
if (!phase) {
  throw new Error("usage: check-phase-map.mjs <phase>");
}

validateManifest(process.cwd(), phase);

const ledgerFilenames = {
  phase0: "phase0_coverage_ledger.md",
  phase2: "phase2_coverage_ledger.md",
  phase3: "phase3_coverage_ledger.md",
  phase4: "phase4_coverage_ledger.md",
};

if (phase in ledgerFilenames) {
  const ledgerPath = path.join(
    process.cwd(),
    "docs",
    "testing",
    ledgerFilenames[phase],
  );
  const committedLedger = readFileSync(ledgerPath, "utf8");
  const renderedLedger = renderPhaseLedger(process.cwd(), phase);
  if (committedLedger !== renderedLedger) {
    throw new Error(
      `${phase} coverage ledger drift: regenerate docs/testing/${ledgerFilenames[phase]} from tools/${phase}_test_map.json`,
    );
  }
}

console.log(`${phase} traceability map verified`);
