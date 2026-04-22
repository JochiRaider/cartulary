import { readFileSync } from "node:fs";
import path from "node:path";

import { validateManifest } from "./lib/phase-manifest.mjs";
import { renderPhaseLedger } from "./render-phase-ledger.mjs";

const phase = process.argv[2];
if (!phase) {
  throw new Error("usage: check-phase-map.mjs <phase>");
}

validateManifest(process.cwd(), phase);

if (phase === "phase3") {
  const ledgerPath = path.join(
    process.cwd(),
    "docs",
    "testing",
    "phase3_coverage_ledger.md",
  );
  const committedLedger = readFileSync(ledgerPath, "utf8");
  const renderedLedger = renderPhaseLedger(process.cwd(), phase);
  if (committedLedger !== renderedLedger) {
    throw new Error(
      `phase3 coverage ledger drift: regenerate docs/testing/phase3_coverage_ledger.md from tools/phase3_test_map.json`,
    );
  }
}

console.log(`${phase} traceability map verified`);
