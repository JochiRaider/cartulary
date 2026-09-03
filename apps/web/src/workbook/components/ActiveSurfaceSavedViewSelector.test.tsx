import {
  savedViewActionMenuTriggerTestId,
  savedViewCreateButtonTestId,
  savedViewStatusTestId,
  savedViewUpdateButtonTestId,
} from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import type { WorkbookSavedViewsResource } from "../models/workbookSavedViewControl";
import type { SavedViewResource } from "../models/workbookSavedViews";
import { ActiveSurfaceSavedViewSelector } from "./ActiveSurfaceSavedViewSelector";

const surface = "cartulary.view.timeline.v2";
const selectedSavedView: SavedViewResource = {
  saved_view_id: "saved-1",
  view_schema_id: surface,
  display_name: "Timeline view",
  scope: "private",
  query_json: {},
  layout_json: {},
  owner_user_id: "user-1",
  saved_view_version: 3,
};

function renderSelector({
  activeViewSchemaId = surface,
  currentIncidentRole = "member",
  currentUserId = "user-1",
  resource = { kind: "ready", savedViews: [selectedSavedView] },
  selected = selectedSavedView,
  onCreateSavedView = vi.fn(async () => selectedSavedView),
  onSelectBaseSurface = vi.fn(),
}: {
  readonly activeViewSchemaId?: string;
  readonly currentIncidentRole?: string | null;
  readonly currentUserId?: string | null;
  readonly resource?: WorkbookSavedViewsResource;
  readonly selected?: SavedViewResource | null;
  readonly onCreateSavedView?: (
    input: Readonly<{ displayName: string; scope: "private" | "shared" }>,
  ) => Promise<SavedViewResource>;
  readonly onSelectBaseSurface?: (viewSchemaId: string) => void;
} = {}) {
  return render(
    <ActiveSurfaceSavedViewSelector
      activeViewSchemaId={activeViewSchemaId}
      chromeMode="base"
      currentIncidentRole={currentIncidentRole}
      currentUserId={currentUserId}
      isModified
      onCreateSavedView={onCreateSavedView}
      onDeleteSavedView={vi.fn(async () => undefined)}
      onDuplicateSavedView={vi.fn(async () => selectedSavedView)}
      onResetToSavedView={vi.fn()}
      onSelectBaseSurface={onSelectBaseSurface}
      onSelectSavedView={vi.fn()}
      onSetDefaultSheetRef={vi.fn(async () => undefined)}
      onSetHomeSheetRef={vi.fn(async () => undefined)}
      onUpdateSavedView={vi.fn(async () => selectedSavedView)}
      savedViewsResource={resource}
      selectedSheetRef={
        selected === null
          ? { kind: "view_schema", id: activeViewSchemaId }
          : { kind: "saved_view", id: selected.saved_view_id }
      }
    />,
  );
}

describe("ActiveSurfaceSavedViewSelector", () => {
  afterEach(() => {
    cleanup();
  });

  it("renders explicit loading and unavailable feedback", () => {
    const { rerender } = render(
      <ActiveSurfaceSavedViewSelector
        activeViewSchemaId={surface}
        chromeMode="base"
        currentIncidentRole="member"
        currentUserId="user-1"
        onCreateSavedView={vi.fn(async () => selectedSavedView)}
        onDeleteSavedView={vi.fn(async () => undefined)}
        onDuplicateSavedView={vi.fn(async () => selectedSavedView)}
        onResetToSavedView={vi.fn()}
        onSelectBaseSurface={vi.fn()}
        onSelectSavedView={vi.fn()}
        onSetDefaultSheetRef={vi.fn(async () => undefined)}
        onSetHomeSheetRef={vi.fn(async () => undefined)}
        onUpdateSavedView={vi.fn(async () => selectedSavedView)}
        savedViewsResource={{ kind: "loading" }}
        selectedSheetRef={{ kind: "view_schema", id: surface }}
      />,
    );
    expect(
      (screen.getByLabelText("Saved view") as HTMLSelectElement).disabled,
    ).toBe(true);
    expect(screen.getByTestId(savedViewStatusTestId(surface)).textContent).toBe(
      "Loading saved views…",
    );

    rerender(
      <ActiveSurfaceSavedViewSelector
        activeViewSchemaId={surface}
        chromeMode="base"
        currentIncidentRole="member"
        currentUserId="user-1"
        onCreateSavedView={vi.fn(async () => selectedSavedView)}
        onDeleteSavedView={vi.fn(async () => undefined)}
        onDuplicateSavedView={vi.fn(async () => selectedSavedView)}
        onResetToSavedView={vi.fn()}
        onSelectBaseSurface={vi.fn()}
        onSelectSavedView={vi.fn()}
        onSetDefaultSheetRef={vi.fn(async () => undefined)}
        onSetHomeSheetRef={vi.fn(async () => undefined)}
        onUpdateSavedView={vi.fn(async () => selectedSavedView)}
        savedViewsResource={{
          kind: "unavailable",
          message: "Saved-view listing failed.",
        }}
        selectedSheetRef={{ kind: "view_schema", id: surface }}
      />,
    );
    expect(
      (screen.getByLabelText("Saved view") as HTMLSelectElement).disabled,
    ).toBe(true);
    expect(screen.getByText("Saved-view listing failed.")).not.toBeNull();
  });

  it("uses a labelled dialog, removes redundant actions, and restores focus on Escape", async () => {
    renderSelector();
    const trigger = screen.getByTestId(
      savedViewActionMenuTriggerTestId(surface),
    );
    fireEvent.click(trigger);

    const dialog = screen.getByRole("dialog", { name: "Saved view" });
    expect(dialog).toBeInstanceOf(HTMLElement);
    await waitFor(() =>
      expect(document.activeElement).toBe(
        screen.getByLabelText("Saved view name"),
      ),
    );
    expect(screen.queryByRole("button", { name: "Rename" })).toBeNull();
    expect(screen.queryByRole("button", { name: "Manage sharing" })).toBeNull();

    fireEvent.keyDown(screen.getByLabelText("Saved view name"), {
      key: "Escape",
    });
    expect(screen.queryByRole("dialog", { name: "Saved view" })).toBeNull();
    expect(document.activeElement).toBe(trigger);
  });

  it("admits only one action and ignores its completion after a surface switch", async () => {
    const pendingCreate = deferred<SavedViewResource>();
    const onCreateSavedView = vi.fn(() => pendingCreate.promise);
    const { rerender } = renderSelector({ onCreateSavedView, selected: null });
    fireEvent.click(
      screen.getByTestId(savedViewActionMenuTriggerTestId(surface)),
    );
    const create = screen.getByTestId(savedViewCreateButtonTestId(surface));
    fireEvent.click(create);
    fireEvent.click(create);
    expect(onCreateSavedView).toHaveBeenCalledOnce();

    const nextSurface = "cartulary.view.evidence.v1";
    rerender(
      <ActiveSurfaceSavedViewSelector
        activeViewSchemaId={nextSurface}
        chromeMode="base"
        currentIncidentRole="member"
        currentUserId="user-1"
        onCreateSavedView={onCreateSavedView}
        onDeleteSavedView={vi.fn(async () => undefined)}
        onDuplicateSavedView={vi.fn(async () => selectedSavedView)}
        onResetToSavedView={vi.fn()}
        onSelectBaseSurface={vi.fn()}
        onSelectSavedView={vi.fn()}
        onSetDefaultSheetRef={vi.fn(async () => undefined)}
        onSetHomeSheetRef={vi.fn(async () => undefined)}
        onUpdateSavedView={vi.fn(async () => selectedSavedView)}
        savedViewsResource={{ kind: "ready", savedViews: [] }}
        selectedSheetRef={{ kind: "view_schema", id: nextSurface }}
      />,
    );
    await act(async () => {
      pendingCreate.resolve(selectedSavedView);
      await pendingCreate.promise;
    });
    expect(
      screen.getByTestId(savedViewStatusTestId(nextSurface)).textContent,
    ).not.toContain("created");
  });

  it("falls back from an invalid selection with inline notice", async () => {
    const onSelectBaseSurface = vi.fn();
    renderSelector({
      onSelectBaseSurface,
      resource: {
        kind: "invalid_selection",
        savedViews: [selectedSavedView],
        selectedSavedViewId: "removed-view",
      },
      selected: {
        ...selectedSavedView,
        saved_view_id: "removed-view",
      },
    });
    await waitFor(() =>
      expect(onSelectBaseSurface).toHaveBeenCalledWith(surface),
    );
    expect(screen.getByText(/no longer available/i)).not.toBeNull();
  });

  it("revalidates owner and role before update dispatch", () => {
    renderSelector({
      currentIncidentRole: "member",
      currentUserId: "another-user",
    });
    fireEvent.click(
      screen.getByTestId(savedViewActionMenuTriggerTestId(surface)),
    );
    expect(
      (
        screen.getByTestId(
          savedViewUpdateButtonTestId(surface, selectedSavedView.saved_view_id),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
  });
});
