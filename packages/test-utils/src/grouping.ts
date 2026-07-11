import type { WorkbookSurface } from "@cartulary/ui-contracts";

import { type BrowserPageLike, requireEvaluate } from "./browser";

export async function setGridGroupExpanded(options: {
  expanded: boolean;
  groupTestId: string;
  page: BrowserPageLike;
  surface: WorkbookSurface;
}) {
  const group = options.page.getByTestId(options.groupTestId);
  const evaluate = requireEvaluate(
    group,
    `setGridGroupExpanded(${options.surface}) requires locator.evaluate() support`,
  );
  const current = await evaluate((element) =>
    element.getAttribute("aria-expanded"),
  );
  if (current !== String(options.expanded)) {
    await group.click();
  }
  const next = await evaluate((element) =>
    element.getAttribute("aria-expanded"),
  );
  if (next !== String(options.expanded)) {
    throw new Error(
      `Expected group ${options.groupTestId} on ${options.surface} to have aria-expanded=${String(options.expanded)}, received ${String(next)}`,
    );
  }
}

export function collapseGridGroup(
  options: Omit<Parameters<typeof setGridGroupExpanded>[0], "expanded">,
) {
  return setGridGroupExpanded({ ...options, expanded: false });
}

export function expandGridGroup(
  options: Omit<Parameters<typeof setGridGroupExpanded>[0], "expanded">,
) {
  return setGridGroupExpanded({ ...options, expanded: true });
}

export async function assertGroupRowPresentationOnly(options: {
  groupTestId: string;
  page: BrowserPageLike;
  surface: WorkbookSurface;
}) {
  const group = options.page.getByTestId(options.groupTestId);
  const evaluate = requireEvaluate(
    group,
    `assertGroupRowPresentationOnly(${options.surface}) requires locator.evaluate() support`,
  );
  const state = (await evaluate((element) => {
    const row = element.closest('[role="row"]');
    if (row === null) {
      return { hasRow: false };
    }
    return {
      buttonCount: row.querySelectorAll("button").length,
      editableControlCount: row.querySelectorAll(
        'input, textarea, select, [contenteditable="true"]',
      ).length,
      hasRow: true,
      interactiveCount: row.querySelectorAll(
        'a[href], button, input, textarea, select, [role="button"], [role="textbox"], [contenteditable="true"]',
      ).length,
      recordId: row.getAttribute("data-grid-record-id"),
      rowKind: row.getAttribute("data-grid-row-kind"),
    };
  })) as
    | { readonly hasRow: false }
    | {
        readonly buttonCount: number;
        readonly editableControlCount: number;
        readonly hasRow: true;
        readonly interactiveCount: number;
        readonly recordId: string | null;
        readonly rowKind: string | null;
      };
  if (!state.hasRow) {
    throw new Error(
      `Expected group ${options.groupTestId} on ${options.surface} to have an ARIA row ancestor`,
    );
  }
  if (state.rowKind !== "group") {
    throw new Error(
      `Expected group ${options.groupTestId} on ${options.surface} to be marked data-grid-row-kind=group, received ${String(state.rowKind)}`,
    );
  }
  if (state.recordId !== null && state.recordId !== "") {
    throw new Error(
      `Expected group ${options.groupTestId} on ${options.surface} to omit data-grid-record-id, received ${state.recordId}`,
    );
  }
  if (state.editableControlCount !== 0) {
    throw new Error(
      `Expected group ${options.groupTestId} on ${options.surface} to expose no editable controls, received ${state.editableControlCount}`,
    );
  }
  if (state.buttonCount !== 1 || state.interactiveCount !== 1) {
    throw new Error(
      `Expected group ${options.groupTestId} on ${options.surface} to expose exactly one expand/collapse control, received ${state.buttonCount} buttons and ${state.interactiveCount} interactive elements`,
    );
  }
}
