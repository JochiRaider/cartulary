import type { InspectorDisabledCondition } from "@cartulary/view-contracts";
import {
  type ComponentPropsWithRef,
  type CSSProperties,
  type ReactNode,
  useId,
} from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type { WorkbookInspectorActionBinding } from "./workbookInspectorPresentationModel";
import { workbookInspectorDisabledReason } from "./workbookInspectorPresentationModel";

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
  onInvoke,
}: {
  readonly binding: WorkbookInspectorActionBinding;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledCondition>;
  readonly onInvoke: () => void;
}) {
  const reasonId = useId();
  const reason = workbookInspectorDisabledReason({
    currentIncidentRole,
    featureGroup: binding.featureGroup,
    stateTokens: disabledTokens,
  });
  return (
    <div style={contextualActionStyle}>
      <WorkbookInspectorActionButton
        {...workbookInspectorActionSemanticProps(
          binding,
          reason === null ? undefined : reasonId,
        )}
        disabled={reason !== null}
        tone="secondary"
        onClick={onInvoke}
      >
        {binding.featureGroup.label}
      </WorkbookInspectorActionButton>
      {reason === null ? null : (
        <WorkbookInspectorDisabledReason id={reasonId}>
          {reason}
        </WorkbookInspectorDisabledReason>
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

function WorkbookInspectorDisabledReason({
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
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "var(--ct-spacing-xs)",
  margin: 0,
  padding: 0,
  border: 0,
} satisfies CSSProperties;
const contextualActionStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xxs)",
} satisfies CSSProperties;
const baseButtonStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  color: "inherit",
  font: "inherit",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
} satisfies CSSProperties;
const buttonStyleByTone = {
  ordinary: {
    ...baseButtonStyle,
    background: "var(--ct-colors-surface-1)",
  },
  secondary: {
    ...baseButtonStyle,
    background: "var(--ct-colors-surface-1)",
  },
  primary: {
    ...baseButtonStyle,
    background: "var(--ct-colors-accent)",
    color: "var(--ct-colors-on-accent)",
  },
  destructive: {
    ...baseButtonStyle,
    borderColor: "var(--ct-colors-semantic-destructive)",
    background: "var(--ct-component-button-danger-backgroundColor)",
    color: "var(--ct-component-button-danger-textColor)",
  },
} as const satisfies Record<string, CSSProperties>;
const disabledReasonStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-metadata-fontSize)",
} satisfies CSSProperties;
