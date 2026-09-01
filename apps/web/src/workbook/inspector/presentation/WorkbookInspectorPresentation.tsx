import {
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureGroupTestId,
  workbookInspectorPanelTestId,
} from "@cartulary/ui-contracts";
import type {
  InspectorConfig,
  InspectorPanelId,
} from "@cartulary/view-contracts";
import { X } from "lucide-react";
import {
  type CSSProperties,
  type KeyboardEvent,
  type ReactNode,
  useEffect,
  useId,
  useMemo,
  useRef,
  useState,
} from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import { workbookSurfaceInspectorPanelStyle } from "../../layout/WorkbookSurfaceLayout";
import {
  type InspectorDisabledToken,
  resolveSemanticInspectorFeature,
} from "../semanticInspectorDispatcher";
import {
  bindWorkbookInspectorAction,
  type WorkbookInspectorActionBinding,
  type WorkbookInspectorSubjectPresentation,
  type WorkbookInspectorTechnicalField,
  workbookInspectorDisabledReason,
  workbookInspectorNoRowMessage,
  workbookInspectorSafePublicMessage,
} from "./workbookInspectorPresentationModel";

export function WorkbookInspectorShell({
  accessibleLabel,
  children,
  eyebrow = "Inspector",
  heading,
  mode = "record",
  noRowHeading,
  onClose,
  subject,
  testId,
  viewSchemaId,
}: {
  readonly accessibleLabel: string;
  readonly children?: ReactNode;
  readonly eyebrow?: string | undefined;
  readonly heading?: string | undefined;
  readonly mode?: "record" | "creation" | undefined;
  readonly noRowHeading: string;
  readonly onClose: () => void;
  readonly subject: WorkbookInspectorSubjectPresentation | null;
  readonly testId?: string | undefined;
  readonly viewSchemaId: string;
}) {
  const headingId = useId();
  return (
    <aside
      aria-label={accessibleLabel}
      aria-labelledby={headingId}
      data-inspector-state={
        subject === null
          ? mode === "creation"
            ? "creation"
            : "no_row_selected"
          : "ready"
      }
      data-record-id={subject?.recordId}
      data-row-version={subject?.rowVersion}
      data-testid={testId}
      data-view-schema-id={viewSchemaId}
      style={shellStyle}
    >
      <header style={headerStyle}>
        <div style={titleRowStyle}>
          <div style={titleStackStyle}>
            <p style={eyebrowStyle}>{eyebrow}</p>
            <h2 id={headingId} style={titleStyle}>
              {heading ?? subject?.label ?? noRowHeading}
            </h2>
          </div>
          <button
            aria-label="Close inspector"
            data-testid={workbookInspectorCloseButtonTestId(viewSchemaId)}
            style={closeButtonStyle}
            type="button"
            onClick={onClose}
          >
            <X aria-hidden="true" size={16} />
          </button>
        </div>
        {subject === null ? (
          mode === "record" ? (
            <p style={messageStyle}>{workbookInspectorNoRowMessage}</p>
          ) : null
        ) : (
          <WorkbookInspectorRecordContext subject={subject} />
        )}
      </header>
      {children}
      {subject === null ? null : (
        <section aria-label="Record technical metadata" style={metadataStyle}>
          <WorkbookInspectorTechnicalDetails
            fields={workbookInspectorSubjectTechnicalFields(subject)}
          />
        </section>
      )}
    </aside>
  );
}

export function WorkbookInspectorRecordContext({
  subject,
}: {
  readonly subject: WorkbookInspectorSubjectPresentation;
}) {
  return (
    <div style={recordContextStyle}>
      <WorkbookInspectorCompactMetadata>
        <span>{subject.surfaceLabel}</span>
        {subject.stateLabel ? <span>{subject.stateLabel}</span> : null}
      </WorkbookInspectorCompactMetadata>
    </div>
  );
}

function workbookInspectorSubjectTechnicalFields(
  subject: WorkbookInspectorSubjectPresentation,
): WorkbookInspectorTechnicalField[] {
  return [
    { label: "Record ID", value: subject.recordId },
    ...(subject.rowVersion === null
      ? []
      : [{ label: "Row version", value: String(subject.rowVersion) }]),
  ];
}

export function WorkbookInspectorPanelSection({
  children,
  config,
  panelId,
}: {
  readonly children?: ReactNode;
  readonly config: InspectorConfig;
  readonly panelId: InspectorPanelId;
}) {
  const panel = config.panels.find(
    (candidate) => candidate.panelId === panelId,
  );
  if (panel === undefined) return null;
  for (const featureGroup of config.featureGroups) {
    if (featureGroup.panelId !== panelId) continue;
    resolveSemanticInspectorFeature(config, featureGroup);
  }
  return (
    <section
      data-testid={workbookInspectorPanelTestId(config.viewSchemaId, panelId)}
      style={panelSectionStyle}
    >
      <h3 style={panelTitleStyle}>{panel.label}</h3>
      {children}
    </section>
  );
}

export function WorkbookInspectorActionGroup({
  bindings,
  children,
  label = "Actions",
}: {
  readonly bindings: readonly WorkbookInspectorActionBinding[];
  readonly children: ReactNode;
  readonly label?: string | undefined;
}) {
  if (bindings.length === 0) return null;
  return (
    <fieldset aria-label={label} style={actionGroupStyle}>
      {bindings.map((binding) => (
        <span
          aria-hidden="true"
          data-feature-group-key={binding.featureGroup.featureGroupKey}
          data-route-kind={binding.featureGroup.routeBinding.kind}
          data-route-owner={binding.featureGroup.routeBinding.owner}
          data-testid={workbookInspectorFeatureGroupTestId(
            binding.viewSchemaId,
            binding.featureGroup.featureGroupKey,
          )}
          key={binding.semanticKey}
          style={semanticMarkerStyle}
        />
      ))}
      {children}
    </fieldset>
  );
}

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
  readonly disabledTokens: ReadonlySet<InspectorDisabledToken>;
  readonly featureGroups: InspectorConfig["featureGroups"];
  readonly onAction: (
    featureGroup: InspectorConfig["featureGroups"][number],
  ) => void;
  readonly subject: WorkbookInspectorSubjectPresentation;
}) {
  const subjectIdentity = `${config.viewSchemaId}:${subject.recordId}:${subject.rowVersion}`;
  return (
    <WorkbookInspectorContextualActionsForSubject
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

function WorkbookInspectorContextualActionsForSubject({
  config,
  currentIncidentRole,
  disabledTokens,
  featureGroups,
  onAction,
  subject,
}: {
  readonly config: InspectorConfig;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledToken>;
  readonly featureGroups: InspectorConfig["featureGroups"];
  readonly onAction: (
    featureGroup: InspectorConfig["featureGroups"][number],
  ) => void;
  readonly subject: WorkbookInspectorSubjectPresentation;
}) {
  const bindings = useMemo(
    () =>
      featureGroups.flatMap((featureGroup) => {
        const binding = bindWorkbookInspectorAction(config, featureGroup);
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

function WorkbookInspectorContextualAction({
  binding,
  currentIncidentRole,
  disabledTokens,
  onInvoke,
}: {
  readonly binding: WorkbookInspectorActionBinding;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledToken>;
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
      <button
        {...workbookInspectorActionSemanticProps(
          binding,
          reason === null ? undefined : reasonId,
        )}
        disabled={reason !== null}
        style={secondaryButtonStyle}
        type="button"
        onClick={onInvoke}
      >
        {binding.featureGroup.label}
      </button>
      {reason === null ? null : (
        <WorkbookInspectorDisabledReason id={reasonId}>
          {reason}
        </WorkbookInspectorDisabledReason>
      )}
    </div>
  );
}

export function WorkbookInspectorStatus({
  children,
  tone = "neutral",
}: {
  readonly children: ReactNode;
  readonly tone?: "neutral" | "pending" | "success" | "error" | undefined;
}) {
  return (
    <p
      aria-live={tone === "error" ? "assertive" : "polite"}
      role={tone === "error" ? "alert" : "status"}
      style={{ ...messageStyle, color: statusColor[tone] }}
    >
      {children}
    </p>
  );
}

export function WorkbookInspectorCompactMetadata({
  children,
}: {
  readonly children: ReactNode;
}) {
  return <div style={compactMetadataStyle}>{children}</div>;
}

export function WorkbookInspectorTechnicalDetails({
  fields,
}: {
  readonly fields: readonly WorkbookInspectorTechnicalField[];
}) {
  if (fields.length === 0) return null;
  return (
    <details style={technicalDetailsStyle}>
      <summary>Technical details</summary>
      <dl style={technicalListStyle}>
        {fields.map((field) => (
          <div key={field.label}>
            <dt style={technicalTermStyle}>{field.label}</dt>
            <dd style={technicalValueStyle}>{field.value}</dd>
          </div>
        ))}
      </dl>
    </details>
  );
}

export function WorkbookInspectorMessage({
  children,
  error = false,
  testId,
}: {
  readonly children: ReactNode;
  readonly error?: boolean | undefined;
  readonly testId?: string | undefined;
}) {
  return (
    <p
      aria-live={error ? "assertive" : "polite"}
      data-testid={testId}
      role={error ? "alert" : "status"}
      style={{
        ...messageStyle,
        color: error
          ? "var(--ct-colors-semantic-conflict)"
          : "var(--ct-colors-ink-muted)",
      }}
    >
      {children}
    </p>
  );
}

export function WorkbookInspectorPublicError({
  message,
  testId,
}: {
  readonly message: string;
  readonly testId?: string | undefined;
}) {
  const primaryMessage = workbookInspectorSafePublicMessage(message);
  return (
    <div
      aria-live="assertive"
      data-testid={testId}
      role="alert"
      style={publicErrorStyle}
    >
      <p style={messageStyle}>{primaryMessage}</p>
      {primaryMessage === message ? null : (
        <WorkbookInspectorTechnicalDetails
          fields={[{ label: "Public error code", value: message }]}
        />
      )}
    </div>
  );
}

export function WorkbookInspectorConfirmation({
  cancelLabel = "Cancel",
  cancelTestId,
  confirmLabel,
  confirmTestId,
  destructive = false,
  onCancel,
  onConfirm,
  operation,
  subject,
  technicalFields = [],
  testId,
}: {
  readonly cancelLabel?: string | undefined;
  readonly cancelTestId?: string | undefined;
  readonly confirmLabel: string;
  readonly confirmTestId?: string | undefined;
  readonly destructive?: boolean | undefined;
  readonly onCancel: () => void;
  readonly onConfirm: () => void;
  readonly operation: string;
  readonly subject: string;
  readonly technicalFields?: readonly WorkbookInspectorTechnicalField[];
  readonly testId?: string | undefined;
}) {
  const safeControlRef = useRef<HTMLButtonElement>(null);
  const invokingControlRef = useRef<HTMLElement | null>(null);
  useEffect(() => {
    invokingControlRef.current =
      document.activeElement instanceof HTMLElement
        ? document.activeElement
        : null;
    safeControlRef.current?.focus({ preventScroll: true });
    return () => {
      if (invokingControlRef.current?.isConnected) {
        invokingControlRef.current.focus({ preventScroll: true });
      }
    };
  }, []);
  const cancel = () => onCancel();
  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key !== "Escape") return;
    event.preventDefault();
    event.stopPropagation();
    cancel();
  };
  return (
    <div
      aria-label={`${operation} confirmation`}
      data-testid={testId}
      role="alertdialog"
      style={confirmationStyle}
      onKeyDown={handleKeyDown}
    >
      <p style={confirmationTextStyle}>
        {operation} <strong>{subject}</strong>?
      </p>
      <WorkbookInspectorTechnicalDetails fields={technicalFields} />
      <div style={confirmationActionsStyle}>
        <button
          data-testid={cancelTestId}
          ref={safeControlRef}
          style={secondaryButtonStyle}
          type="button"
          onClick={cancel}
        >
          {cancelLabel}
        </button>
        <button
          data-testid={confirmTestId}
          style={destructive ? destructiveButtonStyle : primaryButtonStyle}
          type="button"
          onClick={onConfirm}
        >
          {confirmLabel}
        </button>
      </div>
    </div>
  );
}

export function workbookInspectorActionSemanticProps(
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

export function WorkbookInspectorDisabledReason({
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

const shellStyle = {
  ...workbookSurfaceInspectorPanelStyle,
  display: "grid",
  alignContent: "start",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;

const headerStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const titleRowStyle = {
  display: "flex",
  alignItems: "start",
  justifyContent: "space-between",
  gap: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
const titleStackStyle = { minWidth: 0 } satisfies CSSProperties;
const eyebrowStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-metadata-fontSize)",
  letterSpacing: "var(--ct-typography-metadata-letterSpacing)",
  textTransform: "uppercase" as const,
} satisfies CSSProperties;
const titleStyle = {
  margin: 0,
  overflowWrap: "anywhere" as const,
} satisfies CSSProperties;
const closeButtonStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "inherit",
  minInlineSize: "var(--ct-density-default-rowHeight)",
  minBlockSize: "var(--ct-density-default-rowHeight)",
} satisfies CSSProperties;
const recordContextStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const metadataStyle = {
  borderBlockStart: "var(--ct-border-hairline)",
  paddingBlockStart: "var(--ct-spacing-sm)",
} satisfies CSSProperties;
const compactMetadataStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "var(--ct-spacing-xs)",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-compact-metadata-fontSize)",
} satisfies CSSProperties;
const panelSectionStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  paddingBlock: "var(--ct-spacing-sm)",
  borderBlockStart: "var(--ct-border-hairline)",
} satisfies CSSProperties;
const panelTitleStyle = { margin: 0 } satisfies CSSProperties;
const actionGroupStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "var(--ct-spacing-xs)",
  margin: 0,
  padding: 0,
  border: 0,
} satisfies CSSProperties;
const semanticMarkerStyle = { display: "none" } satisfies CSSProperties;
const contextualActionStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xxs)",
} satisfies CSSProperties;
const messageStyle = { margin: 0 } satisfies CSSProperties;
const publicErrorStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  color: "var(--ct-colors-semantic-conflict)",
} satisfies CSSProperties;
const technicalDetailsStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-compact-metadata-fontSize)",
} satisfies CSSProperties;
const technicalListStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  marginBlockEnd: 0,
} satisfies CSSProperties;
const technicalTermStyle = {
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;
const technicalValueStyle = {
  margin: 0,
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  overflowWrap: "anywhere" as const,
} satisfies CSSProperties;
const confirmationStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-sm)",
  border: "var(--ct-border-hairline)",
  borderColor: "var(--ct-colors-semantic-caution)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
} satisfies CSSProperties;
const confirmationTextStyle = { margin: 0 } satisfies CSSProperties;
const confirmationActionsStyle = {
  display: "flex",
  justifyContent: "flex-end",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const baseButtonStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  color: "inherit",
  font: "inherit",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
} satisfies CSSProperties;
const secondaryButtonStyle = {
  ...baseButtonStyle,
  background: "var(--ct-colors-surface-1)",
} satisfies CSSProperties;
const primaryButtonStyle = {
  ...baseButtonStyle,
  background: "var(--ct-colors-accent)",
  color: "var(--ct-colors-on-accent)",
} satisfies CSSProperties;
const destructiveButtonStyle = {
  ...baseButtonStyle,
  borderColor: "var(--ct-colors-semantic-destructive)",
  background: "var(--ct-component-button-danger-backgroundColor)",
  color: "var(--ct-component-button-danger-textColor)",
} satisfies CSSProperties;
const disabledReasonStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-metadata-fontSize)",
} satisfies CSSProperties;
const statusColor = {
  neutral: "var(--ct-colors-ink-muted)",
  pending: "var(--ct-colors-semantic-caution)",
  success: "var(--ct-colors-semantic-success)",
  error: "var(--ct-colors-semantic-conflict)",
} as const;
