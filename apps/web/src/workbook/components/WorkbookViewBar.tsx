import {
  workbookAddRowButtonTestId,
  workbookInspectorToggleTestId,
} from "@cartulary/ui-contracts";
import { Plus, SlidersHorizontal } from "lucide-react";
import {
  type CSSProperties,
  type ReactNode,
  type Ref,
  type RefObject,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import { useWorkbookCommand } from "../commands/WorkbookCommands";
import { useWorkbookMenuPlacement } from "../layout/useWorkbookMenuPlacement";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import { useWorkbookWorkbench } from "../navigation/WorkbookWorkbenchContext";
import { WorkbookReturnControl } from "../navigation/WorkbookWorkPanel";
import { useWorkbookBrowsingRegistry } from "../query/WorkbookQueryBrowsingContext";
import {
  ActiveSurfaceSavedViewSelector,
  type ActiveSurfaceSavedViewSelectorProps,
} from "./ActiveSurfaceSavedViewSelector";
import {
  WorkbookGridControls,
  type WorkbookGridControlsProps,
} from "./WorkbookGridControls";
import { WorkbookTimelinePresets } from "./WorkbookTimelinePresets";
import { workbookQuietCommandStyle } from "./workbookFormStyles";

export type WorkbookViewBarWorkingSetBinding = {
  readonly query: Omit<
    WorkbookGridControlsProps,
    "chromeMode" | "composeControls"
  > | null;
  readonly savedView: Omit<
    ActiveSurfaceSavedViewSelectorProps,
    "chromeMode"
  > | null;
};

type WorkbookViewBarProps = {
  readonly addRowDisabled?: boolean | undefined;
  readonly addRowLabel?: string | undefined;
  readonly chromeMode?: WorkbookChromeMode | undefined;
  readonly iconOnlyActions?: boolean | undefined;
  readonly inspectorButtonRef?: Ref<HTMLButtonElement> | undefined;
  readonly onAddRow?: (() => void) | undefined;
  readonly onInspectorOpen?: (() => void) | undefined;
  readonly findControls?: ReactNode | undefined;
  readonly surface: string;
  readonly workingSet?: WorkbookViewBarWorkingSetBinding | undefined;
};

export function WorkbookViewBar({
  addRowDisabled = false,
  addRowLabel = "Add row",
  chromeMode = "base",
  iconOnlyActions = false,
  inspectorButtonRef,
  onAddRow,
  onInspectorOpen,
  findControls,
  surface,
  workingSet,
}: WorkbookViewBarProps) {
  const workbench = useWorkbookWorkbench();
  const browsing = useWorkbookBrowsingRegistry();
  const queryMenu = useRef<((focusTarget?: HTMLElement) => void) | null>(null);
  const viewMenu = useRef<((focusTarget?: HTMLElement) => void) | null>(null);
  const registerInspector = workbench?.registerInspector;
  useLayoutEffect(
    () =>
      onInspectorOpen
        ? registerInspector?.(surface, onInspectorOpen)
        : undefined,
    [registerInspector, surface, onInspectorOpen],
  );
  useWorkbookCommand(
    onAddRow
      ? {
          id: "capture.add_row",
          family: "Capture",
          label: addRowLabel,
          terms: ["capture", "new", "insert"],
          targetKind: "surface",
          availability: (target) =>
            target.kind !== "surface" || target.viewSchemaId !== surface
              ? "The selected surface changed."
              : addRowDisabled
                ? "This surface is read-only."
                : null,
          invoke: (target) => {
            if (
              target.kind !== "surface" ||
              target.viewSchemaId !== surface ||
              addRowDisabled
            )
              return false;
            onAddRow();
            return true;
          },
        }
      : null,
  );
  useWorkbookCommand(
    onInspectorOpen
      ? {
          id: "inspect.record",
          family: "Inspect",
          label: "Inspect selected record",
          terms: [
            "details",
            "full value",
            "source",
            "relationships",
            "evidence",
            "workflow",
            "history",
          ],
          targetKind: "record",
          availability: (target) =>
            target.kind === "record" && target.viewSchemaId === surface
              ? null
              : "Select a saved record on the current surface.",
          invoke: (target) => {
            if (target.kind !== "record" || target.viewSchemaId !== surface)
              return false;
            if (!browsing.selectLoadedRecord(surface, target.recordId))
              return false;
            onInspectorOpen();
            return true;
          },
        }
      : null,
  );
  useWorkbookCommand(
    onInspectorOpen && workbench
      ? {
          id: "inspect.value",
          family: "Inspect",
          label: "Inspect value",
          terms: ["source", "raw", "read", "full"],
          targetKind: "cell",
          availability: (target) =>
            target.kind === "cell" && target.viewSchemaId === surface
              ? null
              : "Select a committed cell.",
          invoke: () => {
            if (!workbench.requestInspectValue()) return false;
            onInspectorOpen();
            return true;
          },
        }
      : null,
  );
  const queryInMenu =
    chromeMode === "compact_desktop" ||
    chromeMode === "below_supported_minimum";
  const compactActions =
    iconOnlyActions ||
    chromeMode === "compact_desktop" ||
    chromeMode === "below_supported_minimum";
  return (
    <section
      aria-label="Workbook query and action controls"
      data-chrome-mode={chromeMode}
      style={viewBarStyle}
    >
      <div style={controlRailStyle}>
        {chromeMode === "base" ? <WorkbookReturnControl /> : null}
        {workingSet?.savedView && chromeMode !== "below_supported_minimum" ? (
          <div style={savedViewAllocationStyleFor(chromeMode)}>
            <ActiveSurfaceSavedViewSelector
              {...workingSet.savedView}
              chromeMode={chromeMode}
              presets={
                workingSet.query?.onApplyPreset ? (
                  <WorkbookTimelinePresets
                    onApply={workingSet.query.onApplyPreset}
                  />
                ) : undefined
              }
            />
          </div>
        ) : null}
        {workingSet?.query ? (
          <div style={queryAllocationStyle}>
            <WorkbookGridControls
              {...workingSet.query}
              chromeMode={chromeMode}
              menu={queryInMenu}
              onRequestMenu={(panel, focusTarget) => {
                if (panel === "columns") {
                  if (chromeMode !== "base") viewMenu.current?.(focusTarget);
                } else if (queryInMenu) queryMenu.current?.(focusTarget);
              }}
              composeControls={({ query, columns }) => (
                <>
                  <ViewBarMenu
                    label="Query"
                    enabled={queryInMenu}
                    menuRef={queryMenu}
                  >
                    {query}
                  </ViewBarMenu>
                  <ViewBarMenu
                    label="View options"
                    enabled={chromeMode !== "base"}
                    menuRef={viewMenu}
                  >
                    {chromeMode === "below_supported_minimum" &&
                    workingSet.savedView ? (
                      <ActiveSurfaceSavedViewSelector
                        {...workingSet.savedView}
                        chromeMode="base"
                        presets={
                          workingSet.query?.onApplyPreset ? (
                            <WorkbookTimelinePresets
                              onApply={workingSet.query.onApplyPreset}
                            />
                          ) : undefined
                        }
                      />
                    ) : null}
                    {chromeMode !== "base" ? <WorkbookReturnControl /> : null}
                    {columns}
                    {chromeMode !== "base" && onInspectorOpen && workbench ? (
                      <button
                        type="button"
                        style={toolbarButtonStyle}
                        data-grid-editor-external-action="true"
                        onClick={() => {
                          if (workbench.requestInspectValue())
                            onInspectorOpen();
                        }}
                      >
                        Inspect value
                      </button>
                    ) : null}
                  </ViewBarMenu>
                </>
              )}
            />
          </div>
        ) : null}
      </div>
      <div style={rightRailStyle}>
        {findControls}
        {chromeMode === "base" && onInspectorOpen && workbench ? (
          <button
            type="button"
            style={toolbarButtonStyle}
            data-grid-editor-external-action="true"
            onClick={() => {
              if (workbench.requestInspectValue()) onInspectorOpen();
            }}
          >
            Inspect value
          </button>
        ) : null}
        {onInspectorOpen ? (
          <button
            aria-label="Open inspector"
            data-testid={workbookInspectorToggleTestId(surface)}
            ref={inspectorButtonRef}
            style={toolbarButtonStyle}
            title={compactActions ? "Open inspector" : undefined}
            type="button"
            onClick={onInspectorOpen}
          >
            <SlidersHorizontal aria-hidden="true" size={16} />
            {compactActions ? null : "Inspect"}
          </button>
        ) : null}
        {onAddRow ? (
          <button
            aria-label={addRowLabel}
            data-testid={workbookAddRowButtonTestId(surface)}
            disabled={addRowDisabled}
            style={primaryToolbarButtonStyle}
            title={compactActions ? addRowLabel : undefined}
            type="button"
            onClick={onAddRow}
          >
            <Plus aria-hidden="true" size={16} />
            {compactActions ? null : addRowLabel}
          </button>
        ) : null}
      </div>
    </section>
  );
}

const toolbarButtonStyle = {
  ...workbookQuietCommandStyle,
} satisfies CSSProperties;

const primaryToolbarButtonStyle = {
  ...toolbarButtonStyle,
  borderColor: "var(--ct-colors-accent-active)",
  background: "var(--ct-colors-accent)",
  color: "var(--ct-colors-on-accent)",
  fontWeight: 700,
} satisfies CSSProperties;

const viewBarStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) auto",
  alignItems: "center",
  gap: "0.5rem",
  blockSize: "var(--ct-layout-viewBarHeight)",
  minWidth: 0,
  padding: "0 var(--ct-spacing-sm)",
  borderBlockEnd: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  overflow: "visible",
} satisfies CSSProperties;

const controlRailStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.45rem",
  inlineSize: "100%",
  maxInlineSize: "100%",
  minWidth: 0,
  overflow: "visible",
} satisfies CSSProperties;

function savedViewAllocationStyleFor(
  chromeMode: WorkbookChromeMode,
): CSSProperties {
  if (chromeMode === "below_supported_minimum") return { display: "none" };
  const inlineSize =
    chromeMode === "base"
      ? "var(--ct-layout-viewBarSavedViewBaseMinInlineSize)"
      : chromeMode === "narrow_desktop"
        ? "var(--ct-layout-viewBarSavedViewNarrowMinInlineSize)"
        : "var(--ct-layout-viewBarSavedViewCompactMinInlineSize)";
  return {
    display: "flex",
    alignItems: "center",
    flex: `1 0 ${inlineSize}`,
    minInlineSize: inlineSize,
    overflow: "visible",
  };
}

const queryAllocationStyle = {
  display: "flex",
  alignItems: "center",
  flex: "0 0 auto",
  maxInlineSize: "100%",
  minInlineSize: 0,
  overflow: "visible",
} satisfies CSSProperties;

const rightRailStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "end",
  gap: "0.45rem",
  flex: "0 0 auto",
  minInlineSize: "max-content",
} satisfies CSSProperties;

function ViewBarMenu({
  label,
  enabled,
  children,
  menuRef,
}: {
  readonly menuRef?: RefObject<((focusTarget?: HTMLElement) => void) | null>;
  readonly label: string;
  readonly enabled: boolean;
  readonly children: ReactNode;
}) {
  const root = useRef<HTMLFieldSetElement>(null);
  const panel = useRef<HTMLDivElement>(null);
  const [pendingFocus, setPendingFocus] = useState<HTMLElement | null>(null);
  const [open, setOpen] = useState(false);
  useWorkbookMenuPlacement(enabled && open, panel, root);
  useLayoutEffect(() => {
    if (menuRef)
      menuRef.current = (focusTarget) => {
        setPendingFocus(focusTarget ?? null);
        setOpen(true);
      };
    return () => {
      if (menuRef) menuRef.current = null;
    };
  }, [menuRef]);
  useLayoutEffect(() => {
    setOpen(
      (current) =>
        enabled &&
        (current || !!root.current?.contains(document.activeElement)),
    );
  }, [enabled]);
  useLayoutEffect(() => {
    if (!pendingFocus) return;
    const target = pendingFocus;
    setPendingFocus(null);
    if (!enabled || !open) return;
    if (target?.isConnected && panel.current?.contains(target))
      target.focus({ preventScroll: true });
  }, [enabled, open, pendingFocus]);
  useEffect(() => {
    if (!enabled) return;
    const closeOutside = (event: Event) => {
      if (event.target instanceof Node && !root.current?.contains(event.target))
        setOpen(false);
    };
    document.addEventListener("pointerdown", closeOutside, true);
    document.addEventListener("focusin", closeOutside, true);
    return () => {
      document.removeEventListener("pointerdown", closeOutside, true);
      document.removeEventListener("focusin", closeOutside, true);
    };
  }, [enabled]);
  return (
    <fieldset
      aria-label={`${label} commands`}
      ref={root}
      style={{
        overflow: "visible",
        position: "relative",
        display: enabled ? "block" : "contents",
        border: 0,
        margin: 0,
        padding: 0,
        minInlineSize: 0,
      }}
      data-grid-editor-external-action="true"
      onKeyDown={(event) => {
        if (
          enabled &&
          event.key === "Escape" &&
          !event.defaultPrevented &&
          !event.nativeEvent.isComposing
        ) {
          event.preventDefault();
          event.stopPropagation();
          setOpen(false);
          root.current?.querySelector("button")?.focus();
        }
      }}
    >
      <button
        type="button"
        aria-label={`${label} controls`}
        aria-expanded={open}
        onClick={() => setOpen((value) => !value)}
        style={{ ...toolbarButtonStyle, display: enabled ? undefined : "none" }}
      >
        {label} ▾
      </button>
      <div
        ref={panel}
        popover={enabled ? "manual" : undefined}
        style={
          enabled
            ? {
                color: "inherit",
                position: "absolute",
                insetBlockStart: "100%",
                insetInlineStart: 0,
                inlineSize:
                  "min(var(--ct-layout-viewBarOverlayMaxInlineSize), 85vw)",
                maxBlockSize: "70vh",
                overflow: "auto",
                boxSizing: "border-box",
                padding: "var(--ct-spacing-sm)",
                display: open ? "grid" : "none",
                gap: "var(--ct-spacing-sm)",
                border: "var(--ct-border-hairline)",
                background: "var(--ct-colors-surface-1)",
                boxShadow: "var(--ct-elevation-popover)",
                zIndex: 30,
              }
            : { display: "contents", overflow: "visible" }
        }
      >
        {children}
      </div>
    </fieldset>
  );
}
