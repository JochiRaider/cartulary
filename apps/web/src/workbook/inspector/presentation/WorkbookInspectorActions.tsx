import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import type { InspectorDisabledCondition } from "@cartulary/view-contracts";
import {
  type ComponentPropsWithRef,
  type CSSProperties,
  type ReactNode,
  useContext,
  useId,
} from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import { useWorkbookCommand } from "../../commands/WorkbookCommands";
import {
  workbookFormButtonStyle,
  workbookQuietCommandStyle,
  workbookTypography,
} from "../../components/workbookFormStyles";
import { WorkbookInspectorCommandSubject } from "./WorkbookInspectorCommandSubject";
import type { WorkbookInspectorActionBinding } from "./workbookInspectorPresentationModel";
import {
  type WorkbookInspectorDisabledReason,
  workbookInspectorDisabledReason,
  workbookInspectorDisabledReasonText,
} from "./workbookInspectorPresentationModel";

export function WorkbookInspectorActionGroup({
  children,
  label = "Actions",
}: {
  readonly children: ReactNode;
  readonly label?: string | undefined;
}) {
  return (
    <fieldset aria-label={label} style={actionGroupStyle}>
      {children}
    </fieldset>
  );
}

export function WorkbookInspectorContextualAction({
  binding,
  currentIncidentRole,
  disabledTokens,
  additionalDisabledReason,
  descriptionId,
  outcomeDescriptionId,
  onInvoke,
}: {
  readonly binding: WorkbookInspectorActionBinding;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly additionalDisabledReason?:
    | WorkbookInspectorDisabledReason
    | undefined;
  readonly descriptionId?: string | undefined;
  readonly outcomeDescriptionId?: string | undefined;
  readonly onInvoke: () => void;
}) {
  const subject = useContext(WorkbookInspectorCommandSubject);
  const reasonId = useId();
  const outcomeId = useId();
  const reason =
    workbookInspectorDisabledReason({
      currentIncidentRole,
      featureGroup: binding.featureGroup,
      stateTokens: disabledTokens,
    }) ??
    additionalDisabledReason ??
    null;
  useWorkbookCommand(
    subject
      ? {
          id: `record.${binding.semanticKey}`,
          family:
            binding.capability.kind === "create_related" ||
            binding.capability.kind === "note_create"
              ? "Follow up"
              : binding.capability.kind === "indicator"
                ? "Relate"
                : "Review",
          label: binding.featureGroup.label,
          terms: [binding.featureGroup.panelId],
          targetKind: "record",
          availability: (target) =>
            target.kind !== "record" ||
            target.recordId !== subject.recordId ||
            target.viewSchemaId !== subject.viewSchemaId
              ? "The selected record changed."
              : reason
                ? workbookInspectorDisabledReasonText(reason)
                : null,
          invoke: (target) => {
            if (
              target.kind !== "record" ||
              target.recordId !== subject.recordId ||
              target.viewSchemaId !== subject.viewSchemaId ||
              reason
            )
              return false;
            onInvoke();
            return true;
          },
        }
      : null,
  );
  return (
    <div style={contextualActionStyle}>
      <WorkbookInspectorActionButton
        {...workbookInspectorActionSemanticProps(
          binding,
          [
            outcomeDescriptionId ?? outcomeId,
            ...(reason === null ? [] : [descriptionId ?? reasonId]),
          ].join(" "),
        )}
        disabled={reason !== null}
        tone="quiet"
        style={{
          inlineSize: "100%",
          justifyContent: "flex-start",
          textAlign: "start",
        }}
        onClick={onInvoke}
      >
        {binding.featureGroup.label}
      </WorkbookInspectorActionButton>
      {outcomeDescriptionId ? null : (
        <p id={outcomeId} style={outcomeStyle}>
          {
            cartularyDesignPresentation.inspector.actionOutcomes[
              binding.outcome
            ]
          }
        </p>
      )}
      {reason === null || descriptionId ? null : (
        <WorkbookInspectorDisabledReasonMessage id={reasonId}>
          {workbookInspectorDisabledReasonText(reason)}
        </WorkbookInspectorDisabledReasonMessage>
      )}
    </div>
  );
}

export function WorkbookInspectorActionButton({
  children,
  tone = "ordinary",
  type = "button",
  ...props
}: ComponentPropsWithRef<"button"> & {
  readonly tone?:
    | "quiet"
    | "ordinary"
    | "primary"
    | "secondary"
    | "destructive"
    | undefined;
}) {
  return (
    <button
      {...props}
      data-inspector-action-tone={tone}
      style={{ ...buttonStyleByTone[tone], ...props.style }}
      type={type}
    >
      {children}
    </button>
  );
}

function workbookInspectorActionSemanticProps(
  binding: WorkbookInspectorActionBinding,
  descriptionId?: string,
) {
  return {
    "aria-describedby": descriptionId,
    "data-feature-group-key": binding.featureGroup.featureGroupKey,
    "data-route-kind": binding.featureGroup.routeBinding.kind,
    "data-route-owner": binding.featureGroup.routeBinding.owner,
    "data-testid": binding.testId,
  } as const;
}

export function WorkbookInspectorDisabledReasonMessage({
  children,
  id,
}: {
  readonly children: ReactNode;
  readonly id: string;
}) {
  return (
    <p id={id} style={disabledReasonStyle}>
      {children}
    </p>
  );
}

const actionGroupStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  margin: 0,
  padding: 0,
  border: 0,
} satisfies CSSProperties;
const contextualActionStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  justifyItems: "stretch",
  paddingBlock: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const outcomeStyle = {
  ...workbookTypography("compact-metadata"),
  color: "var(--ct-colors-ink-muted)",
  margin: 0,
} satisfies CSSProperties;
const baseButtonStyle = {
  ...workbookFormButtonStyle,
} satisfies CSSProperties;
const buttonStyleByTone = {
  quiet: { ...workbookQuietCommandStyle, whiteSpace: "normal" },
  ordinary: {
    ...baseButtonStyle,
  },
  secondary: {
    ...baseButtonStyle,
  },
  primary: {
    ...baseButtonStyle,
    background: "var(--ct-component-button-primary-backgroundColor)",
    color: "var(--ct-component-button-primary-textColor)",
    padding: "var(--ct-component-button-primary-padding)",
    borderRadius: "var(--ct-component-button-primary-rounded)",
  },
  destructive: {
    ...baseButtonStyle,
    borderColor: "var(--ct-colors-semantic-destructive)",
    background: "var(--ct-component-button-danger-backgroundColor)",
    color: "var(--ct-component-button-danger-textColor)",
    padding: "var(--ct-component-button-danger-padding)",
    borderRadius: "var(--ct-component-button-danger-rounded)",
  },
} as const satisfies Record<string, CSSProperties>;
const disabledReasonStyle = {
  ...workbookTypography("metadata"),
  flexBasis: "100%",
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;
