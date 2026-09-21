import type {
  InspectorConfig,
  InspectorDisabledCondition,
  InspectorPanelId,
} from "@cartulary/view-contracts";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import {
  type InspectorContextualCapability,
  inspectorContextualCapabilities,
} from "./inspectorCapabilityResolver";
import {
  WorkbookInspectorPanelContent,
  type WorkbookInspectorPanelModel,
} from "./presentation/WorkbookInspectorPanelContent";
import { WorkbookInspectorPanelSection } from "./presentation/WorkbookInspectorShell";
import type { WorkbookInspectorDisabledReason } from "./presentation/workbookInspectorPresentationModel";
import { WorkbookInspectorContextualActions } from "./WorkbookInspectorContextualActions";
import type { WorkbookInspectorSubject } from "./workbookInspectorSubject";

export function WorkbookInspectorDeclaredPanelList({
  config,
  modelsByPanel,
  currentIncidentRole,
  disabledTokens,
  additionalDisabledReasons,
  panelRef,
  onContextualAction,
  subject,
  creationAttachment,
}: {
  readonly config: InspectorConfig;
  readonly modelsByPanel: Partial<
    Record<InspectorPanelId, WorkbookInspectorPanelModel | undefined>
  >;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly additionalDisabledReasons?:
    | ReadonlyMap<string, WorkbookInspectorDisabledReason>
    | undefined;
  readonly panelRef?:
    | ((panelId: InspectorPanelId, element: HTMLElement | null) => void)
    | undefined;
  readonly onContextualAction: (
    capability: InspectorContextualCapability,
  ) => void;
  readonly subject: WorkbookInspectorSubject | null;
  readonly creationAttachment?:
    | { readonly id: string; readonly viewSchemaId: string }
    | undefined;
}) {
  return config.panels.map((panel) => {
    const model = modelsByPanel[panel.panelId];
    if (subject?.kind === "deleted" && panel.panelId !== "history") {
      return null;
    }
    if (
      subject === null &&
      (panel.panelId === "history" || model === undefined)
    ) {
      return null;
    }
    if (!model)
      throw new Error(
        `Missing inspector panel contribution: ${config.viewSchemaId}/${panel.panelId}`,
      );
    if (
      subject === null &&
      (!creationAttachment?.id ||
        creationAttachment.viewSchemaId !== config.viewSchemaId)
    )
      throw new Error(
        `Missing inspector creation attachment: ${config.viewSchemaId}/${panel.panelId}`,
      );
    if (model.access === "concealed" || currentIncidentRole === null)
      return null;
    const capabilities =
      subject?.kind === "live"
        ? inspectorContextualCapabilities({
            config,
            panelId: panel.panelId,
          })
        : [];
    return (
      <WorkbookInspectorPanelSection
        elementRef={(element) => panelRef?.(panel.panelId, element)}
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
            additionalDisabledReasons={additionalDisabledReasons}
            onAction={onContextualAction}
          />
        )}
        <WorkbookInspectorPanelContent model={model} />
      </WorkbookInspectorPanelSection>
    );
  });
}
