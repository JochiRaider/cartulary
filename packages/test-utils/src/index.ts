export type WorkbookSurface = "timeline" | "hosts" | "identities";

export function gridShellTestId(surface: WorkbookSurface): string {
  return `${surface}-grid-shell`;
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

export function pasteMatrixText(matrix: readonly (readonly string[])[]): string {
  return matrix.map((row) => row.join("\t")).join("\n");
}

type BrowserLocator = {
  click: () => Promise<void>;
  fill: (value: string) => Promise<void>;
  press?: (value: string) => Promise<void>;
  scrollIntoViewIfNeeded?: () => Promise<void>;
  selectOption?: (value: string) => Promise<void>;
};

type BrowserPageLike = {
  getByTestId: (value: string) => BrowserLocator;
};

export async function sortByHeader(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  fieldKey: string,
) {
  await page.getByTestId(gridSortHeaderTestId(surface, fieldKey)).click();
}

export async function applyFilterChip(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  fieldKey: string,
  value: string,
) {
  await page.getByTestId(gridFilterFieldTestId(surface)).selectOption?.(
    fieldKey,
  );
  const valueControl = page.getByTestId(gridFilterValueTestId(surface));
  try {
    await valueControl.selectOption?.(value);
  } catch {
    await valueControl.fill(value);
  }
  await page.getByTestId(gridFilterApplyTestId(surface)).click();
}

export async function removeFilterChip(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  fieldKey: string,
) {
  await page.getByTestId(gridFilterChipTestId(surface, fieldKey)).click();
}

export async function changeGrouping(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  fieldKey: string,
) {
  await page.getByTestId(gridGroupingSelectTestId(surface)).selectOption?.(
    fieldKey,
  );
}

export async function scrollToCell(
  page: BrowserPageLike,
  recordId: string,
  fieldKey: string,
) {
  await page
    .getByTestId(rowCellTestId(recordId, fieldKey))
    .scrollIntoViewIfNeeded?.();
}

export function assertAnchorTestId(recordId: string, fieldKey: string): string {
  return rowCellTestId(recordId, fieldKey);
}

function sanitizeToken(value: string): string {
  return value.replace(/[^a-zA-Z0-9_-]+/gu, "-");
}
