import type {
  InspectorConfig,
  InspectorDisabledCondition,
  InspectorPanelId,
} from "@cartulary/view-contracts";
import type { ReactNode } from "react";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import { workbookInspectorSectionFocusDestination } from "../layout/workbookInspectorNavigation";
import {
  type InspectorContextualCapability,
  inspectorContextualCapabilities,
} from "./inspectorCapabilityResolver";
import {
  WorkbookInspectorPanelContent,
  type WorkbookInspectorPanelModel,
} from "./presentation/WorkbookInspectorPanelContent";
import type { WorkbookInspectorSection } from "./presentation/WorkbookInspectorShell";
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
  children,
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
  readonly children: (
    sections: readonly WorkbookInspectorSection[],
  ) => ReactNode;
}) {
  const sections = config.panels.flatMap(
    (panel): WorkbookInspectorSection[] => {
      const model = modelsByPanel[panel.panelId];
      if (subject?.kind === "deleted" && panel.panelId !== "history") {
        return [];
      }
      if (
        subject === null &&
        (panel.panelId === "history" || model === undefined)
      ) {
        return [];
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
        return [];
      const capabilities =
        subject?.kind === "live"
          ? inspectorContextualCapabilities({
              config,
              panelId: panel.panelId,
            })
          : [];
      return [
        {
          panel,
          focusDestination: workbookInspectorSectionFocusDestination,
          elementRef: (element) => panelRef?.(panel.panelId, element),
          content: (
            <>
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
            </>
          ),
        },
      ];
    },
  );
  return children(sections);
}
