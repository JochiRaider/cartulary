import { workbookInspectorFeatureActionTestId } from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { WorkbookInspectorPanelSection } from "./WorkbookInspectorFeatureGroups";

const contract = requireViewContract("cartulary.view.hosts.v1");

describe("WorkbookInspectorPanelSection", () => {
  it("confirms against stable subject identity and invalidates on row-version change", async () => {
    const onFeatureAction = vi.fn();
    const rendered = render(
      <WorkbookInspectorPanelSection
        config={contract.inspectorConfig}
        currentIncidentRole="editor"
        disabledTokens={new Set(["record_not_deleted"])}
        panelId="history"
        subjectRecordId="host-1"
        subjectRowVersion={1}
        onFeatureAction={onFeatureAction}
      />,
    );
    fireEvent.click(
      screen.getByTestId(
        workbookInspectorFeatureActionTestId(
          contract.viewSchemaId,
          "record.delete",
        ),
      ),
    );
    expect(screen.getByRole("alertdialog").textContent).toContain(
      "host-1 at row version 1",
    );

    rendered.rerender(
      <WorkbookInspectorPanelSection
        config={contract.inspectorConfig}
        currentIncidentRole="editor"
        disabledTokens={new Set(["record_not_deleted"])}
        panelId="history"
        subjectRecordId="host-1"
        subjectRowVersion={2}
        onFeatureAction={onFeatureAction}
      />,
    );
    await waitFor(() => {
      expect(screen.queryByRole("alertdialog")).toBeNull();
    });
    expect(onFeatureAction).not.toHaveBeenCalled();
  });

  it("renders but disables role-restricted actions after authorization loss", () => {
    render(
      <WorkbookInspectorPanelSection
        config={contract.inspectorConfig}
        currentIncidentRole="viewer"
        disabledTokens={new Set(["record_not_deleted"])}
        panelId="history"
        subjectRecordId="host-1"
        subjectRowVersion={1}
        onFeatureAction={vi.fn()}
      />,
    );
    expect(
      screen
        .getByTestId(
          workbookInspectorFeatureActionTestId(
            contract.viewSchemaId,
            "record.delete",
          ),
        )
        .hasAttribute("disabled"),
    ).toBe(true);
  });
});
