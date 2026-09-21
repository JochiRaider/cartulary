import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import type { InspectorDisabledCondition } from "@cartulary/view-contracts";
import {
  type ComponentPropsWithRef,
  type CSSProperties,
  type ReactNode,
  useId,
} from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import {
  workbookFormButtonStyle,
  workbookTypography,
} from "../../components/workbookFormStyles";
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
        tone="secondary"
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
    | "ordinary"
    | "primary"
    | "secondary"
    | "destructive"
    | undefined;
}) {
  return (
    <button
      {...props}
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
  justifyItems: "start",
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
