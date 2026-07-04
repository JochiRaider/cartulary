import {
  collectTargetPlanRows,
  findTargetDescriptor,
} from "../../backend/backend-target-plan.mjs";

export function targetStartStats(root, target, children = []) {
  const childSet = new Set(children);
  const rows = collectTargetPlanRows(root).filter((row) => {
    if (childSet.size > 0) {
      return childSet.has(row.target);
    }
    return row.target === target;
  });
  const descriptor = findTargetDescriptor(target, root);
  const serviceBacked =
    descriptor?.serviceBacked ?? rows.some((row) => row.service_backed);
  const manifestPhases = new Set(
    rows.map((row) => row.manifest_phase).filter(Boolean),
  );
  const rawRows = rows.filter((row) => row.manifest_phase === "").length;
  return {
    serviceBacked,
    expectedPhases: manifestPhases.size + rawRows,
    expectedTests: rows.length,
  };
}
