import type {
  InspectorConfig,
  InspectorDisabledCondition,
} from "@cartulary/view-contracts";
import { type CSSProperties, useId, useMemo } from "react";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { InspectorContextualCapability } from "./inspectorCapabilityResolver";
import {
  WorkbookInspectorActionGroup,
  WorkbookInspectorContextualAction,
  WorkbookInspectorDisabledReasonMessage,
} from "./presentation/WorkbookInspectorActions";
import {
  bindWorkbookInspectorAction,
  type WorkbookInspectorDisabledReason,
  workbookInspectorDisabledReason,
  workbookInspectorDisabledReasonKey,
  workbookInspectorDisabledReasonText,
} from "./presentation/workbookInspectorPresentationModel";

export function WorkbookInspectorContextualActions({
  config,
  currentIncidentRole,
  disabledTokens,
  additionalDisabledReasons,
  capabilities,
  onAction,
}: {
  readonly config: InspectorConfig;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly additionalDisabledReasons?:
    | ReadonlyMap<string, WorkbookInspectorDisabledReason>
    | undefined;
  readonly capabilities: readonly InspectorContextualCapability[];
  readonly onAction: (capability: InspectorContextualCapability) => void;
}) {
  const groupId = useId();
  const bindings = useMemo(() => {
    const unique = new Map(
      capabilities.map((capability) => [capability.semanticKey, capability]),
    );
    const order = new Map(
      config.featureGroups.map((feature, index) => [
        feature.featureGroupKey,
        index,
      ]),
    );
    return [...unique.values()]
      .sort(
        (a, b) =>
          (order.get(a.featureGroup.featureGroupKey) ?? 0) -
          (order.get(b.featureGroup.featureGroupKey) ?? 0),
      )
      .map((capability) => bindWorkbookInspectorAction(config, capability));
  }, [capabilities, config]);
  const reasons = new Map<
    string,
    { id: string; reason: WorkbookInspectorDisabledReason }
  >();
  const descriptions = new Map<string, string>();
  for (const binding of bindings) {
    const reason =
      workbookInspectorDisabledReason({
        currentIncidentRole,
        featureGroup: binding.featureGroup,
        stateTokens: disabledTokens,
      }) ??
      additionalDisabledReasons?.get(binding.featureGroup.featureGroupKey);
    if (!reason) continue;
    const key = workbookInspectorDisabledReasonKey(reason);
    let shared = reasons.get(key);
    if (!shared) {
      shared = { id: `${groupId}-${reasons.size}`, reason };
      reasons.set(key, shared);
    }
    descriptions.set(binding.semanticKey, shared.id);
  }
  if (bindings.length === 0) return null;
  return (
    <WorkbookInspectorActionGroup label="Contextual actions">
      <ol style={listStyle}>
        {bindings.map((binding) => (
          <li key={binding.semanticKey}>
            <WorkbookInspectorContextualAction
              binding={binding}
              descriptionId={descriptions.get(binding.semanticKey)}
              currentIncidentRole={currentIncidentRole}
              disabledTokens={disabledTokens}
              additionalDisabledReason={additionalDisabledReasons?.get(
                binding.featureGroup.featureGroupKey,
              )}
              onInvoke={() => onAction(binding.capability)}
            />
          </li>
        ))}
      </ol>
      {[...reasons.values()].map(({ id, reason }) => (
        <WorkbookInspectorDisabledReasonMessage key={id} id={id}>
          {workbookInspectorDisabledReasonText(reason)}
        </WorkbookInspectorDisabledReasonMessage>
      ))}
    </WorkbookInspectorActionGroup>
  );
}
const listStyle = {
  listStyle: "none",
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  margin: 0,
  padding: 0,
} satisfies CSSProperties;
