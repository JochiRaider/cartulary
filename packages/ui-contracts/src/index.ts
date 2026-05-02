export type WorkbookSurface = string;

export function gridShellTestId(surface: WorkbookSurface): string {
  return `${surface}-grid-shell`;
}

/**
 * Scope this selector through `gridShellTestId(surface)` when targeting
 * workbook rows. Do not rely on raw table markup or renderer classes.
 */
export function gridSavedRowsSelector(): string {
  return '[role="row"][data-grid-record-id]:not([data-grid-record-id=""])';
}

/**
 * Scope this selector through `gridShellTestId(surface)` when targeting the
 * workbook draft row. Do not rely on raw table markup or renderer classes.
 */
export function gridDraftRowSelector(): string {
  return '[role="row"][data-grid-record-id=""]';
}

export function gridSortHeaderTestId(
  surface: WorkbookSurface,
  fieldKey: string,
): string {
  return `${surface}-sort-${sanitizeToken(fieldKey)}`;
}

export function gridFilterChipTestId(
  surface: WorkbookSurface,
  fieldKey: string,
): string {
  return `${surface}-filter-chip-${sanitizeToken(fieldKey)}`;
}

export function gridFilterFieldTestId(surface: WorkbookSurface): string {
  return `${surface}-filter-field`;
}

export function gridFilterValueTestId(surface: WorkbookSurface): string {
  return `${surface}-filter-value`;
}

export function gridFilterApplyTestId(surface: WorkbookSurface): string {
  return `${surface}-filter-apply`;
}

export function gridGroupingSelectTestId(surface: WorkbookSurface): string {
  return `${surface}-group-by`;
}

export function gridGroupRowTestId(
  surface: WorkbookSurface,
  fieldKey: string,
  value: string,
): string {
  return `${surface}-group-${sanitizeToken(fieldKey)}-${sanitizeToken(value)}`;
}

export function rowCellTestId(recordId: string, fieldKey: string): string {
  return `row-${recordId}-${fieldKey}`;
}

export function rowInspectorFieldTestId(
  recordId: string,
  fieldKey: string,
): string {
  return `${rowCellTestId(recordId, fieldKey)}-inspector`;
}

export function rowInspectButtonTestId(recordId: string): string {
  return `row-${recordId}-inspect`;
}

export function draftCellTestId(fieldKey: string): string {
  return `draft-row-${fieldKey}`;
}

export function relationshipItemsTestId(
  recordId: string,
  relationshipKey: string,
): string {
  return `row-${recordId}-${relationshipKey}-items`;
}

export function timelineRowVersionTestId(recordId: string): string {
  return `row-${recordId}-row-version`;
}

function sanitizeToken(value: string): string {
  return value.replace(/[^a-zA-Z0-9_-]+/gu, "-");
}
