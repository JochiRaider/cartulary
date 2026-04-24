import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";

import { validateManifest } from "./lib/phase-manifest.mjs";
import { phaseLedgerOutputs, renderPhaseLedger } from "./render-phase-ledger.mjs";

function main() {
  const root = process.cwd();

  for (const { phase, outputPath } of phaseLedgerOutputs) {
    validateManifest(root, phase);
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
