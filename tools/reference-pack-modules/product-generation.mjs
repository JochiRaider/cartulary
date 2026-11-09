// One declared sequence for ordinary generation and its isolated qualification.
import { readFileSync, realpathSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { pathToFileURL } from "node:url";

export const profilePath = "tools/reference-pack-modules/product-generation.json";
export function generationProfile(file = profilePath) {
  const profile = JSON.parse(readFileSync(file));
  if (profile.schema_id !== "cartulary.reference_pack_product_generation.v1" || !profile.steps?.length) throw new Error("invalid product generation profile");
  for (const step of profile.steps) {
    if (!["node", "go"].includes(step.runtime) || !Array.isArray(step.arguments) || !step.arguments.length || step.arguments.some(value => typeof value !== "string" || !value)) throw new Error("invalid generation command");
  }
  return profile;
}
export function generationCommands(profile, go) {
  return profile.steps.map(step => [step.runtime === "node" ? process.execPath : go, ...step.arguments]);
}
if (process.argv[1] && import.meta.url === pathToFileURL(realpathSync(process.argv[1])).href) {
  const go = process.argv[2];
  if (!go) throw new Error("explicit Go command required");
  for (const [command, ...args] of generationCommands(generationProfile(), go)) {
    const result = spawnSync(command, args, {stdio: "inherit", env: process.env});
    if (result.error || result.signal || result.status !== 0) throw new Error(`product generation failed: ${command} ${args.join(" ")}`);
  }
}
