import {
  genericWorkbookTestId,
  saveStateActionButtonTestId,
  saveStateTestId,
  workbookPresenceSummaryTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import { WorkbookContinuityAnchorStatus } from "../continuity/useWorkbookGridContinuity";
import type { WorkbookContinuityAnchor } from "../continuity/workbookContinuityPort";
import {
  displayInitials,
  type PresenceRecord,
  visiblePresence,
} from "../utils/workbookPresence";
import {
  presenceAvatarStyle,
  presenceEmptyStyle,
  presenceOverflowStyle,
  statusIconStyle,
  statusStripItemStyle,
  statusStripPresenceStyle,
  statusStripSecondaryItemStyle,
} from "../utils/workbookStyles";

export type WorkbookStatusSaveState = "Syncing" | "Saved" | "Conflict";
export type WorkbookConflictActivation = (invoker: HTMLButtonElement) => void;

export function WorkbookStatusStrip({
  activeSheetPresenceRecords,
  saveState,
  saveStateSecondaryMessage,
  showPresence = true,
  onActivateConflict,
  workbookFocusAnchor,
}: {
  readonly activeSheetPresenceRecords: readonly PresenceRecord[];
  readonly inFlightCount: number;
  readonly queuedCount: number;
  readonly saveState: WorkbookStatusSaveState;
  readonly saveStateSecondaryMessage: string | null;
  readonly showPresence?: boolean | undefined;
  readonly onActivateConflict?: WorkbookConflictActivation | undefined;
  readonly workbookFocusAnchor: WorkbookContinuityAnchor | null;
}) {
  return (
    <>
      <span style={statusStripItemStyle}>
        <span aria-hidden="true" style={statusIconStyle(saveState)} />
        {saveState === "Conflict" && onActivateConflict !== undefined ? (
          <button
            aria-label="Open conflict recovery"
            data-testid={saveStateActionButtonTestId()}
            style={statusStripActionButtonStyle}
            type="button"
            onClick={(event) => onActivateConflict(event.currentTarget)}
          >
            <strong
              aria-live="polite"
              aria-label="Save state"
              data-density-role="narrow-metadata"
              data-testid={saveStateTestId()}
              role="status"
            >
              {saveState}
            </strong>
          </button>
        ) : (
          <strong
            aria-live="polite"
            aria-label="Save state"
            data-density-role="narrow-metadata"
            data-testid={saveStateTestId()}
            role="status"
          >
            {saveState}
          </strong>
        )}
      </span>
      {saveStateSecondaryMessage !== null ? (
        <span style={statusStripSecondaryItemStyle}>
          {saveStateSecondaryMessage}
        </span>
      ) : null}
      {showPresence ? (
        <WorkbookPresenceSummary records={activeSheetPresenceRecords} />
      ) : null}
      <WorkbookContinuityAnchorStatus anchor={workbookFocusAnchor} />
    </>
  );
}

export function WorkbookSurfaceStatusStrip({
  activeSheetPresenceRecords = [],
  mutationError = null,
  mutationState,
  onActivateConflict,
  showPresence = true,
  workbookFocusAnchor,
}: {
  readonly activeSheetPresenceRecords?: readonly PresenceRecord[] | undefined;
  readonly mutationError?: string | null | undefined;
  readonly mutationState: WorkbookStatusSaveState;
  readonly onActivateConflict?: WorkbookConflictActivation | undefined;
  readonly showPresence?: boolean | undefined;
  readonly workbookFocusAnchor: WorkbookContinuityAnchor | null;
}) {
  return (
    <>
      <span style={statusStripItemStyle}>
        <span aria-hidden="true" style={statusIconStyle(mutationState)} />
        {mutationState === "Conflict" && onActivateConflict !== undefined ? (
          <button
            aria-label="Open conflict recovery"
            data-testid={saveStateActionButtonTestId()}
            style={statusStripActionButtonStyle}
            type="button"
            onClick={(event) => onActivateConflict(event.currentTarget)}
          >
            <strong
              aria-live="polite"
              aria-label="Save state"
              data-density-role="narrow-metadata"
              data-testid={saveStateTestId()}
              role="status"
            >
              {mutationState}
            </strong>
          </button>
        ) : (
          <strong
            aria-live="polite"
            aria-label="Save state"
            data-density-role="narrow-metadata"
            data-testid={saveStateTestId()}
            role="status"
          >
            {mutationState}
          </strong>
        )}
      </span>
      {mutationError ? (
        <span
          aria-live="polite"
          data-testid={genericWorkbookTestId("mutation-error")}
          role="status"
          style={surfaceStatusStripErrorStyle}
        >
          {mutationError}
        </span>
      ) : null}
      {showPresence ? (
        <WorkbookPresenceSummary records={activeSheetPresenceRecords} />
      ) : null}
      <WorkbookContinuityAnchorStatus anchor={workbookFocusAnchor} />
    </>
  );
}

export function WorkbookPresenceSummary({
  records,
}: {
  readonly records: readonly PresenceRecord[];
}) {
  const visible = visiblePresence(records, 5);
  return (
    <div
      aria-label={`${records.length} collaborators present on this sheet`}
      data-testid={workbookPresenceSummaryTestId()}
      role="status"
      style={statusStripPresenceStyle}
    >
      <span>Presence</span>
      {visible.shown.length === 0 ? (
        <span style={presenceEmptyStyle}>0</span>
      ) : (
        visible.shown.map((presence) => (
          <span
            key={presence.connection_id}
            title={presence.display_name}
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

const surfaceStatusStripErrorStyle = {
  minWidth: 0,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 700,
};

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
