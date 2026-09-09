import { terminalCommonJob } from "../services/commonJobContract";
import type { ImportFailure } from "../services/importClient";
import type {
  DiscoveredImportPreview,
  DiscoveredImportUnit,
  ImportJobResource,
  ImportSessionResource,
} from "../services/importContractAdapter";

import { workbookImportTargets } from "../services/importTargetContractAdapter";
import type { ImportWriteAttempt } from "./importRequests";
import {
  overlappingImportUnits,
  type WorkbookMappingDraft,
} from "./workbookImportMapping";

export type ImportUnitState = {
  readonly unit: DiscoveredImportUnit;
  readonly preview: DiscoveredImportPreview | null;
  readonly previewLoading: boolean;
  readonly previewFailure: ImportFailure | null;
  readonly draft: WorkbookMappingDraft | null;
  readonly failure: ImportFailure | null;
};
export type ImportOperation = {
  readonly attempt: ImportWriteAttempt;
  readonly phase: "pending" | "uncertain" | "rejected";
  readonly failure: ImportFailure | null;
};
export type WorkbookImportState = {
  readonly revision: number;
  readonly access: "active" | "paused" | "unavailable";
  readonly canWrite: boolean;
  readonly file: File | null;
  readonly session: ImportSessionResource | null;
  readonly units: readonly ImportUnitState[];
  readonly operation: ImportOperation | null;
  readonly job: ImportJobResource | null;
  readonly jobPurpose: "discovery" | "apply" | null;
  readonly jobCurrent: boolean;
  readonly observing: boolean;
  readonly observationFailure: ImportFailure | null;
  readonly cancellation: ImportOperation | null;
  readonly loading: boolean;
  readonly loadFailure: ImportFailure | null;
  readonly actionPending: boolean;
  readonly message: string;
};
export function initialWorkbookImportState(): WorkbookImportState {
  return {
    revision: 0,
    access: "unavailable",
    canWrite: false,
    file: null,
    session: null,
    units: [],
    operation: null,
    job: null,
    jobPurpose: null,
    jobCurrent: false,
    observing: false,
    observationFailure: null,
    cancellation: null,
    loading: false,
    loadFailure: null,
    actionPending: false,
    message: "Choose a CSV or XLSX workbook.",
  };
}
export const terminalImportSession = (session: ImportSessionResource) =>
  ["applied", "partially_applied", "failed", "canceled"].includes(
    session.session_status,
  );
export function importApplyBlocker(state: WorkbookImportState): string | null {
  if (!state.canWrite)
    return "Current write access and an open incident are required.";
  if (
    state.operation?.phase === "pending" ||
    state.operation?.phase === "uncertain" ||
    state.actionPending ||
    state.loading
  )
    return "Resolve the current operation before applying.";
  if (state.loadFailure) return "Refresh the session before applying.";
  if (
    !state.session ||
    terminalImportSession(state.session) ||
    state.session.session_status === "applying" ||
    (state.job && !terminalCommonJob(state.job))
  )
    return "This session is not available for a new apply.";
  const ids = state.session.selected_unit_ids;
  if (!ids.length) return "Select at least one approved unit.";
  const selected = state.units.filter((u) =>
    ids.includes(u.unit.import_unit_id),
  );
  if (
    selected.length !== ids.length ||
    selected.some(
      (u) =>
        u.unit.unit_status !== "ready" ||
        !u.unit.approved_mapping ||
        u.draft?.dirty,
    )
  )
    return "Every selected unit needs its current approved mapping and ready state.";
  if (overlappingImportUnits(selected.map((u) => u.unit)).length)
    return "Selected units overlap. Skip an overlapping unit before applying.";
  return null;
}
export function importOutcomeViews(
  state: WorkbookImportState,
): readonly { readonly id: string; readonly title: string }[] {
  if (
    !state.session ||
    !terminalImportSession(state.session) ||
    !state.jobCurrent ||
    state.loadFailure
  )
    return [];
  const ids = new Set<string>();
  for (const { unit } of state.units)
    if (
      unit.unit_status === "applied" &&
      unit.approved_mapping &&
      "target_view_schema_id" in unit.approved_mapping
    )
      ids.add(unit.approved_mapping.target_view_schema_id);
  return workbookImportTargets
    .filter((t) => ids.has(t.contract.viewSchemaId))
    .map((t) => ({ id: t.contract.viewSchemaId, title: t.contract.title }));
}
