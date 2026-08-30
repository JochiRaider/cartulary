import {
  workbookInspectorFeatureActionTestId,
  workbookInspectorFeatureGroupTestId,
  workbookInspectorPanelTestId,
} from "@cartulary/ui-contracts";
import type {
  InspectorConfig,
  InspectorFeatureGroup,
  InspectorPanelId,
} from "@cartulary/view-contracts";
import { type CSSProperties, type ReactNode, useEffect, useState } from "react";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import {
  type InspectorDisabledToken,
  inspectorFeatureDisabledTokens,
  resolveSemanticInspectorFeature,
} from "../inspector/semanticInspectorDispatcher";
import { inspectorFeatureGroupsForPanel } from "../models/workbookInspectorModel";

export type { InspectorDisabledToken } from "../inspector/semanticInspectorDispatcher";

export function WorkbookInspectorPanelSection({
  children,
  config,
  currentIncidentRole,
  disabledTokens,
  panelId,
  subjectRecordId,
  subjectRowVersion,
  onFeatureAction,
}: {
  readonly children?: ReactNode;
  readonly config: InspectorConfig;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly disabledTokens: ReadonlySet<InspectorDisabledToken>;
  readonly panelId: InspectorPanelId;
  readonly subjectRecordId: string | null;
  readonly subjectRowVersion: number | null;
  readonly onFeatureAction?: (featureGroup: InspectorFeatureGroup) => void;
}) {
  const [pendingConfirmation, setPendingConfirmation] = useState<{
    readonly featureGroup: InspectorFeatureGroup;
    readonly subjectRecordId: string | null;
    readonly subjectRowVersion: number | null;
    readonly viewSchemaId: string;
  } | null>(null);
  useEffect(() => {
    setPendingConfirmation((current) =>
      current?.subjectRecordId === subjectRecordId &&
      current.subjectRowVersion === subjectRowVersion &&
      current.viewSchemaId === config.viewSchemaId
        ? current
        : null,
    );
  }, [config.viewSchemaId, subjectRecordId, subjectRowVersion]);
  const panel = config.panels.find(
    (candidate) => candidate.panelId === panelId,
  );
  if (!panel) {
    return null;
  }
  const featureGroups = inspectorFeatureGroupsForPanel(config, panelId).flatMap(
    (featureGroup) => {
      const resolution = resolveSemanticInspectorFeature(config, featureGroup);
      return resolution.kind === "unsupported" ? [] : [resolution];
    },
  );
  return (
    <section
      data-testid={workbookInspectorPanelTestId(config.viewSchemaId, panelId)}
      style={panelSectionStyle}
    >
      <div style={panelHeaderStyle}>
        <h3 style={panelTitleStyle}>{panel.label}</h3>
        <div style={featureGroupListStyle}>
          {featureGroups.map((resolution) => {
            const featureGroup = resolution.featureGroup;
            if (resolution.kind === "panel_read") {
              return (
                <span
                  data-feature-group-key={featureGroup.featureGroupKey}
                  data-route-kind={featureGroup.routeBinding.kind}
                  data-route-owner={featureGroup.routeBinding.owner}
                  data-testid={workbookInspectorFeatureGroupTestId(
                    config.viewSchemaId,
                    featureGroup.featureGroupKey,
                  )}
                  key={resolution.semanticKey}
                  style={readFeatureStyle}
                >
                  {featureGroup.label}
                </span>
              );
            }
            const activeDisabledTokens = inspectorFeatureDisabledTokens({
              currentIncidentRole,
              featureGroup,
              stateTokens: disabledTokens,
            });
            const disabled =
              onFeatureAction === undefined ||
              activeDisabledTokens.has("authorization_lost") ||
              featureGroup.disabledWhen.some((token) =>
                activeDisabledTokens.has(token),
              );
            return (
              <button
                aria-disabled={disabled}
                data-feature-group-key={featureGroup.featureGroupKey}
                data-route-kind={featureGroup.routeBinding.kind}
                data-route-owner={featureGroup.routeBinding.owner}
                data-testid={workbookInspectorFeatureActionTestId(
                  config.viewSchemaId,
                  featureGroup.featureGroupKey,
                )}
                disabled={disabled}
                key={resolution.semanticKey}
                style={featureButtonStyle}
                type="button"
                onClick={() => {
                  if (featureGroup.requiresConfirmation) {
                    setPendingConfirmation({
                      featureGroup,
                      subjectRecordId,
                      subjectRowVersion,
                      viewSchemaId: config.viewSchemaId,
                    });
                    return;
                  }
                  onFeatureAction?.(featureGroup);
                }}
              >
                <span
                  data-testid={workbookInspectorFeatureGroupTestId(
                    config.viewSchemaId,
                    featureGroup.featureGroupKey,
                  )}
                >
                  {featureGroup.label}
                </span>
              </button>
            );
          })}
        </div>
      </div>
      {pendingConfirmation !== null ? (
        <div
          aria-label={`Confirm ${pendingConfirmation.featureGroup.featureGroupKey}`}
          aria-modal="true"
          data-feature-group-key={
            pendingConfirmation.featureGroup.featureGroupKey
          }
          role="alertdialog"
          style={confirmationStyle}
        >
          <p style={confirmationTextStyle}>
            Confirm {pendingConfirmation.featureGroup.label} for record{" "}
            <code>
              {pendingConfirmation.subjectRecordId ?? "(no selected record)"}
            </code>{" "}
            at row version {pendingConfirmation.subjectRowVersion ?? "unknown"}.
          </p>
          <div style={confirmationActionsStyle}>
            <button
              style={featureButtonStyle}
              type="button"
              onClick={() => {
                const confirmed = pendingConfirmation.featureGroup;
                setPendingConfirmation(null);
                onFeatureAction?.(confirmed);
              }}
            >
              Confirm
            </button>
            <button
              style={featureButtonStyle}
              type="button"
              onClick={() => setPendingConfirmation(null)}
            >
              Cancel
            </button>
          </div>
        </div>
      ) : null}
      {children}
    </section>
  );
}

const panelSectionStyle = {
  display: "grid",
  gap: "0.75rem",
  paddingBlock: "0.75rem",
  borderBlockStart: "var(--ct-border-hairline)",
} satisfies CSSProperties;

const panelHeaderStyle = {
  display: "grid",
  gap: "0.5rem",
} satisfies CSSProperties;

const panelTitleStyle = {
  margin: 0,
  fontSize: "0.95rem",
} satisfies CSSProperties;

const featureGroupListStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "0.35rem",
} satisfies CSSProperties;

const featureButtonStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
  font: "inherit",
  fontSize: "0.75rem",
  lineHeight: 1.2,
  padding: "0.3rem 0.45rem",
} satisfies CSSProperties;

const readFeatureStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.75rem",
  lineHeight: 1.2,
  padding: "0.3rem 0",
} satisfies CSSProperties;

const confirmationStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  display: "grid",
  gap: "0.5rem",
  padding: "0.65rem",
} satisfies CSSProperties;

const confirmationTextStyle = { margin: 0 } satisfies CSSProperties;

const confirmationActionsStyle = {
  display: "flex",
  gap: "0.4rem",
} satisfies CSSProperties;
