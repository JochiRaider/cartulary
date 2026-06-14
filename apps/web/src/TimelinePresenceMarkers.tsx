import {
  cellPresenceMarkerTestId,
  rowPresenceMarkerTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import {
  displayInitials,
  type PresenceRecord,
  visiblePresence,
} from "./workbookPresence";

export function TimelineCellPresenceMarker({
  fieldKey,
  fieldLabel,
  presences,
  recordId,
}: {
  readonly fieldKey: string;
  readonly fieldLabel: string;
  readonly presences: readonly PresenceRecord[];
  readonly recordId: string | null;
}) {
  if (presences.length < 1) {
    return null;
  }
  const visible = visiblePresence(presences, 1);
  return (
    <span
      aria-label={`${presences
        .map((presence) => presence.display_name)
        .join(", ")} editing ${fieldLabel}`}
      data-testid={cellPresenceMarkerTestId(recordId ?? "draft", fieldKey)}
      role="img"
      style={cellPresenceStyle}
    >
      {visible.shown.map((presence) => displayInitials(presence.display_name))}
      {visible.overflow > 0 ? ` +${visible.overflow}` : ""}
    </span>
  );
}

export function TimelineRowGutterContent({
  ordinal,
  presences,
  recordId,
}: {
  readonly ordinal: string;
  readonly presences: readonly PresenceRecord[];
  readonly recordId: string | null;
}) {
  return (
    <span style={rowGutterContentStyle}>
      <span aria-hidden="true">{ordinal}</span>
      <TimelineRowPresenceMarker presences={presences} recordId={recordId} />
    </span>
  );
}

function TimelineRowPresenceMarker({
  presences,
  recordId,
}: {
  readonly presences: readonly PresenceRecord[];
  readonly recordId: string | null;
}) {
  if (presences.length < 1) {
    return null;
  }
  const visible = visiblePresence(presences, 2);
  return (
    <span
      aria-label={`${presences
        .map((presence) => presence.display_name)
        .join(", ")} focused on this row`}
      data-testid={rowPresenceMarkerTestId(recordId ?? "draft")}
      role="img"
      style={rowGutterPresenceStyle}
    >
      {visible.shown
        .map((presence) => displayInitials(presence.display_name))
        .join("")}
      {visible.overflow > 0 ? `+${visible.overflow}` : ""}
    </span>
  );
}

const cellPresenceStyle = {
  position: "absolute",
  insetBlockStart: "4px",
  insetInlineEnd: "6px",
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: "fit-content",
  minHeight: 0,
  height: "18px",
  margin: 0,
  borderRadius: "var(--ct-rounded-pill)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-semantic-presence-other)",
  padding: "0 0.35rem",
  fontSize: "0.68rem",
  fontWeight: 700,
  lineHeight: 1,
} satisfies CSSProperties;

const rowGutterContentStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  gap: "0.2rem",
  minWidth: 0,
} satisfies CSSProperties;

const rowGutterPresenceStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  minWidth: "1rem",
  height: "1rem",
  borderRadius: "var(--ct-rounded-pill)",
  border: "var(--ct-border-hairline)",
  color: "var(--ct-colors-semantic-presence-other)",
  fontSize: "0.62rem",
  lineHeight: 1,
} satisfies CSSProperties;
