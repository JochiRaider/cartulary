import { readdirSync, readFileSync } from "node:fs";
import {
  type InspectorDisabledCondition,
  type InspectorFeatureGroup,
  requireViewContract,
} from "@cartulary/view-contracts";
import { fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";
import { inspectorContextualCapabilities } from "../inspectorCapabilityResolver";
import { WorkbookInspectorDeclaredPanelList } from "../WorkbookInspectorDeclaredPanelList";
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
import { WorkbookInspectorContextualAction } from "./WorkbookInspectorActions";
import {
  WorkbookInspectorConfirmation,
  WorkbookInspectorFeedbackView,
  WorkbookInspectorPublicError,
} from "./WorkbookInspectorFeedback";
import {
  WorkbookInspectorPanelSection,
  WorkbookInspectorShell,
} from "./WorkbookInspectorShell";
import {
  bindWorkbookInspectorAction,
  workbookInspectorDisabledReason,
} from "./workbookInspectorPresentationModel";

const hosts = requireViewContract("cartulary.view.hosts.v1");
const relationshipsPanel = hosts.inspectorConfig.panels.find(
  (panel) => panel.panelId === "relationships",
);
if (relationshipsPanel === undefined) {
  throw new Error("Missing Hosts relationships panel");
}

describe("Workbook Inspector presentation", () => {
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
    const props = {
      config: hosts.inspectorConfig,
      currentIncidentRole: "admin" as const,
      disabledTokens: new Set<InspectorDisabledCondition>(),
      onContextualAction: vi.fn(),
    };
    const { rerender } = render(
      <WorkbookInspectorDeclaredPanelList
        {...props}
        contentByPanel={{
          details: <p>Details content</p>,
          evidence: <p>Evidence content</p>,
          history: <p>History content</p>,
          relationships: <p>Relationships content</p>,
          workflow: <p>Workflow content</p>,
        }}
        subject={live}
      />,
    );
    expect(
      screen
        .getAllByRole("heading", { level: 3 })
        .map((node) => node.textContent),
    ).toEqual(hosts.inspectorConfig.panels.map((panel) => panel.label));

    rerender(
      <WorkbookInspectorDeclaredPanelList
        {...props}
        contentByPanel={{ history: <p>History content</p> }}
        subject={deleted}
      />,
    );
    expect(
      screen
        .getAllByRole("heading", { level: 3 })
        .map((node) => node.textContent),
    ).toEqual(["History"]);

    rerender(
      <WorkbookInspectorDeclaredPanelList
        {...props}
        contentByPanel={{ workflow: <p>Standalone creation</p> }}
        subject={null}
      />,
    );
    expect(screen.getAllByRole("heading", { level: 3 })).toHaveLength(1);
    expect(screen.getByText("Standalone creation")).not.toBeNull();
    expect(screen.queryByText("History content")).toBeNull();
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
        noRowHeading="Hosts inspector"
        subject={null}
        onClose={vi.fn()}
      >
        <button type="button">Create a host</button>
      </WorkbookInspectorShell>,
    );
    expect(
      screen.getByRole("complementary").getAttribute("data-inspector-state"),
    ).toBe("no_row_selected");
    expect(
      screen.getByText("Select a saved row to inspect its details."),
    ).not.toBeNull();
    expect(screen.queryByText("no_row_selected")).toBeNull();
    expect(
      screen.getByRole("button", { name: "Create a host" }),
    ).not.toBeNull();
  });

  it("consumes panel-read groups without rendering their labels", () => {
    render(
      <WorkbookInspectorPanelSection
        panel={relationshipsPanel}
        viewSchemaId={hosts.viewSchemaId}
      >
        <p>Relationship content</p>
      </WorkbookInspectorPanelSection>,
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
      document.getElementById(review.getAttribute("aria-describedby") ?? "")
        ?.textContent,
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
      expect(review.hasAttribute("aria-describedby")).toBe(false);
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
    ).toBe("Requires the reviewer incident role.");
    expect(
      workbookInspectorDisabledReason({
        currentIncidentRole: "reviewer",
        featureGroup: merge,
        stateTokens: new Set(["row_version_changed"]),
      }),
    ).toBe("This row changed; refresh it before retrying.");
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
    expect(
      workbookInspectorDisabledReason({
        currentIncidentRole: "admin",
        featureGroup,
        stateTokens: new Set([token]),
      }),
    ).toBe(expected);
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
