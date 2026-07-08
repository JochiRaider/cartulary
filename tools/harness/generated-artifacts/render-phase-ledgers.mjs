import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";

import { validateManifest } from "../phase-accounting/index.mjs";
import { phaseLedgerOutputs, renderPhaseLedger } from "./render-phase-ledger.mjs";

function main() {
  const root = process.cwd();

  for (const { phase, outputPath } of phaseLedgerOutputs(root)) {
    if (!phase.startsWith("FE-P")) {
      validateManifest(root, phase, { allowPlanned: true });
    }
    const rendered = renderPhaseLedger(root, phase);
    const absoluteOutputPath = path.join(root, outputPath);
    mkdirSync(path.dirname(absoluteOutputPath), { recursive: true });
    writeFileSync(absoluteOutputPath, rendered, "utf8");
    console.log(`rendered ${outputPath}`);
  }
}

try {
  main();
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
