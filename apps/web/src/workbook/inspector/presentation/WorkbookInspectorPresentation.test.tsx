import {
  type InspectorFeatureGroup,
  requireViewContract,
} from "@cartulary/view-contracts";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { InspectorDisabledToken } from "../semanticInspectorDispatcher";
import {
  WorkbookInspectorConfirmation,
  WorkbookInspectorDisabledReason,
  WorkbookInspectorPanelSection,
  WorkbookInspectorPublicError,
  WorkbookInspectorShell,
  workbookInspectorActionSemanticProps,
} from "./WorkbookInspectorPresentation";
import {
  bindWorkbookInspectorAction,
  workbookInspectorDisabledReason,
  workbookInspectorSafePublicMessage,
} from "./workbookInspectorPresentationModel";

const hosts = requireViewContract("cartulary.view.hosts.v1");

describe("Workbook Inspector presentation", () => {
  it("keeps the machine no-row state while presenting ordinary-user copy", () => {
    render(
      <WorkbookInspectorShell
        accessibleLabel="Hosts inspector"
        noRowHeading="Hosts inspector"
        subject={null}
        viewSchemaId={hosts.viewSchemaId}
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
        config={hosts.inspectorConfig}
        panelId="relationships"
      >
        <p>Relationship content</p>
      </WorkbookInspectorPanelSection>,
    );
    expect(screen.getByText("Relationships")).not.toBeNull();
    expect(screen.getByText("Relationship content")).not.toBeNull();
    expect(screen.queryByText("Relationships Read")).toBeNull();
    expect(screen.queryByText("Entity Aliases Read")).toBeNull();
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
    InspectorDisabledToken,
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

  it("replaces a raw row-version code with safe typed conflict copy", () => {
    expect(workbookInspectorSafePublicMessage("row_version_conflict")).toBe(
      "This row changed; refresh it before retrying.",
    );
  });

  it("keeps a safe public code available in technical details", () => {
    render(<WorkbookInspectorPublicError message="row_version_conflict" />);
    expect(screen.getByRole("alert").textContent).toContain(
      "This row changed; refresh it before retrying.",
    );
    expect(screen.getByText("Public error code")).not.toBeNull();
    expect(screen.getByText("row_version_conflict")).not.toBeNull();
  });

  it("binds semantic identity to an owner control and describes it", () => {
    const merge = hosts.inspectorConfig.featureGroups.find(
      (feature) => feature.featureGroupKey === "entity.merge",
    );
    expect(merge).toBeDefined();
    if (!merge) return;
    const binding = bindWorkbookInspectorAction(hosts.inspectorConfig, merge);
    expect(binding).not.toBeNull();
    if (!binding) return;
    render(
      <>
        <button
          {...workbookInspectorActionSemanticProps(binding, "merge-reason")}
          disabled
          type="button"
        >
          Merge entities
        </button>
        <WorkbookInspectorDisabledReason id="merge-reason">
          Requires the reviewer incident role.
        </WorkbookInspectorDisabledReason>
      </>,
    );
    expect(screen.getByRole("button").getAttribute("aria-describedby")).toBe(
      "merge-reason",
    );
    expect(
      screen.getByText("Requires the reviewer incident role."),
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
