import {
  dataTestIdPrefixSelector,
  encodeSelectorSegment,
  requireFieldKey,
  requireRecordId,
} from "./selectorCore";
import { viewFirstTestId, viewScopedTestId } from "./viewSchemaSelectors";

export function gridScrollportClassName(): string {
  return "cartulary-grid-scrollport";
}

export function gridScrollportSelector(): string {
  return `.${gridScrollportClassName()}`;
}

export function gridActionsHeaderTestId(viewSchemaId: string): string {
  return viewFirstTestId(viewSchemaId, "actions-header");
}

export function gridRowGutterTestId(
  viewSchemaId: string,
  recordId: string,
): string {
  return viewFirstTestId(
    viewSchemaId,
    `row-gutter-${requireRecordId(recordId)}`,
  );
}

/**
 * Scope this selector through `gridShellTestId(surface)` when targeting
 * workbook rows. Do not rely on raw table markup or renderer classes.
 */
export function gridSavedRowsSelector(): string {
  return '[role="row"][data-grid-record-id]:not([data-grid-record-id=""])';
}

/**
 * Scope these selectors through an owner grid shell. They describe adapter-
 * owned semantic data rows and cells without exposing vendor classes or
 * positional coordinates.
 */
export function gridDataRowsSelector(): string {
  return '[role="row"][data-cartulary-grid-row-kind="data"]';
}

export function gridDataCellsSelector(): string {
  return `${gridDataRowsSelector()} [role="gridcell"]`;
}

/**
 * Scope this selector through `gridShellTestId(surface)` when targeting the
 * workbook draft row. Do not rely on raw table markup or renderer classes.
 */
export function gridDraftRowSelector(): string {
  return '[role="row"][data-cartulary-grid-draft-row="true"]';
}

export function gridSortHeaderTestId(
  viewSchemaId: string,
  fieldKey: string,
): string {
  return viewFirstTestId(viewSchemaId, `sort-${requireFieldKey(fieldKey)}`);
}

export function gridFilterChipTestId(
  viewSchemaId: string,
  fieldKey: string,
): string {
  return viewFirstTestId(
    viewSchemaId,
    `filter-chip-${requireFieldKey(fieldKey)}`,
  );
}

export function gridFilterFieldTestId(viewSchemaId: string): string {
  return viewFirstTestId(viewSchemaId, "filter-field");
}

export function gridFilterValueTestId(viewSchemaId: string): string {
  return viewFirstTestId(viewSchemaId, "filter-value");
}

export function gridFilterApplyTestId(viewSchemaId: string): string {
  return viewFirstTestId(viewSchemaId, "filter-apply");
}

export function gridGroupingSelectTestId(viewSchemaId: string): string {
  return viewFirstTestId(viewSchemaId, "group-by");
}

export function gridGroupRowTestId(
  viewSchemaId: string,
  fieldKey: string,
  value: string,
): string {
  return `${gridGroupRowPrefix(viewSchemaId, fieldKey)}${encodeSelectorSegment(value, "group value")}`;
}

function gridGroupRowPrefix(viewSchemaId: string, fieldKey: string): string {
  return `${viewFirstTestId(viewSchemaId, `group-${requireFieldKey(fieldKey)}`)}-`;
}

export function gridGroupRowsSelector(
  viewSchemaId: string,
  fieldKey: string,
): string {
  return dataTestIdPrefixSelector(gridGroupRowPrefix(viewSchemaId, fieldKey));
}

export function gridRowTestId(viewSchemaId: string, recordId: string): string {
  return `${viewScopedTestId("grid-row", viewSchemaId)}-${requireRecordId(recordId)}`;
}
