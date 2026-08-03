import { readFileSync } from "node:fs";
import path from "node:path";

import { loadTestCatalog, targetForCatalogRow } from "../../test-catalog/index.mjs";

export function targetStartStats(root, target, children = []) {
  const taskSurface = JSON.parse(
    readFileSync(path.join(root, "tools/task_surface_owner.json"), "utf8"),
  );
  const commandTargetByID = new Map(
    taskSurface.targets
      .filter((entry) => entry.command_id)
      .map((entry) => [entry.command_id, entry.name]),
  );
  const selectedTargets = new Set(children.length > 0 ? children : [target]);
  const rows = loadTestCatalog(root).rows.filter((row) =>
    selectedTargets.has(targetForCatalogRow(row, { commandTargetByID })),
  );
  return {
    serviceBacked: rows.some((row) => row.fixture_capability !== "none"),
    expectedSteps: new Set(rows.map((row) => row.owner_id)).size,
    expectedTests: rows.length,
  };
}
