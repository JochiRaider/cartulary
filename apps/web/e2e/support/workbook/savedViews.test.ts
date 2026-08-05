// @vitest-environment jsdom

import {
  savedViewActionMenuTestId,
  savedViewActionMenuTriggerTestId,
  savedViewSelectorTestId,
  savedViewSetHomeButtonTestId,
} from "@cartulary/ui-contracts";
import { timelineViewSchemaId } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";

import {
  readSavedViewSelectionState,
  selectSavedView,
  setCurrentSavedViewAsHomeAndWait,
} from "./savedViews";

describe("saved-view workbook support", () => {
  it("selects by stable ID and reads identity from data attributes", async () => {
    const selector = document.createElement("select");
    selector.dataset.activeViewSchemaId = timelineViewSchemaId;
    selector.dataset.selectedSheetRefKind = "view_schema";
    selector.dataset.selectedSavedViewId = "";
    const page = {
      getByTestId(testId: string) {
        expect(testId).toBe(savedViewSelectorTestId(timelineViewSchemaId));
        return {
          click: async () => undefined,
          evaluate: async (
            callback: (element: Element, argument?: unknown) => unknown,
            argument?: unknown,
          ) => callback(selector, argument),
          fill: async () => undefined,
          selectOption: async (savedViewId: string | readonly string[]) => {
            selector.dataset.selectedSheetRefKind = "saved_view";
            selector.dataset.selectedSavedViewId = String(savedViewId);
          },
        };
      },
    };

    await selectSavedView(page, timelineViewSchemaId, "saved-view-1");

    await expect(
      readSavedViewSelectionState(page, timelineViewSchemaId),
    ).resolves.toEqual({
      activeViewSchemaId: timelineViewSchemaId,
      selectedSavedViewId: "saved-view-1",
      selectedSheetRefKind: "saved_view",
    });
  });

  it("matches the exact preference route and preserves request/response envelopes", async () => {
    type TestRequest = {
      method: () => string;
      postDataJSON: () => unknown;
      url: () => string;
    };
    type TestResponse = {
      json: () => Promise<unknown>;
      ok: () => boolean;
      request: () => TestRequest;
      status: () => number;
      url: () => string;
    };
    const incidentId = "incident-1";
    const path = `/api/v1/incidents/${incidentId}/workbook-preferences/me`;
    const expectedSheetRef = {
      id: "saved-view-1",
      kind: "saved_view" as const,
    };
    let requestPredicate: ((request: TestRequest) => boolean) | undefined;
    let responsePredicate: ((response: TestResponse) => boolean) | undefined;
    let resolveRequest: ((request: TestRequest) => void) | undefined;
    let resolveResponse: ((response: TestResponse) => void) | undefined;
    const request: TestRequest = {
      method: () => "PUT",
      postDataJSON: () => ({ home_sheet_ref: expectedSheetRef }),
      url: () => `https://cartulary.test${path}`,
    };
    const response: TestResponse = {
      json: async () => ({
        data: { home_sheet_ref: expectedSheetRef },
        meta: { request_id: "request-1" },
      }),
      ok: () => true,
      request: () => request,
      status: () => 200,
      url: () => `https://cartulary.test${path}`,
    };
    let menuOpen = false;
    const page = {
      getByTestId(testId: string) {
        return {
          click: async () => {
            if (
              testId === savedViewActionMenuTriggerTestId(timelineViewSchemaId)
            ) {
              menuOpen = true;
            }
            if (testId === savedViewSetHomeButtonTestId(timelineViewSchemaId)) {
              menuOpen = false;
              if (requestPredicate?.(request)) {
                resolveRequest?.(request);
              }
              if (responsePredicate?.(response)) {
                resolveResponse?.(response);
              }
            }
          },
          fill: async () => undefined,
          isVisible: async () =>
            testId === savedViewActionMenuTestId(timelineViewSchemaId)
              ? menuOpen
              : true,
        };
      },
      waitForRequest(predicate: (request: TestRequest) => boolean) {
        requestPredicate = predicate;
        return new Promise<TestRequest>((resolve) => {
          resolveRequest = resolve;
        });
      },
      waitForResponse(predicate: (response: TestResponse) => boolean) {
        responsePredicate = predicate;
        return new Promise<TestResponse>((resolve) => {
          resolveResponse = resolve;
        });
      },
    };

    await expect(
      setCurrentSavedViewAsHomeAndWait(page, timelineViewSchemaId, {
        expectedSheetRef,
        incidentId,
      }),
    ).resolves.toEqual({
      field: "home_sheet_ref",
      requestBody: { home_sheet_ref: expectedSheetRef },
      responseBody: {
        data: { home_sheet_ref: expectedSheetRef },
        meta: { request_id: "request-1" },
      },
      status: 200,
    });
    expect(savedViewActionMenuTriggerTestId(timelineViewSchemaId)).not.toBe(
      savedViewSetHomeButtonTestId(timelineViewSchemaId),
    );
  });

  it("fails before waiting when request interception is unavailable", async () => {
    const page = {
      getByTestId() {
        return {
          click: async () => undefined,
          fill: async () => undefined,
        };
      },
    };

    await expect(
      setCurrentSavedViewAsHomeAndWait(page, timelineViewSchemaId, {
        expectedSheetRef: { id: "saved-view-1", kind: "saved_view" },
        incidentId: "incident-1",
      }),
    ).rejects.toThrow(/requires page\.waitForRequest\(\) support/);
  });
});
