import type {
  InspectorConfig,
  InspectorDisabledCondition,
} from "@cartulary/view-contracts";
import { useMemo, useState } from "react";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import {
  WorkbookInspectorActionGroup,
  WorkbookInspectorContextualAction,
} from "./presentation/WorkbookInspectorActions";
import { WorkbookInspectorConfirmation } from "./presentation/WorkbookInspectorFeedback";
import {
  bindWorkbookInspectorAction,
  type WorkbookInspectorActionBinding,
  type WorkbookInspectorSubjectPresentation,
} from "./presentation/workbookInspectorPresentationModel";
import type { InspectorContextualCapability } from "./semanticInspectorDispatcher";

export function WorkbookInspectorContextualActions({
  config,
  currentIncidentRole,
  disabledTokens,
  capabilities,
  onAction,
  subject,
}: {
  readonly config: InspectorConfig;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly capabilities: readonly InspectorContextualCapability[];
  readonly onAction: (capability: InspectorContextualCapability) => void;
  readonly subject: WorkbookInspectorSubjectPresentation;
}) {
  const subjectIdentity = `${config.viewSchemaId}:${subject.recordId}:${subject.rowVersion}`;
  return (
    <ContextualActionsForSubject
      config={config}
      currentIncidentRole={currentIncidentRole}
      disabledTokens={disabledTokens}
      capabilities={capabilities}
      key={subjectIdentity}
      subject={subject}
      onAction={onAction}
    />
  );
}

function ContextualActionsForSubject({
  config,
  currentIncidentRole,
  disabledTokens,
  capabilities,
  onAction,
  subject,
}: {
  readonly config: InspectorConfig;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly capabilities: readonly InspectorContextualCapability[];
  readonly onAction: (capability: InspectorContextualCapability) => void;
  readonly subject: WorkbookInspectorSubjectPresentation;
}) {
  const bindings = useMemo(
    () =>
      capabilities.map((capability) =>
        bindWorkbookInspectorAction(config, capability),
      ),
    [capabilities, config],
  );
  const [pending, setPending] = useState<WorkbookInspectorActionBinding | null>(
    null,
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
          onInvoke={() => {
            if (binding.featureGroup.requiresConfirmation) {
              setPending(binding);
            } else {
              onAction(binding.capability);
            }
          }}
        />
      ))}
      {pending === null ? null : (
        <WorkbookInspectorConfirmation
          confirmLabel={`Confirm ${pending.featureGroup.label}`}
          destructive={pending.featureGroup.mutates}
          operation={pending.featureGroup.label}
          subject={subject.label}
          technicalFields={[
            { label: "Record ID", value: subject.recordId },
            ...(subject.rowVersion === null
              ? []
              : [
                  {
                    label: "Row version",
                    value: String(subject.rowVersion),
                  },
                ]),
          ]}
          onCancel={() => setPending(null)}
          onConfirm={() => {
            const capability = pending.capability;
            setPending(null);
            onAction(capability);
          }}
        />
      )}
    </WorkbookInspectorActionGroup>
  );
}
