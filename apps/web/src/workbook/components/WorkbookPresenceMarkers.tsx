import {
  cellPresenceMarkerTestId,
  rowPresenceMarkerTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties, ReactElement, ReactNode } from "react";
import type { PresenceScope } from "../collaboration/workbookPresencePresentation";
import { displayInitials } from "../utils/workbookPresence";

// Existing workbook gutter geometry, shared by the base surface renderers.
export const workbookPresenceRowGutter = { width: 58, minWidth: 58 };

export function presenceAccessibleLabel(
  scope: PresenceScope,
  context: string,
): string {
  const identities = scope.users
    .map(
      (user) => `${user.display_name || "Unnamed collaborator"} ${user.mode}`,
    )
    .join(", ");
  return `${scope.users.length} collaborator${scope.users.length === 1 ? "" : "s"} ${context}${identities ? `: ${identities}` : ""}${scope.overflow > 0 ? `; ${scope.overflow} additional collaborators` : ""}`;
}

function PresenceGlyphs({
  scope,
  row = false,
}: {
  readonly scope: PresenceScope;
  readonly row?: boolean;
}) {
  return (
    <>
      {scope.shown.map((presence) => (
        <span
          aria-hidden="true"
          key={presence.user_id}
          style={row ? rowIdentityStyle : cellIdentityStyle}
        >
          {row
            ? Array.from(displayInitials(presence.display_name))[0]
            : displayInitials(presence.display_name)}
        </span>
      ))}
      {scope.overflow > 0 ? (
        <span aria-hidden="true" style={overflowStyle}>
          +{scope.overflow}
        </span>
      ) : null}
    </>
  );
}

export function WorkbookCellPresenceMarker({
  fieldKey,
  fieldLabel,
  presences,
  recordId,
}: {
  readonly fieldKey: string;
  readonly fieldLabel: string;
  readonly presences: PresenceScope;
  readonly recordId: string | null;
}) {
  if (recordId === null || presences.users.length === 0) return null;
  return (
    <span
      aria-label={presenceAccessibleLabel(
        presences,
        `editing ${fieldLabel} on this row`,
      )}
      data-testid={cellPresenceMarkerTestId(recordId, fieldKey)}
      role="img"
      style={cellPresenceStyle}
    >
      <PresenceGlyphs scope={presences} />
    </span>
  );
}

/** A layout slot, not an overlay: values, editors and focus outlines retain their own box. */
export function WorkbookPresenceCellLayout({
  children,
  marker,
  editing = false,
}: {
  readonly children: ReactNode;
  readonly marker: ReactElement<{ presences: PresenceScope }>;
  readonly editing?: boolean;
}) {
  // Keep the editor's DOM identity when peers arrive or leave. Replacing the
  // wrapper would unmount its input, lose focus and publish an editing stop.
  if (!editing && marker.props.presences.users.length === 0)
    return <>{children}</>;
  return (
    <span style={cellLayoutStyle}>
      <span
        style={{
          ...cellContentStyle,
          ...(editing
            ? { overflow: "visible", position: "relative", blockSize: "100%" }
            : {}),
        }}
      >
        {children}
      </span>
      {marker}
    </span>
  );
}

export function WorkbookRowGutterContent({
  ordinal,
  presences,
  recordId,
}: {
  readonly ordinal: string;
  readonly presences: PresenceScope;
  readonly recordId: string | null;
}) {
  if (recordId === null || presences.users.length === 0)
    return <span aria-hidden="true">{ordinal}</span>;
  return (
    <span
      aria-label={presenceAccessibleLabel(presences, "on this row")}
      data-testid={rowPresenceMarkerTestId(recordId)}
      role="img"
      style={rowPresenceStyle}
    >
      <PresenceGlyphs scope={presences} row />
    </span>
  );
}

const identityStyle = {
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  fontSize: "var(--ct-typography-compact-metadata-fontSize)",
  lineHeight: 1,
  flex: "0 0 auto",
  textAlign: "center",
} satisfies CSSProperties;
const rowIdentityStyle = {
  ...identityStyle,
  inlineSize: "1ch",
} satisfies CSSProperties;
const cellIdentityStyle = {
  ...identityStyle,
  inlineSize: "2ch",
} satisfies CSSProperties;
const overflowStyle = {
  ...identityStyle,
  marginInlineStart: "0.25ch",
} satisfies CSSProperties;
const rowPresenceStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  maxInlineSize: "100%",
  color: "var(--ct-colors-semantic-presence-other)",
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  fontSize: "var(--ct-typography-compact-metadata-fontSize)",
  borderBlockEnd: "var(--ct-border-hairline)",
  lineHeight: 1,
} satisfies CSSProperties;
const cellPresenceStyle = {
  ...rowPresenceStyle,
  flex: "0 0 auto",
  gap: "0.25ch",
  borderRadius: "var(--ct-rounded-xs)",
} satisfies CSSProperties;
const cellLayoutStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.5ch",
  minInlineSize: 0,
  inlineSize: "100%",
  blockSize: "100%",
} satisfies CSSProperties;
const cellContentStyle = {
  flex: "1 1 auto",
  minInlineSize: 0,
  maxBlockSize: "100%",
  overflow: "hidden",
  textOverflow: "ellipsis",
} satisfies CSSProperties;
