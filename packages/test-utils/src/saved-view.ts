import {
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
} from "@cartulary/ui-contracts";

import {
  type BrowserNetworkRequestLike,
  type BrowserNetworkResponseLike,
  type BrowserPageLike,
  delay,
  isLocatorVisible,
  requireEvaluate,
  requireSelectOption,
  supportsVisibilityCheck,
} from "./browser";

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
