import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup, renderHook, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import {
  createSavedViewTestController,
  savedViewTestResource,
} from "../../testing/workbookSavedViewTestSupport";
import { defaultWorkbookLayoutState } from "../layout/workbookColumnLayout";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import type { SavedViewResource } from "../models/workbookSavedViews";
import type {
  SavedViewResult,
  WorkbookSavedViewPort,
} from "../ports/WorkbookSavedViewPort";
import { useWorkbookSavedViewController } from "./useWorkbookSavedViewController";

const saved = savedViewTestResource();
const contract = requireViewContract(saved.view_schema_id);
function setup(overrides: Partial<WorkbookSavedViewPort> = {}) {
  const port: WorkbookSavedViewPort = {
    getResource: vi.fn<WorkbookSavedViewPort["getResource"]>(
      async ({ savedViewId }) => {
        const resource = [saved].find((r) => r.saved_view_id === savedViewId);
        return resource
          ? { kind: "accepted", value: resource }
          : {
              kind: "rejected",
              failure: { kind: "unavailable_target", message: "Unavailable" },
            };
      },
    ),
    listPage: vi.fn<WorkbookSavedViewPort["listPage"]>(async () => ({
      kind: "accepted",
      value: { nextCursor: null, savedViews: [saved] },
    })),
    create: vi.fn<WorkbookSavedViewPort["create"]>(),
    patch: vi.fn<WorkbookSavedViewPort["patch"]>(),
    delete: vi.fn<WorkbookSavedViewPort["delete"]>(),
    ...overrides,
  };
  const controller = createSavedViewTestController(port);
  const options: Parameters<typeof useWorkbookSavedViewController>[0] = {
    controller,
    bindWorkbook: controller.setBinding,
    incidentId: saved.incident_id,
    activeContract: contract,
    startupSheetRef: { kind: "saved_view", id: saved.saved_view_id },
    selectionGeneration: 0,
    authorizationRecovered: vi.fn(),
    applyWorkbookIdentity: vi.fn(),
    applyLayoutStateForSurface: vi.fn(),
    applyQueryStateForSurface: vi.fn(),
    currentLayoutStateForSurface: () => defaultWorkbookLayoutState(contract),
    currentQueryStateForSurface: emptyWorkbookQueryState,
  };
  const hook = renderHook((input) => useWorkbookSavedViewController(input), {
    initialProps: options,
  });
  return { ...hook, options, controller, port };
}
const subject = {
  viewSchemaId: saved.view_schema_id,
  savedViewId: saved.saved_view_id,
  savedViewVersion: saved.saved_view_version,
};

describe("useWorkbookSavedViewController", () => {
  afterEach(cleanup);
  it("retains selected configuration while discovery is incomplete or fails", async () => {
    const listPage = vi
      .fn<WorkbookSavedViewPort["listPage"]>()
      .mockResolvedValue({
        kind: "rejected",
        failure: { kind: "transport", message: "Offline" },
      });
    const h = setup({ listPage });
    await waitFor(() =>
      expect(
        h.result.current.snapshot.savedViewsResource.selectedSavedView,
      ).toEqual(saved),
    );
    expect(listPage).not.toHaveBeenCalled();
    act(() => h.controller.openDiscovery());
    await waitFor(() =>
      expect(h.controller.getSnapshot().discovery.problem).not.toBeNull(),
    );
    expect(
      h.result.current.snapshot.savedViewsResource.selectedSavedView,
    ).toEqual(saved);
    expect(h.options.applyWorkbookIdentity).not.toHaveBeenCalled();
    h.controller.dispose();
  });
  it("selects saved query and layout explicitly while Reset never changes identity", async () => {
    const h = setup();
    await waitFor(() =>
      expect(h.result.current.snapshot.savedViewsResource.kind).toBe("ready"),
    );
    act(() => h.controller.run({ kind: "reset" }, subject));
    expect(h.options.applyQueryStateForSurface).toHaveBeenCalledOnce();
    expect(h.options.applyLayoutStateForSurface).toHaveBeenCalledOnce();
    expect(h.options.applyWorkbookIdentity).not.toHaveBeenCalled();
    await act(async () => {
      await h.controller.activateResource(saved.saved_view_id);
    });
    expect(h.options.applyWorkbookIdentity).toHaveBeenCalledWith(
      {
        sheetRef: { kind: "saved_view", id: saved.saved_view_id },
        viewSchemaId: saved.view_schema_id,
      },
      { reloadSheet: true },
    );
    expect(h.port.patch).not.toHaveBeenCalled();
    h.controller.dispose();
  });
  it("captures exact working changes and preserves newer working edits after acknowledgement", async () => {
    const pending = deferred<SavedViewResult<SavedViewResource>>();
    const h = setup({ patch: vi.fn(() => pending.promise) });
    await waitFor(() =>
      expect(h.result.current.snapshot.savedViewsResource.kind).toBe("ready"),
    );
    act(() => h.controller.changeDraft(subject, { displayName: "Renamed" }));
    act(() => h.controller.run({ kind: "update" }, subject));
    await waitFor(() => expect(h.port.patch).toHaveBeenCalledOnce());
    h.rerender({
      ...h.options,
      currentQueryStateForSurface: () => ({
        ...emptyWorkbookQueryState(),
        groupBy: "timeline.capture_state",
      }),
    });
    await act(async () =>
      pending.resolve({
        kind: "accepted",
        value: {
          ...saved,
          display_name: "Renamed",
          saved_view_version: 2,
          updated_at: "2026-08-01T00:01:00Z",
        },
      }),
    );
    expect(h.controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(h.options.applyQueryStateForSurface).not.toHaveBeenCalled();
    expect(h.options.applyWorkbookIdentity).not.toHaveBeenCalled();
    h.controller.dispose();
  });
  it("detaches an unmounted workbook while retaining a valid late receipt", async () => {
    const pending = deferred<SavedViewResult<SavedViewResource>>();
    const h = setup({ create: vi.fn(() => pending.promise) });
    await waitFor(() =>
      expect(h.result.current.snapshot.savedViewsResource.kind).toBe("ready"),
    );
    act(() => h.controller.run({ kind: "create" }, subject));
    await waitFor(() => expect(h.port.create).toHaveBeenCalledOnce());
    h.unmount();
    await act(async () =>
      pending.resolve({
        kind: "accepted",
        value: savedViewTestResource({ saved_view_id: "new-view" }),
      }),
    );
    expect(h.controller.getSnapshot().operation.kind).toBe("confirmed");
    expect(h.options.applyWorkbookIdentity).not.toHaveBeenCalled();
    h.controller.dispose();
  });
});
