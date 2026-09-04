import {
  savedViewModifiedTestId,
  savedViewOptionTestId,
  savedViewSelectorTestId,
  savedViewStatusTestId,
} from "@cartulary/ui-contracts";
import {
  type Dispatch,
  type RefObject,
  useEffect,
  useMemo,
  useReducer,
  useRef,
} from "react";
import type { SheetRef } from "../../shared/sheetRef";
import {
  type SavedViewActionIntent,
  useActiveSurfaceSavedViewActions,
} from "../hooks/useActiveSurfaceSavedViewActions";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import {
  type ActiveSurfaceSavedViewProjection,
  createSavedViewControlState,
  projectActiveSurfaceSavedViews,
  reduceSavedViewControl,
  type SavedViewControlEvent,
  type SavedViewSurfaceControlState,
  savedViewSurfaceControlState,
  type WorkbookSavedViewsResource,
} from "../models/workbookSavedViewControl";
import type { SavedViewResource } from "../models/workbookSavedViews";
import { visuallyHiddenStyle } from "../utils/workbookStyles";
import { SavedViewActionPanel } from "./SavedViewActionPanel";

export type ActiveSurfaceSavedViewSelectorProps = {
  readonly activeViewSchemaId: string;
  readonly chromeMode: WorkbookChromeMode;
  readonly currentIncidentRole: string | null;
  readonly currentUserId: string | null;
  readonly isModified?: boolean | undefined;
  readonly savedViewsResource: WorkbookSavedViewsResource;
  readonly selectedSheetRef: SheetRef;
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
};

export function ActiveSurfaceSavedViewSelector({
  activeViewSchemaId,
  chromeMode,
  currentIncidentRole,
  currentUserId,
  isModified = false,
  savedViewsResource,
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
}: ActiveSurfaceSavedViewSelectorProps) {
  const projection = useMemo(
    () =>
      projectActiveSurfaceSavedViews(
        savedViewsResource,
        activeViewSchemaId,
        selectedSheetRef,
      ),
    [activeViewSchemaId, savedViewsResource, selectedSheetRef],
  );
  const [state, dispatch] = useReducer(
    reduceSavedViewControl,
    createSavedViewControlState(
      activeViewSchemaId,
      projection.selectedSavedView,
    ),
  );
  const control = savedViewSurfaceControlState(
    state,
    activeViewSchemaId,
    projection.selectedSavedView,
  );
  const selectorRef = useRef<HTMLSelectElement>(null);

  useEffect(() => {
    dispatch({
      type: "activate",
      surface: activeViewSchemaId,
      selectedSavedView: projection.selectedSavedView,
    });
  }, [activeViewSchemaId, projection.selectedSavedView]);

  useInvalidSavedViewFallback({
    activeViewSchemaId,
    dispatch,
    onSelectBaseSurface,
    savedViewsResource,
  });

  const { runAction } = useActiveSurfaceSavedViewActions({
    activeViewSchemaId,
    currentIncidentRole,
    currentUserId,
    dispatch,
    isModified,
    ports: {
      create: onCreateSavedView,
      delete: onDeleteSavedView,
      duplicate: onDuplicateSavedView,
      reset: onResetToSavedView,
      setDefault: onSetDefaultSheetRef,
      setHome: onSetHomeSheetRef,
      update: onUpdateSavedView,
    },
    projection,
  });

  return (
    <SavedViewControlPresentation
      activeViewSchemaId={activeViewSchemaId}
      chromeMode={chromeMode}
      control={control}
      currentIncidentRole={currentIncidentRole}
      currentUserId={currentUserId}
      dispatch={dispatch}
      isModified={isModified}
      onSelectBaseSurface={onSelectBaseSurface}
      onSelectSavedView={onSelectSavedView}
      projection={projection}
      runAction={runAction}
      selectorRef={selectorRef}
    />
  );
}

function useInvalidSavedViewFallback({
  activeViewSchemaId,
  dispatch,
  onSelectBaseSurface,
  savedViewsResource,
}: {
  readonly activeViewSchemaId: string;
  readonly dispatch: Dispatch<SavedViewControlEvent>;
  readonly onSelectBaseSurface: (viewSchemaId: string) => void;
  readonly savedViewsResource: WorkbookSavedViewsResource;
}) {
  const handledSelectionRef = useRef<string | null>(null);
  useEffect(() => {
    if (savedViewsResource.kind !== "invalid_selection") {
      handledSelectionRef.current = null;
      return;
    }
    const invalidKey = `${activeViewSchemaId}:${savedViewsResource.selectedSavedViewId}`;
    if (handledSelectionRef.current === invalidKey) return;
    handledSelectionRef.current = invalidKey;
    dispatch({
      type: "publish_notice",
      surface: activeViewSchemaId,
      message:
        "The selected saved view is no longer available. Showing the base surface.",
    });
    onSelectBaseSurface(activeViewSchemaId);
  }, [activeViewSchemaId, dispatch, onSelectBaseSurface, savedViewsResource]);
}

function SavedViewControlPresentation({
  activeViewSchemaId,
  chromeMode,
  control,
  currentIncidentRole,
  currentUserId,
  dispatch,
  isModified,
  onSelectBaseSurface,
  onSelectSavedView,
  projection,
  runAction,
  selectorRef,
}: {
  readonly activeViewSchemaId: string;
  readonly chromeMode: WorkbookChromeMode;
  readonly control: SavedViewSurfaceControlState;
  readonly currentIncidentRole: string | null;
  readonly currentUserId: string | null;
  readonly dispatch: Dispatch<SavedViewControlEvent>;
  readonly isModified: boolean;
  readonly onSelectBaseSurface: (viewSchemaId: string) => void;
  readonly onSelectSavedView: (savedView: SavedViewResource) => void;
  readonly projection: ActiveSurfaceSavedViewProjection;
  readonly runAction: (intent: SavedViewActionIntent) => void;
  readonly selectorRef: RefObject<HTMLSelectElement | null>;
}) {
  const condensedControls = chromeMode !== "base";
  const compactControls =
    chromeMode === "compact_desktop" ||
    chromeMode === "below_supported_minimum";
  return (
    <div
      style={{
        ...savedViewControlGroupStyle,
        ...(condensedControls ? condensedSavedViewControlGroupStyle : null),
        ...(!condensedControls && projection.selectedSavedView !== null
          ? selectedBaseSavedViewControlGroupStyle
          : null),
        ...(compactControls ? compactSavedViewControlGroupStyle : null),
      }}
    >
      <SavedViewSelectionField
        activeViewSchemaId={activeViewSchemaId}
        condensedControls={condensedControls}
        dispatch={dispatch}
        onSelectBaseSurface={onSelectBaseSurface}
        onSelectSavedView={onSelectSavedView}
        projection={projection}
        selectorRef={selectorRef}
      />
      <SavedViewModifiedBadge
        activeViewSchemaId={activeViewSchemaId}
        condensed={condensedControls}
        isModified={isModified}
        selectedSavedView={projection.selectedSavedView}
      />
      <SavedViewActionPanel
        activeViewSchemaId={activeViewSchemaId}
        control={control}
        currentIncidentRole={currentIncidentRole}
        currentUserId={currentUserId}
        dispatch={dispatch}
        fallbackFocusRef={selectorRef}
        isModified={isModified}
        resourceKind={projection.resourceKind}
        runAction={runAction}
        selectedSavedView={projection.selectedSavedView}
      />
      <SavedViewStatus
        activeViewSchemaId={activeViewSchemaId}
        compact={condensedControls || projection.selectedSavedView !== null}
        control={control}
        resourceMessage={projection.resourceMessage}
      />
    </div>
  );
}

function SavedViewSelectionField({
  activeViewSchemaId,
  condensedControls,
  dispatch,
  onSelectBaseSurface,
  onSelectSavedView,
  projection,
  selectorRef,
}: {
  readonly activeViewSchemaId: string;
  readonly condensedControls: boolean;
  readonly dispatch: Dispatch<SavedViewControlEvent>;
  readonly onSelectBaseSurface: (viewSchemaId: string) => void;
  readonly onSelectSavedView: (savedView: SavedViewResource) => void;
  readonly projection: ActiveSurfaceSavedViewProjection;
  readonly selectorRef: RefObject<HTMLSelectElement | null>;
}) {
  const descriptionId = `${savedViewSelectorTestId(activeViewSchemaId)}-description`;
  const disabled =
    projection.resourceKind === "loading" ||
    projection.resourceKind === "unavailable";
  return (
    <label
      style={{
        ...savedViewSelectorFrameStyle,
        ...(condensedControls || projection.selectedSavedView !== null
          ? condensedSavedViewSelectorFrameStyle
          : null),
      }}
    >
      {condensedControls ? null : (
        <span style={savedViewSelectorLabelStyle}>View:</span>
      )}
      <select
        ref={selectorRef}
        aria-label="Saved view"
        aria-describedby={descriptionId}
        aria-busy={projection.resourceKind === "loading" || undefined}
        data-active-view-schema-id={activeViewSchemaId}
        data-resource-kind={projection.resourceKind}
        data-selected-saved-view-id={projection.selectedSavedViewId}
        data-selected-sheet-ref-kind={
          projection.selectedSavedViewId === "" ? "view_schema" : "saved_view"
        }
        data-testid={savedViewSelectorTestId(activeViewSchemaId)}
        disabled={disabled}
        style={{
          ...savedViewSelectStyle,
          ...(condensedControls ? compactSavedViewSelectStyle : null),
          ...(!condensedControls && projection.selectedSavedView !== null
            ? allocatedBaseSavedViewSelectStyle
            : null),
        }}
        title={projection.selectedSavedView?.display_name ?? "Unsaved view"}
        value={projection.selectedSavedViewId}
        onChange={(event) => {
          selectSavedView({
            activeViewSchemaId,
            dispatch,
            nextSavedViewId: event.currentTarget.value,
            onSelectBaseSurface,
            onSelectSavedView,
            projection,
          });
        }}
      >
        <option value="">
          {projection.resourceKind === "loading"
            ? "Loading saved views…"
            : "Unsaved view"}
        </option>
        <SavedViewOptionGroup
          activeViewSchemaId={activeViewSchemaId}
          label="Private"
          savedViews={projection.privateSavedViews}
        />
        <SavedViewOptionGroup
          activeViewSchemaId={activeViewSchemaId}
          label="Shared"
          savedViews={projection.sharedSavedViews}
        />
        <SavedViewOptionGroup
          activeViewSchemaId={activeViewSchemaId}
          label="System"
          savedViews={projection.systemSavedViews}
        />
      </select>
      <span id={descriptionId} style={visuallyHiddenStyle}>
        {projection.selectedSavedView === null
          ? "Base surface configuration"
          : `Selected saved view ${projection.selectedSavedView.display_name}, ${projection.selectedSavedView.scope} scope`}
      </span>
    </label>
  );
}

function selectSavedView({
  activeViewSchemaId,
  dispatch,
  nextSavedViewId,
  onSelectBaseSurface,
  onSelectSavedView,
  projection,
}: {
  readonly activeViewSchemaId: string;
  readonly dispatch: Dispatch<SavedViewControlEvent>;
  readonly nextSavedViewId: string;
  readonly onSelectBaseSurface: (viewSchemaId: string) => void;
  readonly onSelectSavedView: (savedView: SavedViewResource) => void;
  readonly projection: ActiveSurfaceSavedViewProjection;
}) {
  dispatch({ type: "clear_feedback", surface: activeViewSchemaId });
  if (nextSavedViewId === "") {
    onSelectBaseSurface(activeViewSchemaId);
    return;
  }
  const savedView = projection.savedViews.find(
    (candidate) => candidate.saved_view_id === nextSavedViewId,
  );
  if (savedView !== undefined) onSelectSavedView(savedView);
}

function SavedViewModifiedBadge({
  activeViewSchemaId,
  condensed,
  isModified,
  selectedSavedView,
}: {
  readonly activeViewSchemaId: string;
  readonly condensed: boolean;
  readonly isModified: boolean;
  readonly selectedSavedView: SavedViewResource | null;
}) {
  if (selectedSavedView === null || !isModified) return null;
  return (
    <span
      data-testid={savedViewModifiedTestId(activeViewSchemaId)}
      style={{
        ...modifiedBadgeStyle,
        ...(condensed ? visuallyHiddenStyle : null),
      }}
      title="Saved view modified"
    >
      Modified
    </span>
  );
}

function SavedViewStatus({
  activeViewSchemaId,
  compact,
  control,
  resourceMessage,
}: {
  readonly activeViewSchemaId: string;
  readonly compact: boolean;
  readonly control: SavedViewSurfaceControlState;
  readonly resourceMessage: string | null;
}) {
  const status = control.feedback?.message ?? resourceMessage ?? "";
  return (
    <span
      aria-live={control.feedback?.kind === "error" ? "assertive" : "polite"}
      data-feedback-kind={control.feedback?.kind}
      data-testid={savedViewStatusTestId(activeViewSchemaId)}
      style={{
        ...savedViewStatusStyle,
        ...(compact ? condensedSavedViewStatusStyle : null),
        ...(status === "" ? emptySavedViewStatusStyle : null),
      }}
      title={status || undefined}
    >
      {status}
    </span>
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
  if (savedViews.length === 0) return null;
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
  flex: "0 1 auto",
  minWidth: 0,
  overflow: "visible",
};

const condensedSavedViewControlGroupStyle = {
  flex: "1 1 auto",
  inlineSize: "100%",
  maxInlineSize: "100%",
};

const selectedBaseSavedViewControlGroupStyle = {
  flex: "0 1 var(--ct-layout-viewBarSavedViewMaxInlineSize)",
  inlineSize: "var(--ct-layout-viewBarSavedViewMaxInlineSize)",
  maxInlineSize: "var(--ct-layout-viewBarSavedViewMaxInlineSize)",
};

const savedViewSelectorFrameStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.4rem",
  flex: "0 1 auto",
  minWidth: 0,
};

const condensedSavedViewSelectorFrameStyle = {
  flex: "1 1 6.5rem",
  minInlineSize: "6.5rem",
  maxInlineSize: "100%",
};

const compactSavedViewControlGroupStyle = { gap: "0.15rem" };

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
  inlineSize: "min(var(--ct-layout-viewBarSavedViewMaxInlineSize), 36vw)",
  minInlineSize: "10rem",
};

const allocatedBaseSavedViewSelectStyle = {
  inlineSize: "100%",
  minInlineSize: "6.5rem",
  maxInlineSize: "100%",
};

const compactSavedViewSelectStyle = {
  inlineSize: "100%",
  minInlineSize: "6.5rem",
  maxInlineSize: "6.5rem",
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

const savedViewStatusStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.78rem",
  maxWidth: "12rem",
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

const condensedSavedViewStatusStyle = visuallyHiddenStyle;

const emptySavedViewStatusStyle = {
  flex: "0 0 0",
  maxInlineSize: 0,
};
