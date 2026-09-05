import {
  saveStateActionButtonTestId,
  saveStateTestId,
  workbookPresenceSummaryTestId,
} from "@cartulary/ui-contracts";
import { type CSSProperties, useId } from "react";
import {
  emptyPresenceScope,
  type PresenceScope,
} from "../collaboration/workbookPresencePresentation";
import { WorkbookContinuityAnchorStatus } from "../continuity/useWorkbookGridContinuity";
import type { WorkbookContinuityAnchor } from "../continuity/workbookContinuityPort";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import type { WorkbookStatusPresentation } from "../runtime/workbookMutationStatusProjector";
import { displayInitials } from "../utils/workbookPresence";
import type { WorkbookStatusAction } from "../utils/workbookStatusSecondary";
import {
  presenceAvatarStyle,
  presenceEmptyStyle,
  presenceOverflowStyle,
  statusIconStyle,
  statusStripItemStyle,
  statusStripPresenceStyle,
  statusStripSecondaryItemStyle,
  visuallyHiddenStyle,
} from "../utils/workbookStyles";
import { presenceAccessibleLabel } from "./WorkbookPresenceMarkers";

export type WorkbookConflictActivation = (
  invoker: HTMLButtonElement,
  action: WorkbookStatusAction,
) => void;

/** Visible status is a projection. Save announcements belong to the shell host. */
export function WorkbookStatusStrip({
  presence = emptyPresenceScope,
  status,
  chromeMode,
  showPresence = true,
  onActivateConflict,
  workbookFocusAnchor,
}: {
  readonly presence?: PresenceScope | undefined;
  readonly status: WorkbookStatusPresentation;
  readonly chromeMode: WorkbookChromeMode;
  readonly showPresence?: boolean | undefined;
  readonly onActivateConflict?: WorkbookConflictActivation | undefined;
  readonly workbookFocusAnchor: WorkbookContinuityAnchor | null;
}) {
  const detailId = useId();
  const primaryDescriptionId = useId();
  const message = status.secondary?.message ?? null;
  const action = status.action;
  const canActivate = action !== null && onActivateConflict !== undefined;
  const label = (
    <strong data-density-role="narrow-metadata" data-testid={saveStateTestId()}>
      {status.primaryLabel}
    </strong>
  );
  const narrowText =
    message === null
      ? null
      : Array.from(message).slice(0, 40).join("") +
        (Array.from(message).length > 40 ? "…" : "");
  const showSecondary =
    chromeMode !== "below_supported_minimum" && message !== null;
  return (
    <>
      <span style={{ ...statusStripItemStyle, flex: "0 0 auto" }}>
        <span aria-hidden="true" style={statusIconStyle(status.primaryLabel)} />
        {canActivate ? (
          <button
            aria-label={
              status.primaryLabel === "Conflict"
                ? "Open conflict recovery"
                : "Open save status details"
            }
            aria-describedby={`${primaryDescriptionId}${showSecondary ? ` ${detailId}` : ""}`}
            data-testid={saveStateActionButtonTestId()}
            style={statusStripActionButtonStyle}
            type="button"
            onClick={(event) => onActivateConflict(event.currentTarget, action)}
          >
            {label}
          </button>
        ) : (
          label
        )}
        {canActivate ? (
          <span id={primaryDescriptionId} style={visuallyHiddenStyle}>
            {status.primaryLabel}
            {status.unresolvedConflictCount > 0
              ? `. ${status.unresolvedConflictCount} unresolved`
              : ""}
          </span>
        ) : null}
      </span>
      {showSecondary ? (
        <span
          id={detailId}
          style={
            chromeMode === "compact_desktop"
              ? visuallyHiddenStyle
              : statusStripSecondaryItemStyle
          }
        >
          {chromeMode === "narrow_desktop" ? (
            <>
              <span aria-hidden="true">{narrowText}</span>
              <span style={visuallyHiddenStyle}>{message}</span>
            </>
          ) : (
            message
          )}
        </span>
      ) : null}
      {showPresence && chromeMode !== "below_supported_minimum" ? (
        <WorkbookPresenceSummary records={presence} />
      ) : null}
      <WorkbookContinuityAnchorStatus anchor={workbookFocusAnchor} />
    </>
  );
}

export function WorkbookPresenceSummary({
  records,
}: {
  readonly records: PresenceScope;
}) {
  const visible = records;
  return (
    <div
      aria-label={presenceAccessibleLabel(records, "present on this sheet")}
      data-testid={workbookPresenceSummaryTestId()}
      role="img"
      style={statusStripPresenceStyle}
    >
      <span>Presence</span>
      {visible.shown.length === 0 ? (
        <span style={presenceEmptyStyle}>0</span>
      ) : (
        visible.shown.map((presence) => (
          <span
            key={presence.user_id}
            aria-hidden="true"
            style={presenceAvatarStyle}
          >
            {displayInitials(presence.display_name)}
          </span>
        ))
      )}
      {visible.overflow > 0 ? (
        <span
          aria-label={`${visible.overflow} additional collaborators`}
          role="img"
          style={presenceOverflowStyle}
        >
          +{visible.overflow}
        </span>
      ) : null}
    </div>
  );
}

const statusStripActionButtonStyle = {
  appearance: "none",
  border: 0,
  borderRadius: "var(--ct-rounded-xs)",
  background: "transparent",
  color: "inherit",
  cursor: "pointer",
  font: "inherit",
  padding: "0.1rem 0.2rem",
} satisfies CSSProperties;
