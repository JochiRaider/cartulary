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
  statusStripMutedItemStyle,
  statusStripPresenceStyle,
  statusStripSecondaryItemStyle,
  statusStripSpacerStyle,
} from "../utils/workbookStyles";

export type WorkbookStatusSaveState = "Syncing" | "Saved" | "Conflict";

export function WorkbookStatusStrip({
  activeSheetPresenceRecords,
  incidentId,
  inFlightCount,
  queuedCount,
  saveState,
  saveStateSecondaryMessage,
  workbookFocusAnchor,
}: {
  readonly activeSheetPresenceRecords: readonly PresenceRecord[];
  readonly incidentId: string;
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
      <span style={statusStripSpacerStyle} />
      <span style={statusStripMutedItemStyle}>{incidentId}</span>
      <WorkbookFocusAnchorStatus anchor={workbookFocusAnchor} />
    </>
  );
}
