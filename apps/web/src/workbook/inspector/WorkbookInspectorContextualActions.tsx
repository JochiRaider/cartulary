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

export function WorkbookInspectorContextualActions({
  config,
  currentIncidentRole,
  disabledTokens,
  featureGroups,
  onAction,
  subject,
}: {
  readonly config: InspectorConfig;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly featureGroups: InspectorConfig["featureGroups"];
  readonly onAction: (
    featureGroup: InspectorConfig["featureGroups"][number],
  ) => void;
  readonly subject: WorkbookInspectorSubjectPresentation;
}) {
  const subjectIdentity = `${config.viewSchemaId}:${subject.recordId}:${subject.rowVersion}`;
  return (
    <ContextualActionsForSubject
      config={config}
      currentIncidentRole={currentIncidentRole}
      disabledTokens={disabledTokens}
      featureGroups={featureGroups}
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
  featureGroups,
  onAction,
  subject,
}: {
  readonly config: InspectorConfig;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly featureGroups: InspectorConfig["featureGroups"];
  readonly onAction: (
    featureGroup: InspectorConfig["featureGroups"][number],
  ) => void;
  readonly subject: WorkbookInspectorSubjectPresentation;
}) {
  const bindings = useMemo(
    () =>
      featureGroups.flatMap((featureGroup) => {
        const binding = bindWorkbookInspectorAction(
          config,
          featureGroup.featureGroupKey,
        );
        return binding === null ? [] : [binding];
      }),
    [config, featureGroups],
  );
  const [pending, setPending] = useState<WorkbookInspectorActionBinding | null>(
    null,
  );
  if (bindings.length === 0) return null;
  return (
    <WorkbookInspectorActionGroup
      bindings={bindings}
      label="Contextual actions"
    >
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
              onAction(binding.featureGroup);
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
            const featureGroup = pending.featureGroup;
            setPending(null);
            onAction(featureGroup);
          }}
        />
      )}
    </WorkbookInspectorActionGroup>
  );
}
