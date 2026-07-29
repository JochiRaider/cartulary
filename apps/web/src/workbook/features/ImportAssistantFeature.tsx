import { workbookImportAssistantTestId } from "@cartulary/ui-contracts";
import {
  listViewContracts,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import { useEffect, useMemo, useRef, useState } from "react";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import {
  importProfileId,
  importRouteFamily,
} from "../../extensions/extensionWorkspaceIdentities";
import {
  applyWorkbookImport,
  approveWorkbookImportMapping,
  cancelImportJob,
  createWorkbookImportRegion,
  type DiscoveredImportColumn,
  type ImportJobResource,
  setWorkbookImportUnitSelection,
  uploadAndDiscoverWorkbookImport,
  type WorkbookImportDiscovery,
  type WorkbookImportUnitDiscovery,
  type WorkbookSourceColumnMapping,
} from "../../imports/importCoordinator";
import {
  type SelectableViewImportTarget,
  selectableViewImportTargets,
} from "../../services/importTargetContractAdapter";
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

type ImportViewTarget = {
  readonly semantics: SelectableViewImportTarget;
  readonly contract: ViewContract;
};

const viewContracts = listViewContracts();
const importTargets: readonly ImportViewTarget[] =
  selectableViewImportTargets.map((semantics) => {
    const matches = viewContracts.filter(
      (contract) => contract.viewSchemaId === semantics.target_view_schema_id,
    );
    const contract = matches[0];
    if (matches.length !== 1 || contract === undefined) {
      throw new Error(
        `missing or duplicate view contract for ${semantics.target_id}`,
      );
    }
    return { semantics, contract };
  });

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
          targetViewSchemaId: target.contract.viewSchemaId,
          fieldsByOrdinal: suggestFields(item.preview.columns, target.contract),
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

  async function createOperatorRegion(
    item: WorkbookImportUnitDiscovery,
    sourceRect: SourceRect,
  ) {
    if (discovery === null || busy) {
      return;
    }
    const operation = operationRef.current + 1;
    operationRef.current = operation;
    setBusy(true);
    setMessage("Creating a durable operator region.");
    try {
      const created = await createWorkbookImportRegion({
        availability,
        apiBase,
        sessionId: discovery.session.import_session_id,
        baseUnitId: item.unit.import_unit_id,
        sourceRect,
        transactionPrefix: `workbook-import-${item.unit.import_unit_id}`,
      });
      if (operation !== operationRef.current) {
        return;
      }
      const target = suggestedTarget(created);
      setDiscovery((current) => {
        if (current === null) {
          return current;
        }
        const existingIndex = current.units.findIndex(
          ({ unit }) => unit.import_unit_id === created.unit.import_unit_id,
        );
        return {
          ...current,
          units:
            existingIndex === -1
              ? [...current.units, created]
              : current.units.map((candidate, index) =>
                  index === existingIndex ? created : candidate,
                ),
        };
      });
      setDrafts((current) => {
        if (current[created.unit.import_unit_id] !== undefined) {
          return current;
        }
        return {
          ...current,
          [created.unit.import_unit_id]: {
            targetViewSchemaId: target.contract.viewSchemaId,
            fieldsByOrdinal: suggestFields(
              created.preview.columns,
              target.contract,
            ),
            approved: false,
            selected: false,
            error: null,
          },
        };
      });
      setMessage("Operator region created. Review its mapping separately.");
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

  async function approveAndSelect(item: WorkbookImportUnitDiscovery) {
    const draft = drafts[item.unit.import_unit_id];
    if (discovery === null || draft === undefined || busy) {
      return;
    }
    setBusy(true);
    let mappingApproved = draft.approved;
    setMessage(
      mappingApproved
        ? "Reselecting unit with its approved mapping."
        : "Approving mapping and selecting unit.",
    );
    try {
      if (!mappingApproved) {
        const target = requireImportTarget(draft.targetViewSchemaId);
        const sourceColumns = mappingColumns(
          item.preview.columns,
          target.contract,
          draft.fieldsByOrdinal,
        );
        const hasUnmapped = sourceColumns.some(
          (column) => column.field_key === null,
        );
        const unknownColumnPolicy =
          target.semantics.default_unknown_column_policy;
        if (hasUnmapped && unknownColumnPolicy === "reject_if_unmapped") {
          throw new Error("Map every source column for this target.");
        }
        await approveWorkbookImportMapping({
          availability,
          apiBase,
          sessionId: discovery.session.import_session_id,
          discovery: item,
          targetViewSchemaId: target.contract.viewSchemaId,
          unknownColumnPolicy,
          sourceColumns,
          transactionPrefix: `workbook-import-${item.unit.import_unit_id}`,
        });
        mappingApproved = true;
      }
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
      setMessage(
        draft.approved
          ? "Unit reselected with its approved mapping."
          : "Mapping approved. The unit is ready to apply.",
      );
    } catch (error) {
      updateDraft(item.unit.import_unit_id, {
        approved: mappingApproved,
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
        const target = requireImportTarget(draft.targetViewSchemaId);
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
            {item.unit.locator_kind === "xlsx_used_range" ? (
              <OperatorRegionForm
                busy={busy}
                discovery={item}
                onCreate={(sourceRect) =>
                  createOperatorRegion(item, sourceRect)
                }
              />
            ) : null}
            <label>
              Target view
              <select
                disabled={busy || draft.selected}
                value={draft.targetViewSchemaId}
                onChange={(event) => {
                  const nextTarget = importTargets.find(
                    (candidate) =>
                      candidate.contract.viewSchemaId ===
                      event.currentTarget.value,
                  );
                  if (nextTarget !== undefined) {
                    updateDraft(item.unit.import_unit_id, {
                      targetViewSchemaId: nextTarget.contract.viewSchemaId,
                      fieldsByOrdinal: suggestFields(
                        item.preview.columns,
                        nextTarget.contract,
                      ),
                      error: null,
                    });
                  }
                }}
              >
                {importTargets.map((candidate) => (
                  <option
                    key={candidate.contract.viewSchemaId}
                    value={candidate.contract.viewSchemaId}
                  >
                    {candidate.contract.title}
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
                    {target.contract.fields
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
                {draft.selected
                  ? "Mapping approved and selected"
                  : draft.approved
                    ? "Reselect unit"
                    : "Approve mapping and select"}
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
              (candidate) => candidate.contract.viewSchemaId === viewSchemaId,
            )?.contract;
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

type SourceRect = {
  readonly startRow: number;
  readonly startColumn: number;
  readonly endRow: number;
  readonly endColumn: number;
};

function OperatorRegionForm({
  busy,
  discovery,
  onCreate,
}: {
  readonly busy: boolean;
  readonly discovery: WorkbookImportUnitDiscovery;
  readonly onCreate: (sourceRect: SourceRect) => void | Promise<void>;
}) {
  const baseRect = decodeRect(discovery.unit.source_rect_a1);
  const [sourceRect, setSourceRect] = useState<SourceRect>(() => ({
    startRow: baseRect?.top ?? 1,
    startColumn: baseRect?.left ?? 1,
    endRow: baseRect?.bottom ?? 1,
    endColumn: baseRect?.right ?? 1,
  }));
  if (baseRect === null) {
    return <p role="alert">This worksheet range cannot define a region.</p>;
  }
  const valid =
    sourceRect.startRow >= baseRect.top &&
    sourceRect.startColumn >= baseRect.left &&
    sourceRect.endRow <= baseRect.bottom &&
    sourceRect.endColumn <= baseRect.right &&
    sourceRect.startRow <= sourceRect.endRow &&
    sourceRect.startColumn <= sourceRect.endColumn;

  function update(member: keyof SourceRect, value: string) {
    setSourceRect((current) => ({
      ...current,
      [member]: Number(value),
    }));
  }

  return (
    <fieldset style={regionStyle}>
      <legend>Create operator-selected region</legend>
      <p style={introStyle}>
        Enter one-based inclusive coordinates inside this worksheet used range.
      </p>
      <div style={mappingStyle}>
        <label>
          Region start row
          <input
            min={baseRect.top}
            type="number"
            value={sourceRect.startRow}
            onChange={(event) => update("startRow", event.currentTarget.value)}
          />
        </label>
        <label>
          Region start column
          <input
            min={baseRect.left}
            type="number"
            value={sourceRect.startColumn}
            onChange={(event) =>
              update("startColumn", event.currentTarget.value)
            }
          />
        </label>
        <label>
          Region end row
          <input
            max={baseRect.bottom}
            min={baseRect.top}
            type="number"
            value={sourceRect.endRow}
            onChange={(event) => update("endRow", event.currentTarget.value)}
          />
        </label>
        <label>
          Region end column
          <input
            max={baseRect.right}
            min={baseRect.left}
            type="number"
            value={sourceRect.endColumn}
            onChange={(event) => update("endColumn", event.currentTarget.value)}
          />
        </label>
      </div>
      <button
        disabled={busy || !valid}
        type="button"
        onClick={() => onCreate(sourceRect)}
      >
        Create operator region
      </button>
    </fieldset>
  );
}

function suggestedTarget(item: WorkbookImportUnitDiscovery): ImportViewTarget {
  const normalizedHeaders = new Set(
    item.preview.columns
      .map((column) => normalizeMappingToken(column.source_header_text))
      .filter((header) => header !== ""),
  );
  let best = requireImportTarget();
  let bestScore = -1;
  for (const target of importTargets) {
    const fieldTokens = new Set(
      target.contract.fields
        .filter((field) => field.writeKind === "direct_value")
        .flatMap((field) => [
          normalizeMappingToken(field.label),
          normalizeMappingToken(field.fieldKey.split(".").at(-1) ?? ""),
        ]),
    );
    const score = [...normalizedHeaders].filter((header) =>
      fieldTokens.has(header),
    ).length;
    if (score > bestScore) {
      best = target;
      bestScore = score;
    }
  }
  return best;
}

function requireImportTarget(viewSchemaId?: string): ImportViewTarget {
  const target =
    viewSchemaId === undefined
      ? importTargets[0]
      : importTargets.find(
          (candidate) => candidate.contract.viewSchemaId === viewSchemaId,
        );
  if (target === undefined) {
    throw new Error(
      viewSchemaId === undefined
        ? "import target registry is empty"
        : `unsupported import target ${viewSchemaId}`,
    );
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

function normalizeMappingToken(
  value: DiscoveredImportColumn["source_header_text"],
): string {
  return (value === null ? "" : String(value))
    .toLowerCase()
    .replaceAll(/[^a-z0-9]/gu, "");
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

const regionStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  margin: 0,
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
