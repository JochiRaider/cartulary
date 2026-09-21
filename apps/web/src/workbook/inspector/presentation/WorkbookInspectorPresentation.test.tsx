import { readdirSync, readFileSync } from "node:fs";
import {
  type InspectorDisabledCondition,
  type InspectorFeatureGroup,
  listViewContracts,
  requireViewContract,
} from "@cartulary/view-contracts";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ComponentProps } from "react";
import { afterEach, describe, expect, expectTypeOf, it, vi } from "vitest";
import type { RecordHistoryItem } from "../../adapters/workbookHistoryResponse";

import {
  WorkbookInspectorNavigationContext,
  type WorkbookInspectorNavigationSelection,
  workbookInspectorSectionFocusDestination,
} from "../../layout/workbookInspectorNavigation";
import { inspectorContextualCapabilities } from "../inspectorCapabilityResolver";
import { WorkbookInspectorContextualActions } from "../WorkbookInspectorContextualActions";
import { WorkbookInspectorDeclaredPanelList } from "../WorkbookInspectorDeclaredPanelList";
import { WorkbookInspectorSavedDetails } from "../WorkbookInspectorSavedDetails";
import { workbookHistoryEventPresentation } from "../workbookHistoryPresentationModel";
import {
  workbookInspectorErrorPresentation,
  workbookInspectorMessageFeedback,
  workbookInspectorOperationFailureFeedback,
} from "../workbookInspectorErrorModel";
import {
  buildWorkbookInspectorSubject,
  updateWorkbookInspectorSubject,
  workbookInspectorSubjectsEqual,
} from "../workbookInspectorSubject";
import {
  WorkbookHistoryEvent,
  WorkbookHistoryList,
} from "./WorkbookHistoryPresentation";
import { WorkbookInspectorContextualAction } from "./WorkbookInspectorActions";
import {
  WorkbookInspectorConfirmation,
  WorkbookInspectorFeedbackView,
  WorkbookInspectorPublicError,
} from "./WorkbookInspectorFeedback";
import {
  inspectorPanel,
  savedInspectorRegion,
} from "./WorkbookInspectorPanelContent";
import { WorkbookInspectorShell } from "./WorkbookInspectorShell";
import {
  bindWorkbookInspectorAction,
  ownerInspectorDisabledReason,
  workbookInspectorDisabledReason,
  workbookInspectorDisabledReasonText,
} from "./workbookInspectorPresentationModel";

afterEach(cleanup);

const hosts = requireViewContract("cartulary.view.hosts.v1");
const relationshipsPanel = hosts.inspectorConfig.panels.find(
  (panel) => panel.panelId === "relationships",
);
if (relationshipsPanel === undefined) {
  throw new Error("Missing Hosts relationships panel");
}

describe("Workbook Inspector presentation", () => {
  it("orders admitted action rows once with explicit outcomes across all configurations", () => {
    let admitted = 0;
    for (const contract of listViewContracts()) {
      for (const panel of contract.inspectorConfig.panels) {
        const capabilities = inspectorContextualCapabilities({
          config: contract.inspectorConfig,
          panelId: panel.panelId,
        });
        if (!capabilities.length) continue;
        const action = vi.fn();
        const { container, unmount } = render(
          <WorkbookInspectorContextualActions
            config={contract.inspectorConfig}
            capabilities={[...capabilities].reverse().concat(capabilities)}
            currentIncidentRole="admin"
            disabledTokens={new Set()}
            onAction={action}
          />,
        );
        const buttons = [...container.querySelectorAll("button")];
        expect(buttons.map((button) => button.textContent)).toEqual(
          capabilities.map((capability) => capability.featureGroup.label),
        );
        expect(container.querySelectorAll("li")).toHaveLength(
          capabilities.length,
        );
        for (const [index, button] of buttons.entries()) {
          const description = button.getAttribute("aria-describedby");
          expect(description).not.toBeNull();
          expect(
            document.getElementById(description ?? "")?.textContent,
          ).toMatch(/^Opens (authoring|review|the related workbook view)/);
          fireEvent.click(button);
          expect(action).toHaveBeenNthCalledWith(
            index + 1,
            capabilities[index],
          );
        }
        expect(action).toHaveBeenCalledTimes(capabilities.length);
        admitted += capabilities.length;
        unmount();
      }
    }
    expect(listViewContracts()).toHaveLength(17);
    expect(admitted).toBe(48);
  });
  it("renders every semantic History family with complete values and nested disclosure focus", () => {
    const kinds: RecordHistoryItem["diff_summary"]["units"][number]["kind"][] =
      [
        "field",
        "link",
        "mention",
        "tag",
        "evidence_association",
        "capture_state",
        "record",
        "entity_identifier",
        "indicator_observation",
        "indicator_interval",
      ];
    const values = [
      false,
      0,
      "",
      "Long narrative\nwith a second line",
      ["record-a", "record-b"],
    ] as const;
    const [firstUnit, ...otherUnits] = kinds.map(
      (kind, index): RecordHistoryItem["diff_summary"]["units"][number] => ({
        unit_ref: `unit-${kind}`,
        kind,
        operation: "update",
        record_ids: ["source-record"],
        changes: [
          {
            field_key: `${kind}.detail`,
            before: { state: index % 2 ? "null" : "absent" },
            after: {
              state: "present",
              value: values[index % values.length] ?? false,
            },
          },
        ],
      }),
    );
    if (!firstUnit) throw new Error("Missing semantic fixture");
    const event = workbookHistoryEventPresentation({
      actor_user_id: "local-actor",
      source_actor_id: "imported-actor",
      committed_at: "2026-09-21T00:12:34.123456Z",
      history_item_ref: "item",
      operation: "patch",
      change_set_id: "change",
      reversible: false,
      available_rollback_actions: [],
      diff_summary: {
        schema_id: "cartulary.history_diff.v1",
        summary: "Untrusted display summary",
        units: [firstUnit, ...otherUnits],
      },
    });
    const closed = vi.fn();
    const { container } = render(
      <WorkbookHistoryList>
        <WorkbookHistoryEvent
          event={event}
          onClose={closed}
          actions={<button type="button">Review reversal</button>}
        />
      </WorkbookHistoryList>,
    );
    expect(container.textContent).toContain(
      "Changed by Identifier imported-actor",
    );
    expect(container.textContent).toContain("2026-09-21 00:12:34 UTC +00:00");
    const disclosure = container.querySelector("li > details");
    if (!(disclosure instanceof HTMLDetailsElement))
      throw new Error("Missing event disclosure");
    expect(disclosure.open).toBe(false);
    disclosure.open = true;
    for (const kind of kinds)
      expect(container.textContent).toContain(`${kind}.detail`);
    for (const value of [
      "Not present",
      "No value (null)",
      "False",
      "0",
      "Empty text",
      "Long narrative",
      "record-a",
      "record-b",
      "2026-09-21T00:12:34.123456Z",
    ])
      expect(container.textContent).toContain(value);
    expect(container.querySelectorAll("section")).toHaveLength(kinds.length);
    const review = screen.getByRole("button", { name: "Review reversal" });
    review.focus();
    fireEvent.keyDown(review, { key: "Escape" });
    expect(disclosure.open).toBe(false);
    expect(document.activeElement).toBe(
      disclosure.querySelector(":scope > summary"),
    );
  });
  it("preserves declared saved field order and value distinctions across all configurations", () => {
    const contracts = listViewContracts();
    expect(contracts).toHaveLength(17);
    for (const contract of contracts) {
      const cells = Object.fromEntries(
        contract.fields.map((field, index) => [
          field.fieldKey,
          { value: [null, false, 0, "", "Readable value"][index % 5] },
        ]),
      );
      const last = contract.fields.at(-1);
      if (last) delete cells[last.fieldKey];
      const view = render(
        <WorkbookInspectorSavedDetails
          contract={contract}
          row={{ record_id: "saved", row_version: 1, cells }}
        />,
      );
      const rows = [
        ...view.container.querySelectorAll("[data-inspector-saved-field]"),
      ];
      expect(
        rows.map((row) => row.getAttribute("data-inspector-saved-field")),
      ).toEqual(contract.fields.map((field) => field.fieldKey));
      rows.forEach((row, index) => {
        expect(row.querySelector("dd")?.textContent).toBe(
          index === rows.length - 1
            ? "Not loaded"
            : ["Not set", "No", "0", "Empty text", "Readable value"][index % 5],
        );
      });
      view.unmount();
    }
  });
  it("shares one permission explanation across affected actions", () => {
    const capabilities = inspectorContextualCapabilities({
      config: hosts.inspectorConfig,
      panelId: "workflow",
    });
    render(
      <WorkbookInspectorContextualActions
        capabilities={capabilities}
        config={hosts.inspectorConfig}
        currentIncidentRole="viewer"
        disabledTokens={new Set()}
        onAction={vi.fn()}
      />,
    );
    const actions = screen.getAllByRole("button") as HTMLButtonElement[];
    expect(actions.length).toBeGreaterThan(1);
    expect(actions.every((action) => action.disabled)).toBe(true);
    const descriptions = new Set(
      actions.map((action) =>
        action.getAttribute("aria-describedby")?.split(" ").at(-1),
      ),
    );
    expect(descriptions.size).toBe(1);
    expect(descriptions.has(undefined)).toBe(false);
    expect(
      screen.getAllByText("Requires the editor incident role."),
    ).toHaveLength(1);
  });
  it("keeps different causes distinct without changing action order or enabled actions", () => {
    const capabilities = inspectorContextualCapabilities({
      config: hosts.inspectorConfig,
      panelId: "workflow",
    });
    const [first, second] = capabilities;
    if (!first || !second) throw new Error("Missing workflow actions");
    const reasons = new Map([
      [
        first.featureGroup.featureGroupKey,
        ownerInspectorDisabledReason(
          "test",
          "first_cause",
          "Wait for recovery.",
        ),
      ],
      [
        second.featureGroup.featureGroupKey,
        ownerInspectorDisabledReason(
          "test",
          "second_cause",
          "Wait for recovery.",
        ),
      ],
    ]);
    const { rerender } = render(
      <WorkbookInspectorContextualActions
        config={hosts.inspectorConfig}
        capabilities={capabilities}
        currentIncidentRole="editor"
        disabledTokens={new Set()}
        additionalDisabledReasons={reasons}
        onAction={vi.fn()}
      />,
    );
    const actions = screen.getAllByRole("button") as HTMLButtonElement[];
    expect(actions.map((action) => action.textContent)).toEqual(
      capabilities.map((capability) => capability.featureGroup.label),
    );
    expect(actions[0]?.getAttribute("aria-describedby")).not.toBe(
      actions[1]?.getAttribute("aria-describedby"),
    );
    expect(screen.getAllByText("Wait for recovery.")).toHaveLength(2);
    reasons.delete(second.featureGroup.featureGroupKey);
    rerender(
      <WorkbookInspectorContextualActions
        config={hosts.inspectorConfig}
        capabilities={capabilities}
        currentIncidentRole="editor"
        disabledTokens={new Set()}
        additionalDisabledReasons={reasons}
        onAction={vi.fn()}
      />,
    );
    expect(actions[0]?.disabled).toBe(true);
    expect(actions[1]?.disabled).toBe(false);
    expect(
      actions[1]?.getAttribute("aria-describedby")?.split(" "),
    ).toHaveLength(1);
  });
  it("keeps record context and Close outside the scrolling section body", async () => {
    const user = userEvent.setup();
    const label = "A long record label that remains available in full";
    const close = vi.fn();
    const openHistory = vi.fn();
    const retention = {
      current: null as WorkbookInspectorNavigationSelection | null,
    };
    const details = hosts.inspectorConfig.panels.find(
      (panel) => panel.panelId === "details",
    );
    const history = hosts.inspectorConfig.panels.find(
      (panel) => panel.panelId === "history",
    );
    if (!details || !history) throw new Error("Missing declared sections");
    const view = (recordId = "host-a", concealHistory = false, show = true) => (
      <WorkbookInspectorNavigationContext value={retention}>
        {show ? (
          <WorkbookInspectorShell
            accessibleLabel="Hosts inspector"
            config={hosts.inspectorConfig}
            mode="saved"
            subject={required(
              buildWorkbookInspectorSubject({
                config: hosts.inspectorConfig,
                kind: "live",
                label,
                recordId,
                rowVersion: 3,
                surfaceLabel: "Hosts",
              }),
            )}
            onClose={close}
            sections={[
              {
                panel: details,
                focusDestination: workbookInspectorSectionFocusDestination,
                content: (
                  <input
                    aria-label="Unfinished field"
                    defaultValue="Retained value"
                  />
                ),
              },
              ...(concealHistory
                ? []
                : [
                    {
                      panel: history,
                      focusDestination:
                        workbookInspectorSectionFocusDestination,
                      content: (
                        <button
                          data-inspector-section-entry
                          type="button"
                          onClick={openHistory}
                        >
                          Open history
                        </button>
                      ),
                    },
                  ]),
            ]}
            feedback={<button type="button">Last section action</button>}
          />
        ) : null}
      </WorkbookInspectorNavigationContext>
    );
    const { rerender } = render(view());
    const shell = screen.getByRole("complementary");
    const body = shell.querySelector("[data-inspector-scroll-body]");
    expect(body).not.toBeNull();
    expect(
      body?.contains(screen.getByRole("button", { name: "Close inspector" })),
    ).toBe(false);
    expect(
      body?.contains(screen.getByRole("button", { name: "Sections" })),
    ).toBe(false);
    expect(body?.contains(screen.getByRole("heading", { name: label }))).toBe(
      false,
    );
    expect(
      body?.contains(
        screen.getByRole("button", { name: "Last section action" }),
      ),
    ).toBe(true);
    expect(body?.querySelector("details")?.textContent).toContain(label);
    const field = screen.getByRole("textbox", { name: "Unfinished field" });
    await user.type(field, " plus draft");
    await user.click(screen.getByRole("button", { name: "Sections" }));
    await user.click(screen.getByRole("button", { name: "History" }));
    expect(screen.getByRole("button", { name: "Open history" })).toBe(
      document.activeElement,
    );
    expect(openHistory).not.toHaveBeenCalled();
    expect(field).toBe(
      screen.getByRole("textbox", { name: "Unfinished field" }),
    );
    expect((field as HTMLInputElement).value).toBe("Retained value plus draft");
    expect(screen.getByText("Current section: History")).not.toBeNull();
    if (body) fireEvent.scroll(body);
    expect(openHistory).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Open history" })).toBe(
      document.activeElement,
    );
    await user.click(screen.getByRole("button", { name: "Sections" }));
    await user.tab();
    await user.keyboard("{Escape}");
    expect(
      screen.queryByRole("navigation", { name: "Inspector sections" }),
    ).toBeNull();
    expect(screen.getByRole("button", { name: "Sections" })).toBe(
      document.activeElement,
    );
    expect(close).not.toHaveBeenCalled();
    rerender(view("host-a", false, false));
    rerender(view());
    expect(screen.getByText("Current section: History")).not.toBeNull();
    const reopenedField = screen.getByRole("textbox", {
      name: "Unfinished field",
    });
    rerender(view("host-b"));
    expect(screen.getByRole("textbox", { name: "Unfinished field" })).toBe(
      reopenedField,
    );
    expect(screen.getByText("Current section: Details")).not.toBeNull();
    // An explicit destination wins without going through the navigation control.
    fireEvent.focus(screen.getByRole("button", { name: "Open history" }));
    expect(screen.getByText("Current section: History")).not.toBeNull();
    screen.getByRole("button", { name: "Open history" }).focus();
    rerender(view("host-b", true));
    expect(screen.queryByRole("button", { name: "Open history" })).toBeNull();
    expect(screen.getByRole("button", { name: "Close inspector" })).toBe(
      document.activeElement,
    );
    await user.click(screen.getByRole("button", { name: "Sections" }));
    expect(screen.queryByRole("button", { name: "History" })).toBeNull();
    expect(screen.getByText("Current section: Details")).not.toBeNull();
  });
  it("validates one live or deleted subject boundary and rejects invalid identity", () => {
    const live = buildWorkbookInspectorSubject({
      config: hosts.inspectorConfig,
      kind: "live",
      label: "Host alpha",
      recordId: "  host-a  ",
      rowVersion: 3,
      surfaceLabel: "Hosts",
    });
    expect(live).toEqual({
      kind: "live",
      label: "Host alpha",
      recordId: "host-a",
      rowVersion: 3,
      surfaceLabel: "Hosts",
      viewSchemaId: hosts.viewSchemaId,
    });
    expect(
      buildWorkbookInspectorSubject({
        config: hosts.inspectorConfig,
        kind: "deleted",
        label: "Deleted host",
        recordId: "host-a",
        rowVersion: 4,
        surfaceLabel: "Hosts",
      }),
    ).toMatchObject({ kind: "deleted", stateLabel: "Deleted" });
    for (const identity of [
      { recordId: "", rowVersion: 1 },
      { recordId: "   ", rowVersion: 1 },
      { recordId: "host-a", rowVersion: 0 },
      { recordId: "host-a", rowVersion: -1 },
      { recordId: "host-a", rowVersion: 1.5 },
    ]) {
      expect(
        buildWorkbookInspectorSubject({
          config: hosts.inspectorConfig,
          kind: "live",
          label: "Host alpha",
          surfaceLabel: "Hosts",
          ...identity,
        }),
      ).toBeNull();
    }
    expect(live).not.toBeNull();
    if (live === null) return;
    expect(
      workbookInspectorSubjectsEqual(live, {
        ...live,
        label: "Renamed host",
        surfaceLabel: "Host records",
      }),
    ).toBe(true);
    expect(
      updateWorkbookInspectorSubject(live, {
        kind: "live",
        recordId: " host-a ",
        rowVersion: 3,
      }),
    ).toBe(live);
    expect(
      workbookInspectorSubjectsEqual(live, {
        ...live,
        kind: "deleted",
        stateLabel: "Deleted",
      }),
    ).toBe(false);
  });

  it("renders declared panels once in order for live, only History for deleted, and only explicit creation content for null", () => {
    const live = buildWorkbookInspectorSubject({
      config: hosts.inspectorConfig,
      kind: "live",
      label: "Host alpha",
      recordId: "host-a",
      rowVersion: 3,
      surfaceLabel: "Hosts",
    });
    const deleted = buildWorkbookInspectorSubject({
      config: hosts.inspectorConfig,
      kind: "deleted",
      label: "Deleted host",
      recordId: "host-a",
      rowVersion: 4,
      surfaceLabel: "Hosts",
    });
    if (!live || !deleted) throw new Error("Missing subject fixture");
    const props = {
      config: hosts.inspectorConfig,
      currentIncidentRole: "admin" as const,
      disabledTokens: new Set<InspectorDisabledCondition>(),
      onContextualAction: vi.fn(),
    };
    const { rerender } = render(
      <WorkbookInspectorDeclaredPanelList
        {...props}
        modelsByPanel={{
          details: inspectorPanel(
            savedInspectorRegion("saved", {
              kind: "populated",
              content: <p>Details content</p>,
            }),
          ),
          evidence: inspectorPanel(
            savedInspectorRegion("saved", {
              kind: "populated",
              content: <p>Evidence content</p>,
            }),
          ),
          history: inspectorPanel(
            savedInspectorRegion("saved", {
              kind: "populated",
              content: <p>History content</p>,
            }),
          ),
          relationships: inspectorPanel(
            savedInspectorRegion("saved", {
              kind: "populated",
              content: <p>Relationships content</p>,
            }),
          ),
          workflow: inspectorPanel(
            savedInspectorRegion("saved", {
              kind: "populated",
              content: <p>Workflow content</p>,
            }),
          ),
        }}
        subject={live}
      >
        {(sections) => (
          <WorkbookInspectorShell
            config={props.config}
            accessibleLabel="Inspector"
            mode="saved"
            subject={live}
            onClose={vi.fn()}
            sections={sections}
          />
        )}
      </WorkbookInspectorDeclaredPanelList>,
    );
    expect(screen.getByRole("heading", { level: 2 }).textContent).toBe(
      "Host alpha",
    );
    expect(
      screen
        .getAllByRole("heading", { level: 3 })
        .map((node) => node.textContent),
    ).toEqual(hosts.inspectorConfig.panels.map((panel) => panel.label));

    rerender(
      <WorkbookInspectorDeclaredPanelList
        {...props}
        modelsByPanel={{
          history: inspectorPanel(
            savedInspectorRegion("saved", {
              kind: "populated",
              content: <p>History content</p>,
            }),
          ),
        }}
        subject={deleted}
      >
        {(sections) => (
          <WorkbookInspectorShell
            config={props.config}
            accessibleLabel="Inspector"
            mode="saved"
            subject={deleted}
            onClose={vi.fn()}
            sections={sections}
          />
        )}
      </WorkbookInspectorDeclaredPanelList>,
    );
    expect(
      screen
        .getAllByRole("heading", { level: 3 })
        .map((node) => node.textContent),
    ).toEqual(["History"]);

    rerender(
      <WorkbookInspectorDeclaredPanelList
        {...props}
        creationAttachment={{
          id: "assessment-create",
          viewSchemaId: props.config.viewSchemaId,
        }}
        modelsByPanel={{
          workflow: inspectorPanel(
            savedInspectorRegion("saved", {
              kind: "populated",
              content: <p>Standalone creation</p>,
            }),
          ),
        }}
        subject={null}
      >
        {(sections) => (
          <WorkbookInspectorShell
            config={props.config}
            accessibleLabel="Inspector"
            mode="creation"
            context={{
              id: "assessment-create",
              viewSchemaId: props.config.viewSchemaId,
              heading: "Append assessment",
            }}
            onClose={vi.fn()}
            sections={sections}
          />
        )}
      </WorkbookInspectorDeclaredPanelList>,
    );
    expect(screen.getAllByRole("heading", { level: 3 })).toHaveLength(1);
    expect(screen.getByText("Standalone creation")).not.toBeNull();
    expect(screen.queryByText("History content")).toBeNull();
    expect(screen.getByRole("heading", { level: 2 }).textContent).toBe(
      "Append assessment",
    );
    expect(
      screen.getByRole("complementary").getAttribute("data-inspector-state"),
    ).toBe("creation");
    expect(
      screen.queryByText("Select a saved row to inspect its details."),
    ).toBeNull();
    expectTypeOf<
      Extract<
        ComponentProps<typeof WorkbookInspectorShell>,
        { mode: "creation" }
      >["subject"]
    >().toEqualTypeOf<undefined>();
    expectTypeOf<
      Extract<
        ComponentProps<typeof WorkbookInspectorShell>,
        { mode: "empty" }
      >["sections"]
    >().toEqualTypeOf<undefined>();
    type Saved = Extract<
      ComponentProps<typeof WorkbookInspectorShell>,
      { mode: "saved" }
    >;
    expectTypeOf<
      Partial<Pick<Saved, "sections">> extends Pick<Saved, "sections">
        ? true
        : false
    >().toEqualTypeOf<false>();
  });

  it("keeps presentation source free of state orchestration hooks", () => {
    const presentationDirectory = new URL(".", import.meta.url);
    for (const filename of readdirSync(presentationDirectory)) {
      if (
        (!filename.endsWith(".ts") && !filename.endsWith(".tsx")) ||
        filename.endsWith(".test.ts") ||
        filename.endsWith(".test.tsx")
      ) {
        continue;
      }
      const source = readFileSync(
        new URL(filename, presentationDirectory),
        "utf8",
      );
      expect(source, filename).not.toMatch(/\buse(?:State|Reducer)\b/u);
    }
  });

  it("keeps the machine no-row state while presenting ordinary-user copy", () => {
    render(
      <WorkbookInspectorShell
        accessibleLabel="Hosts inspector"
        config={hosts.inspectorConfig}
        mode="empty"
        heading="Hosts inspector"
        onClose={vi.fn()}
      />,
    );
    expect(
      screen.getByRole("complementary").getAttribute("data-inspector-state"),
    ).toBe("no_row_selected");
    expect(
      screen.getByText("Select a saved row to inspect its details."),
    ).not.toBeNull();
    expect(screen.queryByText("no_row_selected")).toBeNull();
    expect(screen.queryByRole("button", { name: "Sections" })).toBeNull();
  });

  it("consumes panel-read groups without rendering their labels", () => {
    render(
      <WorkbookInspectorShell
        accessibleLabel="Hosts inspector"
        config={hosts.inspectorConfig}
        mode="creation"
        context={{
          id: "test",
          viewSchemaId: hosts.viewSchemaId,
          heading: "Host",
        }}
        onClose={vi.fn()}
        sections={[
          {
            panel: relationshipsPanel,
            content: <p>Relationship content</p>,
            focusDestination: (section) => section,
          },
        ]}
      />,
    );
    expect(screen.getByText("Relationships")).not.toBeNull();
    expect(screen.getByText("Relationship content")).not.toBeNull();
    expect(screen.queryByText("Relationships Read")).toBeNull();
    expect(screen.queryByText("Entity Aliases Read")).toBeNull();
  });

  it("keeps Timeline review discoverable without permitting unauthorized pointer or keyboard activation", async () => {
    const user = userEvent.setup();
    const timeline = requireViewContract("cartulary.view.timeline.v2");
    const capability = inspectorContextualCapabilities({
      config: timeline.inspectorConfig,
      panelId: "history",
    }).find(
      (entry) =>
        entry.featureGroup.featureGroupKey === "timeline.mark_reviewed",
    );
    if (!capability) throw new Error("Missing declared Timeline review action");
    const binding = bindWorkbookInspectorAction(
      timeline.inspectorConfig,
      capability,
    );
    const onInvoke = vi.fn();
    const action = (role: "editor" | "reviewer" | "admin") => (
      <>
        <button type="button">Before review</button>
        <WorkbookInspectorContextualAction
          binding={binding}
          currentIncidentRole={role}
          disabledTokens={new Set()}
          onInvoke={onInvoke}
        />
        <button type="button">After review</button>
      </>
    );
    const { rerender } = render(action("editor"));
    const review = screen.getByRole("button", {
      name: binding.featureGroup.label,
    }) as HTMLButtonElement;
    expect(review.disabled).toBe(true);
    expect(
      document.getElementById(
        review.getAttribute("aria-describedby")?.split(" ").at(-1) ?? "",
      )?.textContent,
    ).toBe("Requires the reviewer incident role.");
    await user.click(review);
    screen.getByRole("button", { name: "Before review" }).focus();
    await user.tab();
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "After review" }),
    );
    await user.keyboard("{Enter} ");
    expect(onInvoke).not.toHaveBeenCalled();
    for (const role of ["reviewer", "admin"] as const) {
      rerender(action(role));
      expect(review.disabled).toBe(false);
      expect(review.getAttribute("aria-describedby")?.split(" ")).toHaveLength(
        1,
      );
      await user.click(review);
    }
    expect(onInvoke).toHaveBeenCalledTimes(2);
  });

  it("derives closed owner-backed disabled reasons in contract order", () => {
    const merge = hosts.inspectorConfig.featureGroups.find(
      (feature) => feature.featureGroupKey === "entity.merge",
    );
    expect(merge).toBeDefined();
    if (!merge) return;
    expect(
      workbookInspectorDisabledReason({
        currentIncidentRole: "editor",
        featureGroup: merge,
        stateTokens: new Set(["row_version_changed"]),
      }),
    ).toEqual({ kind: "minimum_role", role: "reviewer" });
    expect(
      workbookInspectorDisabledReason({
        currentIncidentRole: "reviewer",
        featureGroup: merge,
        stateTokens: new Set(["row_version_changed"]),
      }),
    ).toEqual({ kind: "condition", condition: "row_version_changed" });
  });

  it.each([
    ["no_row_selected", "Select a saved row to use this action."],
    ["incident_closed", "This incident is closed and read-only."],
    ["authorization_lost", "You no longer have access to this action."],
    ["row_version_changed", "This row changed; refresh it before retrying."],
    ["record_deleted", "This action is unavailable for a deleted record."],
    ["record_merged", "This record was merged and can no longer be changed."],
    [
      "evidence_preview_unavailable",
      "Preview is unavailable for this evidence.",
    ],
    ["merge_target_unavailable", "Select a valid merge target."],
    [
      "record_not_deleted",
      "This action is available only for deleted records.",
    ],
    [
      "rollback_target_unavailable",
      "Select an available history change to roll back.",
    ],
    ["party_text_unavailable", "No party reference text is available to link."],
    ["pivot_target_unavailable", "No matching destination is available."],
  ] satisfies readonly (readonly [
    InspectorDisabledCondition,
    string,
  ])[])("presents deterministic copy for %s", (token, expected) => {
    const template = hosts.inspectorConfig.featureGroups.find(
      (feature) => feature.featureGroupKey === "entity.merge",
    );
    expect(template).toBeDefined();
    if (!template) return;
    const featureGroup = {
      ...template,
      disabledWhen: [token],
      minimumIncidentRole: null,
    } satisfies InspectorFeatureGroup;
    const reason = workbookInspectorDisabledReason({
      currentIncidentRole: "admin",
      featureGroup,
      stateTokens: new Set([token]),
    });
    expect(reason).not.toBeNull();
    if (reason)
      expect(workbookInspectorDisabledReasonText(reason)).toBe(expected);
  });

  it("keeps a safe public code available in technical details", () => {
    render(
      <WorkbookInspectorPublicError
        error={workbookInspectorErrorPresentation({
          kind: "stale_target",
          message: "The record is stale.",
          publicCode: "row_version_conflict",
        })}
      />,
    );
    expect(screen.getByRole("alert").textContent).toContain(
      "This row changed; refresh it before retrying.",
    );
    expect(screen.getByText("Public error code")).not.toBeNull();
    expect(screen.getByText("row_version_conflict")).not.toBeNull();
    expect(screen.getByText("The record is stale.")).not.toBeNull();
  });

  it("renders declared neutral announcements and assertive typed failures", () => {
    const { rerender } = render(
      <WorkbookInspectorFeedbackView
        feedback={workbookInspectorMessageFeedback("Timeline ready.", "none")}
      />,
    );
    expect(screen.getByText("Timeline ready.").getAttribute("role")).toBeNull();
    expect(
      screen.getByText("Timeline ready.").getAttribute("aria-live"),
    ).toBeNull();

    rerender(
      <WorkbookInspectorFeedbackView
        feedback={workbookInspectorMessageFeedback(
          "Indicator ready.",
          "polite",
        )}
      />,
    );
    expect(screen.getByRole("status").getAttribute("aria-live")).toBe("polite");

    rerender(
      <WorkbookInspectorFeedbackView
        feedback={workbookInspectorOperationFailureFeedback({
          kind: "retryable",
          message: "Try again.",
        })}
      />,
    );
    expect(screen.getByRole("alert").getAttribute("aria-live")).toBe(
      "assertive",
    );
    expect(screen.getByText("Try again.")).not.toBeNull();
  });

  it("binds semantic identity to an owner control and describes it", () => {
    const createCapability = inspectorContextualCapabilities({
      config: hosts.inspectorConfig,
      panelId: "workflow",
    }).find(
      (capability) =>
        capability.featureGroup.featureGroupKey === "create_related.note",
    );
    expect(createCapability).toBeDefined();
    if (createCapability === undefined) return;
    const binding = bindWorkbookInspectorAction(
      hosts.inspectorConfig,
      createCapability,
    );
    render(
      <WorkbookInspectorContextualAction
        binding={binding}
        currentIncidentRole="editor"
        disabledTokens={new Set()}
        onInvoke={vi.fn()}
      />,
    );
    expect(
      screen.getByRole("button").getAttribute("aria-describedby"),
    ).not.toBeNull();
    expect(
      screen.getByText("Requires the editor incident role."),
    ).not.toBeNull();
  });

  it("focuses the safe confirmation action and consumes Escape locally", () => {
    const onCancel = vi.fn();
    render(
      <WorkbookInspectorConfirmation
        confirmLabel="Delete record"
        destructive
        operation="Delete"
        subject="Host alpha"
        onCancel={onCancel}
        onConfirm={vi.fn()}
      />,
    );
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Cancel" }),
    );
    fireEvent.keyDown(screen.getByRole("alertdialog"), { key: "Escape" });
    expect(onCancel).toHaveBeenCalledTimes(1);
  });
});

function required<T>(value: T | null): T {
  if (value === null) throw new Error("Missing fixture value");
  return value;
}
