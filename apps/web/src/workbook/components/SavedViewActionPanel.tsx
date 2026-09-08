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
  workbookPreferenceTestId,
} from "@cartulary/ui-contracts";
import { MoreHorizontal } from "lucide-react";
import { type RefObject, useRef, useSyncExternalStore } from "react";
import { useRegisteredOverlayNavigation } from "../../shared/useRegisteredOverlayNavigation";
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
import type { WorkbookPreferenceController } from "../preferences/WorkbookPreferenceController";
import {
  preferenceOutcome,
  useWorkbookPreferencesSnapshot,
} from "../preferences/WorkbookPreferencesPanel";
import type { WorkbookSavedViewController } from "../savedviews/WorkbookSavedViewController";
import {
  SavedViewRecovery,
  type SavedViewRecoveryKey,
  savedViewRecoveryKeys,
} from "./SavedViewRecovery";

type SavedViewActionControlKey =
  | SavedViewRecoveryKey
  | "create"
  | "delete"
  | "duplicate"
  | "name"
  | "reset"
  | "scope"
  | "set_default"
  | "preferences"
  | "set_home"
  | "update";

const savedViewActionControlKeys: readonly SavedViewActionControlKey[] = [
  "name",
  "scope",
  "create",
  "update",
  "reset",
  "duplicate",
  "set_home",
  "set_default",
  "preferences",
  "delete",
  ...savedViewRecoveryKeys,
];

export function SavedViewActionPanel({
  controller,
  activeViewSchemaId,
  control,
  currentIncidentRole,
  currentUserId,
  dispatch,
  fallbackFocusRef,
  isModified,
  resourceKind,
  runAction,
  preferenceController,
  onInspectPreferences,
  selectedSavedView,
}: {
  readonly controller: WorkbookSavedViewController;
  readonly activeViewSchemaId: string;
  readonly control: SavedViewSurfaceControlState;
  readonly currentIncidentRole: string | null;
  readonly currentUserId: string | null;
  readonly dispatch: (event: SavedViewControlEvent) => void;
  readonly fallbackFocusRef: RefObject<HTMLElement | null>;
  readonly isModified: boolean;
  readonly resourceKind: WorkbookSavedViewsResource["kind"];
  readonly runAction: (intent: SavedViewActionIntent) => void;
  readonly preferenceController?: WorkbookPreferenceController | undefined;
  readonly onInspectPreferences?:
    | ((target?: HTMLElement | null) => void)
    | undefined;
  readonly selectedSavedView: SavedViewResource | null;
}) {
  const preferences = useWorkbookPreferencesSnapshot(preferenceController);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const panelId = savedViewActionMenuTestId(activeViewSchemaId);
  const titleId = `${panelId}-title`;
  const selectedSavedViewMutable = canMutateSavedView(
    selectedSavedView,
    currentUserId,
    currentIncidentRole,
  );
  const operationState = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const subject = {
    viewSchemaId: activeViewSchemaId,
    savedViewId: selectedSavedView?.saved_view_id ?? null,
    savedViewVersion: selectedSavedView?.saved_view_version ?? null,
  };
  const unavailable = (kind: SavedViewActionIntent["kind"]) =>
    controller.unavailableReason(kind, subject) !== null;
  const actionPending = control.busy;
  const operation = operationState.operation;
  const validation =
    operation.kind === "rejected" &&
    operation.problem.field === "display_name" &&
    operation.attempt.kind !== "duplicate" &&
    operation.attempt.subject.viewSchemaId === subject.viewSchemaId &&
    operation.attempt.subject.savedViewId === subject.savedViewId &&
    operation.attempt.definition.displayName === control.displayName
      ? operation.problem.message
      : null;
  const validationId = `${panelId}-name-error`;
  const resourceReady = resourceKind === "ready";

  const controls = useRef(new Map<SavedViewActionControlKey, HTMLElement>());
  const navigation = useRegisteredOverlayNavigation({
    fallbackFocusRef,
    initialItemKey: "name",
    isOpen: control.panelOpen,
    itemKeys: savedViewActionControlKeys,
    onRequestClose: () => {
      dispatch({ type: "close_panel", surface: activeViewSchemaId });
    },
    subjectKey: `${activeViewSchemaId}:${control.selectionKey}`,
    trapTab: true,
    restoreFocusOnSubjectChange: false,
    triggerRef,
  });

  const registerControl =
    (key: SavedViewActionControlKey) => (element: HTMLElement | null) => {
      navigation.registerItem(key)(element);
      if (element) controls.current.set(key, element);
      else controls.current.delete(key);
    };
  return (
    <div style={actionPanelFrameStyle}>
      <button
        ref={triggerRef}
        aria-controls={control.panelOpen ? panelId : undefined}
        aria-expanded={control.panelOpen}
        aria-haspopup="dialog"
        aria-label="Saved view actions"
        data-testid={savedViewActionMenuTriggerTestId(activeViewSchemaId)}
        disabled={resourceKind === "loading" && operation.kind === "idle"}
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
          aria-busy={actionPending || undefined}
          aria-label="Saved view"
          data-testid={panelId}
          id={panelId}
          role="dialog"
          style={actionPanelStyle}
          tabIndex={-1}
          onBlur={navigation.onOverlayBlur}
          onFocusCapture={(event) => {
            const key = savedViewActionControlKeys.find(
              (candidate) => controls.current.get(candidate) === event.target,
            );
            if (key) navigation.onItemFocus(key);
          }}
          onKeyDown={(event) => {
            if (navigation.activeKey === null) return;
            if (
              (event.target instanceof HTMLInputElement ||
                event.target instanceof HTMLSelectElement) &&
              event.key !== "Escape" &&
              event.key !== "Tab"
            )
              return;
            navigation.onItemKeyDown(event, navigation.activeKey);
          }}
        >
          <strong id={titleId} style={panelTitleStyle}>
            {selectedSavedView === null
              ? "Unsaved view configuration"
              : `Saved view: ${selectedSavedView.display_name}`}
          </strong>
          <section aria-label="Saved view configuration" style={sectionStyle}>
            <label style={panelLabelStyle}>
              Name
              <input
                ref={registerControl("name")}
                aria-label="Saved view name"
                aria-invalid={validation ? true : undefined}
                aria-describedby={validation ? validationId : undefined}
                data-testid={savedViewNameInputTestId(activeViewSchemaId)}
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
            {validation ? (
              <p id={validationId} role="alert" style={panelLabelStyle}>
                {validation}
              </p>
            ) : null}
            <label style={panelLabelStyle}>
              Scope
              <select
                ref={registerControl("scope")}
                aria-label="Saved view scope"
                data-testid={savedViewScopeSelectTestId(activeViewSchemaId)}
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
          </section>
          <section aria-label="Save as new view" style={sectionStyle}>
            <strong style={sectionTitleStyle}>Create</strong>
            <button
              ref={registerControl("create")}
              data-testid={savedViewCreateButtonTestId(activeViewSchemaId)}
              disabled={!resourceReady || unavailable("create")}
              style={
                selectedSavedViewMutable ? panelActionStyle : primaryActionStyle
              }
              type="button"
              onClick={() => {
                runAction({
                  kind: "create",
                });
              }}
            >
              Save current configuration as new view
            </button>
          </section>
          {selectedSavedView === null ? (
            <button
              ref={registerControl("reset")}
              type="button"
              style={panelActionStyle}
              disabled={unavailable("reset")}
              onClick={() => runAction({ kind: "reset" })}
            >
              Reset to default configuration
            </button>
          ) : (
            <>
              <section aria-label="Selected view actions" style={sectionStyle}>
                <strong style={sectionTitleStyle}>Selected view</strong>
                <button
                  ref={registerControl("update")}
                  data-testid={savedViewUpdateButtonTestId(
                    activeViewSchemaId,
                    selectedSavedView.saved_view_id,
                  )}
                  disabled={unavailable("update") || !selectedSavedViewMutable}
                  style={
                    selectedSavedViewMutable
                      ? primaryActionStyle
                      : panelActionStyle
                  }
                  type="button"
                  onClick={() => {
                    runAction({
                      kind: "update",
                    });
                  }}
                >
                  Update selected view
                </button>
                <button
                  ref={registerControl("reset")}
                  data-testid={savedViewResetButtonTestId(
                    activeViewSchemaId,
                    selectedSavedView.saved_view_id,
                  )}
                  disabled={unavailable("reset") || !isModified}
                  style={panelActionStyle}
                  type="button"
                  onClick={() => runAction({ kind: "reset" })}
                >
                  Reset to saved configuration
                </button>
              </section>
              <section aria-label="Duplicate view" style={sectionStyle}>
                <strong style={sectionTitleStyle}>Duplicate</strong>
                <button
                  ref={registerControl("duplicate")}
                  data-testid={savedViewDuplicateButtonTestId(
                    activeViewSchemaId,
                    selectedSavedView.saved_view_id,
                  )}
                  style={panelActionStyle}
                  type="button"
                  onClick={() => runAction({ kind: "duplicate" })}
                >
                  Duplicate selected view
                </button>
              </section>
            </>
          )}
          <SavedViewRecovery
            controller={controller}
            snapshot={operationState}
            registerItem={registerControl}
          />
          <section aria-label="Startup view references" style={sectionStyle}>
            <strong style={sectionTitleStyle}>Startup</strong>
            <button
              ref={registerControl("set_home")}
              data-testid={savedViewSetHomeButtonTestId(activeViewSchemaId)}
              aria-disabled={!preferenceController?.canSetCurrent("home")}
              style={panelActionStyle}
              type="button"
              onClick={() => {
                if (preferenceController?.canSetCurrent("home")) {
                  preferenceController.setCurrent("home");
                }
              }}
            >
              Set as my home
            </button>
            <button
              ref={registerControl("set_default")}
              data-testid={savedViewSetDefaultButtonTestId(activeViewSchemaId)}
              aria-disabled={!preferenceController?.canSetCurrent("default")}
              style={panelActionStyle}
              type="button"
              onClick={() => {
                if (preferenceController?.canSetCurrent("default")) {
                  preferenceController.setCurrent("default");
                }
              }}
            >
              Set as incident default
            </button>
            {preferences ? (
              <>
                <p
                  data-testid={workbookPreferenceTestId(
                    "home",
                    "shortcut-outcome",
                  )}
                  style={panelLabelStyle}
                >
                  {preferenceOutcome("home", preferences.home)}
                </p>
                <p
                  data-testid={workbookPreferenceTestId(
                    "default",
                    "shortcut-outcome",
                  )}
                  style={panelLabelStyle}
                >
                  {preferenceOutcome("default", preferences.default)}
                </p>
              </>
            ) : (
              <p style={panelLabelStyle}>
                Workbook preference controls are unavailable.
              </p>
            )}
            {onInspectPreferences ? (
              <button
                ref={registerControl("preferences")}
                type="button"
                style={panelActionStyle}
                onClick={() => {
                  dispatch({
                    type: "close_panel",
                    surface: activeViewSchemaId,
                  });
                  onInspectPreferences(triggerRef.current);
                }}
              >
                Inspect and recover workbook preferences…
              </button>
            ) : null}
          </section>
          {selectedSavedView === null ? null : (
            <section aria-label="Delete saved view" style={dangerSectionStyle}>
              <strong style={sectionTitleStyle}>Delete</strong>
              <button
                ref={registerControl("delete")}
                data-testid={savedViewDeleteButtonTestId(
                  activeViewSchemaId,
                  selectedSavedView.saved_view_id,
                )}
                disabled={unavailable("delete") || !selectedSavedViewMutable}
                style={dangerPanelActionStyle}
                type="button"
                onClick={() => runAction({ kind: "delete" })}
              >
                Delete selected view
              </button>
            </section>
          )}
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
  padding: "var(--ct-component-text-input-padding)",
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
  insetBlockStart: "calc(100% + var(--ct-spacing-xs))",
  insetInlineStart: 0,
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  inlineSize: "min(var(--ct-layout-viewBarOverlayMaxInlineSize), 92vw)",
  maxBlockSize: "70dvh",
  overflowY: "auto" as const,
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
  padding: "var(--ct-spacing-sm)",
};
const panelTitleStyle = {
  color: "var(--ct-colors-ink)",
  fontSize: "0.9rem",
  overflowWrap: "anywhere" as const,
};
const sectionStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  borderBlockStart: "var(--ct-border-hairline)",
  paddingBlockStart: "var(--ct-spacing-sm)",
};
const dangerSectionStyle = {
  ...sectionStyle,
  borderColor: "var(--ct-colors-semantic-conflict)",
};
const sectionTitleStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.78rem",
};
const panelLabelStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
};
const panelActionStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-xs)",
  background: "transparent",
  color: "var(--ct-colors-ink)",
  cursor: "pointer",
  font: "inherit",
  padding: "var(--ct-component-button-secondary-padding)",
  textAlign: "left" as const,
};
const primaryActionStyle = {
  ...panelActionStyle,
  borderColor: "var(--ct-colors-accent-active)",
  background: "var(--ct-colors-accent)",
  color: "var(--ct-colors-on-accent)",
  fontWeight: 700,
};
const dangerPanelActionStyle = {
  ...panelActionStyle,
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 700,
};
