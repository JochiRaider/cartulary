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
import type { CSSProperties, ReactNode } from "react";
import { inspectorFeatureGroupsForPanel } from "../models/workbookInspectorModel";

export type InspectorDisabledToken =
  | "no_row_selected"
  | "incident_closed"
  | "authorization_lost"
  | "row_version_changed"
  | "record_deleted"
  | "record_merged"
  | "evidence_preview_unavailable"
  | "merge_target_unavailable"
  | "record_not_deleted"
  | "rollback_target_unavailable"
  | "party_text_unavailable"
  | "pivot_target_unavailable";

export function WorkbookInspectorPanelSection({
  children,
  config,
  disabledTokens,
  panelId,
  onFeatureAction,
}: {
  readonly children?: ReactNode;
  readonly config: InspectorConfig;
  readonly disabledTokens: ReadonlySet<InspectorDisabledToken>;
  readonly panelId: InspectorPanelId;
  readonly onFeatureAction?: (featureGroup: InspectorFeatureGroup) => void;
}) {
  const panel = config.panels.find(
    (candidate) => candidate.panelId === panelId,
  );
  if (!panel) {
    return null;
  }
  const featureGroups = inspectorFeatureGroupsForPanel(config, panelId);
  return (
    <section
      data-testid={workbookInspectorPanelTestId(config.viewSchemaId, panelId)}
      style={panelSectionStyle}
    >
      <div style={panelHeaderStyle}>
        <h3 style={panelTitleStyle}>{panel.label}</h3>
        <div style={featureGroupListStyle}>
          {featureGroups.map((featureGroup) => {
            const disabled = featureGroup.disabledWhen.some((token) =>
              disabledTokens.has(token),
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
                key={featureGroup.featureGroupKey}
                style={featureButtonStyle}
                type="button"
                onClick={() => {
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
