import {
  savedViewActionMenuTestId,
  savedViewActionMenuTriggerTestId,
  savedViewCreateButtonTestId,
  savedViewDeleteButtonTestId,
  savedViewDuplicateButtonTestId,
  savedViewManageSharingButtonTestId,
  savedViewModifiedTestId,
  savedViewNameInputTestId,
  savedViewOptionTestId,
  savedViewRenameButtonTestId,
  savedViewResetButtonTestId,
  savedViewScopeSelectTestId,
  savedViewSelectorTestId,
  savedViewSetDefaultButtonTestId,
  savedViewSetHomeButtonTestId,
  savedViewStatusTestId,
  savedViewUpdateButtonTestId,
} from "@cartulary/ui-contracts";
import { MoreHorizontal } from "lucide-react";
import { useEffect, useMemo, useState } from "react";
import {
  canMutateSavedView,
  type SavedViewResource,
} from "../models/workbookSavedViews";
import type { WorkbookSheetRef } from "../models/workbookStartup";

export function ActiveSurfaceSavedViewSelector({
  activeViewSchemaId,
  currentIncidentRole,
  currentUserId,
  isModified = false,
  savedViews,
  selectedSheetRef,
  onCreateSavedView,
  onDeleteSavedView,
  onDuplicateSavedView,
  onResetToSavedView,
  onSelectBaseSurface,
  onSelectSavedView,
  onSetDefaultSheetRef,
  onSetHomeSheetRef,
  onUpdateSavedView,
}: {
  readonly activeViewSchemaId: string;
  readonly currentIncidentRole: string | null;
  readonly currentUserId: string | null;
  readonly isModified?: boolean | undefined;
  readonly savedViews: readonly SavedViewResource[];
  readonly selectedSheetRef: WorkbookSheetRef;
  readonly onCreateSavedView: (input: {
    readonly displayName: string;
    readonly scope: "private" | "shared";
  }) => Promise<SavedViewResource>;
  readonly onDeleteSavedView: (savedView: SavedViewResource) => Promise<void>;
  readonly onDuplicateSavedView: (
    savedView: SavedViewResource,
  ) => Promise<SavedViewResource>;
  readonly onResetToSavedView: (savedView: SavedViewResource) => void;
  readonly onSelectBaseSurface: (viewSchemaId: string) => void;
  readonly onSelectSavedView: (savedView: SavedViewResource) => void;
  readonly onSetDefaultSheetRef: () => Promise<void>;
  readonly onSetHomeSheetRef: () => Promise<void>;
  readonly onUpdateSavedView: (
    savedView: SavedViewResource,
    input: {
      readonly displayName: string;
      readonly scope: "private" | "shared";
    },
  ) => Promise<SavedViewResource>;
}) {
  const [displayName, setDisplayName] = useState("Saved view");
  const [scope, setScope] = useState<"private" | "shared">("private");
  const [status, setStatus] = useState("");
  const [isActionMenuOpen, setIsActionMenuOpen] = useState(false);
  const activeSavedViews = useMemo(
    () =>
      savedViews.filter(
        (savedView) => savedView.view_schema_id === activeViewSchemaId,
      ),
    [activeViewSchemaId, savedViews],
  );
  const groupedSavedViews = useMemo(
    () => ({
      private: activeSavedViews.filter((savedView) => savedView.scope === "private"),
      shared: activeSavedViews.filter((savedView) => savedView.scope === "shared"),
      system: activeSavedViews.filter((savedView) => savedView.scope === "system"),
    }),
    [activeSavedViews],
  );
  const selectedSavedViewId =
    selectedSheetRef.kind === "saved_view" &&
    activeSavedViews.some(
      (savedView) => savedView.saved_view_id === selectedSheetRef.id,
    )
      ? selectedSheetRef.id
      : "";
  const selectedSavedView =
    activeSavedViews.find(
      (savedView) => savedView.saved_view_id === selectedSavedViewId,
    ) ?? null;
  const selectedSavedViewMutable = canMutateSavedView(
    selectedSavedView,
    currentUserId,
    currentIncidentRole,
  );
  const trimmedDisplayName = displayName.trim();
  const canSetDefault = currentIncidentRole === "admin";

  useEffect(() => {
    if (selectedSavedView === null) {
      setDisplayName("Saved view");
      setScope("private");
      return;
    }
    setDisplayName(selectedSavedView.display_name);
    setScope(selectedSavedView.scope === "shared" ? "shared" : "private");
  }, [selectedSavedView]);

  const runSavedViewAction = async (
    action: () => Promise<void> | void,
    successMessage: string,
  ) => {
    setStatus("");
    try {
      await action();
      setStatus(successMessage);
      setIsActionMenuOpen(false);
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Saved view failed.");
    }
  };

  return (
    <div style={savedViewControlGroupStyle}>
      <label style={savedViewSelectorFrameStyle}>
        <span style={savedViewSelectorLabelStyle}>View:</span>
        <select
          aria-label="Saved view"
          data-active-view-schema-id={activeViewSchemaId}
          data-selected-saved-view-id={selectedSavedViewId}
          data-selected-sheet-ref-kind={
            selectedSavedViewId === "" ? "view_schema" : "saved_view"
          }
          data-testid={savedViewSelectorTestId(activeViewSchemaId)}
          style={savedViewSelectStyle}
          value={selectedSavedViewId}
          onChange={(event) => {
            const nextSavedViewId = event.currentTarget.value;
            setStatus("");
            if (nextSavedViewId === "") {
              onSelectBaseSurface(activeViewSchemaId);
              return;
            }
            const savedView = activeSavedViews.find(
              (candidate) => candidate.saved_view_id === nextSavedViewId,
            );
            if (savedView !== undefined) {
              onSelectSavedView(savedView);
            }
          }}
        >
          <option value="">Unsaved view</option>
          <SavedViewOptionGroup
            activeViewSchemaId={activeViewSchemaId}
            label="Private"
            savedViews={groupedSavedViews.private}
          />
          <SavedViewOptionGroup
            activeViewSchemaId={activeViewSchemaId}
            label="Shared"
            savedViews={groupedSavedViews.shared}
          />
          <SavedViewOptionGroup
            activeViewSchemaId={activeViewSchemaId}
            label="System"
            savedViews={groupedSavedViews.system}
          />
        </select>
      </label>
      {selectedSavedView !== null && isModified ? (
        <span
          aria-label="Saved view modified"
          data-testid={savedViewModifiedTestId(activeViewSchemaId)}
          style={modifiedBadgeStyle}
        >
          Modified
        </span>
      ) : null}
      <div style={actionMenuFrameStyle}>
        <button
          aria-controls={
            isActionMenuOpen ? savedViewActionMenuTestId(activeViewSchemaId) : undefined
          }
          aria-expanded={isActionMenuOpen}
          aria-haspopup="menu"
          aria-label="Saved view actions"
          data-testid={savedViewActionMenuTriggerTestId(activeViewSchemaId)}
          style={iconButtonStyle}
          type="button"
          onClick={() => {
            setIsActionMenuOpen((current) => !current);
          }}
        >
          <MoreHorizontal aria-hidden="true" size={16} />
        </button>
        {isActionMenuOpen ? (
          <div
            data-testid={savedViewActionMenuTestId(activeViewSchemaId)}
            id={savedViewActionMenuTestId(activeViewSchemaId)}
            role="menu"
            style={actionMenuStyle}
          >
            <label style={menuLabelStyle}>
              Name
              <input
                aria-label="Saved view name"
                data-testid={savedViewNameInputTestId(activeViewSchemaId)}
                style={inputStyle}
                type="text"
                value={displayName}
                onChange={(event) => {
                  setDisplayName(event.currentTarget.value);
                }}
              />
            </label>
            <label style={menuLabelStyle}>
              Scope
              <select
                aria-label="Saved view scope"
                data-testid={savedViewScopeSelectTestId(activeViewSchemaId)}
                style={inputStyle}
                value={scope}
                onChange={(event) => {
                  setScope(
                    event.currentTarget.value === "shared" ? "shared" : "private",
                  );
                }}
              >
                <option value="private">Private</option>
                <option value="shared">Shared</option>
              </select>
            </label>
            <button
              data-testid={savedViewCreateButtonTestId(activeViewSchemaId)}
              disabled={trimmedDisplayName === ""}
              role="menuitem"
              style={menuActionStyle}
              type="button"
              onClick={() => {
                void runSavedViewAction(async () => {
                  await onCreateSavedView({
                    displayName: trimmedDisplayName,
                    scope,
                  });
                }, "Saved view created.");
              }}
            >
              Save as new view
            </button>
            {selectedSavedView ? (
              <>
                <button
                  data-testid={savedViewUpdateButtonTestId(
                    activeViewSchemaId,
                    selectedSavedView.saved_view_id,
                  )}
                  disabled={!selectedSavedViewMutable}
                  role="menuitem"
                  style={menuActionStyle}
                  title={
                    selectedSavedViewMutable
                      ? undefined
                      : "Only the owner or incident administrator can update this view."
                  }
                  type="button"
                  onClick={() => {
                    void runSavedViewAction(async () => {
                      await onUpdateSavedView(selectedSavedView, {
                        displayName: trimmedDisplayName || selectedSavedView.display_name,
                        scope,
                      });
                    }, "Saved view updated.");
                  }}
                >
                  Update view
                </button>
                <button
                  data-testid={savedViewDuplicateButtonTestId(
                    activeViewSchemaId,
                    selectedSavedView.saved_view_id,
                  )}
                  role="menuitem"
                  style={menuActionStyle}
                  type="button"
                  onClick={() => {
                    void runSavedViewAction(async () => {
                      await onDuplicateSavedView(selectedSavedView);
                    }, "Saved view duplicated.");
                  }}
                >
                  Duplicate
                </button>
                <button
                  data-testid={savedViewRenameButtonTestId(
                    activeViewSchemaId,
                    selectedSavedView.saved_view_id,
                  )}
                  disabled={!selectedSavedViewMutable || trimmedDisplayName === ""}
                  role="menuitem"
                  style={menuActionStyle}
                  type="button"
                  onClick={() => {
                    void runSavedViewAction(async () => {
                      await onUpdateSavedView(selectedSavedView, {
                        displayName: trimmedDisplayName,
                        scope,
                      });
                    }, "Saved view renamed.");
                  }}
                >
                  Rename
                </button>
                <button
                  data-testid={savedViewManageSharingButtonTestId(
                    activeViewSchemaId,
                    selectedSavedView.saved_view_id,
                  )}
                  disabled={!selectedSavedViewMutable}
                  role="menuitem"
                  style={menuActionStyle}
                  type="button"
                  onClick={() => {
                    void runSavedViewAction(async () => {
                      await onUpdateSavedView(selectedSavedView, {
                        displayName:
                          trimmedDisplayName || selectedSavedView.display_name,
                        scope,
                      });
                    }, "Saved view sharing updated.");
                  }}
                >
                  Manage sharing
                </button>
                <button
                  data-testid={savedViewResetButtonTestId(
                    activeViewSchemaId,
                    selectedSavedView.saved_view_id,
                  )}
                  disabled={!isModified}
                  role="menuitem"
                  style={menuActionStyle}
                  type="button"
                  onClick={() => {
                    void runSavedViewAction(() => {
                      onResetToSavedView(selectedSavedView);
                    }, "Saved configuration restored.");
                  }}
                >
                  Reset to saved configuration
                </button>
                <button
                  data-testid={savedViewDeleteButtonTestId(
                    activeViewSchemaId,
                    selectedSavedView.saved_view_id,
                  )}
                  disabled={!selectedSavedViewMutable}
                  role="menuitem"
                  style={dangerMenuActionStyle}
                  type="button"
                  onClick={() => {
                    void runSavedViewAction(async () => {
                      await onDeleteSavedView(selectedSavedView);
                    }, "Saved view deleted.");
                  }}
                >
                  Delete
                </button>
              </>
            ) : null}
            <button
              data-testid={savedViewSetHomeButtonTestId(activeViewSchemaId)}
              role="menuitem"
              style={menuActionStyle}
              type="button"
              onClick={() => {
                void runSavedViewAction(onSetHomeSheetRef, "Home view updated.");
              }}
            >
              Set as my home
            </button>
            <button
              data-testid={savedViewSetDefaultButtonTestId(activeViewSchemaId)}
              disabled={!canSetDefault}
              role="menuitem"
              style={menuActionStyle}
              title={
                canSetDefault
                  ? undefined
                  : "Only incident administrators can set the incident default."
              }
              type="button"
              onClick={() => {
                void runSavedViewAction(
                  onSetDefaultSheetRef,
                  "Default view updated.",
                );
              }}
            >
              Set as incident default
            </button>
          </div>
        ) : null}
      </div>
      <span
        aria-live="polite"
        data-testid={savedViewStatusTestId(activeViewSchemaId)}
        style={savedViewStatusStyle}
      >
        {status}
      </span>
    </div>
  );
}

function SavedViewOptionGroup({
  activeViewSchemaId,
  label,
  savedViews,
}: {
  readonly activeViewSchemaId: string;
  readonly label: string;
  readonly savedViews: readonly SavedViewResource[];
}) {
  if (savedViews.length === 0) {
    return null;
  }
  return (
    <optgroup label={label}>
      {savedViews.map((savedView) => (
        <option
          key={savedView.saved_view_id}
          data-saved-view-id={savedView.saved_view_id}
          data-testid={savedViewOptionTestId(
            activeViewSchemaId,
            savedView.saved_view_id,
          )}
          data-view-schema-id={activeViewSchemaId}
          value={savedView.saved_view_id}
        >
          {savedView.display_name}
        </option>
      ))}
    </optgroup>
  );
}

const savedViewControlGroupStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.4rem",
  flex: "1 1 auto",
  minWidth: 0,
  overflow: "hidden",
};

const savedViewSelectorFrameStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.4rem",
  flex: "0 1 auto",
  minWidth: 0,
};

const savedViewSelectorLabelStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
  whiteSpace: "nowrap" as const,
};

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

const savedViewSelectStyle = {
  ...inputStyle,
  appearance: "auto" as const,
  inlineSize: "min(18rem, 36vw)",
  minInlineSize: "10rem",
};

const modifiedBadgeStyle = {
  borderRadius: "var(--ct-rounded-xs)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink)",
  fontSize: "0.78rem",
  fontWeight: 700,
  padding: "0.25rem 0.4rem",
  whiteSpace: "nowrap" as const,
};

const actionMenuFrameStyle = {
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

const actionMenuStyle = {
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

const menuLabelStyle = {
  display: "grid",
  gap: "0.25rem",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
};

const menuActionStyle = {
  border: 0,
  borderRadius: "var(--ct-rounded-xs)",
  background: "transparent",
  color: "var(--ct-colors-ink)",
  cursor: "pointer",
  font: "inherit",
  padding: "0.45rem 0.5rem",
  textAlign: "left" as const,
};

const dangerMenuActionStyle = {
  ...menuActionStyle,
  color: "var(--ct-colors-semantic-conflict)",
  fontWeight: 700,
};

const savedViewStatusStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.78rem",
  maxWidth: "12rem",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};
