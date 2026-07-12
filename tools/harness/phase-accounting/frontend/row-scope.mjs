import { loadFrontendPhaseMap } from "./registry-loader.mjs";
import {
  frontendPhaseNumber,
  frontendRowIDPattern,
} from "./phase-ids.mjs";
import { PhaseSliceSelectionError } from "../phase-row-selector.mjs";

function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

export function uniqueSorted(values) {
  return Array.from(new Set(values.filter(Boolean))).sort(compareStrings);
}

export function parseFrontendRowIDs(value) {
  if (Array.isArray(value)) {
    return uniqueSorted(value.map((item) => String(item).trim()));
  }
  return uniqueSorted(
    String(value ?? "")
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean),
  );
}

export function validateFrontendRowIDs(rowIDs, invalidMessage) {
  for (const rowID of rowIDs) {
    if (!frontendRowIDPattern.test(rowID)) {
      throw new Error(invalidMessage(rowID));
    }
  }
}

export function frontendMapRef(phaseID) {
  return `tools/frontend_phase_maps/fe_p${phaseID.slice("FE-P".length)}_test_map.json`;
}

export function selectedFrontendRowAccountingScope(phase, rowIDs) {
  return {
    mode: "selected_rows",
    invocation_kind: "frontend_phase_slice",
    phase_namespace: "frontend",
    phase,
    selection_policy: "frontend_rows_through_selected_phase",
    selected_row_ids: uniqueSorted(rowIDs),
  };
}

export function rowIsInActiveTargetScope(phase, row) {
  return phase.status === "active" && row.claim_status === "implemented";
}

export function rowsThroughSelectedActiveFrontendPhase(
  root,
  registry,
  selectedPhase,
) {
  const selectedOrder = frontendPhaseNumber(selectedPhase);
  const rows = [];
  for (const entry of registry.phases) {
    const order = frontendPhaseNumber(entry.phase_id);
    if (!Number.isFinite(order) || order > selectedOrder) {
      continue;
    }
    if (entry.status !== "active") {
      continue;
    }
    const { manifest } = loadFrontendPhaseMap(root, entry.phase_id);
    rows.push(...manifest.rows);
  }
  return rows;
}

export function selectedFrontendRows(
  root,
  registry,
  selectedPhase,
  selectedRowIDs,
) {
  const selectedIDSet = new Set(selectedRowIDs);
  const found = new Map();
  const selectedEntry = registry.phases.find(
    (entry) => entry.phase_id === selectedPhase,
  );
  const { manifest } = loadFrontendPhaseMap(root, selectedEntry.phase_id);
  for (const row of manifest.rows) {
    if (!selectedIDSet.has(row.id)) {
      continue;
    }
    if (row.claim_status !== "implemented") {
      throw new PhaseSliceSelectionError(
        `selected frontend row ${row.id} is ${row.claim_status} and is not executable`,
      );
    }
    found.set(row.id, row);
  }
  const missing = selectedRowIDs.filter((rowID) => !found.has(rowID));
  if (missing.length > 0) {
    for (const rowID of missing) {
      for (const entry of registry.phases) {
        if (entry.phase_id === selectedPhase) {
          continue;
        }
        const { manifest: candidate } = loadFrontendPhaseMap(root, entry.phase_id);
        if (candidate.rows.some((row) => row.id === rowID)) {
          throw new PhaseSliceSelectionError(
            `selected frontend row ${rowID} belongs to ${entry.phase_id}, not ${selectedPhase}`,
          );
        }
      }
    }
    throw new PhaseSliceSelectionError(
      `selected frontend row id(s) not found in ${selectedPhase}: ${missing.join(",")}`,
    );
  }
  return selectedRowIDs.map((rowID) => found.get(rowID));
}

export function frontendRowsForAccountingTarget({
  root,
  registry,
  target,
  scope,
}) {
  if (scope.mode === "disabled") {
    return { rows: [], phaseMapRefs: [], explicitScope: true };
  }

  const rows = [];
  const selectedIDs = new Set(scope.selected_row_ids);
  const knownSelectedIDs = new Set();
  const mappedSelectedIDs = new Set();
  const selectedPhaseNumber = frontendPhaseNumber(scope.phase);
  for (const phase of registry.phases) {
    if (scope.mode === "selected_rows") {
      const currentPhaseNumber = frontendPhaseNumber(phase.phase_id);
      if (
        !Number.isFinite(currentPhaseNumber) ||
        currentPhaseNumber > selectedPhaseNumber
      ) {
        continue;
      }
    }

    const { manifest } = loadFrontendPhaseMap(root, phase.phase_id);
    for (const row of manifest.rows) {
      if (scope.mode === "active_target" && !rowIsInActiveTargetScope(phase, row)) {
        continue;
      }
      if (scope.mode === "selected_rows") {
        if (selectedIDs.has(row.id)) {
          knownSelectedIDs.add(row.id);
        } else {
          continue;
        }
      }

      const targetRefs = row.targets.filter(
        (targetRef) => targetRef.target_name === target,
      );
      const closingTargetRefs = targetRefs.filter(
        (targetRef) =>
          targetRef.required_for_closure ||
          targetRef.frontend_row_accounting_required,
      );
      if (closingTargetRefs.length === 0) {
        continue;
      }
      if (scope.mode === "selected_rows") {
        mappedSelectedIDs.add(row.id);
      }
      rows.push({
        phase_id: phase.phase_id,
        phase_status: phase.status,
        row_rollup_state: phase.row_rollup_state,
        row_id: row.id,
        layer: row.layer,
        evidence_class: row.evidence_class,
        claim_status: row.claim_status,
        claim: row.claim,
        blockers: row.blockers,
        required_for_closure: closingTargetRefs.some(
          (targetRef) => targetRef.required_for_closure,
        ),
        scenario_titles: row.scenario_titles,
      });
    }
  }

  if (scope.mode === "selected_rows") {
    const unknown = [...selectedIDs]
      .filter((rowID) => !knownSelectedIDs.has(rowID))
      .sort(compareStrings);
    if (unknown.length > 0) {
      throw new Error(
        `selected frontend row id(s) not found: ${unknown.join(",")}`,
      );
    }
    const unmapped = [...selectedIDs]
      .filter((rowID) => !mappedSelectedIDs.has(rowID))
      .sort(compareStrings);
    if (unmapped.length > 0) {
      throw new Error(
        `selected frontend row id(s) not mapped to ${target}: ${unmapped.join(",")}`,
      );
    }
  }

  return {
    rows,
    phaseMapRefs: uniqueSorted(rows.map((row) => frontendMapRef(row.phase_id))),
    explicitScope: scope.mode !== "active_target",
  };
}

export function frontendTargetHasClosureRows({ root, registry, target }) {
  for (const phase of registry.phases) {
    if (phase.status !== "active") {
      continue;
    }
    const { manifest } = loadFrontendPhaseMap(root, phase.phase_id);
    for (const row of manifest.rows) {
      if (row.claim_status !== "implemented") {
        continue;
      }
      if (
        row.targets.some(
          (targetRef) =>
            targetRef.target_name === target &&
            (targetRef.required_for_closure ||
              targetRef.frontend_row_accounting_required),
        )
      ) {
        return true;
      }
    }
  }
  return false;
}
