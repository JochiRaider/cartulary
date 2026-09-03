import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  fireEvent,
  render,
  renderHook,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import { defaultWorkbookLayoutState } from "../layout/workbookColumnLayout";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookSavedViewPort } from "../ports/WorkbookSavedViewPort";
import { useWorkbookSavedViewController } from "./useWorkbookSavedViewController";

const savedView = {
  saved_view_id: "saved-1",
  view_schema_id: "cartulary.view.timeline.v2",
  display_name: "Analyst timeline",
  scope: "private",
  query_json: { filters: [], sort: [] },
  layout_json: {
    layout_schema_id: "cartulary.layout.v1",
    column_order: ["row_version", "timeline.activity_synopsis_text"],
    hidden_field_keys: ["timeline.activity_synopsis_text"],
    column_widths: [{ field_key: "row_version", width_px: 96 }],
  },
  owner_user_id: "user-1",
  saved_view_version: 1,
} as const;

const savedViewPort: WorkbookSavedViewPort = {
  listPage: async () => ({
    kind: "accepted",
    value: { nextCursor: null, savedViews: [savedView] },
  }),
  create: async () => {
    throw new Error("Unexpected saved-view create");
  },
  patch: async () => {
    throw new Error("Unexpected saved-view patch");
  },
  delete: async () => {
    throw new Error("Unexpected saved-view delete");
  },
};

function emptySavedViewPort(
  overrides: Partial<WorkbookSavedViewPort> = {},
): WorkbookSavedViewPort {
  return {
    listPage: async () => ({
      kind: "accepted",
      value: { nextCursor: null, savedViews: [] },
    }),
    create: async () => {
      throw new Error("Unexpected saved-view create");
    },
    patch: async () => {
      throw new Error("Unexpected saved-view patch");
    },
    delete: async () => {
      throw new Error("Unexpected saved-view delete");
    },
    ...overrides,
  };
}

type SavedViewControllerOptions = Parameters<
  typeof useWorkbookSavedViewController
>[0];

function loadedSavedViews(
  resource: ReturnType<
    typeof useWorkbookSavedViewController
  >["snapshot"]["savedViewsResource"],
) {
  return resource.kind === "ready" || resource.kind === "invalid_selection"
    ? resource.savedViews
    : [];
}

function controllerOptions(
  port: WorkbookSavedViewPort,
  overrides: Partial<SavedViewControllerOptions> = {},
): SavedViewControllerOptions {
  return {
    activeContract: requireViewContract("cartulary.view.timeline.v2"),
    applyLayoutStateForSurface: vi.fn(),
    applyQueryStateForSurface: vi.fn(),
    applyWorkbookIdentity: vi.fn(),
    currentLayoutStateForSurface: () =>
      defaultWorkbookLayoutState(
        requireViewContract("cartulary.view.timeline.v2"),
      ),
    currentQueryStateForSurface: emptyWorkbookQueryState,
    savedViewPort: port,
    startupSheetRef: { kind: "view_schema", id: savedView.view_schema_id },
    ...overrides,
  };
}

function SavedViewControllerHarness({
  applyIdentity,
  applyLayout,
  applyQuery,
}: {
  readonly applyIdentity: SavedViewControllerOptions["applyWorkbookIdentity"];
  readonly applyLayout: SavedViewControllerOptions["applyLayoutStateForSurface"];
  readonly applyQuery: SavedViewControllerOptions["applyQueryStateForSurface"];
}) {
  const controller = useWorkbookSavedViewController({
    activeContract: requireViewContract("cartulary.view.timeline.v2"),
    applyLayoutStateForSurface: applyLayout,
    applyQueryStateForSurface: applyQuery,
    applyWorkbookIdentity: applyIdentity,
    currentLayoutStateForSurface: () =>
      defaultWorkbookLayoutState(
        requireViewContract("cartulary.view.timeline.v2"),
      ),
    currentQueryStateForSurface: emptyWorkbookQueryState,
    savedViewPort,
    startupSheetRef: { kind: "view_schema", id: savedView.view_schema_id },
  });
  const loaded = loadedSavedViews(controller.snapshot.savedViewsResource);
  return (
    <>
      <button
        disabled={loaded.length === 0}
        onClick={() =>
          controller.commands.selectSavedView(loaded[0] ?? savedView)
        }
        type="button"
      >
        Select Saved View
      </button>
      <output aria-label="saved-view-count">{loaded.length}</output>
    </>
  );
}

describe("useWorkbookSavedViewController", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("owns saved-view loading and query/identity precedence", async () => {
    const applyIdentity =
      vi.fn<SavedViewControllerOptions["applyWorkbookIdentity"]>();
    const applyQuery =
      vi.fn<SavedViewControllerOptions["applyQueryStateForSurface"]>();
    const applyLayout =
      vi.fn<SavedViewControllerOptions["applyLayoutStateForSurface"]>();
    render(
      <SavedViewControllerHarness
        applyIdentity={applyIdentity}
        applyLayout={applyLayout}
        applyQuery={applyQuery}
      />,
    );

    await waitFor(() =>
      expect(screen.getByLabelText("saved-view-count").textContent).toBe("1"),
    );
    fireEvent.click(screen.getByRole("button", { name: "Select Saved View" }));
    expect(applyQuery).toHaveBeenCalledWith(
      savedView.view_schema_id,
      expect.any(Object),
    );
    expect(applyLayout).toHaveBeenCalledWith(
      savedView.view_schema_id,
      expect.objectContaining({
        columnOrder: ["row_version", "timeline.activity_synopsis_text"],
        hiddenFieldKeys: ["timeline.activity_synopsis_text"],
      }),
    );
    expect(applyIdentity).toHaveBeenCalledWith(
      {
        sheetRef: { kind: "saved_view", id: savedView.saved_view_id },
        viewSchemaId: savedView.view_schema_id,
      },
      { reloadSheet: true },
    );
  });

  it("accumulates explicit pages without publishing a partial list", async () => {
    const secondPage =
      deferred<Awaited<ReturnType<WorkbookSavedViewPort["listPage"]>>>();
    const secondSavedView = {
      ...savedView,
      display_name: "Second saved view",
      saved_view_id: "saved-2",
    };
    const listPage = vi
      .fn<WorkbookSavedViewPort["listPage"]>()
      .mockResolvedValueOnce({
        kind: "accepted",
        value: { nextCursor: "cursor-2", savedViews: [savedView] },
      })
      .mockReturnValueOnce(secondPage.promise);
    const port = emptySavedViewPort({ listPage });
    const { result } = renderHook(() =>
      useWorkbookSavedViewController(controllerOptions(port)),
    );

    await waitFor(() => expect(listPage).toHaveBeenCalledTimes(2));
    expect(result.current.snapshot.savedViewsResource).toEqual({
      kind: "loading",
    });
    await act(async () => {
      secondPage.resolve({
        kind: "accepted",
        value: { nextCursor: null, savedViews: [secondSavedView] },
      });
      await secondPage.promise;
    });
    await waitFor(() =>
      expect(
        loadedSavedViews(result.current.snapshot.savedViewsResource).map(
          (view) => view.saved_view_id,
        ),
      ).toEqual(["saved-1", "saved-2"]),
    );
    expect(listPage.mock.calls.map(([input]) => input.cursorToken)).toEqual([
      null,
      "cursor-2",
    ]);
  });

  it("fails closed on cyclic cursors and duplicate resources", async () => {
    const listPage = vi
      .fn<WorkbookSavedViewPort["listPage"]>()
      .mockResolvedValueOnce({
        kind: "accepted",
        value: { nextCursor: "cursor-cycle", savedViews: [savedView] },
      })
      .mockResolvedValueOnce({
        kind: "accepted",
        value: { nextCursor: "cursor-cycle", savedViews: [savedView] },
      });
    const port = emptySavedViewPort({ listPage });
    const { result } = renderHook(() =>
      useWorkbookSavedViewController(controllerOptions(port)),
    );

    await waitFor(() => expect(listPage).toHaveBeenCalledTimes(2));
    await waitFor(() =>
      expect(result.current.snapshot.savedViewsResource).toEqual({
        kind: "unavailable",
        message: "Saved-view listing returned a duplicate resource.",
      }),
    );
  });

  it("rejects a late create after the incident-bound port is replaced", async () => {
    const pendingCreate =
      deferred<Awaited<ReturnType<WorkbookSavedViewPort["create"]>>>();
    const firstPort = emptySavedViewPort({
      create: vi.fn(() => pendingCreate.promise),
    });
    const secondPort = emptySavedViewPort();
    const { result, rerender } = renderHook(
      ({ port }: { readonly port: WorkbookSavedViewPort }) =>
        useWorkbookSavedViewController(controllerOptions(port)),
      { initialProps: { port: firstPort } },
    );

    let createPromise!: ReturnType<
      typeof result.current.commands.createSavedView
    >;
    act(() => {
      createPromise = result.current.commands.createSavedView({
        displayName: "Late create",
        scope: "private",
      });
    });
    rerender({ port: secondPort });
    await act(async () => {
      pendingCreate.resolve({ kind: "accepted", value: savedView });
      await expect(createPromise).rejects.toThrow(
        "Saved-view create was superseded.",
      );
    });
    await waitFor(() =>
      expect(result.current.snapshot.savedViewsResource).toEqual({
        kind: "ready",
        savedViews: [],
      }),
    );
  });

  it("admits only one mutation action at a time", async () => {
    const pendingCreate =
      deferred<Awaited<ReturnType<WorkbookSavedViewPort["create"]>>>();
    const port = emptySavedViewPort({
      create: vi.fn(() => pendingCreate.promise),
    });
    const { result } = renderHook(() =>
      useWorkbookSavedViewController(controllerOptions(port)),
    );

    let firstPromise!: ReturnType<
      typeof result.current.commands.createSavedView
    >;
    let secondPromise!: ReturnType<
      typeof result.current.commands.createSavedView
    >;
    act(() => {
      firstPromise = result.current.commands.createSavedView({
        displayName: "First",
        scope: "private",
      });
      secondPromise = result.current.commands.createSavedView({
        displayName: "Second",
        scope: "private",
      });
    });
    await expect(secondPromise).rejects.toThrow(
      "Another saved-view action is already in progress.",
    );
    await act(async () => {
      pendingCreate.resolve({ kind: "accepted", value: savedView });
      await firstPromise;
    });
  });

  it("ignores a late accepted mutation after the active surface changes", async () => {
    const pendingCreate =
      deferred<Awaited<ReturnType<WorkbookSavedViewPort["create"]>>>();
    const port = emptySavedViewPort({
      create: vi.fn(() => pendingCreate.promise),
    });
    const applyWorkbookIdentity = vi.fn();
    const initialActiveViewSchemaId: string = savedView.view_schema_id;
    const { result, rerender } = renderHook(
      ({ activeViewSchemaId }: { readonly activeViewSchemaId: string }) =>
        useWorkbookSavedViewController(
          controllerOptions(port, {
            activeContract: requireViewContract(activeViewSchemaId),
            applyWorkbookIdentity,
          }),
        ),
      { initialProps: { activeViewSchemaId: initialActiveViewSchemaId } },
    );

    let createPromise!: ReturnType<
      typeof result.current.commands.createSavedView
    >;
    act(() => {
      createPromise = result.current.commands.createSavedView({
        displayName: "Late surface create",
        scope: "private",
      });
    });
    rerender({ activeViewSchemaId: "cartulary.view.evidence.v1" });
    await act(async () => {
      pendingCreate.resolve({ kind: "accepted", value: savedView });
      await expect(createPromise).rejects.toThrow(
        "Saved-view create was superseded.",
      );
    });
    expect(applyWorkbookIdentity).not.toHaveBeenCalled();
  });

  it("routes semantic access loss and publishes no rejected list state", async () => {
    const onIncidentAccessLost = vi.fn();
    const port = emptySavedViewPort({
      listPage: async () => ({
        kind: "rejected",
        failure: {
          kind: "authorization_lost",
          message: "Saved views are unavailable.",
        },
      }),
    });
    const { result } = renderHook(() =>
      useWorkbookSavedViewController(
        controllerOptions(port, { onIncidentAccessLost }),
      ),
    );

    await waitFor(() => expect(onIncidentAccessLost).toHaveBeenCalledOnce());
    expect(result.current.snapshot.savedViewsResource).toEqual({
      kind: "unavailable",
      message: "Saved views are unavailable.",
    });
  });

  it("publishes an invalid-selection resource only after complete loading", async () => {
    const { result } = renderHook(() =>
      useWorkbookSavedViewController(
        controllerOptions(savedViewPort, {
          startupSheetRef: { kind: "saved_view", id: "missing-saved-view" },
        }),
      ),
    );

    expect(result.current.snapshot.savedViewsResource).toEqual({
      kind: "loading",
    });
    await waitFor(() =>
      expect(result.current.snapshot.savedViewsResource).toEqual({
        kind: "invalid_selection",
        savedViews: [savedView],
        selectedSavedViewId: "missing-saved-view",
      }),
    );
  });
});
