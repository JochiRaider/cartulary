function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

export class PhaseSliceSelectionError extends Error {
  constructor(message) {
    super(message);
    this.name = "PhaseSliceSelectionError";
    this.exitCode = 2;
  }
}

export function parsePhaseRowIDs(value) {
  if (Array.isArray(value)) {
    value = value.join(",");
  }
  const raw = String(value ?? "");
  if (raw.trim() === "") {
    return [];
  }
  const tokens = raw.split(",").map((token) => token.trim());
  if (tokens.some((token) => token === "")) {
    throw new PhaseSliceSelectionError(
      "ROWS contains an empty row id; provide a comma-separated list of non-empty phase row ids",
    );
  }
  const seen = new Set();
  for (const token of tokens) {
    if (seen.has(token)) {
      throw new PhaseSliceSelectionError(
        `ROWS contains duplicate row id after normalization: ${token}`,
      );
    }
    seen.add(token);
  }
  return [...seen].sort(compareStrings);
}

export function phaseSliceSelection({
  phaseNamespace,
  mode,
  requestedRowIDs,
  resolvedRowIDs,
}) {
  const exact = requestedRowIDs.length > 0;
  return {
    mode: exact ? "exact_rows" : "default",
    phase_span:
      phaseNamespace === "frontend" && !exact
        ? "through_phase"
        : "exact_phase",
    dependency_scope: mode === "service_backed" ? "service_backed" : "all",
    completion_scope:
      exact || mode === "service_backed" ? "selected_subset" : "full_phase",
    requested_row_ids: [...requestedRowIDs],
    resolved_row_ids: [...new Set(resolvedRowIDs)].sort(compareStrings),
  };
}
