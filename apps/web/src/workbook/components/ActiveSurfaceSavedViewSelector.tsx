import {
  savedViewCreateButtonTestId,
  savedViewDeleteButtonTestId,
  savedViewDuplicateButtonTestId,
  savedViewNameInputTestId,
  savedViewOptionTestId,
  savedViewScopeSelectTestId,
  savedViewSelectorTestId,
  savedViewSetDefaultButtonTestId,
  savedViewSetHomeButtonTestId,
  savedViewStatusTestId,
  savedViewUpdateButtonTestId,
} from "@cartulary/ui-contracts";
import { type ReactNode, useEffect, useMemo, useState } from "react";
import {
  canMutateSavedView,
  type SavedViewResource,
} from "../models/workbookSavedViews";
import type { WorkbookSheetRef } from "../models/workbookStartup";

export function ActiveSurfaceSavedViewSelector({
  activeViewSchemaId,
  afterActions,
  currentIncidentRole,
  currentUserId,
  savedViews,
  selectedSheetRef,
  onCreateSavedView,
  onDeleteSavedView,
  onDuplicateSavedView,
  onSelectBaseSurface,
  onSelectSavedView,
  onSetDefaultSheetRef,
  onSetHomeSheetRef,
  onUpdateSavedView,
}: {
  readonly activeViewSchemaId: string;
  readonly afterActions?: ReactNode | undefined;
  readonly currentIncidentRole: string | null;
  readonly currentUserId: string | null;
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
  const activeSavedViews = useMemo(
    () =>
      savedViews.filter(
        (savedView) => savedView.view_schema_id === activeViewSchemaId,
      ),
    [activeViewSchemaId, savedViews],
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
    action: () => Promise<void>,
    successMessage: string,
  ) => {
    setStatus("");
    try {
      await action();
      setStatus(successMessage);
    } catch (error) {
      setStatus(error instanceof Error ? error.message : "Saved view failed.");
    }
  };

  return (
    <div style={savedViewControlGroupStyle}>
      <div style={savedViewInputGroupStyle}>
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
            <option value="">Base view</option>
            {activeSavedViews.map((savedView) => (
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
          </select>
        </label>
        <input
          aria-label="Saved view name"
          data-testid={savedViewNameInputTestId(activeViewSchemaId)}
          style={savedViewNameInputStyle}
          type="text"
          value={displayName}
          onChange={(event) => {
            setDisplayName(event.currentTarget.value);
          }}
        />
        <select
          aria-label="Saved view scope"
          data-testid={savedViewScopeSelectTestId(activeViewSchemaId)}
          style={savedViewScopeSelectStyle}
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
      </div>
      <div style={savedViewActionGroupStyle}>
        <button
          data-testid={savedViewCreateButtonTestId(activeViewSchemaId)}
          disabled={trimmedDisplayName === ""}
          style={secondaryActionButtonStyle}
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
          Create
        </button>
        {selectedSavedView ? (
          <>
            <button
              data-testid={savedViewDuplicateButtonTestId(
                activeViewSchemaId,
                selectedSavedView.saved_view_id,
              )}
              style={secondaryActionButtonStyle}
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
              data-testid={savedViewUpdateButtonTestId(
                activeViewSchemaId,
                selectedSavedView.saved_view_id,
              )}
              disabled={!selectedSavedViewMutable || trimmedDisplayName === ""}
              style={secondaryActionButtonStyle}
              type="button"
              onClick={() => {
                void runSavedViewAction(async () => {
                  await onUpdateSavedView(selectedSavedView, {
                    displayName: trimmedDisplayName,
                    scope,
                  });
                }, "Saved view updated.");
              }}
            >
              Update
            </button>
            <button
              data-testid={savedViewDeleteButtonTestId(
                activeViewSchemaId,
                selectedSavedView.saved_view_id,
              )}
              disabled={!selectedSavedViewMutable}
              style={secondaryActionButtonStyle}
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
          style={secondaryActionButtonStyle}
          type="button"
          onClick={() => {
            void runSavedViewAction(onSetHomeSheetRef, "Home view updated.");
          }}
        >
          Home
        </button>
        <button
          data-testid={savedViewSetDefaultButtonTestId(activeViewSchemaId)}
          disabled={currentIncidentRole !== "admin"}
          style={secondaryActionButtonStyle}
          type="button"
          onClick={() => {
            void runSavedViewAction(
              onSetDefaultSheetRef,
              "Default view updated.",
            );
          }}
        >
          Default
        </button>
      </div>
      {afterActions === undefined || afterActions === null ? null : (
        <div style={savedViewAfterActionsStyle}>{afterActions}</div>
      )}
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

const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};

const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
};

const inputStyle = {
  boxSizing: "border-box" as const,
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
};

const selectStyle = {
  ...inputStyle,
  appearance: "auto" as const,
};

const savedViewSelectorFrameStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.4rem",
  flex: "0 0 auto",
  minWidth: 0,
};

const savedViewControlGroupStyle = {
  display: "flex",
  alignItems: "center",
  flexWrap: "wrap" as const,
  gap: "0.4rem",
  flex: "1 1 auto",
  minWidth: 0,
};

const savedViewInputGroupStyle = {
  display: "inline-flex",
  alignItems: "center",
  flexWrap: "nowrap" as const,
  gap: "0.4rem",
  minWidth: 0,
  whiteSpace: "nowrap" as const,
};

const savedViewActionGroupStyle = {
  display: "inline-flex",
  alignItems: "center",
  flexWrap: "nowrap" as const,
  gap: "0.4rem",
  whiteSpace: "nowrap" as const,
  minWidth: 0,
};

const savedViewAfterActionsStyle = {
  display: "inline-flex",
  alignItems: "center",
  flex: "1 1 34rem",
  minWidth: 0,
};

const savedViewSelectorLabelStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
  whiteSpace: "nowrap" as const,
};

const savedViewSelectStyle = {
  ...selectStyle,
  borderRadius: "var(--ct-rounded-sm)",
  padding: "0.42rem 0.55rem",
  maxWidth: "10rem",
};

const savedViewNameInputStyle = {
  ...inputStyle,
  width: "7.5rem",
  minWidth: "6rem",
  padding: "0.42rem 0.55rem",
};

const savedViewScopeSelectStyle = {
  ...savedViewSelectStyle,
  maxWidth: "7rem",
};

const savedViewStatusStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.78rem",
  maxWidth: "12rem",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};
