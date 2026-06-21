import {
  gridFilterApplyTestId,
  gridFilterChipTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridSavedRowsSelector,
  gridScrollportSelector,
  gridShellTestId,
  gridSortHeaderTestId,
  rowCellTestId,
  savedViewActionMenuTestId,
  savedViewActionMenuTriggerTestId,
  savedViewCreateButtonTestId,
  savedViewDeleteButtonTestId,
  savedViewDuplicateButtonTestId,
  savedViewNameInputTestId,
  savedViewScopeSelectTestId,
  savedViewSelectorTestId,
  savedViewSetDefaultButtonTestId,
  savedViewSetHomeButtonTestId,
  savedViewUpdateButtonTestId,
  type WorkbookSurface,
  workbookFilterPopoverTriggerTestId,
} from "@cartulary/ui-contracts";

export function pasteMatrixText(
  matrix: readonly (readonly string[])[],
): string {
  return matrix.map((row) => row.join("\t")).join("\n");
}

type BrowserLocator = {
  blur?: () => Promise<void>;
  click: () => Promise<void>;
  dispatchEvent?: (type: string) => Promise<void>;
  evaluate?: (
    pageFunction: (element: Element, arg?: unknown) => unknown,
    arg?: unknown,
  ) => Promise<unknown>;
  fill: (value: string) => Promise<void>;
  isVisible?: () => Promise<boolean>;
  press?: (value: string) => Promise<void>;
  scrollIntoViewIfNeeded?: () => Promise<void>;
  selectOption?: (value: string | readonly string[]) => Promise<unknown>;
};

export type SavedViewSelectionState = {
  readonly activeViewSchemaId: string | null;
  readonly selectedSavedViewId: string | null;
  readonly selectedSheetRefKind: string | null;
};

export type WorkbookSheetRef = {
  readonly id: string;
  readonly kind: "saved_view" | "view_schema";
};

export type SavedViewPreferenceActionResult = {
  readonly field: "default_sheet_ref" | "home_sheet_ref";
  readonly requestBody: Record<string, unknown>;
  readonly responseBody: unknown;
  readonly status: number | null;
};

type BrowserResponseLike = {
  ok: () => boolean;
  status?: () => number;
};

type BrowserNetworkRequestLike = {
  method: () => string;
  postData?: () => string | null;
  postDataJSON?: () => unknown;
  url: () => string;
};

type BrowserNetworkResponseLike = BrowserResponseLike & {
  json?: () => Promise<unknown>;
  request: () => BrowserNetworkRequestLike;
  url: () => string;
};

type BrowserRequestLike = {
  post: (
    url: string,
    options: {
      data?: unknown;
      headers?: Record<string, string>;
    },
  ) => Promise<BrowserResponseLike>;
};

type BrowserPageLike = {
  evaluate?: (
    pageFunction: (arg?: unknown) => unknown,
    arg?: unknown,
  ) => Promise<unknown>;
  getByTestId: (value: string) => BrowserLocator;
  request?: BrowserRequestLike;
  waitForRequest?: (
    predicate: (request: BrowserNetworkRequestLike) => boolean,
  ) => Promise<BrowserNetworkRequestLike>;
  waitForResponse?: (
    predicate: (response: BrowserNetworkResponseLike) => boolean,
  ) => Promise<BrowserNetworkResponseLike>;
};

export type GridAnchorCommandScenario = {
  commit: (context: {
    input: BrowserLocator;
    page: BrowserPageLike;
    surface: WorkbookSurface;
  }) => Promise<void>;
  name: string;
};

export function gridAnchorCommandScenarios(
  surface: WorkbookSurface,
): readonly GridAnchorCommandScenario[] {
  return [
    {
      commit: async ({ input }) => {
        await requirePress(input, "Enter");
      },
      name: "enter",
    },
    {
      commit: async ({ input }) => {
        await requirePress(input, "Tab");
      },
      name: "tab",
    },
    {
      commit: async ({ input, page }) => {
        await requireBlur(input);
        await page.getByTestId(gridShellTestId(surface)).click();
      },
      name: "blur",
    },
    {
      commit: async ({ input }) => {
        await requireDispatchEvent(input, "paste");
      },
      name: "single-cell-paste",
    },
  ];
}

export async function resizeGridColumn(options: {
  deltaPx: number;
  fieldKey: string;
  page: BrowserPageLike;
  surface: WorkbookSurface;
}) {
  void options.deltaPx;
  const header = options.page.getByTestId(
    gridSortHeaderTestId(options.surface, options.fieldKey),
  );
  const evaluate = requireEvaluate(
    header,
    `resizeGridColumn(${options.surface}) requires locator.evaluate() support`,
  );
  return evaluate((element) => element.getBoundingClientRect().width);
}

export async function selectSavedView(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  savedViewId: string,
) {
  const selector = page.getByTestId(savedViewSelectorTestId(surface));
  const selectOption = requireSelectOption(
    selector,
    `selectSavedView(${surface}) requires locator.selectOption() support`,
  );
  await selectOption(savedViewId);
}

export async function readSavedViewSelectionState(
  page: BrowserPageLike,
  surface: WorkbookSurface,
): Promise<SavedViewSelectionState> {
  const selector = page.getByTestId(savedViewSelectorTestId(surface));
  const evaluate = requireEvaluate(
    selector,
    `readSavedViewSelectionState(${surface}) requires locator.evaluate() support`,
  );
  return (await evaluate((element) => ({
    activeViewSchemaId: element.getAttribute("data-active-view-schema-id"),
    selectedSavedViewId: element.getAttribute("data-selected-saved-view-id"),
    selectedSheetRefKind: element.getAttribute("data-selected-sheet-ref-kind"),
  }))) as SavedViewSelectionState;
}

export async function openSavedViewActionMenu(
  page: BrowserPageLike,
  surface: WorkbookSurface,
) {
  const menu = page.getByTestId(savedViewActionMenuTestId(surface));
  const canVerifyVisibility = supportsVisibilityCheck(menu);
  if (canVerifyVisibility) {
    if (await isLocatorVisible(menu)) {
      return;
    }
  }
  await page.getByTestId(savedViewActionMenuTriggerTestId(surface)).click();
  if (canVerifyVisibility) {
    if (!(await isLocatorVisible(menu))) {
      throw new Error(`Saved-view action menu for ${surface} did not open`);
    }
  }
}

async function clickSavedViewMenuActionAndWaitForClose(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  actionTestId: string,
) {
  await openSavedViewActionMenu(page, surface);
  await page.getByTestId(actionTestId).click();
  await waitForSavedViewActionMenuClose(page, surface);
}

async function waitForSavedViewActionMenuClose(
  page: BrowserPageLike,
  surface: WorkbookSurface,
) {
  const menu = page.getByTestId(savedViewActionMenuTestId(surface));
  if (!supportsVisibilityCheck(menu)) {
    return;
  }
  const deadline = Date.now() + 10_000;
  while (Date.now() < deadline) {
    if (!(await isLocatorVisible(menu))) {
      return;
    }
    await delay(50);
  }
  throw new Error(`Saved-view action menu for ${surface} did not close`);
}

export async function setSavedViewDraftName(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  displayName: string,
) {
  await openSavedViewActionMenu(page, surface);
  await page.getByTestId(savedViewNameInputTestId(surface)).fill(displayName);
}

export async function selectSavedViewScope(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  scope: "private" | "shared",
) {
  await openSavedViewActionMenu(page, surface);
  const scopeSelect = page.getByTestId(savedViewScopeSelectTestId(surface));
  const selectOption = requireSelectOption(
    scopeSelect,
    `selectSavedViewScope(${surface}) requires locator.selectOption() support`,
  );
  await selectOption(scope);
}

export async function createSavedViewFromCurrentSurface(
  page: BrowserPageLike,
  surface: WorkbookSurface,
) {
  await clickSavedViewMenuActionAndWaitForClose(
    page,
    surface,
    savedViewCreateButtonTestId(surface),
  );
}

export async function updateSavedViewFromCurrentSurface(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  savedViewId: string,
) {
  await clickSavedViewMenuActionAndWaitForClose(
    page,
    surface,
    savedViewUpdateButtonTestId(surface, savedViewId),
  );
}

export async function duplicateSavedViewFromCurrentSurface(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  savedViewId: string,
) {
  await clickSavedViewMenuActionAndWaitForClose(
    page,
    surface,
    savedViewDuplicateButtonTestId(surface, savedViewId),
  );
}

export async function deleteSavedViewFromCurrentSurface(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  savedViewId: string,
) {
  await clickSavedViewMenuActionAndWaitForClose(
    page,
    surface,
    savedViewDeleteButtonTestId(surface, savedViewId),
  );
}

export async function setCurrentSavedViewAsHome(
  page: BrowserPageLike,
  surface: WorkbookSurface,
) {
  await clickSavedViewMenuActionAndWaitForClose(
    page,
    surface,
    savedViewSetHomeButtonTestId(surface),
  );
}

export async function setCurrentSavedViewAsHomeAndWait(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  options: {
    expectedSheetRef: WorkbookSheetRef;
    incidentId: string;
  },
): Promise<SavedViewPreferenceActionResult> {
  return setCurrentSavedViewPreferenceAndWait(page, surface, {
    buttonTestId: savedViewSetHomeButtonTestId(surface),
    expectedSheetRef: options.expectedSheetRef,
    field: "home_sheet_ref",
    incidentId: options.incidentId,
    routeSuffix: "me",
  });
}

export async function setCurrentSavedViewAsDefault(
  page: BrowserPageLike,
  surface: WorkbookSurface,
) {
  await clickSavedViewMenuActionAndWaitForClose(
    page,
    surface,
    savedViewSetDefaultButtonTestId(surface),
  );
}

export async function setCurrentSavedViewAsDefaultAndWait(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  options: {
    expectedSheetRef: WorkbookSheetRef;
    incidentId: string;
  },
): Promise<SavedViewPreferenceActionResult> {
  return setCurrentSavedViewPreferenceAndWait(page, surface, {
    buttonTestId: savedViewSetDefaultButtonTestId(surface),
    expectedSheetRef: options.expectedSheetRef,
    field: "default_sheet_ref",
    incidentId: options.incidentId,
    routeSuffix: "default",
  });
}

export async function assertActiveFilterChipVisible(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  fieldKey: string,
) {
  const chip = page.getByTestId(gridFilterChipTestId(surface, fieldKey));
  if (!(await isLocatorVisible(chip))) {
    throw new Error(
      `Expected active filter chip for ${fieldKey} on ${surface} to be visible`,
    );
  }
}

export async function fillDownGridCells(options: {
  apiBase: string;
  csrfHeaders?: Record<string, string> | undefined;
  fieldKey: string;
  incidentId: string;
  page: BrowserPageLike;
  targetRecords: readonly {
    readonly baseRowVersion: number;
    readonly recordId: string;
  }[];
  surface: WorkbookSurface;
  value: string;
}) {
  if (options.page.request === undefined) {
    if (options.page.evaluate === undefined) {
      throw new Error(
        `fillDownGridCells(${options.surface}) requires page.evaluate() or page.request.post() support`,
      );
    }
  }
  const path = `/api/v1/incidents/${options.incidentId}/views/${options.surface}/bulk-mutations`;
  const apiURL = `${options.apiBase}${path}`;
  const data = {
    view_schema_id: options.surface,
    client_txn_id: `${options.surface}-fill-down-${Date.now()}`,
    kind: "fill_down_v1",
    field_key: options.fieldKey,
    value: options.value,
    targets: options.targetRecords.map((target) => ({
      record_id: target.recordId,
      base_row_version: target.baseRowVersion,
    })),
  };
  const headers = {
    "content-type": "application/json",
    ...(options.csrfHeaders ?? {}),
  };
  if (options.page.evaluate !== undefined) {
    const response = (await options.page.evaluate(
      async (arg) => {
        const request = arg as {
          data: unknown;
          headers: Record<string, string>;
          url: string;
        };
        const result = await fetch(request.url, {
          method: "POST",
          credentials: "include",
          headers: request.headers,
          body: JSON.stringify(request.data),
        });
        return { ok: result.ok, status: result.status };
      },
      { data, headers, url: path },
    )) as { ok?: unknown; status?: unknown };
    return {
      ok: () => response.ok === true,
      status: () =>
        typeof response.status === "number" ? response.status : Number.NaN,
    };
  }
  if (options.page.request === undefined) {
    throw new Error(
      `fillDownGridCells(${options.surface}) requires page.request.post() support`,
    );
  }
  const requestOptions: {
    data: unknown;
    headers?: Record<string, string>;
  } = { data };
  if (options.csrfHeaders !== undefined) {
    requestOptions.headers = options.csrfHeaders;
  }
  return options.page.request.post(apiURL, requestOptions);
}

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

export async function pasteGridMatrix(options: {
  fieldKey: string;
  matrix: readonly (readonly string[])[];
  page: BrowserPageLike;
  recordId: string;
  surface: WorkbookSurface;
}) {
  const cell = options.page.getByTestId(
    rowCellTestId(options.recordId, options.fieldKey),
  );
  await cell.click();
  const evaluate = requireEvaluate(
    cell,
    `pasteGridMatrix(${options.surface}) requires locator.evaluate() support`,
  );
  await evaluate((element, clipboardText) => {
    const data = new DataTransfer();
    data.setData("text/plain", String(clipboardText));
    element.dispatchEvent(
      new ClipboardEvent("paste", {
        bubbles: true,
        cancelable: true,
        clipboardData: data,
      }),
    );
  }, pasteMatrixText(options.matrix));
}

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
  await page.getByTestId(workbookFilterPopoverTriggerTestId(surface)).click();
  await page
    .getByTestId(gridFilterFieldTestId(surface))
    .selectOption?.(fieldKey);
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
  await page
    .getByTestId(gridGroupingSelectTestId(surface))
    .selectOption?.(fieldKey);
}

export function assertAnchorTestId(recordId: string, fieldKey: string): string {
  return rowCellTestId(recordId, fieldKey);
}

export function assertRecordFieldMutationAnchor(options: {
  actualRecordId: string;
  body: Record<string, unknown>;
  expectedRecordId: string;
  expectedValue?: unknown;
  fieldKey: string;
}) {
  const { actualRecordId, body, expectedRecordId, expectedValue, fieldKey } =
    options;
  if (actualRecordId !== expectedRecordId) {
    throw new Error(
      `Expected mutation for record_id ${expectedRecordId}, received ${actualRecordId}`,
    );
  }
  const changes = Array.isArray(body.changes) ? body.changes : [];
  const change = changes.find(
    (candidate): candidate is { field_key: string; value?: unknown } =>
      typeof candidate === "object" &&
      candidate !== null &&
      "field_key" in candidate &&
      candidate.field_key === fieldKey,
  );
  if (!change) {
    throw new Error(`Expected mutation body to include field_key ${fieldKey}`);
  }
  if ("expectedValue" in options && change.value !== expectedValue) {
    throw new Error(
      `Expected ${fieldKey} mutation value ${String(expectedValue)}, received ${String(change.value)}`,
    );
  }
}

export async function readGridScroll(
  page: BrowserPageLike,
  surface: WorkbookSurface,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `readGridScroll(${surface}) requires locator.evaluate() support`,
  );
  return readScrollSnapshot(evaluate, surface);
}

export async function scrollGridToBottom(
  page: BrowserPageLike,
  surface: WorkbookSurface,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `scrollGridToBottom(${surface}) requires locator.evaluate() support`,
  );
  return readScrollSnapshot(evaluate, surface, { kind: "bottom" });
}

export async function scrollGridToOffset(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  top: number,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `scrollGridToOffset(${surface}) requires locator.evaluate() support`,
  );
  return readScrollSnapshot(evaluate, surface, { kind: "offset", top });
}

export async function scrollGridCellIntoView(options: {
  cellKey: string;
  intervalMs?: number;
  page: BrowserPageLike;
  recordId: string;
  surface: WorkbookSurface;
  timeoutMs?: number;
}) {
  const scanOptions: Parameters<typeof scrollGridTargetIntoView>[0] = {
    page: options.page,
    surface: options.surface,
    targetTestId: rowCellTestId(options.recordId, options.cellKey),
  };
  if (options.intervalMs !== undefined) {
    scanOptions.intervalMs = options.intervalMs;
  }
  if (options.timeoutMs !== undefined) {
    scanOptions.timeoutMs = options.timeoutMs;
  }
  return scrollGridTargetIntoView(scanOptions);
}

export async function scrollGridTargetIntoView(options: {
  intervalMs?: number;
  page: BrowserPageLike;
  surface: WorkbookSurface;
  targetTestId: string;
  timeoutMs?: number;
}) {
  const {
    intervalMs = 50,
    page,
    surface,
    targetTestId,
    timeoutMs = 3_000,
  } = options;
  const target = page.getByTestId(targetTestId);
  if (await isLocatorVisible(target)) {
    return readGridScroll(page, surface);
  }

  const retryIntervalMs = Math.max(intervalMs, 0);
  const deadline = Date.now() + Math.max(timeoutMs, 0);
  const observation = createGridTargetScanObservation();

  for (;;) {
    if (await isLocatorVisible(target)) {
      await target.scrollIntoViewIfNeeded?.();
      return readGridScroll(page, surface);
    }

    let state = await readGridScrollDiagnostics(page, surface);
    observeGridTargetScanState(observation, state);
    let scanRangeGrew = false;
    const scanMaxTop = state.maxTop;
    const scanOffsets = buildGridScanOffsets(state);

    for (const top of scanOffsets) {
      await scrollGridToOffset(page, surface, top);
      observation.scrollAttempts += 1;
      await waitForGridTargetRetry(retryIntervalMs);
      if (await isLocatorVisible(target)) {
        await target.scrollIntoViewIfNeeded?.();
        return readGridScroll(page, surface);
      }
      state = await readGridScrollDiagnostics(page, surface);
      observeGridTargetScanState(observation, state);
      if (state.maxTop > scanMaxTop) {
        observation.scrollRangeGrowths += 1;
        scanRangeGrew = true;
        break;
      }
    }

    if (scanRangeGrew) {
      continue;
    }

    observation.completedScanCycles += 1;
    observation.completedScanMaxTop = Math.max(
      observation.completedScanMaxTop,
      scanMaxTop,
    );
    if (Date.now() > deadline) {
      break;
    }
    await waitForGridTargetRetry(retryIntervalMs);
  }

  const finalState = await readGridScrollDiagnostics(page, surface);
  observeGridTargetScanState(observation, finalState);
  throw new Error(
    [
      `Expected ${targetTestId} to become visible in the ${surface} grid viewport after scanning virtualized rows.`,
      `scrollTop=${finalState.top}`,
      `scrollLeft=${finalState.left}`,
      `clientHeight=${finalState.clientHeight}`,
      `scrollHeight=${finalState.scrollHeight}`,
      `maxTop=${finalState.maxTop}`,
      `mountedRowIds=${finalState.mountedRowIds.join(",") || "(none)"}`,
      `scanCycles=${observation.scanCycles}`,
      `completedScanCycles=${observation.completedScanCycles}`,
      `scrollAttempts=${observation.scrollAttempts}`,
      `scrollRangeGrowths=${observation.scrollRangeGrowths}`,
      `observedScrollable=${observation.scrollableScanCycles > 0}`,
      `observedMaxTop=${observation.maxTop}`,
      `completedScanMaxTop=${observation.completedScanMaxTop}`,
      `observedMountedRowIds=${
        Array.from(observation.mountedRowIds).join(",") || "(none)"
      }`,
    ].join(" "),
  );
}

export async function assertMountedGridRowCountAtMost(options: {
  maxRows: number;
  page: BrowserPageLike;
  surface: WorkbookSurface;
}) {
  const { maxRows, page, surface } = options;
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `assertMountedGridRowCountAtMost(${surface}) requires locator.evaluate() support`,
  );
  const mountedRows = (await evaluate(
    (element, selector) =>
      element.querySelectorAll(typeof selector === "string" ? selector : "")
        .length,
    gridSavedRowsSelector(),
  )) as number;
  if (mountedRows > maxRows) {
    throw new Error(
      `Expected ${surface} to mount at most ${maxRows} saved rows, received ${mountedRows}`,
    );
  }
}

export async function assertMarkerAnchoredToGridTarget(options: {
  anchorKind: "cell" | "row-gutter";
  markerTestId: string;
  page: BrowserPageLike;
  surface: WorkbookSurface;
  targetTestId: string;
}) {
  const { anchorKind, markerTestId, page, surface, targetTestId } = options;
  const markerVisible = await isTestIdVisibleWithinGridViewport(
    page,
    surface,
    markerTestId,
  );
  if (!markerVisible) {
    throw new Error(
      `Expected marker ${markerTestId} to be visible in the ${surface} grid viewport`,
    );
  }
  const targetVisible = await isTestIdVisibleWithinGridViewport(
    page,
    surface,
    targetTestId,
  );
  if (!targetVisible) {
    throw new Error(
      `Expected target ${targetTestId} to be visible in the ${surface} grid viewport`,
    );
  }
  const state = await readMarkerAnchorState({
    markerTestId,
    page,
    surface,
    targetTestId,
  });
  if (state.markerRowRecordId !== state.targetRowRecordId) {
    throw new Error(
      `Expected marker ${markerTestId} to share row record_id ${state.targetRowRecordId} with target ${targetTestId}, received ${state.markerRowRecordId}`,
    );
  }
  if (anchorKind === "cell") {
    if (state.markerCellFieldKey !== state.targetCellFieldKey) {
      throw new Error(
        `Expected marker ${markerTestId} to share cell field_key ${state.targetCellFieldKey} with target ${targetTestId}, received ${state.markerCellFieldKey}`,
      );
    }
    if (
      !containsRect(state.targetCellRect, state.markerRect, anchorTolerancePx)
    ) {
      throw new Error(
        `Expected marker ${markerTestId} to be geometrically inside target cell ${targetTestId} (marker=${formatRect(state.markerRect)} targetCell=${formatRect(state.targetCellRect)})`,
      );
    }
    return;
  }

  if (state.markerCellFieldKey !== state.targetCellFieldKey) {
    throw new Error(
      `Expected row-gutter marker ${markerTestId} to be inside target gutter field_key ${state.targetCellFieldKey}, received ${state.markerCellFieldKey}`,
    );
  }
  const markerCenterY = (state.markerRect.top + state.markerRect.bottom) / 2;
  if (
    markerCenterY < state.targetRowRect.top - anchorTolerancePx ||
    markerCenterY > state.targetRowRect.bottom + anchorTolerancePx
  ) {
    throw new Error(
      `Expected row-gutter marker ${markerTestId} to be vertically anchored to row ${state.targetRowRecordId} (marker=${formatRect(state.markerRect)} targetRow=${formatRect(state.targetRowRect)})`,
    );
  }
  if (
    !containsRect(state.targetCellRect, state.markerRect, anchorTolerancePx)
  ) {
    throw new Error(
      `Expected row-gutter marker ${markerTestId} to be geometrically inside target gutter cell ${targetTestId} (marker=${formatRect(state.markerRect)} targetCell=${formatRect(state.targetCellRect)})`,
    );
  }
}

export async function isTestIdVisibleWithinGridViewport(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  testId: string,
) {
  const state = await readTestIdGridViewportState(page, surface, testId);
  return (
    state.top >= -viewportVisibilityTolerancePx &&
    state.left >= -viewportVisibilityTolerancePx &&
    state.bottom <= state.containerHeight + viewportVisibilityTolerancePx &&
    state.right <= state.containerWidth + viewportVisibilityTolerancePx
  );
}

type GridRectSnapshot = {
  readonly bottom: number;
  readonly height: number;
  readonly left: number;
  readonly right: number;
  readonly top: number;
  readonly width: number;
};

type MarkerAnchorState = {
  readonly markerCellFieldKey: string;
  readonly markerRect: GridRectSnapshot;
  readonly markerRowRecordId: string;
  readonly targetCellFieldKey: string;
  readonly targetCellRect: GridRectSnapshot;
  readonly targetRowRect: GridRectSnapshot;
  readonly targetRowRecordId: string;
};

async function readMarkerAnchorState(options: {
  markerTestId: string;
  page: BrowserPageLike;
  surface: WorkbookSurface;
  targetTestId: string;
}) {
  const { markerTestId, page, surface, targetTestId } = options;
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `assertMarkerAnchoredToGridTarget(${surface}, ${markerTestId}, ${targetTestId}) requires locator.evaluate() support`,
  );
  return (await evaluate(readMarkerAnchorStateInGrid, {
    markerTestId,
    surface,
    targetTestId,
  })) as MarkerAnchorState;
}

function readMarkerAnchorStateInGrid(
  element: Element,
  rawOptions?: unknown,
): MarkerAnchorState {
  const options =
    typeof rawOptions === "object" && rawOptions !== null
      ? (rawOptions as {
          markerTestId?: unknown;
          surface?: unknown;
          targetTestId?: unknown;
        })
      : {};
  const markerTestId =
    typeof options.markerTestId === "string" ? options.markerTestId : "";
  const targetTestId =
    typeof options.targetTestId === "string" ? options.targetTestId : "";
  const surface =
    typeof options.surface === "string" ? options.surface : "workbook";
  function findElementByTestId(
    root: Element,
    testId: string,
    role: "marker" | "target",
  ) {
    const match = Array.from(
      root.querySelectorAll<HTMLElement>("[data-testid]"),
    )
      .filter((candidate) => candidate.getAttribute("data-testid") === testId)
      .at(0);
    if (match === undefined) {
      throw new Error(`Expected ${surface} grid ${role} ${testId} to exist`);
    }
    return match;
  }
  function requireClosestElement(
    candidate: Element,
    selector: string,
    description: string,
  ) {
    const match = candidate.closest<HTMLElement>(selector);
    if (match === null) {
      throw new Error(`Expected ${description} to have closest ${selector}`);
    }
    return match;
  }
  function rectSnapshot(rect: DOMRect) {
    return {
      bottom: rect.bottom,
      height: rect.height,
      left: rect.left,
      right: rect.right,
      top: rect.top,
      width: rect.width,
    };
  }
  function effectiveRowRect(row: HTMLElement) {
    const ownRect = row.getBoundingClientRect();
    if (ownRect.width > 0 && ownRect.height > 0) {
      return rectSnapshot(ownRect);
    }
    const childRects = Array.from(row.querySelectorAll<HTMLElement>("*"))
      .map((child) => child.getBoundingClientRect())
      .filter((rect) => rect.width > 0 && rect.height > 0);
    if (childRects.length === 0) {
      return rectSnapshot(ownRect);
    }
    return {
      bottom: Math.max(...childRects.map((rect) => rect.bottom)),
      height:
        Math.max(...childRects.map((rect) => rect.bottom)) -
        Math.min(...childRects.map((rect) => rect.top)),
      left: Math.min(...childRects.map((rect) => rect.left)),
      right: Math.max(...childRects.map((rect) => rect.right)),
      top: Math.min(...childRects.map((rect) => rect.top)),
      width:
        Math.max(...childRects.map((rect) => rect.right)) -
        Math.min(...childRects.map((rect) => rect.left)),
    };
  }

  const marker = findElementByTestId(element, markerTestId, "marker");
  const target = findElementByTestId(element, targetTestId, "target");
  const markerRow = requireClosestElement(
    marker,
    '[role="row"][data-grid-record-id]',
    `${surface} marker ${markerTestId} row`,
  );
  const targetRow = requireClosestElement(
    target,
    '[role="row"][data-grid-record-id]',
    `${surface} target ${targetTestId} row`,
  );
  const markerCell = requireClosestElement(
    marker,
    "[data-grid-field-key]",
    `${surface} marker ${markerTestId} cell`,
  );
  const targetCell = requireClosestElement(
    target,
    "[data-grid-field-key]",
    `${surface} target ${targetTestId} cell`,
  );
  return {
    markerCellFieldKey:
      markerCell.getAttribute("data-grid-field-key") ?? "(missing)",
    markerRect: rectSnapshot(marker.getBoundingClientRect()),
    markerRowRecordId:
      markerRow.getAttribute("data-grid-record-id") ?? "(missing)",
    targetCellFieldKey:
      targetCell.getAttribute("data-grid-field-key") ?? "(missing)",
    targetCellRect: rectSnapshot(targetCell.getBoundingClientRect()),
    targetRowRect: effectiveRowRect(targetRow),
    targetRowRecordId:
      targetRow.getAttribute("data-grid-record-id") ?? "(missing)",
  };
}

async function readTestIdGridViewportState(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  testId: string,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const target = page.getByTestId(testId);
  const evaluateGrid = requireEvaluate(
    grid,
    `isTestIdVisibleWithinGridViewport(${surface}, ${testId}) requires locator.evaluate() support`,
  );
  const evaluateTarget = requireEvaluate(
    target,
    `isTestIdVisibleWithinGridViewport(${surface}, ${testId}) requires locator.evaluate() support`,
  );
  const containerRect = (await evaluateGrid(
    (element, options) => {
      const { scrollportSelector, surface } =
        typeof options === "object" && options !== null
          ? (options as { scrollportSelector?: unknown; surface?: unknown })
          : {};
      const selector =
        typeof scrollportSelector === "string" ? scrollportSelector : "";
      const scrollports = Array.from(
        element.querySelectorAll<HTMLElement>(selector),
      );
      if (scrollports.length !== 1) {
        throw new Error(
          `Expected ${typeof surface === "string" ? surface : "workbook"} grid shell to contain exactly one ${selector} scrollport, received ${scrollports.length}`,
        );
      }
      const gridScrollport = scrollports[0];
      if (gridScrollport === undefined) {
        throw new Error(
          `Expected ${typeof surface === "string" ? surface : "workbook"} grid shell to contain exactly one ${selector} scrollport, received 0`,
        );
      }
      const rect = gridScrollport.getBoundingClientRect();
      return {
        bottom: rect.bottom,
        height: rect.height,
        left: rect.left,
        right: rect.right,
        top: rect.top,
        width: rect.width,
      };
    },
    { scrollportSelector: gridScrollportSelector(), surface },
  )) as {
    bottom: number;
    height: number;
    left: number;
    right: number;
    top: number;
    width: number;
  };
  const elementRect = (await evaluateTarget((element) => {
    const rect = element.getBoundingClientRect();
    return {
      bottom: rect.bottom,
      height: rect.height,
      left: rect.left,
      right: rect.right,
      top: rect.top,
      width: rect.width,
    };
  })) as {
    bottom: number;
    height: number;
    left: number;
    right: number;
    top: number;
    width: number;
  };

  const top = elementRect.top - containerRect.top;
  const left = elementRect.left - containerRect.left;
  const bottom = elementRect.bottom - containerRect.top;
  const right = elementRect.right - containerRect.left;
  return {
    bottom,
    containerHeight: containerRect.height,
    containerWidth: containerRect.width,
    left,
    right,
    top,
  };
}

/**
 * Base continuity requires the focused control to remain focused and fully
 * visible. Exact scroll preservation is stricter and must be opted into.
 */
export async function assertGridFocusContinuity(options: {
  focusTestId: string;
  intervalMs?: number;
  page: BrowserPageLike;
  preservedScroll: { left: number; top: number };
  requireExactHorizontalScroll?: boolean;
  requireExactVerticalScroll?: boolean;
  surface: WorkbookSurface;
  timeoutMs?: number;
}) {
  const {
    focusTestId,
    intervalMs = 50,
    page,
    preservedScroll,
    requireExactHorizontalScroll = false,
    requireExactVerticalScroll = false,
    surface,
    timeoutMs = 3_000,
  } = options;
  const retryIntervalMs = Math.max(intervalMs, 0);
  const maxAttempts = Math.max(
    1,
    Math.ceil(timeoutMs / Math.max(retryIntervalMs, 1)) + 1,
  );
  const deadline = Date.now() + timeoutMs;
  let lastError: Error | null = null;

  for (let attempt = 0; attempt < maxAttempts; attempt += 1) {
    try {
      await assertGridFocusContinuityOnce({
        focusTestId,
        page,
        preservedScroll,
        requireExactHorizontalScroll,
        requireExactVerticalScroll,
        surface,
      });
      return;
    } catch (error) {
      lastError =
        error instanceof Error ? error : new Error(String(error ?? "unknown"));
      if (attempt === maxAttempts - 1 || Date.now() > deadline) {
        break;
      }
      await waitForGridContinuityRetry(retryIntervalMs);
    }
  }
  throw lastError ?? new Error("Grid continuity assertion timed out");
}

async function assertGridFocusContinuityOnce(options: {
  focusTestId: string;
  page: BrowserPageLike;
  preservedScroll: { left: number; top: number };
  requireExactHorizontalScroll: boolean;
  requireExactVerticalScroll: boolean;
  surface: WorkbookSurface;
}) {
  const {
    focusTestId,
    page,
    preservedScroll,
    requireExactHorizontalScroll,
    requireExactVerticalScroll,
    surface,
  } = options;
  const focusTarget = page.getByTestId(focusTestId);
  const evaluateFocusTarget = requireEvaluate(
    focusTarget,
    `assertGridFocusContinuity(${surface}, ${focusTestId}) requires locator.evaluate() support`,
  );
  const isFocused = (await evaluateFocusTarget(
    (element) => document.activeElement === element,
  )) as boolean;
  if (!isFocused) {
    throw new Error(
      `Expected ${focusTestId} to be focused within the ${surface} grid continuity restore`,
    );
  }
  const viewportState = await readTestIdGridViewportState(
    page,
    surface,
    focusTestId,
  );
  const isVisibleWithinViewport =
    viewportState.top >= -viewportVisibilityTolerancePx &&
    viewportState.left >= -viewportVisibilityTolerancePx &&
    viewportState.bottom <=
      viewportState.containerHeight + viewportVisibilityTolerancePx &&
    viewportState.right <=
      viewportState.containerWidth + viewportVisibilityTolerancePx;
  if (!isVisibleWithinViewport) {
    throw new Error(
      `Expected ${focusTestId} to remain fully visible within the ${surface} grid viewport (top=${viewportState.top}, bottom=${viewportState.bottom}, left=${viewportState.left}, right=${viewportState.right}, containerHeight=${viewportState.containerHeight}, containerWidth=${viewportState.containerWidth})`,
    );
  }
  const currentScroll = await readGridScroll(page, surface);
  if (requireExactVerticalScroll && currentScroll.top !== preservedScroll.top) {
    throw new Error(
      `Expected ${surface} vertical scroll ${preservedScroll.top}, received ${currentScroll.top}`,
    );
  }
  if (
    requireExactHorizontalScroll &&
    currentScroll.left !== preservedScroll.left
  ) {
    throw new Error(
      `Expected ${surface} horizontal scroll ${preservedScroll.left}, received ${currentScroll.left}`,
    );
  }
}

function delay(durationMs: number) {
  return new Promise<void>((resolve) => {
    setTimeout(resolve, durationMs);
  });
}

function waitForGridContinuityRetry(intervalMs: number) {
  if (intervalMs <= 0) {
    return Promise.resolve();
  }
  return delay(intervalMs);
}

type BrowserEvaluate = NonNullable<BrowserLocator["evaluate"]>;

type GridScrollAction =
  | { kind: "bottom" }
  | { kind: "none" }
  | { kind: "offset"; top: number };

type GridScrollDiagnostics = {
  readonly clientHeight: number;
  readonly clientWidth: number;
  readonly left: number;
  readonly maxTop: number;
  readonly mountedRowIds: readonly string[];
  readonly scrollHeight: number;
  readonly scrollWidth: number;
  readonly top: number;
};

async function readScrollSnapshot(
  evaluate: BrowserEvaluate,
  surface: WorkbookSurface,
  action: GridScrollAction = { kind: "none" },
) {
  const state = (await evaluate(readGridScrollState, {
    ...action,
    savedRowsSelector: gridSavedRowsSelector(),
    scrollportSelector: gridScrollportSelector(),
    surface,
  })) as GridScrollDiagnostics;
  return {
    left: state.left,
    top: state.top,
  };
}

async function readGridScrollDiagnostics(
  page: BrowserPageLike,
  surface: WorkbookSurface,
) {
  const grid = page.getByTestId(gridShellTestId(surface));
  const evaluate = requireEvaluate(
    grid,
    `readGridScrollDiagnostics(${surface}) requires locator.evaluate() support`,
  );
  return (await evaluate(readGridScrollState, {
    kind: "none",
    savedRowsSelector: gridSavedRowsSelector(),
    scrollportSelector: gridScrollportSelector(),
    surface,
  })) as GridScrollDiagnostics;
}

function readGridScrollState(
  element: Element,
  rawAction?: unknown,
): GridScrollDiagnostics {
  const action =
    typeof rawAction === "object" && rawAction !== null
      ? (rawAction as {
          kind?: unknown;
          savedRowsSelector?: unknown;
          scrollportSelector?: unknown;
          surface?: unknown;
          top?: unknown;
        })
      : { kind: "none" };
  const scrollportSelector =
    typeof action.scrollportSelector === "string"
      ? action.scrollportSelector
      : "";
  const surface =
    typeof action.surface === "string" ? action.surface : "workbook";
  const scrollports = Array.from(
    element.querySelectorAll<HTMLElement>(scrollportSelector),
  );
  if (scrollports.length !== 1) {
    throw new Error(
      `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received ${scrollports.length}`,
    );
  }
  const gridScrollport = scrollports[0];
  if (gridScrollport === undefined) {
    throw new Error(
      `Expected ${surface} grid shell to contain exactly one ${scrollportSelector} scrollport, received 0`,
    );
  }
  if (action.kind === "bottom") {
    gridScrollport.scrollTop = gridScrollport.scrollHeight;
  }
  if (action.kind === "offset") {
    const nextTop =
      typeof action.top === "number" && Number.isFinite(action.top)
        ? action.top
        : 0;
    gridScrollport.scrollTop = nextTop;
  }

  const scrollHeight = gridScrollport.scrollHeight;
  const clientHeight = gridScrollport.clientHeight;
  const savedRowsSelector =
    typeof action.savedRowsSelector === "string"
      ? action.savedRowsSelector
      : '[role="row"][data-grid-record-id]:not([data-grid-record-id=""])';
  return {
    clientHeight,
    clientWidth: gridScrollport.clientWidth,
    left: gridScrollport.scrollLeft,
    maxTop: Math.max(0, scrollHeight - clientHeight),
    mountedRowIds: Array.from(
      element.querySelectorAll<HTMLElement>(savedRowsSelector),
    ).map((row) => row.getAttribute("data-grid-record-id") ?? ""),
    scrollHeight,
    scrollWidth: gridScrollport.scrollWidth,
    top: gridScrollport.scrollTop,
  };
}

function buildGridScanOffsets(state: GridScrollDiagnostics) {
  const maxTop = Math.max(0, state.maxTop);
  if (maxTop === 0) {
    return [0];
  }
  const step = Math.max(1, Math.floor(Math.max(state.clientHeight, 1) / 2));
  const offsets = [0];
  for (let top = step; top < maxTop; top += step) {
    offsets.push(top);
  }
  offsets.push(maxTop);
  return Array.from(new Set(offsets));
}

type GridTargetScanObservation = {
  completedScanCycles: number;
  completedScanMaxTop: number;
  maxTop: number;
  mountedRowIds: Set<string>;
  scanCycles: number;
  scrollRangeGrowths: number;
  scrollableScanCycles: number;
  scrollAttempts: number;
};

function createGridTargetScanObservation(): GridTargetScanObservation {
  return {
    completedScanCycles: 0,
    completedScanMaxTop: 0,
    maxTop: 0,
    mountedRowIds: new Set(),
    scanCycles: 0,
    scrollRangeGrowths: 0,
    scrollableScanCycles: 0,
    scrollAttempts: 0,
  };
}

function observeGridTargetScanState(
  observation: GridTargetScanObservation,
  state: GridScrollDiagnostics,
) {
  observation.scanCycles += 1;
  observation.maxTop = Math.max(observation.maxTop, state.maxTop);
  if (state.maxTop > 0) {
    observation.scrollableScanCycles += 1;
  }
  for (const rowId of state.mountedRowIds) {
    if (rowId !== "") {
      observation.mountedRowIds.add(rowId);
    }
  }
}

function waitForGridTargetRetry(intervalMs: number) {
  if (intervalMs <= 0) {
    return Promise.resolve();
  }
  return delay(intervalMs);
}

async function setCurrentSavedViewPreferenceAndWait(
  page: BrowserPageLike,
  surface: WorkbookSurface,
  options: {
    buttonTestId: string;
    expectedSheetRef: WorkbookSheetRef;
    field: "default_sheet_ref" | "home_sheet_ref";
    incidentId: string;
    routeSuffix: "default" | "me";
  },
): Promise<SavedViewPreferenceActionResult> {
  const waitForRequest = requireWaitForRequest(
    page,
    `setCurrentSavedViewPreferenceAndWait(${surface}) requires page.waitForRequest() support`,
  );
  const waitForResponse = requireWaitForResponse(
    page,
    `setCurrentSavedViewPreferenceAndWait(${surface}) requires page.waitForResponse() support`,
  );
  const path = `/api/v1/incidents/${options.incidentId}/workbook-preferences/${options.routeSuffix}`;
  const matchesPreferenceRoute = (method: string, url: string) =>
    method.toUpperCase() === "PUT" && url.endsWith(path);
  const requestPromise = waitForRequest((request) =>
    matchesPreferenceRoute(request.method(), request.url()),
  );
  const responsePromise = waitForResponse((response) =>
    matchesPreferenceRoute(response.request().method(), response.url()),
  );

  await openSavedViewActionMenu(page, surface);
  await page.getByTestId(options.buttonTestId).click();
  const [request, response] = await Promise.all([
    requestPromise,
    responsePromise,
  ]);
  await waitForSavedViewActionMenuClose(page, surface);
  const requestBody = readRequestJSON(request, options.field);
  assertPreferenceBody(options.field, requestBody, options.expectedSheetRef);
  const responseBody = await readResponseJSON(response, options.field);
  if (!response.ok()) {
    throw new Error(
      `${options.field} update failed with status ${
        response.status?.() ?? "unknown"
      }`,
    );
  }
  assertPreferenceResponseBody(
    options.field,
    responseBody,
    options.expectedSheetRef,
  );
  return {
    field: options.field,
    requestBody,
    responseBody,
    status: response.status?.() ?? null,
  };
}

function requireWaitForRequest(
  page: BrowserPageLike,
  message: string,
): NonNullable<BrowserPageLike["waitForRequest"]> {
  if (typeof page.waitForRequest !== "function") {
    throw new Error(message);
  }
  return (predicate) =>
    page.waitForRequest?.(predicate) as Promise<BrowserNetworkRequestLike>;
}

function requireWaitForResponse(
  page: BrowserPageLike,
  message: string,
): NonNullable<BrowserPageLike["waitForResponse"]> {
  if (typeof page.waitForResponse !== "function") {
    throw new Error(message);
  }
  return (predicate) =>
    page.waitForResponse?.(predicate) as Promise<BrowserNetworkResponseLike>;
}

function readRequestJSON(
  request: BrowserNetworkRequestLike,
  field: string,
): Record<string, unknown> {
  if (request.postDataJSON !== undefined) {
    try {
      return requireRecord(request.postDataJSON(), `${field} request body`);
    } catch {
      // Fall back to raw body parsing for runtimes that reject postDataJSON().
    }
  }
  const raw = request.postData?.();
  if (raw === undefined || raw === null || raw === "") {
    throw new Error(`${field} request did not expose a JSON body`);
  }
  return requireRecord(JSON.parse(raw) as unknown, `${field} request body`);
}

async function readResponseJSON(
  response: BrowserNetworkResponseLike,
  field: string,
) {
  if (response.json === undefined) {
    throw new Error(`${field} response did not expose a JSON body`);
  }
  return response.json();
}

function assertPreferenceBody(
  field: "default_sheet_ref" | "home_sheet_ref",
  body: Record<string, unknown>,
  expectedSheetRef: WorkbookSheetRef,
) {
  const keys = Object.keys(body);
  if (keys.length !== 1 || keys[0] !== field) {
    throw new Error(
      `${field} request body must contain only ${field}; got ${keys.join(",")}`,
    );
  }
  assertSheetRef(body[field], expectedSheetRef, `${field} request`);
}

function assertPreferenceResponseBody(
  field: "default_sheet_ref" | "home_sheet_ref",
  body: unknown,
  expectedSheetRef: WorkbookSheetRef,
) {
  const envelope = requireRecord(body, `${field} response envelope`);
  const data = requireRecord(envelope.data, `${field} response data`);
  assertSheetRef(data[field], expectedSheetRef, `${field} response`);
}

function assertSheetRef(
  actual: unknown,
  expected: WorkbookSheetRef,
  label: string,
) {
  const record = requireRecord(actual, label);
  if (record.kind !== expected.kind || record.id !== expected.id) {
    throw new Error(
      `${label} sheet ref mismatch: expected ${expected.kind}:${expected.id}, got ${String(
        record.kind,
      )}:${String(record.id)}`,
    );
  }
}

function requireRecord(value: unknown, label: string): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be a JSON object`);
  }
  return value as Record<string, unknown>;
}

function requirePress(locator: BrowserLocator, value: string) {
  if (typeof locator.press !== "function") {
    throw new Error(`Grid browser-command ${value} requires locator.press()`);
  }
  return locator.press(value);
}

function requireBlur(locator: BrowserLocator) {
  if (typeof locator.blur !== "function") {
    throw new Error("Grid browser-command blur requires locator.blur()");
  }
  return locator.blur();
}

function requireDispatchEvent(locator: BrowserLocator, type: string) {
  if (typeof locator.dispatchEvent !== "function") {
    throw new Error(
      `Grid browser-command ${type} requires locator.dispatchEvent()`,
    );
  }
  return locator.dispatchEvent(type);
}

function containsRect(
  outer: GridRectSnapshot,
  inner: GridRectSnapshot,
  tolerancePx: number,
) {
  return (
    inner.top >= outer.top - tolerancePx &&
    inner.left >= outer.left - tolerancePx &&
    inner.bottom <= outer.bottom + tolerancePx &&
    inner.right <= outer.right + tolerancePx
  );
}

function formatRect(rect: GridRectSnapshot) {
  return `top=${rect.top},right=${rect.right},bottom=${rect.bottom},left=${rect.left},width=${rect.width},height=${rect.height}`;
}

async function isLocatorVisible(locator: BrowserLocator) {
  if (typeof locator.isVisible === "function") {
    return locator.isVisible();
  }
  try {
    const evaluate = requireEvaluate(
      locator,
      "isLocatorVisible requires locator.evaluate() support",
    );
    return Boolean(
      await evaluate((element) => {
        const rect = element.getBoundingClientRect();
        return (
          element.isConnected &&
          rect.width > 0 &&
          rect.height > 0 &&
          getComputedStyle(element).visibility !== "hidden"
        );
      }),
    );
  } catch {
    return false;
  }
}

function supportsVisibilityCheck(locator: BrowserLocator) {
  return typeof locator.isVisible === "function";
}

const anchorTolerancePx = 2;
const viewportVisibilityTolerancePx = 1;

function requireEvaluate(
  locator: BrowserLocator,
  message: string,
): BrowserEvaluate {
  if (typeof locator.evaluate !== "function") {
    throw new Error(message);
  }
  return (pageFunction, arg) =>
    locator.evaluate?.(pageFunction, arg) as Promise<unknown>;
}

function requireSelectOption(
  locator: BrowserLocator,
  message: string,
): NonNullable<BrowserLocator["selectOption"]> {
  if (typeof locator.selectOption !== "function") {
    throw new Error(message);
  }
  return (value) => locator.selectOption?.(value) as Promise<unknown>;
}
