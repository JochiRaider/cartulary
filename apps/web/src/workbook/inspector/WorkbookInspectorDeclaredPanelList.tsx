import type {
  InspectorConfig,
  InspectorDisabledCondition,
  InspectorPanelId,
} from "@cartulary/view-contracts";
import type { ReactNode } from "react";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import {
  type InspectorContextualCapability,
  inspectorContextualCapabilities,
} from "./inspectorCapabilityResolver";
import { WorkbookInspectorPanelSection } from "./presentation/WorkbookInspectorShell";
import { WorkbookInspectorContextualActions } from "./WorkbookInspectorContextualActions";
import type { WorkbookInspectorSubject } from "./workbookInspectorSubject";

export function WorkbookInspectorDeclaredPanelList({
  config,
  contentByPanel,
  currentIncidentRole,
  disabledTokens,
  onContextualAction,
  subject,
}: {
  readonly config: InspectorConfig;
  readonly contentByPanel: Partial<Record<InspectorPanelId, ReactNode>>;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly onContextualAction: (
    capability: InspectorContextualCapability,
  ) => void;
  readonly subject: WorkbookInspectorSubject | null;
}) {
  return config.panels.map((panel) => {
    const content = contentByPanel[panel.panelId];
    if (subject?.kind === "deleted" && panel.panelId !== "history") {
      return null;
    }
    if (
      subject === null &&
      (panel.panelId === "history" || content === null || content === undefined)
    ) {
      return null;
    }
    const capabilities =
      subject?.kind === "live"
        ? inspectorContextualCapabilities({
            config,
            panelId: panel.panelId,
          })
        : [];
    return (
      <WorkbookInspectorPanelSection
        key={panel.panelId}
        panel={panel}
        viewSchemaId={config.viewSchemaId}
      >
        {subject?.kind !== "live" || capabilities.length === 0 ? null : (
          <WorkbookInspectorContextualActions
            capabilities={capabilities}
            config={config}
            currentIncidentRole={currentIncidentRole}
            disabledTokens={disabledTokens}
            onAction={onContextualAction}
          />
        )}
        {content}
      </WorkbookInspectorPanelSection>
    );
  });
}
