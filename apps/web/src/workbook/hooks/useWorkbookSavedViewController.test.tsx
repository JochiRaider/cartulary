import { requireViewContract } from "@cartulary/view-contracts";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { jsonResponse } from "../../testing/fetchMockTestSupport";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import { useWorkbookSavedViewController } from "./useWorkbookSavedViewController";

const savedView = {
  saved_view_id: "saved-1",
  view_schema_id: "cartulary.view.timeline.v2",
  display_name: "Analyst timeline",
  scope: "private",
  query_json: { filters: [], sort: [] },
  layout_json: {},
  owner_user_id: "user-1",
  saved_view_version: 1,
} as const;

type SavedViewControllerOptions = Parameters<
  typeof useWorkbookSavedViewController
>[0];

function SavedViewControllerHarness({
  applyIdentity,
  applyQuery,
}: {
  readonly applyIdentity: SavedViewControllerOptions["applyWorkbookIdentity"];
  readonly applyQuery: SavedViewControllerOptions["applyQueryStateForSurface"];
}) {
  const controller = useWorkbookSavedViewController({
    activeContract: requireViewContract("cartulary.view.timeline.v2"),
    applyQueryStateForSurface: applyQuery,
    applyWorkbookIdentity: applyIdentity,
    currentQueryStateForSurface: emptyWorkbookQueryState,
    incidentId: "incident-1",
    startupSheetRef: { kind: "view_schema", id: savedView.view_schema_id },
  });
  return (
    <>
      <button
        disabled={controller.snapshot.savedViews.length === 0}
        onClick={() =>
          controller.commands.selectSavedView(
            controller.snapshot.savedViews[0] ?? savedView,
          )
        }
        type="button"
      >
        Select Saved View
      </button>
      <output aria-label="saved-view-count">
        {controller.snapshot.savedViews.length}
      </output>
    </>
  );
}

describe("useWorkbookSavedViewController", () => {
  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("owns saved-view loading and query/identity precedence", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(
        jsonResponse({
          data: { saved_views: [savedView] },
          meta: { paging: { has_more: false, next_cursor: null } },
        }),
      ),
    );
    const applyIdentity =
      vi.fn<SavedViewControllerOptions["applyWorkbookIdentity"]>();
    const applyQuery =
      vi.fn<SavedViewControllerOptions["applyQueryStateForSurface"]>();
    render(
      <SavedViewControllerHarness
        applyIdentity={applyIdentity}
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
    expect(applyIdentity).toHaveBeenCalledWith(
      {
        sheetRef: { kind: "saved_view", id: savedView.saved_view_id },
        viewSchemaId: savedView.view_schema_id,
      },
      { reloadSheet: true },
    );
  });
});
