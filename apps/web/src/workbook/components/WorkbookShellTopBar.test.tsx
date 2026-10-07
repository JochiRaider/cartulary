import {
  surfaceTabTestId,
  workbookSurfacesMenuOptionTestId,
  workbookSurfacesMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { type ComponentProps, createRef } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { emptyWorkbookPresence } from "../collaboration/workbookPresencePresentation";
import {
  evidenceViewSchemaId,
  hostsViewSchemaId,
  indicatorsViewSchemaId,
  notesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { WorkbookShellTopBar } from "./WorkbookShellTopBar";

afterEach(cleanup);

function setup(
  overrides: Partial<ComponentProps<typeof WorkbookShellTopBar>> = {},
) {
  const props: ComponentProps<typeof WorkbookShellTopBar> = {
    account: {
      applicationMenu: null,
      displayName: "Selector actor",
      title: "Actor",
    },
    activeSurfaceFocusRef: createRef<HTMLElement>(),
    activeSystemSurfaceTitle: null,
    collaboration: {
      presence: emptyWorkbookPresence,
      connectionId: null,
      status: "connected",
    },
    incidentIdentity: null,
    incidentIdentityError: null,
    layout: {
      blockMode: "base_height",
      chromeMode: "base",
      density: "compact",
      showStatusPresence: false,
    },
    networkAnalysisActive: false,
    networkAnalysisAvailable: false,
    onSelectNetworkAnalysis: vi.fn(),
    onSelectSurface: vi.fn(),
    surface: evidenceViewSchemaId,
    ...overrides,
  };
  const view = render(<WorkbookShellTopBar {...props} />);
  return {
    props,
    user: userEvent.setup(),
    update: (changes: Partial<typeof props>) => {
      Object.assign(props, changes);
      view.rerender(<WorkbookShellTopBar {...props} />);
    },
  };
}

async function enter(user: ReturnType<typeof userEvent.setup>) {
  await user.tab();
  expect(document.activeElement?.getAttribute("aria-label")).toMatch(
    /^Incident details:/,
  );
  await user.tab();
}

function tab(id: string) {
  return screen.getByTestId(surfaceTabTestId(id));
}

describe("Workbook desktop surface selector", () => {
  it("enters once and navigates canonical surfaces without activation", async () => {
    const h = setup();
    await enter(h.user);
    expect(document.activeElement).toBe(tab(evidenceViewSchemaId));
    const controls = within(
      screen.getByRole("tablist", { name: "Built-in workbook surfaces" }),
    ).getAllByRole("tab");
    expect(controls.map((control) => control.dataset.viewSchemaId)).toEqual(
      requiredBuiltInWorkbookSurfaceIds,
    );
    expect(controls.filter((control) => control.tabIndex === 0)).toEqual([
      tab(evidenceViewSchemaId),
    ]);
    await h.user.keyboard("{ArrowRight}{ArrowRight}");
    expect(document.activeElement).toBe(tab(timelineViewSchemaId));
    await h.user.keyboard("{ArrowLeft}");
    expect(document.activeElement).toBe(tab(notesViewSchemaId));
    await h.user.keyboard("{Home}{ArrowRight}");
    expect(document.activeElement).toBe(tab(hostsViewSchemaId));
    await h.user.keyboard("{End}{Escape}");
    expect(document.activeElement).toBe(tab(notesViewSchemaId));
    expect(
      controls.filter(
        (control) => control.getAttribute("aria-selected") === "true",
      ),
    ).toEqual([tab(evidenceViewSchemaId)]);
    expect(h.props.onSelectSurface).not.toHaveBeenCalled();
    await h.user.tab();
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "More views" }),
    );
    await h.user.tab({ shift: true });
    expect(document.activeElement).toBe(tab(evidenceViewSchemaId));
    await h.user.tab({ shift: true });
    expect(document.activeElement?.getAttribute("aria-label")).toMatch(
      /^Incident details:/,
    );
  });

  it("activates Enter Space and pointer exactly once and follows actual focus", async () => {
    const h = setup();
    await enter(h.user);
    await h.user.keyboard("{ArrowRight}{Enter}");
    expect(h.props.onSelectSurface).toHaveBeenCalledTimes(1);
    expect(h.props.onSelectSurface).toHaveBeenLastCalledWith(
      notesViewSchemaId,
      { focusFirstGridTarget: true },
    );
    await h.user.keyboard("{Home} ");
    expect(h.props.onSelectSurface).toHaveBeenCalledTimes(2);
    expect(h.props.onSelectSurface).toHaveBeenLastCalledWith(
      timelineViewSchemaId,
      { focusFirstGridTarget: true },
    );
    await h.user.click(tab(hostsViewSchemaId));
    expect(h.props.onSelectSurface).toHaveBeenCalledTimes(3);
    await h.user.keyboard("{ArrowRight}");
    expect(document.activeElement?.getAttribute("data-view-schema-id")).toBe(
      requiredBuiltInWorkbookSurfaceIds[2],
    );
    expect(h.props.onSelectSurface).toHaveBeenCalledTimes(3);
    h.update({ surface: timelineViewSchemaId });
    expect(document.activeElement?.getAttribute("data-view-schema-id")).toBe(
      requiredBuiltInWorkbookSurfaceIds[2],
    );
    await h.user.tab();
    await h.user.tab({ shift: true });
    expect(document.activeElement).toBe(tab(timelineViewSchemaId));
  });

  async function enterWithoutSelection(
    state: Partial<ComponentProps<typeof WorkbookShellTopBar>>,
  ) {
    const h = setup(state);
    await enter(h.user);
    expect(document.activeElement).toBe(tab(timelineViewSchemaId));
    for (const id of requiredBuiltInWorkbookSurfaceIds)
      expect(tab(id).getAttribute("aria-selected")).toBe("false");
    await h.user.keyboard("{ArrowRight}");
    expect(document.activeElement).toBe(tab(hostsViewSchemaId));
    expect(h.props.onSelectSurface).not.toHaveBeenCalled();
  }
  it("enters a System view without selecting a built-in", async () => {
    await enterWithoutSelection({ surface: indicatorsViewSchemaId });
  });
  it("enters Network Analysis without selecting a built-in", async () => {
    await enterWithoutSelection({ networkAnalysisActive: true });
  });

  it("preserves owned selector focus across responsive replacement without stealing external focus", async () => {
    const h = setup();
    await enter(h.user);
    await h.user.keyboard("{ArrowRight}");
    h.update({ layout: { ...h.props.layout, chromeMode: "narrow_desktop" } });
    const trigger = screen.getByTestId(workbookSurfacesMenuTriggerTestId());
    expect(document.activeElement).toBe(trigger);
    await h.user.keyboard("{Enter}{ArrowDown}");
    expect(document.activeElement).toBe(
      screen.getByTestId(workbookSurfacesMenuOptionTestId(notesViewSchemaId)),
    );
    h.update({ layout: { ...h.props.layout, chromeMode: "base" } });
    expect(document.activeElement).toBe(tab(evidenceViewSchemaId));
    expect(screen.queryByRole("menu")).toBeNull();
    await h.user.tab();
    const external = document.activeElement;
    h.update({ layout: { ...h.props.layout, chromeMode: "compact_desktop" } });
    expect(document.activeElement).toBe(external);
    h.update({ layout: { ...h.props.layout, chromeMode: "base" } });
    expect(document.activeElement).toBe(external);
    expect(h.props.onSelectSurface).not.toHaveBeenCalled();
  });

  it("dismisses compact navigation to its trigger and explicitly selects once", async () => {
    const h = setup({
      layout: {
        blockMode: "compact_height",
        chromeMode: "narrow_desktop",
        density: "compact",
        showStatusPresence: false,
      },
    });
    await enter(h.user);
    const trigger = screen.getByTestId(workbookSurfacesMenuTriggerTestId());
    await h.user.keyboard("{Enter}{ArrowDown}{Escape}");
    expect(document.activeElement).toBe(trigger);
    expect(h.props.onSelectSurface).not.toHaveBeenCalled();
    await h.user.keyboard("{ArrowDown}{ArrowDown} ");
    expect(h.props.onSelectSurface).toHaveBeenCalledTimes(1);
    expect(h.props.onSelectSurface).toHaveBeenCalledWith(notesViewSchemaId, {
      focusFirstGridTarget: true,
    });
  });
});
