import { frontendPhaseNamespace, loadFrontendPhaseMap } from "./registry.mjs";

function targetDisplayName(target) {
  return target.target_name ? `make ${target.target_name}` : String(target);
}

function ownerRefDisplay(ownerRef) {
  if (ownerRef.resolution_status === "resolved") {
    return `${ownerRef.path}#${ownerRef.section_ref}`;
  }
  return `${ownerRef.path}#${ownerRef.section_ref} (${ownerRef.resolution_status})`;
}

function claimStatement(row) {
  return row.claim.statement;
}

function outOfScopeText(row) {
  return row.out_of_scope.join(" ");
}

export function frontendLedgerOutputPath(entry) {
  return entry.ledger_path;
}

export function renderFrontendPhaseLedger(root, phaseID) {
  const { registryEntry, manifest } = loadFrontendPhaseMap(root, phaseID);
  const lines = [
    `# ${phaseID} Frontend Coverage Ledger`,
    "",
    `This ledger is generated from \`${registryEntry.manifest_path}\`. Update the frontend phase map first, then regenerate this file.`,
    "",
    `- Namespace: \`${frontendPhaseNamespace}\``,
    `- Status: \`${registryEntry.status}\``,
    `- Row rollup state: \`${registryEntry.row_rollup_state}\``,
    `- Owner refs: ${registryEntry.owner_refs.map((owner) => `\`${ownerRefDisplay(owner)}\``).join(", ")}`,
    `- Depends on: ${
      registryEntry.depends_on.length === 0
        ? "`none`"
        : registryEntry.depends_on.map((phase) => `\`${phase}\``).join(", ")
    }`,
    "- Authority: frontend phase maps are implementation-readiness inputs. This rendered ledger does not own product behavior.",
    "",
    "## Rows",
    "",
    "| Row | Layer | Evidence class | Claim status | Targets | Owner refs | Core REQs | Core ACs | Support/design ACs | Scenario titles | Claim | Out of scope |",
    "| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |",
  ];

  for (const row of manifest.rows) {
    lines.push(
      `| \`${row.id}\` | \`${row.layer}\` | \`${row.evidence_class}\` | \`${row.claim_status}\` | ${row.targets.map((target) => `\`${targetDisplayName(target)}\``).join("<br>")} | ${row.owner_refs.map((owner) => `\`${ownerRefDisplay(owner)}\``).join("<br>")} | ${row.core_req_ids.map((id) => `\`${id}\``).join(", ") || "`none`"} | ${row.core_ac_ids.map((id) => `\`${id}\``).join(", ") || "`none`"} | ${row.support_or_design_ac_ids.map((id) => `\`${id}\``).join(", ") || "`none`"} | ${row.scenario_titles.map((title) => `\`${title}\``).join("<br>") || "`none`"} | ${claimStatement(row)} | ${outOfScopeText(row)} |`,
    );
  }

  return `${lines.join("\n")}\n`;
}

