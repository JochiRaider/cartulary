import {
  cellPresenceMarkerTestId,
  rowPresenceMarkerTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import {
  displayInitials,
  type PresenceRecord,
  visiblePresence,
} from "../utils/workbookPresence";

export function WorkbookCellPresenceMarker({
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
  if (presences.length < 1) return null;
  const visible = visiblePresence(presences, 2);
  return (
    <span
      aria-label={`${presences
        .map((presence) => presence.display_name)
        .join(", ")} editing ${fieldLabel}`}
      data-testid={cellPresenceMarkerTestId(recordId ?? "draft", fieldKey)}
      role="img"
      style={cellPresenceStyle}
    >
      {visible.shown.map((presence, index) => (
        <span
          aria-hidden="true"
          key={presence.connection_id}
          style={{
            ...presenceMarkerAvatarStyle,
            marginInlineStart: index === 0 ? 0 : "-0.2rem",
          }}
        >
          {displayInitials(presence.display_name)}
        </span>
      ))}
      {visible.overflow > 0 ? (
        <span aria-hidden="true" style={presenceMarkerOverflowStyle}>
          +{visible.overflow}
        </span>
      ) : null}
    </span>
  );
}

export function WorkbookRowGutterContent({
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
      {presences.length === 0 ? (
        <span aria-hidden="true">{ordinal}</span>
      ) : null}
      <WorkbookRowPresenceMarker presences={presences} recordId={recordId} />
    </span>
  );
}

function WorkbookRowPresenceMarker({
  presences,
  recordId,
}: {
  readonly presences: readonly PresenceRecord[];
  readonly recordId: string | null;
}) {
  if (presences.length < 1) return null;
  const visible = visiblePresence(presences, 3);
  return (
    <span
      aria-label={`${presences
        .map((presence) => presence.display_name)
        .join(", ")} focused on this row`}
      data-testid={rowPresenceMarkerTestId(recordId ?? "draft")}
      role="img"
      style={rowGutterPresenceStyle}
    >
      {visible.shown.map((presence, index) => (
        <span
          aria-hidden="true"
          key={presence.connection_id}
          style={{
            ...presenceMarkerAvatarStyle,
            marginInlineStart: index === 0 ? 0 : "-0.2rem",
          }}
        >
          {displayInitials(presence.display_name)}
        </span>
      ))}
      {visible.overflow > 0 ? (
        <span aria-hidden="true" style={presenceMarkerOverflowStyle}>
          +{visible.overflow}
        </span>
      ) : null}
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

const presenceMarkerAvatarStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  inlineSize: "0.7rem",
  blockSize: "0.8rem",
  borderRadius: "var(--ct-rounded-pill)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  fontSize: "0.42rem",
  lineHeight: 1,
} satisfies CSSProperties;

const presenceMarkerOverflowStyle = {
  marginInlineStart: "0.1rem",
  fontSize: "0.58rem",
  lineHeight: 1,
} satisfies CSSProperties;
