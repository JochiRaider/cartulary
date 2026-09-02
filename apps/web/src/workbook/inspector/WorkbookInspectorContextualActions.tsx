import type {
  InspectorConfig,
  InspectorDisabledCondition,
} from "@cartulary/view-contracts";
import { useMemo } from "react";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { InspectorContextualCapability } from "./inspectorCapabilityResolver";
import {
  WorkbookInspectorActionGroup,
  WorkbookInspectorContextualAction,
} from "./presentation/WorkbookInspectorActions";
import { bindWorkbookInspectorAction } from "./presentation/workbookInspectorPresentationModel";

export function WorkbookInspectorContextualActions({
  config,
  currentIncidentRole,
  disabledTokens,
  capabilities,
  onAction,
}: {
  readonly config: InspectorConfig;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly capabilities: readonly InspectorContextualCapability[];
  readonly onAction: (capability: InspectorContextualCapability) => void;
}) {
  const bindings = useMemo(
    () =>
      capabilities.map((capability) =>
        bindWorkbookInspectorAction(config, capability),
      ),
    [capabilities, config],
  );
  if (bindings.length === 0) return null;
  return (
    <WorkbookInspectorActionGroup label="Contextual actions">
      {bindings.map((binding) => (
        <WorkbookInspectorContextualAction
          binding={binding}
          currentIncidentRole={currentIncidentRole}
          disabledTokens={disabledTokens}
          key={binding.semanticKey}
          onInvoke={() => onAction(binding.capability)}
        />
      ))}
    </WorkbookInspectorActionGroup>
  );
}
