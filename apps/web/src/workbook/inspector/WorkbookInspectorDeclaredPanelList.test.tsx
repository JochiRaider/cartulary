import { requireViewContract } from "@cartulary/view-contracts";
import { render, screen } from "@testing-library/react";
import type { ComponentProps } from "react";
import { describe, expect, it, vi } from "vitest";
import {
  inspectorPanel,
  savedInspectorRegion,
} from "./presentation/WorkbookInspectorPanelContent";
import type { WorkbookInspectorSection } from "./presentation/WorkbookInspectorShell";
import { WorkbookInspectorDeclaredPanelList } from "./WorkbookInspectorDeclaredPanelList";

type Props = ComponentProps<typeof WorkbookInspectorDeclaredPanelList>;
const config = requireViewContract("cartulary.view.notes.v1").inspectorConfig;
const subject: NonNullable<Props["subject"]> = {
  kind: "live",
  recordId: "note-A",
  rowVersion: 3,
  viewSchemaId: config.viewSchemaId,
  label: "Note A",
  surfaceLabel: "Notes",
};
const models: Props["modelsByPanel"] = Object.fromEntries(
  config.panels.map(({ panelId }) => [
    panelId,
    inspectorPanel(
      savedInspectorRegion(`fixture-${panelId}`, {
        kind: "populated",
        content: <p>{`Content ${panelId}`}</p>,
      }),
    ),
  ]),
);

function renderPanels(overrides: Partial<Props> = {}) {
  let sections: readonly WorkbookInspectorSection[] = [];
  const props: Props = {
    config,
    modelsByPanel: models,
    currentIncidentRole: "editor",
    disabledTokens: new Set(),
    onContextualAction: vi.fn(),
    subject,
    children: (value) => {
      sections = value;
      return value.map((section) => (
        <section key={section.panel.panelId}>{section.content}</section>
      ));
    },
    ...overrides,
  };
  const view = render(<WorkbookInspectorDeclaredPanelList {...props} />);
  return { ...view, sections, props };
}

describe("Declared inspector panels", () => {
  it("preserves declared order and forwards semantic focus and refs without owning data reads", () => {
    const panelRef = vi.fn();
    const { sections, unmount } = renderPanels({
      config: { ...config, panels: [...config.panels].reverse() },
      panelRef,
    });
    expect(sections.map((section) => section.panel.panelId)).toEqual(
      [...config.panels].reverse().map((panel) => panel.panelId),
    );
    expect(
      screen.getAllByText(/^Content /u).map((element) => element.textContent),
    ).toEqual(sections.map((section) => `Content ${section.panel.panelId}`));
    const element = document.createElement("section");
    const entry = document.createElement("button");
    entry.dataset.inspectorSectionEntry = "";
    for (const section of sections) {
      expect(section.focusDestination?.(element)).toBe(element);
      element.append(entry);
      expect(section.focusDestination?.(element)).toBe(entry);
      entry.remove();
      section.elementRef?.(element);
      expect(panelRef).toHaveBeenLastCalledWith(section.panel.panelId, element);
      section.elementRef?.(null);
      expect(panelRef).toHaveBeenLastCalledWith(section.panel.panelId, null);
    }
    unmount();
  });

  it("conceals inaccessible content and limits a deleted subject to history", () => {
    const concealed = config.panels.find(
      (panel) => panel.panelId !== "history",
    );
    if (concealed === undefined) {
      throw new Error("inspector fixture requires a non-history panel");
    }
    const view = renderPanels({
      modelsByPanel: {
        ...models,
        [concealed.panelId]: { access: "concealed" },
      },
    });
    expect(screen.queryByText(`Content ${concealed.panelId}`)).toBeNull();
    expect(view.sections.map((section) => section.panel.panelId)).not.toContain(
      concealed.panelId,
    );
    view.unmount();
    const noRole = renderPanels({ currentIncidentRole: null });
    expect(noRole.sections).toEqual([]);
    expect(screen.queryByText(/^Content /u)).toBeNull();
    noRole.unmount();
    const deleted = renderPanels({
      subject: { ...subject, kind: "deleted", stateLabel: "Deleted" },
    });
    expect(deleted.sections.map((section) => section.panel.panelId)).toEqual([
      "history",
    ]);
    expect(
      screen.getAllByText(/^Content /u).map((element) => element.textContent),
    ).toEqual(["Content history"]);
    expect(screen.queryByRole("button")).toBeNull();
    deleted.unmount();
  });

  it("requires explicit owner models and a matching creation attachment", () => {
    expect(() => renderPanels({ modelsByPanel: {} })).toThrow(
      /Missing inspector panel contribution/u,
    );
    expect(() => renderPanels({ subject: null })).toThrow(
      /Missing inspector creation attachment/u,
    );
    expect(() =>
      renderPanels({
        subject: null,
        creationAttachment: {
          id: "draft",
          viewSchemaId: "cartulary.view.evidence.v1",
        },
      }),
    ).toThrow(/Missing inspector creation attachment/u);
    const creation = renderPanels({
      subject: null,
      creationAttachment: { id: "draft", viewSchemaId: config.viewSchemaId },
    });
    expect(creation.sections.map((section) => section.panel.panelId)).toEqual(
      config.panels
        .filter((panel) => panel.panelId !== "history")
        .map((panel) => panel.panelId),
    );
    expect(screen.queryByText("Content history")).toBeNull();
    expect(screen.queryByRole("button")).toBeNull();
    creation.unmount();
  });
});
