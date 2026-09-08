import { useState } from "react";
import type { WorkbookImportController } from "../../imports/WorkbookImportController";
import {
  type ImportRectangle,
  importRectangle,
  validImportRegion,
  workbookMappingErrors,
} from "../../imports/workbookImportMapping";
import type {
  ImportUnitState,
  WorkbookImportState,
} from "../../imports/workbookImportState";
import { importFailureMessage } from "../../services/importClient";
import type { DiscoveredImportUnit } from "../../services/importContractAdapter";
import { workbookImportTargets } from "../../services/importTargetContractAdapter";
import {
  ImportOperationNotice,
  importActionsStyle,
  importSectionStyle,
} from "./ImportOperationNotice";

const mappingStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(min(12rem, 100%), 1fr))",
  gap: "var(--ct-spacing-sm)",
  minWidth: 0,
};
export function ImportUnitCard({
  item,
  index,
  selected,
  mutable,
  state,
  controller,
}: {
  readonly item: ImportUnitState;
  readonly index: number;
  readonly selected: boolean;
  readonly mutable: boolean;
  readonly state: WorkbookImportState;
  readonly controller: WorkbookImportController;
}) {
  const { unit, preview, draft } = item,
    id = unit.import_unit_id;
  const target = workbookImportTargets.find(
    (t) => t.contract.viewSchemaId === draft?.targetViewSchemaId,
  );
  const errors = draft && preview ? workbookMappingErrors(draft, preview) : {};
  const operation =
    state.operation &&
    "unitId" in state.operation.attempt &&
    state.operation.attempt.unitId === id
      ? state.operation
      : null;
  const terminal =
    ["applied", "failed", "rejected"].includes(unit.unit_status) ||
    ["applied", "partially_applied", "failed", "canceled"].includes(
      state.session?.session_status ?? "",
    );
  return (
    <section aria-label={`Import unit ${index + 1}`} style={importSectionStyle}>
      <h3 style={{ margin: 0 }}>
        Unit {index + 1}:{" "}
        {typeof unit.locator.sheet_name === "string"
          ? unit.locator.sheet_name
          : unit.locator_kind === "csv_file"
            ? "CSV file"
            : unit.locator_kind.replaceAll("_", " ")}
      </h3>
      <p style={{ margin: 0 }}>
        {unit.inferred_row_count} rows × {unit.inferred_column_count} columns;
        source {unit.source_rect_a1}. Outcome:{" "}
        {unit.unit_status.replaceAll("_", " ")}.{" "}
        {selected ? "Selected" : "Not selected"}.
      </p>
      {[...new Set(unit.warning_codes)].map((code) => (
        <p key={code}>Warning: {code}</p>
      ))}
      <p style={{ margin: 0 }}>
        {unit.approved_mapping
          ? "Approved mapping retained."
          : "Mapping has not been approved."}{" "}
        {draft?.dirty ? "Local mapping draft needs explicit approval." : ""}
      </p>
      {item.previewLoading ? <p>Loading unit preview…</p> : null}
      {item.previewFailure ? (
        <p role="alert">Preview: {importFailureMessage(item.previewFailure)}</p>
      ) : null}
      {!preview || item.previewFailure ? (
        <button
          type="button"
          disabled={item.previewLoading}
          onClick={() => controller.loadPreview(id)}
        >
          {item.previewFailure ? "Retry unit preview" : "Load unit preview"}
        </button>
      ) : null}
      {preview ? (
        <details open={index === 0}>
          <summary>Source preview</summary>
          <section
            aria-label={`Unit ${index + 1} source preview`}
            // biome-ignore lint/a11y/noNoninteractiveTabindex: Keyboard users must be able to scroll the bounded preview.
            tabIndex={0}
            style={{
              overflow: "auto",
              maxBlockSize: "14rem",
              maxInlineSize: "100%",
            }}
          >
            <table
              style={{
                borderCollapse: "collapse",
                fontSize: "0.8rem",
                minInlineSize: "100%",
              }}
            >
              <caption>
                Source cells for unit {index + 1}
                {preview.truncated ? " (bounded preview)" : ""}
              </caption>
              <thead>
                <tr>
                  {preview.columns.map((column) => (
                    <th key={column.source_column_ordinal} scope="col">
                      {String(
                        column.source_header_text ??
                          `Column ${column.source_column_ordinal}`,
                      )}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {preview.preview_rows.map((row) => (
                  <tr key={row.source_row_ref}>
                    {row.cells.map((cell) => (
                      <td
                        key={cell.source_column_ordinal}
                        style={{
                          padding: "var(--ct-spacing-xs)",
                          verticalAlign: "top",
                          minWidth: "6rem",
                          maxWidth: "24rem",
                          overflowWrap: "anywhere",
                        }}
                      >
                        {cell.display_text}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </section>
        </details>
      ) : null}
      {preview && draft ? (
        <>
          {!terminal ? (
            <>
              <label style={{ display: "grid", minWidth: 0 }}>
                Target view
                <select
                  style={{ maxWidth: "100%" }}
                  disabled={!mutable || selected}
                  value={draft.targetViewSchemaId}
                  onChange={(event) =>
                    controller.updateMapping(id, {
                      targetViewSchemaId: event.currentTarget.value,
                    })
                  }
                >
                  {workbookImportTargets.map((t) => (
                    <option
                      key={t.contract.viewSchemaId}
                      value={t.contract.viewSchemaId}
                    >
                      {t.contract.title}
                    </option>
                  ))}
                </select>
              </label>
              <p style={{ margin: 0 }}>
                Header matches are suggestions.{" "}
                {target?.semantics.default_unknown_column_policy ===
                "reject_if_unmapped"
                  ? "Every source column must be mapped."
                  : target?.semantics.default_unknown_column_policy ===
                      "preserve_custom_attrs"
                    ? "Unmapped columns are retained as custom attributes."
                    : "Unmapped columns remain in the raw source capture."}
              </p>
              <div style={mappingStyle}>
                {preview.columns.map((column) => {
                  const ordinal = column.source_column_ordinal,
                    errorId = `import-${id}-${ordinal}-error`;
                  const failureField = item.failure?.field;
                  const error =
                    errors[ordinal] ??
                    (failureField === `source_columns[${ordinal - 1}]` ||
                    failureField?.startsWith(`source_columns[${ordinal - 1}].`)
                      ? item.failure
                        ? importFailureMessage(item.failure)
                        : undefined
                      : undefined);
                  return (
                    <div key={ordinal} style={{ display: "grid", minWidth: 0 }}>
                      <label style={{ display: "grid", minWidth: 0 }}>
                        {String(
                          column.source_header_text ?? `Column ${ordinal}`,
                        )}
                        <select
                          style={{ maxWidth: "100%" }}
                          disabled={!mutable || selected}
                          aria-invalid={error ? true : undefined}
                          aria-describedby={error ? errorId : undefined}
                          value={draft.fields[ordinal] ?? ""}
                          onChange={(event) =>
                            controller.updateMapping(id, {
                              ordinal,
                              fieldKey: event.currentTarget.value,
                            })
                          }
                        >
                          <option value="">Unmapped</option>
                          {target?.fields.map((field) => (
                            <option key={field.fieldKey} value={field.fieldKey}>
                              {field.label}
                            </option>
                          ))}
                        </select>
                      </label>
                      {error ? <span id={errorId}>{error}</span> : null}
                    </div>
                  );
                })}
              </div>
              <div style={importActionsStyle}>
                <button
                  type="button"
                  disabled={
                    !mutable || selected || Object.keys(errors).length > 0
                  }
                  onClick={() => void controller.approve(id)}
                >
                  {selected
                    ? "Mapping approved and selected"
                    : unit.approved_mapping && !draft.dirty
                      ? "Reselect unit"
                      : "Approve mapping and select"}
                </button>
                <button
                  type="button"
                  disabled={!mutable || unit.unit_status === "skipped"}
                  onClick={() => void controller.select(id, false)}
                >
                  Skip unit
                </button>
              </div>
              {selected ? (
                <p>Skip this unit before editing its mapping.</p>
              ) : null}
            </>
          ) : (
            <p>
              Approved target:{" "}
              {unit.approved_mapping &&
              "target_view_schema_id" in unit.approved_mapping
                ? (workbookImportTargets.find(
                    (t) =>
                      t.contract.viewSchemaId ===
                      unit.approved_mapping?.target_view_schema_id,
                  )?.contract.title ?? "Unavailable target")
                : "No workbook target"}
              .
            </p>
          )}
          {!state.canWrite ? (
            <details>
              <summary>Copy local mapping draft</summary>
              <textarea
                aria-label={`Unit ${index + 1} local mapping draft`}
                readOnly
                value={JSON.stringify(draft, null, 2)}
                style={{
                  inlineSize: "100%",
                  boxSizing: "border-box",
                  minBlockSize: "6rem",
                }}
              />
            </details>
          ) : null}
        </>
      ) : null}
      {unit.locator_kind === "xlsx_used_range" && preview && !terminal ? (
        <OperatorRegionForm
          unit={unit}
          enabled={mutable}
          create={(rect) => void controller.createRegion(id, rect)}
        />
      ) : null}
      {operation ? (
        <ImportOperationNotice
          operation={operation}
          controller={controller}
          canRetry={state.canWrite}
        />
      ) : item.failure ? (
        <p role="alert">{importFailureMessage(item.failure)}</p>
      ) : null}
    </section>
  );
}

function OperatorRegionForm({
  unit,
  enabled,
  create,
}: {
  readonly unit: DiscoveredImportUnit;
  readonly enabled: boolean;
  readonly create: (rect: ImportRectangle) => void;
}) {
  const [rect, setRect] = useState(() => importRectangle(unit.source_rect_a1));
  if (!rect) return <p>This worksheet range cannot define a region.</p>;
  const fields = [
    { key: "startRow", label: "Region start row" },
    { key: "startColumn", label: "Region start column" },
    { key: "endRow", label: "Region end row" },
    { key: "endColumn", label: "Region end column" },
  ] as const;
  return (
    <fieldset disabled={!enabled} style={importSectionStyle}>
      <legend>Create operator-selected region</legend>
      <p>
        Use inclusive coordinates inside the worksheet range, including a header
        and data row.
      </p>
      <div style={mappingStyle}>
        {fields.map((field) => (
          <label key={field.key} style={{ minWidth: 0, display: "grid" }}>
            {field.label}
            <input
              type="number"
              min={1}
              step={1}
              value={rect[field.key]}
              onChange={(event) =>
                setRect({
                  ...rect,
                  [field.key]: Number(event.currentTarget.value),
                })
              }
            />
          </label>
        ))}
      </div>
      <button
        type="button"
        disabled={!enabled || !validImportRegion(rect, unit)}
        onClick={() => create(rect)}
      >
        Create operator region
      </button>
    </fieldset>
  );
}
