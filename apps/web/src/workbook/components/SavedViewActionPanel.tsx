import {
  savedViewActionMenuTestId,
  savedViewActionMenuTriggerTestId,
  savedViewCreateButtonTestId,
  savedViewDeleteButtonTestId,
  savedViewDuplicateButtonTestId,
  savedViewNameInputTestId,
  savedViewResetButtonTestId,
  savedViewScopeSelectTestId,
  savedViewSetDefaultButtonTestId,
  savedViewSetHomeButtonTestId,
  savedViewUpdateButtonTestId,
} from "@cartulary/ui-contracts";
import { MoreHorizontal } from "lucide-react";
import { type RefObject, useRef } from "react";
import { useRegisteredOverlayNavigation } from "../focus/useRegisteredOverlayNavigation";
import type { SavedViewActionIntent } from "../hooks/useActiveSurfaceSavedViewActions";
import {
  parseSavedViewEditableScope,
  type SavedViewControlEvent,
  type SavedViewSurfaceControlState,
  type WorkbookSavedViewsResource,
} from "../models/workbookSavedViewControl";
import {
  canMutateSavedView,
  type SavedViewResource,
} from "../models/workbookSavedViews";

type SavedViewActionControlKey =
  | "name"
  | "scope"
  | "create"
  | "update"
  | "duplicate"
  | "reset"
  | "delete"
  | "set_home"
  | "set_default";

const savedViewActionControlKeys: readonly SavedViewActionControlKey[] = [
  "name",
  "scope",
  "create",
  "update",
  "duplicate",
  "reset",
  "delete",
  "set_home",
  "set_default",
];

export function SavedViewActionPanel({
  activeViewSchemaId,
  control,
  currentIncidentRole,
  currentUserId,
  dispatch,
  fallbackFocusRef,
  isModified,
  resourceKind,
  runAction,
  selectedSavedView,
}: {
  readonly activeViewSchemaId: string;
  readonly control: SavedViewSurfaceControlState;
  readonly currentIncidentRole: string | null;
  readonly currentUserId: string | null;
  readonly dispatch: (event: SavedViewControlEvent) => void;
  readonly fallbackFocusRef: RefObject<HTMLElement | null>;
  readonly isModified: boolean;
  readonly resourceKind: WorkbookSavedViewsResource["kind"];
  readonly runAction: (intent: SavedViewActionIntent) => void;
  readonly selectedSavedView: SavedViewResource | null;
}) {
  const triggerRef = useRef<HTMLButtonElement>(null);
  const panelId = savedViewActionMenuTestId(activeViewSchemaId);
  const titleId = `${panelId}-title`;
  const selectedSavedViewMutable = canMutateSavedView(
    selectedSavedView,
    currentUserId,
    currentIncidentRole,
  );
  const actionPending = control.activeAction !== null;
  const resourceReady = resourceKind === "ready";
  const trimmedDisplayName = control.displayName.trim();
  const navigation = useRegisteredOverlayNavigation({
    fallbackFocusRef,
    initialItemKey: "name",
    isOpen: control.panelOpen,
    itemKeys: savedViewActionControlKeys,
    onRequestClose: () => {
      dispatch({ type: "close_panel", surface: activeViewSchemaId });
    },
    subjectKey: `${activeViewSchemaId}:${control.selectionKey}`,
    triggerRef,
  });

  return (
    <div style={actionPanelFrameStyle}>
      <button
        ref={triggerRef}
        aria-controls={control.panelOpen ? panelId : undefined}
        aria-expanded={control.panelOpen}
        aria-haspopup="dialog"
        aria-label="Saved view actions"
        data-testid={savedViewActionMenuTriggerTestId(activeViewSchemaId)}
        disabled={resourceKind === "loading" || resourceKind === "unavailable"}
        style={iconButtonStyle}
        type="button"
        onClick={() => {
          if (!control.panelOpen) navigation.prepareOpen("name");
          dispatch({ type: "toggle_panel", surface: activeViewSchemaId });
        }}
      >
        <MoreHorizontal aria-hidden="true" size={16} />
      </button>
      {control.panelOpen ? (
        <div
          aria-labelledby={titleId}
          data-testid={panelId}
          id={panelId}
          role="dialog"
          style={actionPanelStyle}
          tabIndex={-1}
          onBlur={navigation.onOverlayBlur}
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              navigation.onItemKeyDown(event, navigation.activeKey ?? "name");
            }
          }}
        >
          <strong id={titleId} style={panelTitleStyle}>
            Saved view
          </strong>
          <label style={panelLabelStyle}>
            Name
            <input
              ref={navigation.registerItem("name")}
              aria-label="Saved view name"
              data-testid={savedViewNameInputTestId(activeViewSchemaId)}
              disabled={actionPending}
              style={inputStyle}
              type="text"
              value={control.displayName}
              onChange={(event) => {
                dispatch({
                  type: "change_name",
                  surface: activeViewSchemaId,
                  displayName: event.currentTarget.value,
                });
              }}
            />
          </label>
          <label style={panelLabelStyle}>
            Scope
            <select
              ref={navigation.registerItem("scope")}
              aria-label="Saved view scope"
              data-testid={savedViewScopeSelectTestId(activeViewSchemaId)}
              disabled={actionPending}
              style={inputStyle}
              value={control.scope}
              onChange={(event) => {
                const scope = parseSavedViewEditableScope(
                  event.currentTarget.value,
                );
                if (scope === null) return;
                dispatch({
                  type: "change_scope",
                  surface: activeViewSchemaId,
                  scope,
                });
              }}
            >
              <option value="private">Private</option>
              <option value="shared">Shared</option>
            </select>
          </label>
          <button
            ref={navigation.registerItem("create")}
            data-testid={savedViewCreateButtonTestId(activeViewSchemaId)}
            disabled={
              !resourceReady || actionPending || trimmedDisplayName === ""
            }
            style={panelActionStyle}
            type="button"
            onClick={() => {
              runAction({
                kind: "create",
                displayName: trimmedDisplayName,
                scope: control.scope,
              });
            }}
          >
            Save as new view
          </button>
          {selectedSavedView === null ? null : (
            <>
              <button
                ref={navigation.registerItem("update")}
                data-testid={savedViewUpdateButtonTestId(
                  activeViewSchemaId,
                  selectedSavedView.saved_view_id,
                )}
                disabled={
                  actionPending ||
                  !selectedSavedViewMutable ||
                  trimmedDisplayName === ""
                }
                style={panelActionStyle}
                title={
                  selectedSavedViewMutable
                    ? undefined
                    : "Only the owner or incident administrator can update this view."
                }
                type="button"
                onClick={() => {
                  runAction({
                    kind: "update",
                    displayName: trimmedDisplayName,
                    scope: control.scope,
                  });
                }}
              >
                Update view
              </button>
              <button
                ref={navigation.registerItem("duplicate")}
                data-testid={savedViewDuplicateButtonTestId(
                  activeViewSchemaId,
                  selectedSavedView.saved_view_id,
                )}
                disabled={actionPending}
                style={panelActionStyle}
                type="button"
                onClick={() => {
                  runAction({ kind: "duplicate" });
                }}
              >
                Duplicate
              </button>
              <button
                ref={navigation.registerItem("reset")}
                data-testid={savedViewResetButtonTestId(
                  activeViewSchemaId,
                  selectedSavedView.saved_view_id,
                )}
                disabled={actionPending || !isModified}
                style={panelActionStyle}
                type="button"
                onClick={() => {
                  runAction({ kind: "reset" });
                }}
              >
                Reset to saved configuration
              </button>
              <button
                ref={navigation.registerItem("delete")}
                data-testid={savedViewDeleteButtonTestId(
                  activeViewSchemaId,
                  selectedSavedView.saved_view_id,
                )}
                disabled={actionPending || !selectedSavedViewMutable}
                style={dangerPanelActionStyle}
                type="button"
                onClick={() => {
                  runAction({ kind: "delete" });
                }}
              >
                Delete
              </button>
            </>
          )}
          <button
            ref={navigation.registerItem("set_home")}
            data-testid={savedViewSetHomeButtonTestId(activeViewSchemaId)}
            disabled={actionPending}
            style={panelActionStyle}
            type="button"
            onClick={() => {
              runAction({ kind: "set_home" });
            }}
          >
            Set as my home
          </button>
          <button
            ref={navigation.registerItem("set_default")}
            data-testid={savedViewSetDefaultButtonTestId(activeViewSchemaId)}
            disabled={actionPending || currentIncidentRole !== "admin"}
            style={panelActionStyle}
            title={
              currentIncidentRole === "admin"
                ? undefined
                : "Only incident administrators can set the incident default."
            }
            type="button"
            onClick={() => {
              runAction({ kind: "set_default" });
            }}
          >
            Set as incident default
          </button>
        </div>
      ) : null}
    </div>
  );
}

const inputStyle = {
  boxSizing: "border-box" as const,
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.42rem 0.55rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
};

const actionPanelFrameStyle = {
  position: "relative" as const,
  display: "inline-flex",
  flex: "0 0 auto",
};

const iconButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  inlineSize: "1.9rem",
  blockSize: "1.9rem",
  borderRadius: "var(--ct-rounded-xs)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  color: "var(--ct-colors-ink)",
  cursor: "pointer",
};

const actionPanelStyle = {
  position: "absolute" as const,
  zIndex: 22,
  insetBlockStart: "calc(100% + 0.35rem)",
  insetInlineStart: 0,
  display: "grid",
  gap: "0.35rem",
  inlineSize: "min(22rem, 86vw)",
  maxBlockSize: "28rem",
  overflowY: "auto" as const,
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
  padding: "0.55rem",
};

const panelTitleStyle = {
  color: "var(--ct-colors-ink)",
  fontSize: "0.9rem",
};

const panelLabelStyle = {
  display: "grid",
  gap: "0.25rem",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
};

const panelActionStyle = {
  border: 0,
  borderRadius: "var(--ct-rounded-xs)",
  background: "transparent",
  color: "var(--ct-colors-ink)",
  cursor: "pointer",
  font: "inherit",
  padding: "0.45rem 0.5rem",
  textAlign: "left" as const,
};

const dangerPanelActionStyle = {
  ...panelActionStyle,
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 700,
};
