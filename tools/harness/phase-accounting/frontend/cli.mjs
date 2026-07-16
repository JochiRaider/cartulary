import {
  frontendExactTitleGrepForTarget,
  frontendPlaywrightGrepForTarget,
} from "./scenario-grep.mjs";
import { validateFrontendPhaseArtifacts } from "./phase-artifacts.mjs";

const grepUsage =
  "usage: frontend-phase-manifest.mjs playwright-grep|title-grep <target> [layer] [--row-ids <ids>] [--runtime-profile-id <id>]";

export function runFrontendPhaseManifestCLI(
  argv = process.argv.slice(2),
  root = process.cwd(),
) {
  const [command = "", ...rest] = argv;
  if (command === "validate") {
    validateFrontendPhaseArtifacts(root);
    console.log("frontend phase artifacts verified");
    return;
  }
  if (command === "playwright-grep" || command === "title-grep") {
    const args = [...rest];
    const target = args.shift() ?? "";
    let layer = "";
    let rowIDs = null;
    let runtimeProfileID = null;
    if (args[0] && !args[0].startsWith("--")) {
      layer = args.shift();
    }
    while (args.length > 0) {
      const arg = args.shift();
      if (arg === "--row-ids") {
        rowIDs = args.shift() ?? "";
        continue;
      }
      if (arg === "--runtime-profile-id") {
        runtimeProfileID = args.shift() ?? "";
        continue;
      }
      console.error(grepUsage);
      process.exit(2);
    }
    if (!target) {
      console.error(grepUsage);
      process.exit(2);
    }
    console.log(
      command === "title-grep"
        ? frontendExactTitleGrepForTarget(root, target, layer, {
            rowIDs,
            runtimeProfileID,
          })
        : frontendPlaywrightGrepForTarget(root, target, layer, {
            rowIDs,
            runtimeProfileID,
          }),
    );
    return;
  }
  console.error(
    "usage: frontend-phase-manifest.mjs validate|playwright-grep|title-grep <target> [layer]",
  );
  process.exit(2);
}
