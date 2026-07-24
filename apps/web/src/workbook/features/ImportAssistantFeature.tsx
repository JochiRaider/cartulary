import { workbookImportAssistantTestId } from "@cartulary/ui-contracts";
import {
  listViewContracts,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import { useEffect, useMemo, useRef, useState } from "react";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import {
  applyWorkbookImport,
  approveWorkbookImportMapping,
  cancelImportJob,
  type DiscoveredImportColumn,
  type ImportJobResource,
  importProfileId,
  importRouteFamily,
  setWorkbookImportUnitSelection,
  uploadAndDiscoverWorkbookImport,
  type WorkbookImportDiscovery,
  type WorkbookImportUnitDiscovery,
  type WorkbookSourceColumnMapping,
} from "../../shared/importCoordinator";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";

type ImportAssistantFeatureProps = {
  readonly apiBase: string | undefined;
  readonly availability: ExtensionAvailabilityController;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly incidentId: string;
  readonly onNavigateToView: (viewSchemaId: string) => void;
};

type UnitDraft = {
  readonly targetViewSchemaId: string;
  readonly fieldsByOrdinal: Readonly<Record<number, string>>;
  readonly approved: boolean;
  readonly selected: boolean;
  readonly error: string | null;
};

const importableViewSchemaIds = new Set([
  "cartulary.view.timeline.v2",
  "cartulary.view.hosts.v1",
  "cartulary.view.identities.v1",
  "cartulary.view.indicators.v1",
  "cartulary.view.evidence.v1",
  "cartulary.view.notes.v1",
  "cartulary.view.assessments.v1",
  "cartulary.view.task_requests.v1",
  "cartulary.view.decisions.v1",
  "cartulary.view.parties.v1",
  "cartulary.view.comm_log.v1",
  "cartulary.view.handoff.v1",
  "cartulary.view.status_review.v1",
  "cartulary.view.lesson.v1",
]);

const importTargets = listViewContracts().filter((contract) =>
  importableViewSchemaIds.has(contract.viewSchemaId),
);

export function ImportAssistantFeature({
  apiBase,
  availability,
  currentIncidentRole,
  incidentId,
  onNavigateToView,
}: ImportAssistantFeatureProps) {
  const operationRef = useRef(0);
  const [file, setFile] = useState<File | null>(null);
  const [discovery, setDiscovery] = useState<WorkbookImportDiscovery | null>(
    null,
  );
  const [drafts, setDrafts] = useState<Readonly<Record<string, UnitDraft>>>({});
  const [message, setMessage] = useState("Choose a CSV or XLSX workbook.");
  const [busy, setBusy] = useState(false);
  const [currentJob, setCurrentJob] = useState<ImportJobResource | null>(null);
  const [resultViews, setResultViews] = useState<readonly string[]>([]);
  const canManage =
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";
  const available = availability.isRouteAvailable(
    importProfileId,
    importRouteFamily,
  );

  useEffect(() => {
    return () => {
      operationRef.current += 1;
    };
  }, []);

  useEffect(() => {
    if (available) {
      return;
    }
    operationRef.current += 1;
    setDiscovery(null);
    setDrafts({});
    setCurrentJob(null);
    setBusy(false);
    setMessage("The Import profile is no longer available.");
  }, [available]);

  const selectedUnitIds = useMemo(
    () =>
      Object.entries(drafts)
        .filter(([, draft]) => draft.approved && draft.selected)
        .map(([unitId]) => unitId),
    [drafts],
  );
  const overlappingUnits = useMemo(
    () =>
      discovery === null
        ? []
        : findOverlappingUnitIds(
            discovery.units.filter(({ unit }) =>
              selectedUnitIds.includes(unit.import_unit_id),
            ),
          ),
    [discovery, selectedUnitIds],
  );

  async function discover() {
    if (file === null || !canManage || !available) {
      return;
    }
    const operation = operationRef.current + 1;
    operationRef.current = operation;
    setBusy(true);
    setDiscovery(null);
    setDrafts({});
    setCurrentJob(null);
    setResultViews([]);
    try {
      const next = await uploadAndDiscoverWorkbookImport({
        availability,
        apiBase,
        incidentId,
        file,
        transactionPrefix: "workbook-import-assistant",
        onProgress: setMessage,
      });
      if (operation !== operationRef.current) {
        return;
      }
      const initialDrafts: Record<string, UnitDraft> = {};
      for (const item of next.units) {
        const target = suggestedTarget(item);
        initialDrafts[item.unit.import_unit_id] = {
          targetViewSchemaId: target.viewSchemaId,
          fieldsByOrdinal: suggestFields(item.preview.columns, target),
          approved: false,
          selected: false,
          error: null,
        };
      }
      setDiscovery(next);
      setDrafts(initialDrafts);
      setMessage(
        `Discovered ${next.units.length} import unit${next.units.length === 1 ? "" : "s"}. Review every mapping, then select or skip each unit.`,
      );
    } catch (error) {
      if (operation === operationRef.current) {
        setMessage(importErrorMessage(error));
      }
    } finally {
      if (operation === operationRef.current) {
        setBusy(false);
      }
    }
  }

  function updateDraft(unitId: string, change: Partial<UnitDraft>) {
    setDrafts((current) => {
      const draft = current[unitId];
      if (draft === undefined) {
        return current;
      }
      return {
        ...current,
        [unitId]: {
          ...draft,
          ...change,
          approved:
            change.targetViewSchemaId !== undefined ||
            change.fieldsByOrdinal !== undefined
              ? false
              : (change.approved ?? draft.approved),
        },
      };
    });
  }

  async function approveAndSelect(item: WorkbookImportUnitDiscovery) {
    const draft = drafts[item.unit.import_unit_id];
    if (discovery === null || draft === undefined || busy) {
      return;
    }
    const target = importTargets.find(
      (candidate) => candidate.viewSchemaId === draft.targetViewSchemaId,
    );
    if (target === undefined) {
      updateDraft(item.unit.import_unit_id, {
        error: "Choose a supported target view.",
      });
      return;
    }
    const sourceColumns = mappingColumns(
      item.preview.columns,
      target,
      draft.fieldsByOrdinal,
    );
    const hasUnmapped = sourceColumns.some(
      (column) => column.field_key === null,
    );
    const unknownColumnPolicy =
      target.viewSchemaId === "cartulary.view.timeline.v2"
        ? "preserve_raw_capture"
        : "reject_if_unmapped";
    if (hasUnmapped && unknownColumnPolicy === "reject_if_unmapped") {
      updateDraft(item.unit.import_unit_id, {
        error: "Map every source column for this target, or choose Timeline.",
      });
      return;
    }
    setBusy(true);
    setMessage("Approving mapping and selecting unit.");
    try {
      await approveWorkbookImportMapping({
        availability,
        apiBase,
        sessionId: discovery.session.import_session_id,
        discovery: item,
        targetViewSchemaId: target.viewSchemaId,
        unknownColumnPolicy,
        sourceColumns,
        transactionPrefix: `workbook-import-${item.unit.import_unit_id}`,
      });
      await setWorkbookImportUnitSelection({
        availability,
        apiBase,
        sessionId: discovery.session.import_session_id,
        unitId: item.unit.import_unit_id,
        selected: true,
        transactionPrefix: `workbook-import-${item.unit.import_unit_id}`,
      });
      updateDraft(item.unit.import_unit_id, {
        approved: true,
        selected: true,
        error: null,
      });
      setMessage("Mapping approved. The unit is ready to apply.");
    } catch (error) {
      updateDraft(item.unit.import_unit_id, {
        approved: false,
        selected: false,
        error: importErrorMessage(error),
      });
      setMessage(importErrorMessage(error));
    } finally {
      setBusy(false);
    }
  }

  async function skip(item: WorkbookImportUnitDiscovery) {
    if (discovery === null || busy) {
      return;
    }
    setBusy(true);
    try {
      await setWorkbookImportUnitSelection({
        availability,
        apiBase,
        sessionId: discovery.session.import_session_id,
        unitId: item.unit.import_unit_id,
        selected: false,
        transactionPrefix: `workbook-import-${item.unit.import_unit_id}`,
      });
      updateDraft(item.unit.import_unit_id, {
        selected: false,
        error: null,
      });
      setMessage("Unit skipped. Its approved mapping is retained.");
    } catch (error) {
      setMessage(importErrorMessage(error));
    } finally {
      setBusy(false);
    }
  }

  async function applySelected() {
    if (
      discovery === null ||
      selectedUnitIds.length === 0 ||
      overlappingUnits.length !== 0 ||
      busy
    ) {
      return;
    }
    setBusy(true);
    setCurrentJob(null);
    setMessage("Applying selected units in workbook order.");
    try {
      const job = await applyWorkbookImport({
        availability,
        apiBase,
        sessionId: discovery.session.import_session_id,
        selectedUnitIds,
        transactionPrefix: "workbook-import-assistant",
        onJob: setCurrentJob,
      });
      const targets = [
        ...new Set(
          selectedUnitIds
            .map((unitId) => drafts[unitId]?.targetViewSchemaId)
            .filter((value): value is string => value !== undefined),
        ),
      ];
      setResultViews(targets);
      setMessage(
        job.result_summary?.code === "import_session_partially_applied"
          ? "Import completed with partial outcomes. Review the selected views."
          : "Import completed. Open a target view to review the new records.",
      );
    } catch (error) {
      setMessage(importErrorMessage(error));
    } finally {
      setBusy(false);
    }
  }

  async function cancelCurrentJob() {
    if (currentJob === null || currentJob.status === "succeeded") {
      return;
    }
    try {
      const job = await cancelImportJob({
        availability,
        apiBase,
        jobId: currentJob.job_id,
        transactionPrefix: "workbook-import-assistant",
      });
      setCurrentJob(job);
      setMessage("Cancellation requested. Committed units remain committed.");
    } catch (error) {
      setMessage(importErrorMessage(error));
    }
  }

  if (!available) {
    return <p role="status">The Import profile is not available.</p>;
  }

  return (
    <div data-testid={workbookImportAssistantTestId()} style={assistantStyle}>
      <p style={introStyle}>
        Discover bounded CSV or XLSX data, review each unit, approve its field
        mapping, and explicitly select or skip it before applying.
      </p>
      {!canManage ? (
        <p role="status">
          Editor, reviewer, or administrator access is required.
        </p>
      ) : (
        <div style={uploadStyle}>
          <label>
            Source workbook
            <input
              accept=".csv,.xlsx,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
              disabled={busy}
              type="file"
              onChange={(event) => {
                setFile(event.currentTarget.files?.[0] ?? null);
              }}
            />
          </label>
          <button
            disabled={file === null || busy}
            type="button"
            onClick={discover}
          >
            Upload and discover
          </button>
        </div>
      )}
      <p aria-live="polite" role="status">
        {message}
      </p>
      {discovery?.session.nonblocking_warning_codes.map((warning) => (
        <p key={warning}>Warning: {warning}</p>
      ))}
      {discovery?.units.map((item, index) => {
        const draft = drafts[item.unit.import_unit_id];
        if (draft === undefined) {
          return null;
        }
        const target =
          importTargets.find(
            (candidate) => candidate.viewSchemaId === draft.targetViewSchemaId,
          ) ?? importTargets[0];
        return (
          <section
            key={item.unit.import_unit_id}
            aria-label={`Import unit ${index + 1}`}
            style={unitStyle}
          >
            <h3 style={unitHeadingStyle}>
              Unit {index + 1}: {locatorLabel(item)}
            </h3>
            <p>
              {item.preview.inferred_row_count} rows ×{" "}
              {item.preview.inferred_column_count} columns; source{" "}
              {item.preview.source_rect_a1}
            </p>
            {item.preview.warning_codes.map((warning) => (
              <p key={warning}>Warning: {warning}</p>
            ))}
            <PreviewTable discovery={item} />
            <label>
              Target view
              <select
                disabled={busy || draft.selected}
                value={draft.targetViewSchemaId}
                onChange={(event) => {
                  const nextTarget = importTargets.find(
                    (candidate) =>
                      candidate.viewSchemaId === event.currentTarget.value,
                  );
                  if (nextTarget !== undefined) {
                    updateDraft(item.unit.import_unit_id, {
                      targetViewSchemaId: nextTarget.viewSchemaId,
                      fieldsByOrdinal: suggestFields(
                        item.preview.columns,
                        nextTarget,
                      ),
                      error: null,
                    });
                  }
                }}
              >
                {importTargets.map((candidate) => (
                  <option
                    key={candidate.viewSchemaId}
                    value={candidate.viewSchemaId}
                  >
                    {candidate.title}
                  </option>
                ))}
              </select>
            </label>
            <div style={mappingStyle}>
              {item.preview.columns.map((column) => (
                <label key={column.source_column_ordinal}>
                  {column.source_header_text ??
                    `Column ${column.source_column_ordinal}`}
                  <select
                    disabled={busy || draft.selected}
                    value={
                      draft.fieldsByOrdinal[column.source_column_ordinal] ?? ""
                    }
                    onChange={(event) => {
                      updateDraft(item.unit.import_unit_id, {
                        fieldsByOrdinal: {
                          ...draft.fieldsByOrdinal,
                          [column.source_column_ordinal]:
                            event.currentTarget.value,
                        },
                        error: null,
                      });
                    }}
                  >
                    <option value="">Keep as unmapped source data</option>
                    {target?.fields
                      .filter((field) => field.writeKind === "direct_value")
                      .map((field) => (
                        <option key={field.fieldKey} value={field.fieldKey}>
                          {field.label}
                        </option>
                      ))}
                  </select>
                </label>
              ))}
            </div>
            {draft.error ? <p role="alert">{draft.error}</p> : null}
            <div style={actionsStyle}>
              <button
                disabled={busy || draft.selected}
                type="button"
                onClick={() => approveAndSelect(item)}
              >
                Approve mapping and select
              </button>
              <button disabled={busy} type="button" onClick={() => skip(item)}>
                Skip unit
              </button>
              <span>{draft.selected ? "Selected" : "Not selected"}</span>
            </div>
          </section>
        );
      })}
      {overlappingUnits.length > 0 ? (
        <p role="alert">
          Selected units overlap. Skip one of each overlapping pair before
          applying.
        </p>
      ) : null}
      {discovery === null ? null : (
        <div style={actionsStyle}>
          <button
            disabled={
              busy ||
              selectedUnitIds.length === 0 ||
              overlappingUnits.length !== 0
            }
            type="button"
            onClick={applySelected}
          >
            Apply {selectedUnitIds.length} selected unit
            {selectedUnitIds.length === 1 ? "" : "s"}
          </button>
          {currentJob !== null &&
          !["failed", "canceled", "succeeded"].includes(currentJob.status) ? (
            <button type="button" onClick={cancelCurrentJob}>
              Cancel import
            </button>
          ) : null}
          {currentJob?.progress ? (
            <span>
              Progress: {currentJob.progress.completed}/
              {currentJob.progress.total ?? "?"}
            </span>
          ) : null}
        </div>
      )}
      {resultViews.length > 0 ? (
        <nav aria-label="Imported result views" style={actionsStyle}>
          {resultViews.map((viewSchemaId) => {
            const contract = importTargets.find(
              (candidate) => candidate.viewSchemaId === viewSchemaId,
            );
            return (
              <button
                key={viewSchemaId}
                type="button"
                onClick={() => onNavigateToView(viewSchemaId)}
              >
                Open {contract?.title ?? viewSchemaId}
              </button>
            );
          })}
        </nav>
      ) : null}
    </div>
  );
}

function PreviewTable({
  discovery,
}: {
  readonly discovery: WorkbookImportUnitDiscovery;
}) {
  return (
    <div style={previewFrameStyle}>
      <table style={previewTableStyle}>
        <thead>
          <tr>
            {discovery.preview.columns.map((column) => (
              <th key={column.source_column_ordinal}>
                {column.source_header_text ??
                  `Column ${column.source_column_ordinal}`}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {discovery.preview.preview_rows.slice(0, 8).map((row) => (
            <tr key={row.source_row_ref}>
              {row.cells.map((cell) => (
                <td key={cell.source_column_ordinal}>{cell.display_text}</td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function suggestedTarget(item: WorkbookImportUnitDiscovery): ViewContract {
  const headerText = item.preview.columns
    .map((column) => column.source_header_text ?? "")
    .join(" ")
    .toLowerCase();
  const preferred = [
    headerText.includes("host") ? "cartulary.view.hosts.v1" : "",
    headerText.includes("account") || headerText.includes("identity")
      ? "cartulary.view.identities.v1"
      : "",
    headerText.includes("indicator") || headerText.includes("ioc")
      ? "cartulary.view.indicators.v1"
      : "",
    headerText.includes("evidence") ? "cartulary.view.evidence.v1" : "",
    "cartulary.view.timeline.v2",
  ].find((value) => value !== "");
  return (
    importTargets.find((target) => target.viewSchemaId === preferred) ??
    requireImportTarget()
  );
}

function requireImportTarget(): ViewContract {
  const target = importTargets[0];
  if (target === undefined) {
    throw new Error("import target registry is empty");
  }
  return target;
}

function suggestFields(
  columns: readonly DiscoveredImportColumn[],
  target: ViewContract,
): Readonly<Record<number, string>> {
  const fieldsByOrdinal: Record<number, string> = {};
  const fields = target.fields.filter(
    (field) => field.writeKind === "direct_value",
  );
  for (const column of columns) {
    const normalizedHeader = normalizeMappingToken(column.source_header_text);
    const field = fields.find(
      (candidate) =>
        normalizeMappingToken(candidate.label) === normalizedHeader ||
        normalizeMappingToken(candidate.fieldKey.split(".").at(-1) ?? "") ===
          normalizedHeader,
    );
    if (field !== undefined) {
      fieldsByOrdinal[column.source_column_ordinal] = field.fieldKey;
    }
  }
  return fieldsByOrdinal;
}

function mappingColumns(
  columns: readonly DiscoveredImportColumn[],
  target: ViewContract,
  fieldsByOrdinal: Readonly<Record<number, string>>,
): readonly WorkbookSourceColumnMapping[] {
  return columns.map((column) => {
    const fieldKey = fieldsByOrdinal[column.source_column_ordinal] || null;
    const field: ViewFieldContract | undefined =
      fieldKey === null ? undefined : target.fieldMap[fieldKey];
    return {
      source_column_ordinal: column.source_column_ordinal,
      source_header_text: column.source_header_text,
      field_key: fieldKey,
      entity_binding_mode: field?.entityBindingMode ?? null,
      transform_id: null,
      transform_options: {},
      empty_value_policy: "omit_field",
    };
  });
}

function normalizeMappingToken(value: string | null): string {
  return (value ?? "").toLowerCase().replaceAll(/[^a-z0-9]/gu, "");
}

function locatorLabel(item: WorkbookImportUnitDiscovery): string {
  const locator = decodeLocator(item.unit.locator);
  const sheetName = locator?.sheet_name;
  return typeof sheetName === "string"
    ? sheetName
    : item.unit.locator_kind === "csv_file"
      ? "CSV file"
      : item.unit.locator_kind;
}

function decodeLocator(value: unknown): Record<string, unknown> | null {
  if (typeof value === "object" && value !== null && !Array.isArray(value)) {
    return value as Record<string, unknown>;
  }
  if (typeof value !== "string" || !value.startsWith("{")) {
    return null;
  }
  try {
    const decoded: unknown = JSON.parse(value);
    return typeof decoded === "object" &&
      decoded !== null &&
      !Array.isArray(decoded)
      ? (decoded as Record<string, unknown>)
      : null;
  } catch {
    return null;
  }
}

function findOverlappingUnitIds(
  units: readonly WorkbookImportUnitDiscovery[],
): readonly string[] {
  const overlaps = new Set<string>();
  for (let leftIndex = 0; leftIndex < units.length; leftIndex += 1) {
    const left = units[leftIndex];
    if (left === undefined) {
      continue;
    }
    for (
      let rightIndex = leftIndex + 1;
      rightIndex < units.length;
      rightIndex += 1
    ) {
      const right = units[rightIndex];
      if (right === undefined || !unitsOverlap(left, right)) {
        continue;
      }
      overlaps.add(left.unit.import_unit_id);
      overlaps.add(right.unit.import_unit_id);
    }
  }
  return [...overlaps];
}

function unitsOverlap(
  left: WorkbookImportUnitDiscovery,
  right: WorkbookImportUnitDiscovery,
): boolean {
  const leftLocator = decodeLocator(left.unit.locator);
  const rightLocator = decodeLocator(right.unit.locator);
  if (
    left.unit.locator_kind === "csv_file" ||
    right.unit.locator_kind === "csv_file" ||
    leftLocator?.sheet_name !== rightLocator?.sheet_name
  ) {
    return false;
  }
  const leftRect = decodeRect(left.unit.source_rect_a1);
  const rightRect = decodeRect(right.unit.source_rect_a1);
  return (
    leftRect !== null &&
    rightRect !== null &&
    leftRect.left <= rightRect.right &&
    rightRect.left <= leftRect.right &&
    leftRect.top <= rightRect.bottom &&
    rightRect.top <= leftRect.bottom
  );
}

function decodeRect(
  value: string,
): { left: number; right: number; top: number; bottom: number } | null {
  const match = /^([A-Z]+)([1-9][0-9]*):([A-Z]+)([1-9][0-9]*)$/u.exec(value);
  if (match === null) {
    return null;
  }
  return {
    left: columnNumber(match[1] ?? ""),
    top: Number(match[2]),
    right: columnNumber(match[3] ?? ""),
    bottom: Number(match[4]),
  };
}

function columnNumber(value: string): number {
  let result = 0;
  for (const character of value) {
    result = result * 26 + character.charCodeAt(0) - 64;
  }
  return result;
}

function importErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    if (error.name === "AbortError") {
      return "Import availability changed. Stale results were discarded.";
    }
    return error.message.replaceAll("_", " ");
  }
  return "Import failed.";
}

const assistantStyle = {
  display: "grid",
  gap: "var(--ct-spacing-md)",
};

const introStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
};

const uploadStyle = {
  display: "flex",
  alignItems: "end",
  flexWrap: "wrap" as const,
  gap: "var(--ct-spacing-sm)",
};

const unitStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-md)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
};

const unitHeadingStyle = {
  margin: 0,
};

const mappingStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "var(--ct-spacing-sm)",
};

const actionsStyle = {
  display: "flex",
  alignItems: "center",
  flexWrap: "wrap" as const,
  gap: "var(--ct-spacing-sm)",
};

const previewFrameStyle = {
  overflowX: "auto" as const,
  maxBlockSize: "14rem",
};

const previewTableStyle = {
  width: "100%",
  borderCollapse: "collapse" as const,
  fontSize: "0.8rem",
};
