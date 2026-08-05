import type {
  CreateIncidentSavedViewRequest,
  CreateIncidentSavedViewResponse,
  DeleteIncidentSavedViewResponse,
  SheetRef,
} from "@cartulary/protocol-ts/http";
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
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";
import { csrfHeaders } from "../auth/browserSession";
import { apiBase } from "../runtime/configuration";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";
import { createEnvironmentTestControlClient } from "../transport/testControlEnvironment";

type SavedViewLocatorLike = {
  click: () => Promise<void>;
  evaluate?: (
    pageFunction: (element: Element, arg?: unknown) => unknown,
    arg?: unknown,
  ) => Promise<unknown>;
  fill: (value: string) => Promise<void>;
  isVisible?: () => Promise<boolean>;
  selectOption?: (value: string | readonly string[]) => Promise<unknown>;
};

type SavedViewNetworkRequestLike = {
  method: () => string;
  postData?: () => string | null;
  postDataJSON?: () => unknown;
  url: () => string;
};

type SavedViewNetworkResponseLike = {
  json?: () => Promise<unknown>;
  ok: () => boolean;
  request: () => SavedViewNetworkRequestLike;
  status?: () => number;
  url: () => string;
};

type SavedViewPageLike = {
  getByTestId: (value: string) => SavedViewLocatorLike;
  waitForRequest?: (
    predicate: (request: SavedViewNetworkRequestLike) => boolean,
  ) => Promise<SavedViewNetworkRequestLike>;
  waitForResponse?: (
    predicate: (response: SavedViewNetworkResponseLike) => boolean,
  ) => Promise<SavedViewNetworkResponseLike>;
};

type SavedViewEvaluate = NonNullable<SavedViewLocatorLike["evaluate"]>;

function requireSavedViewEvaluate(
  locator: SavedViewLocatorLike,
  message: string,
): SavedViewEvaluate {
  if (typeof locator.evaluate !== "function") {
    throw new Error(message);
  }
  return (pageFunction, arg) =>
    locator.evaluate?.(pageFunction, arg) as Promise<unknown>;
}

function requireSavedViewSelectOption(
  locator: SavedViewLocatorLike,
  message: string,
): NonNullable<SavedViewLocatorLike["selectOption"]> {
  if (typeof locator.selectOption !== "function") {
    throw new Error(message);
  }
  return (value) => locator.selectOption?.(value) as Promise<unknown>;
}

async function isSavedViewLocatorVisible(locator: SavedViewLocatorLike) {
  if (typeof locator.isVisible === "function") {
    return locator.isVisible();
  }
  try {
    const evaluate = requireSavedViewEvaluate(
      locator,
      "isSavedViewLocatorVisible requires locator.evaluate() support",
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

function supportsSavedViewVisibilityCheck(locator: SavedViewLocatorLike) {
  return typeof locator.isVisible === "function";
}

function waitForSavedViewRetry(durationMs: number) {
  return new Promise<void>((resolve) => {
    setTimeout(resolve, durationMs);
  });
}

export type SavedViewApiResource = CreateIncidentSavedViewResponse["data"];

export async function createSavedView(
  page: Page,
  incidentId: string,
  options: {
    display_name: string;
    layout_json?: Record<string, unknown>;
    query_json?: Record<string, unknown>;
    scope?: "private" | "shared";
    view_schema_id: string;
  },
): Promise<SavedViewApiResource> {
  const body = {
    display_name: options.display_name,
    layout_json: options.layout_json ?? {},
    query_json: options.query_json ?? {},
    ...(options.scope === undefined ? {} : { scope: options.scope }),
    view_schema_id: options.view_schema_id,
  } satisfies CreateIncidentSavedViewRequest;
  const response = await publicHttpOperation({
    body,
    headers: await csrfHeaders(page),
    operationID: "createIncidentSavedView",
    pathParameters: { incident_id: incidentId },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok) {
    throw new Error(
      `createIncidentSavedView failed with HTTP ${response.status}: ${JSON.stringify(response.payload)}`,
    );
  }
  return response.payload.data;
}

export async function deleteSavedView(
  page: Page,
  incidentId: string,
  savedViewId: string,
): Promise<DeleteIncidentSavedViewResponse["data"]> {
  const response = await publicHttpOperation({
    headers: await csrfHeaders(page),
    operationID: "deleteIncidentSavedView",
    pathParameters: {
      incident_id: incidentId,
      saved_view_id: savedViewId,
    },
    request: atJsonOrigin(page.request, apiBase),
  });
  if (!response.ok) {
    throw new Error(
      `deleteIncidentSavedView failed with HTTP ${response.status}: ${JSON.stringify(response.payload)}`,
    );
  }
  return response.payload.data;
}

export async function seedSystemSavedView(
  page: Page,
  incidentId: string,
  options: {
    display_name: string;
    layout_json?: Record<string, unknown>;
    query_json?: Record<string, unknown>;
    view_schema_id: string;
  },
): Promise<SavedViewApiResource> {
  const response = await createEnvironmentTestControlClient(page.request, {
    endpointOrigin: apiBase,
  }).request({
    body: {
      display_name: options.display_name,
      layout_json: options.layout_json ?? {},
      query_json: options.query_json ?? {},
      view_schema_id: options.view_schema_id,
    },
    method: "POST",
    path: `/api/v1/test/incidents/${incidentId}/saved-views/system`,
  });
  expect(response.ok).toBeTruthy();
  return (response.body as { data: SavedViewApiResource }).data;
}

type SavedViewSelectionState = {
  readonly activeViewSchemaId: string | null;
  readonly selectedSavedViewId: string | null;
  readonly selectedSheetRefKind: string | null;
};

type SavedViewPreferenceActionResult = {
  readonly field: "default_sheet_ref" | "home_sheet_ref";
  readonly requestBody: Record<string, unknown>;
  readonly responseBody: unknown;
  readonly status: number | null;
};

export async function selectSavedView(
  page: SavedViewPageLike,
  surface: string,
  savedViewId: string,
) {
  const selector = page.getByTestId(savedViewSelectorTestId(surface));
  const selectOption = requireSavedViewSelectOption(
    selector,
    `selectSavedView(${surface}) requires locator.selectOption() support`,
  );
  await selectOption(savedViewId);
}

export async function readSavedViewSelectionState(
  page: SavedViewPageLike,
  surface: string,
): Promise<SavedViewSelectionState> {
  const selector = page.getByTestId(savedViewSelectorTestId(surface));
  const evaluate = requireSavedViewEvaluate(
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
  page: SavedViewPageLike,
  surface: string,
) {
  const menu = page.getByTestId(savedViewActionMenuTestId(surface));
  const canVerifyVisibility = supportsSavedViewVisibilityCheck(menu);
  if (canVerifyVisibility) {
    if (await isSavedViewLocatorVisible(menu)) {
      return;
    }
  }
  await page.getByTestId(savedViewActionMenuTriggerTestId(surface)).click();
  if (canVerifyVisibility) {
    if (!(await isSavedViewLocatorVisible(menu))) {
      throw new Error(`Saved-view action menu for ${surface} did not open`);
    }
  }
}

export async function setSavedViewDraftName(
  page: SavedViewPageLike,
  surface: string,
  displayName: string,
) {
  await openSavedViewActionMenu(page, surface);
  await page.getByTestId(savedViewNameInputTestId(surface)).fill(displayName);
}

export async function selectSavedViewScope(
  page: SavedViewPageLike,
  surface: string,
  scope: "private" | "shared",
) {
  await openSavedViewActionMenu(page, surface);
  const scopeSelect = page.getByTestId(savedViewScopeSelectTestId(surface));
  const selectOption = requireSavedViewSelectOption(
    scopeSelect,
    `selectSavedViewScope(${surface}) requires locator.selectOption() support`,
  );
  await selectOption(scope);
}

export async function createSavedViewFromCurrentSurface(
  page: SavedViewPageLike,
  surface: string,
) {
  await clickSavedViewMenuActionAndWaitForClose(
    page,
    surface,
    savedViewCreateButtonTestId(surface),
  );
}

export async function updateSavedViewFromCurrentSurface(
  page: SavedViewPageLike,
  surface: string,
  savedViewId: string,
) {
  await clickSavedViewMenuActionAndWaitForClose(
    page,
    surface,
    savedViewUpdateButtonTestId(surface, savedViewId),
  );
}

export async function duplicateSavedViewFromCurrentSurface(
  page: SavedViewPageLike,
  surface: string,
  savedViewId: string,
) {
  await clickSavedViewMenuActionAndWaitForClose(
    page,
    surface,
    savedViewDuplicateButtonTestId(surface, savedViewId),
  );
}

export async function deleteSavedViewFromCurrentSurface(
  page: SavedViewPageLike,
  surface: string,
  savedViewId: string,
) {
  await clickSavedViewMenuActionAndWaitForClose(
    page,
    surface,
    savedViewDeleteButtonTestId(surface, savedViewId),
  );
}

export async function setCurrentSavedViewAsHome(
  page: SavedViewPageLike,
  surface: string,
) {
  await clickSavedViewMenuActionAndWaitForClose(
    page,
    surface,
    savedViewSetHomeButtonTestId(surface),
  );
}

export async function setCurrentSavedViewAsHomeAndWait(
  page: SavedViewPageLike,
  surface: string,
  options: {
    expectedSheetRef: SheetRef;
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
  page: SavedViewPageLike,
  surface: string,
) {
  await clickSavedViewMenuActionAndWaitForClose(
    page,
    surface,
    savedViewSetDefaultButtonTestId(surface),
  );
}

export async function setCurrentSavedViewAsDefaultAndWait(
  page: SavedViewPageLike,
  surface: string,
  options: {
    expectedSheetRef: SheetRef;
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
  page: SavedViewPageLike,
  surface: string,
  actionTestId: string,
) {
  await openSavedViewActionMenu(page, surface);
  await page.getByTestId(actionTestId).click();
  await waitForSavedViewActionMenuClose(page, surface);
}

async function waitForSavedViewActionMenuClose(
  page: SavedViewPageLike,
  surface: string,
) {
  const menu = page.getByTestId(savedViewActionMenuTestId(surface));
  if (!supportsSavedViewVisibilityCheck(menu)) {
    return;
  }
  const deadline = Date.now() + 10_000;
  while (Date.now() < deadline) {
    if (!(await isSavedViewLocatorVisible(menu))) {
      return;
    }
    await waitForSavedViewRetry(50);
  }
  throw new Error(`Saved-view action menu for ${surface} did not close`);
}

async function setCurrentSavedViewPreferenceAndWait(
  page: SavedViewPageLike,
  surface: string,
  options: {
    buttonTestId: string;
    expectedSheetRef: SheetRef;
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
  page: SavedViewPageLike,
  message: string,
): NonNullable<SavedViewPageLike["waitForRequest"]> {
  if (typeof page.waitForRequest !== "function") {
    throw new Error(message);
  }
  return (predicate) =>
    page.waitForRequest?.(predicate) as Promise<SavedViewNetworkRequestLike>;
}

function requireWaitForResponse(
  page: SavedViewPageLike,
  message: string,
): NonNullable<SavedViewPageLike["waitForResponse"]> {
  if (typeof page.waitForResponse !== "function") {
    throw new Error(message);
  }
  return (predicate) =>
    page.waitForResponse?.(predicate) as Promise<SavedViewNetworkResponseLike>;
}

function readRequestJSON(
  request: SavedViewNetworkRequestLike,
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
  response: SavedViewNetworkResponseLike,
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
  expectedSheetRef: SheetRef,
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
  expectedSheetRef: SheetRef,
) {
  const envelope = requireRecord(body, `${field} response envelope`);
  const data = requireRecord(envelope.data, `${field} response data`);
  assertSheetRef(data[field], expectedSheetRef, `${field} response`);
}

function assertSheetRef(actual: unknown, expected: SheetRef, label: string) {
  const record = requireRecord(actual, label);
  const matches =
    expected.kind === "extension_workspace"
      ? record.kind === expected.kind &&
        record.extension_profile_id === expected.extension_profile_id &&
        record.workspace_key === expected.workspace_key
      : record.kind === expected.kind && record.id === expected.id;
  if (!matches) {
    const expectedIdentity =
      expected.kind === "extension_workspace"
        ? `${expected.kind}:${expected.extension_profile_id}:${expected.workspace_key}`
        : `${expected.kind}:${expected.id}`;
    throw new Error(
      `${label} sheet ref mismatch: expected ${expectedIdentity}, got ${JSON.stringify(record)}`,
    );
  }
}

function requireRecord(value: unknown, label: string): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${label} must be a JSON object`);
  }
  return value as Record<string, unknown>;
}
