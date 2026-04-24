import { validateManifest } from "./lib/phase-manifest.mjs";

const phase = process.argv[2];
if (!phase) {
  throw new Error("usage: check-phase-map.mjs <phase>");
}

validateManifest(process.cwd(), phase);

console.log(`${phase} traceability map verified`);
