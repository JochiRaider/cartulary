import { workbookInspectorFeatureActionTestId } from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { decisionReview } from "../../../testing/decisionSupersessionTestSupport";
import { createWorkbookDecisionSupersessionAdapter } from "../../adapters/createWorkbookDecisionSupersessionAdapter";
import { inspectorContextualCapabilities } from "../../inspector/inspectorCapabilityResolver";
import { WorkbookInspectorContextualActions } from "../../inspector/WorkbookInspectorContextualActions";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

it("Decision supersession resolves exactly once in its declared History panel", () => {
  const config = requireViewContract(
    "cartulary.view.decisions.v1",
  ).inspectorConfig;
  const actions = config.panels
    .flatMap(({ panelId }) =>
      inspectorContextualCapabilities({ config, panelId }),
    )
    .filter(
      (action) => action.featureGroup.featureGroupKey === "decision.supersede",
    );
  expect(actions).toHaveLength(1);
  expect(actions[0]?.featureGroup).toMatchObject({
    panelId: "history",
    requiresConfirmation: true,
    minimumIncidentRole: "reviewer",
  });
});

it("Decision supersession retains the normalized reason in its complete acknowledgement", async () => {
  const target = "00000000-0000-4000-8000-000000000420";
  const replacement = "00000000-0000-4000-8000-000000000421";
  vi.stubGlobal(
    "fetch",
    vi.fn(
      async () =>
        new Response(
          JSON.stringify({
            data: {
              view_schema_id: "cartulary.view.decisions.v1",
              change_set_id: "00000000-0000-4000-8000-000000000511",
              target_record_id: target,
              superseding_record_id: replacement,
              target_row_version: 5,
              superseding_row_version: 7,
              target_status: "superseded",
              reason: "Later evidence",
            },
            meta: { request_id: "decision-receipt" },
          }),
          { status: 200, headers: { "content-type": "application/json" } },
        ),
    ),
  );
  const port = createWorkbookDecisionSupersessionAdapter({
    apiBase: undefined,
    incidentId: "00000000-0000-4000-8000-000000000001",
  });
  const attempt = port.capture(decisionReview(), "captured-decision");
  const result = await port.send(attempt, new AbortController().signal);
  expect(result).toMatchObject({
    kind: "acknowledged",
    receipt: { reason: "Later evidence" },
  });
});

it("Decision supersession uses declared role and disabled-condition admission", () => {
  const config = requireViewContract(
    "cartulary.view.decisions.v1",
  ).inspectorConfig;
  const capabilities = inspectorContextualCapabilities({
    config,
    panelId: "history",
  });
  const invoked = vi.fn();
  for (const role of [null, "viewer", "editor", "reviewer", "admin"] as const) {
    const view = render(
      <WorkbookInspectorContextualActions
        config={config}
        capabilities={capabilities}
        currentIncidentRole={role}
        disabledTokens={new Set()}
        onAction={invoked}
      />,
    );
    const button = screen.getByTestId(
      workbookInspectorFeatureActionTestId(
        config.viewSchemaId,
        "decision.supersede",
      ),
    );
    expect(button.hasAttribute("disabled")).toBe(
      role !== "reviewer" && role !== "admin",
    );
    fireEvent.click(button);
    view.unmount();
  }
  expect(invoked).toHaveBeenCalledTimes(2);
  for (const token of [
    "no_row_selected",
    "incident_closed",
    "authorization_lost",
    "row_version_changed",
  ] as const) {
    const view = render(
      <WorkbookInspectorContextualActions
        config={config}
        capabilities={capabilities}
        currentIncidentRole="reviewer"
        disabledTokens={new Set([token])}
        onAction={invoked}
      />,
    );
    const button = screen.getByTestId(
      workbookInspectorFeatureActionTestId(
        config.viewSchemaId,
        "decision.supersede",
      ),
    );
    expect(button.hasAttribute("disabled")).toBe(true);
    expect(button.getAttribute("aria-describedby")).toBeTruthy();
    fireEvent.click(button);
    view.unmount();
  }
  expect(invoked).toHaveBeenCalledTimes(2);
});
