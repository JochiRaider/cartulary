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
  const owners = new Set(
    rows.map((row) => row.owner_id).filter(Boolean),
  );
  const unownedRows = rows.filter((row) => row.owner_id === "").length;
  return {
    serviceBacked,
    expectedSteps: owners.size + unownedRows,
    expectedTests: rows.length,
  };
}
