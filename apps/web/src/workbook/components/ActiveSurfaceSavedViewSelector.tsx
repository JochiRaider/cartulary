import {
  savedViewModifiedTestId,
  savedViewStatusTestId,
} from "@cartulary/ui-contracts";
import {
  type Dispatch,
  type RefObject,
  useEffect,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { SheetRef } from "../../shared/sheetRef";
import {
  type SavedViewActionIntent,
  useActiveSurfaceSavedViewActions,
} from "../hooks/useActiveSurfaceSavedViewActions";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import {
  type ActiveSurfaceSavedViewProjection,
  projectActiveSurfaceSavedViews,
  type SavedViewControlEvent,
  type SavedViewSurfaceControlState,
  type WorkbookSavedViewsResource,
} from "../models/workbookSavedViewControl";
import type { SavedViewResource } from "../models/workbookSavedViews";
import type { WorkbookPreferenceController } from "../preferences/WorkbookPreferenceController";
import { savedViewOutcome } from "../savedviews/savedViewOperationModel";
import type { WorkbookSavedViewController } from "../savedviews/WorkbookSavedViewController";
import { visuallyHiddenStyle } from "../utils/workbookStyles";
import { SavedViewActionPanel } from "./SavedViewActionPanel";
import { SavedViewBrowser } from "./SavedViewBrowser";
import {
  workbookQuietCommandStyle,
  workbookTypography,
} from "./workbookFormStyles";

export type ActiveSurfaceSavedViewSelectorProps = {
  readonly activeViewSchemaId: string;
  readonly chromeMode: WorkbookChromeMode;
  readonly currentIncidentRole: string | null;
  readonly currentUserId: string | null;
  readonly isModified?: boolean | undefined;
  readonly savedViewsResource: WorkbookSavedViewsResource;
  readonly selectedSheetRef: SheetRef;
  readonly controller: WorkbookSavedViewController;
  readonly onSelectBaseSurface: (viewSchemaId: string) => void;
  readonly preferenceController?: WorkbookPreferenceController | undefined;
  readonly onInspectPreferences?:
    | ((target?: HTMLElement | null) => void)
    | undefined;
};

export function ActiveSurfaceSavedViewSelector({
  activeViewSchemaId,
  chromeMode,
  controller,
  currentIncidentRole,
  currentUserId,
  isModified = false,
  savedViewsResource,
  selectedSheetRef,
  onSelectBaseSurface,
  preferenceController,
  onInspectPreferences,
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
  const operationState = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const selected = projection.selectedSavedView;
  const subject = {
    viewSchemaId: activeViewSchemaId,
    savedViewId:
      selectedSheetRef.kind === "saved_view" ? selectedSheetRef.id : null,
    savedViewVersion: selected?.saved_view_version ?? null,
  };
  const selectionKey = `${activeViewSchemaId}:${subject.savedViewId ?? "base"}`;
  const [openSelection, setOpenSelection] = useState<string | null>(null);
  useEffect(
    () =>
      setOpenSelection((current) =>
        current === selectionKey ? current : null,
      ),
    [selectionKey],
  );
  const [notice, setNotice] = useState<string | null>(null);
  const draft = controller.draftFor(subject);
  const operation = operationState.operation;
  const message =
    operationState.notice ?? notice ?? savedViewOutcome(operation);
  const control: SavedViewSurfaceControlState = {
    selectionKey,
    displayName: draft.displayName,
    scope: draft.scope,
    panelOpen: openSelection === selectionKey,
    busy: operationState.transportPending || operation.kind === "pending",
    feedback: message
      ? {
          kind: ["conflict", "uncertain", "rejected"].includes(operation.kind)
            ? "error"
            : operation.kind === "confirmed"
              ? "success"
              : "notice",
          message,
        }
      : null,
  };
  const dispatch = (event: SavedViewControlEvent) => {
    if (event.surface !== activeViewSchemaId) return;
    switch (event.type) {
      case "toggle_panel":
        setOpenSelection(control.panelOpen ? null : selectionKey);
        break;
      case "close_panel":
        setOpenSelection(null);
        break;
      case "change_name":
        controller.changeDraft(subject, { displayName: event.displayName });
        break;
      case "change_scope":
        controller.changeDraft(subject, { scope: event.scope });
        break;
      case "publish_notice":
        setNotice(event.message);
        break;
      case "clear_feedback":
        setNotice(null);
        break;
    }
  };
  const selectorRef = useRef<HTMLButtonElement>(null);
  const { runAction } = useActiveSurfaceSavedViewActions(controller, subject);

  return (
    <SavedViewControlPresentation
      activeViewSchemaId={activeViewSchemaId}
      chromeMode={chromeMode}
      control={control}
      controller={controller}
      currentIncidentRole={currentIncidentRole}
      currentUserId={currentUserId}
      dispatch={dispatch}
      isModified={isModified}
      onSelectBaseSurface={onSelectBaseSurface}
      projection={projection}
      runAction={runAction}
      preferenceController={preferenceController}
      onInspectPreferences={onInspectPreferences}
      selectorRef={selectorRef}
    />
  );
}

function SavedViewControlPresentation({
  activeViewSchemaId,
  chromeMode,
  control,
  controller,
  currentIncidentRole,
  currentUserId,
  dispatch,
  isModified,
  onSelectBaseSurface,
  projection,
  runAction,
  preferenceController,
  onInspectPreferences,
  selectorRef,
}: {
  readonly activeViewSchemaId: string;
  readonly chromeMode: WorkbookChromeMode;
  readonly control: SavedViewSurfaceControlState;
  readonly controller: WorkbookSavedViewController;
  readonly currentIncidentRole: string | null;
  readonly currentUserId: string | null;
  readonly dispatch: Dispatch<SavedViewControlEvent>;
  readonly isModified: boolean;
  readonly onSelectBaseSurface: (viewSchemaId: string) => void;
  readonly projection: ActiveSurfaceSavedViewProjection;
  readonly runAction: (intent: SavedViewActionIntent) => void;
  readonly preferenceController?: WorkbookPreferenceController | undefined;
  readonly onInspectPreferences?:
    | ((target?: HTMLElement | null) => void)
    | undefined;
  readonly selectorRef: RefObject<HTMLButtonElement | null>;
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
        ...(compactControls ? compactSavedViewControlGroupStyle : null),
      }}
    >
      <SavedViewSelectionField
        controller={controller}
        activeViewSchemaId={activeViewSchemaId}
        condensedControls={condensedControls}
        dispatch={dispatch}
        onSelectBaseSurface={onSelectBaseSurface}
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
        controller={controller}
        activeViewSchemaId={activeViewSchemaId}
        control={control}
        currentIncidentRole={currentIncidentRole}
        currentUserId={currentUserId}
        dispatch={dispatch}
        fallbackFocusRef={selectorRef}
        isModified={isModified}
        resourceKind={projection.resourceKind}
        runAction={runAction}
        preferenceController={preferenceController}
        onInspectPreferences={onInspectPreferences}
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
  onSelectBaseSurface,
  projection,
  selectorRef,
  controller,
}: {
  readonly activeViewSchemaId: string;
  readonly condensedControls: boolean;
  readonly dispatch: Dispatch<SavedViewControlEvent>;
  readonly onSelectBaseSurface: (viewSchemaId: string) => void;
  readonly projection: ActiveSurfaceSavedViewProjection;
  readonly selectorRef: RefObject<HTMLButtonElement | null>;
  readonly controller: WorkbookSavedViewController;
}) {
  return (
    <div
      style={{
        ...savedViewSelectorFrameStyle,
        ...(condensedControls || projection.selectedSavedView
          ? condensedSavedViewSelectorFrameStyle
          : null),
      }}
    >
      {condensedControls || projection.selectedSavedView ? null : (
        <span style={savedViewSelectorLabelStyle}>View:</span>
      )}
      <SavedViewBrowser
        controller={controller}
        schema={activeViewSchemaId}
        projection={projection}
        triggerRef={selectorRef}
        triggerStyle={savedViewSelectStyle}
        onBase={() => onSelectBaseSurface(activeViewSchemaId)}
      />
    </div>
  );
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

const savedViewControlGroupStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.4rem",
  flex: "1 1 auto",
  inlineSize: "100%",
  minWidth: 0,
  overflow: "visible",
};

const condensedSavedViewControlGroupStyle = {
  flex: "1 1 auto",
  inlineSize: "100%",
  maxInlineSize: "100%",
};

const savedViewSelectorFrameStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.4rem",
  flex: "1 1 auto",
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

const savedViewSelectStyle = {
  ...workbookQuietCommandStyle,
  ...workbookTypography("surface-title"),
  boxSizing: "border-box" as const,
  inlineSize: "100%",
  minInlineSize: 0,
  justifyContent: "flex-start",
  color: "var(--ct-colors-ink)",
};
const modifiedBadgeStyle = {
  ...workbookTypography("metadata"),
  color: "var(--ct-colors-ink-subtle)",
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
