import { readFileSync } from "node:fs";
import path from "node:path";
import { repoRoot } from "./policy.mjs";

const topology = JSON.parse(readFileSync(path.join(repoRoot, "tools/execution_topology_manifest.json"), "utf8"));
export const workClaims = Object.freeze(Object.fromEntries(["raster", "report", "diagnostic"].map((kind) => {
  const matches = topology.resource_profiles.filter((entry) => entry.id === `ui_review_${kind}`);
  if (matches.length !== 1) throw new Error("missing review work policy");
  const claims = matches[0].resource_claims;
  if (Object.keys(claims).sort().join(",") !== "cpu,io,memory_mb,process" || Object.values(claims).some((value) => !Number.isSafeInteger(value) || value < 1)) throw new Error("invalid review work claims");
  return [kind, Object.freeze({ ...claims })];
})));
