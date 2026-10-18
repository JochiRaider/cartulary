/** Resolve the graph owner's explicit assurance dependencies, never owner scope. */
export function policyDependencyRowIDs(owner, policy, rows) {
  if (!Object.hasOwn(owner.policy_units, policy)) throw new Error(`unknown policy operation ${policy}`);
  const declared = owner.policy_units[policy]?.required_rows ?? [];
  const active = new Map(rows.filter((row) => row.status === "active").map((row) => [row.row_id, row]));
  const selected = new Set();
  for (const dependency of declared) {
    if (!active.has(dependency.row_id)) throw new Error(`${policy}: missing active required row ${dependency.row_id}`);
    if (selected.has(dependency.row_id)) throw new Error(`${policy}: duplicate required row ${dependency.row_id}`);
    if (!dependency.rationale?.trim()) throw new Error(`${policy}: required row ${dependency.row_id} needs a rationale`);
    selected.add(dependency.row_id);
  }
  return [...selected].sort();
}

