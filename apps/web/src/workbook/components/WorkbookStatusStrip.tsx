import {
  saveStateTestId,
  statusStripQueueCountTestId,
} from "@cartulary/ui-contracts";
import {
  type WorkbookFocusAnchor,
  WorkbookFocusAnchorStatus,
} from "../utils/workbookGridFocus";
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

export function WorkbookStatusStrip({
  activeSheetPresenceRecords,
  inFlightCount,
  queuedCount,
  saveState,
  saveStateSecondaryMessage,
  workbookFocusAnchor,
}: {
  readonly activeSheetPresenceRecords: readonly PresenceRecord[];
  readonly inFlightCount: number;
  readonly queuedCount: number;
  readonly saveState: WorkbookStatusSaveState;
  readonly saveStateSecondaryMessage: string | null;
  readonly workbookFocusAnchor: WorkbookFocusAnchor | null;
}) {
  const headerPresence = visiblePresence(activeSheetPresenceRecords, 5);
  return (
    <>
      <span style={statusStripItemStyle}>
        <span aria-hidden="true" style={statusIconStyle(saveState)} />
        <strong
          aria-live="polite"
          aria-label="Save state"
          data-density-role="narrow-metadata"
          data-testid={saveStateTestId()}
          role="status"
        >
          {saveState}
        </strong>
      </span>
      {saveStateSecondaryMessage !== null ? (
        <span style={statusStripSecondaryItemStyle}>
          {saveStateSecondaryMessage}
        </span>
      ) : null}
      <span style={statusStripItemStyle}>
        Queue{" "}
        <span data-testid={statusStripQueueCountTestId()}>
          {queuedCount + inFlightCount}
        </span>
      </span>
      <div
        aria-label={`${activeSheetPresenceRecords.length} collaborators present on this sheet`}
        data-testid="presence-header"
        role="status"
        style={statusStripPresenceStyle}
      >
        <span>Presence</span>
        {headerPresence.shown.length === 0 ? (
          <span style={presenceEmptyStyle}>0</span>
        ) : (
          headerPresence.shown.map((presence) => (
            <span
              key={presence.connection_id}
              title={presence.display_name}
              style={presenceAvatarStyle}
            >
              {displayInitials(presence.display_name)}
            </span>
          ))
        )}
        {headerPresence.overflow > 0 ? (
          <span style={presenceOverflowStyle}>+{headerPresence.overflow}</span>
        ) : null}
      </div>
      <WorkbookFocusAnchorStatus anchor={workbookFocusAnchor} />
    </>
  );
}

export function WorkbookSurfaceStatusStrip({
  mutationError = null,
  mutationState,
  workbookFocusAnchor,
}: {
  readonly mutationError?: string | null | undefined;
  readonly mutationState: WorkbookStatusSaveState;
  readonly workbookFocusAnchor: WorkbookFocusAnchor | null;
}) {
  return (
    <>
      <span style={statusStripItemStyle}>
        <span aria-hidden="true" style={statusIconStyle(mutationState)} />
        <strong
          aria-live="polite"
          aria-label="Save state"
          data-density-role="narrow-metadata"
          data-testid={saveStateTestId()}
          role="status"
        >
          {mutationState}
        </strong>
      </span>
      {mutationError ? (
        <span
          aria-live="polite"
          data-testid="generic-mutation-error"
          role="status"
          style={surfaceStatusStripErrorStyle}
        >
          {mutationError}
        </span>
      ) : null}
      <WorkbookFocusAnchorStatus anchor={workbookFocusAnchor} />
    </>
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
