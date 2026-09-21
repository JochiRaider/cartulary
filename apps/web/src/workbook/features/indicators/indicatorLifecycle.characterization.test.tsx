import { requireViewContract } from "@cartulary/view-contracts";
import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { savedInspectorRegion } from "../../inspector/presentation/WorkbookInspectorPanelContent";
import { GenericWorkbookInspector } from "../generic/GenericWorkbookInspector";

vi.mock("../../inspector/WorkbookInspectorRecordHistory", () => ({
  WorkbookInspectorRecordHistory: () => <p>Existing record history</p>,
}));
vi.mock("./IndicatorLifecycleWorkflow", () => ({
  IndicatorLifecycleWorkflow: () => <p>Lifecycle interval content</p>,
}));
afterEach(cleanup);

function fixture(
  config: ReturnType<typeof requireViewContract>["inspectorConfig"],
  action: "indicator.lifecycle.read" | "indicator.lifecycle.manage",
) {
  return (
    <GenericWorkbookInspector
      config={config}
      currentIncidentRole="editor"
      disabledTokens={new Set()}
      subject={{
        kind: "live",
        recordId: "00000000-0000-4000-8000-000000000001",
        rowVersion: 1,
        viewSchemaId: config.viewSchemaId,
        label: "example.test",
        surfaceLabel: "Indicators",
      }}
      surfaceTitle="Indicators"
      detailsContent={null}
      evidenceContent={[
        savedInspectorRegion("evidence", {
          kind: "empty",
          message: "No evidence.",
        }),
      ]}
      relationshipsContent={[
        savedInspectorRegion("relationships", {
          kind: "empty",
          message: "No relationships.",
        }),
      ]}
      workflowContent={null}
      onClose={vi.fn()}
      mutationError={null}
      relatedFeedback={null}
      history={{
        actions: new Set(),
        canMutate: true,
        effects: {
          deleteAccepted: () => {},
          restoreAccepted: () => {},
          rollbackAccepted: () => {},
          refresh: () => {},
        },
      }}
      related={{
        begin: () => false,
        state: null,
        cancel: vi.fn(),
        submit: async () => {},
        updateDraft: vi.fn(),
      }}
      indicator={{
        handler: { action, panelId: "history" },
        onMutationCommitted: vi.fn(),
        recordId: "00000000-0000-4000-8000-000000000001",
        rowVersion: 1,
        select: vi.fn(),
      }}
    />
  );
}
it("Indicator lifecycle content is reachable in the canonical History panel", () => {
  const config = requireViewContract(
    "cartulary.view.indicators.v1",
  ).inspectorConfig;
  for (const action of [
    "indicator.lifecycle.read",
    "indicator.lifecycle.manage",
  ] as const) {
    render(fixture(config, action));
    expect(screen.getAllByText("Existing record history")).toHaveLength(1);
    expect(screen.getAllByText("Lifecycle interval content")).toHaveLength(1);
    cleanup();
  }
  const changed = {
    ...config,
    featureGroups: config.featureGroups.map((feature) =>
      feature.routeBinding.kind === "indicator_lifecycle"
        ? {
            ...feature,
            routeBinding: {
              ...feature.routeBinding,
              owner: "record_patch_route" as const,
            },
          }
        : feature,
    ),
  };
  render(fixture(changed, "indicator.lifecycle.manage"));
  expect(screen.queryByText("Lifecycle interval content")).toBeNull();
});
