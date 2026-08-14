import type {
  GridColumn,
  GridEditCommitOutcome,
  GridEditorFocusTarget,
} from "@cartulary/grid-adapter";
import type {
  ClipboardEvent as ReactClipboardEvent,
  KeyboardEvent as ReactKeyboardEvent,
  ReactNode,
} from "react";
import type {
  CollectionDraftKey,
  CollectionFieldKey,
  FocusFieldKey,
  RowValues,
  TimelineCollectionBinding,
  TimelineScalarBinding,
  TimelineScalarEditorSurface,
  WorkbookRow,
} from "../models/workbookTimelineModel";

export type TimelineEntityIndex = Record<string, { label: string }>;

export type TimelineScalarBlurCommit = (
  rowKey: string,
  field: keyof RowValues,
  surface: TimelineScalarEditorSurface,
  value: string,
) => void;

export type TimelineScalarKeyCommit = (
  event: ReactKeyboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  rowKey: string,
  field: keyof RowValues,
  surface: TimelineScalarEditorSurface,
) => void;

export type TimelineScalarPasteCommit = (
  event: ReactClipboardEvent<HTMLInputElement | HTMLTextAreaElement>,
  rowKey: string,
  field: keyof RowValues,
  surface: TimelineScalarEditorSurface,
) => void;

export type TimelineScalarGridCommit = (
  rowKey: string,
  field: keyof RowValues,
  value: string,
) => Promise<GridEditCommitOutcome>;

export type RegisterTimelineInput = (
  rowKey: string,
  field: FocusFieldKey,
  surface: TimelineScalarEditorSurface,
  dataTestId: string,
  element: HTMLInputElement | HTMLTextAreaElement | null,
) => void;

export type TimelineCollectionSave = (
  rowKey: string,
  fieldKey: CollectionFieldKey,
  focusField: CollectionDraftKey,
  draftValueOverride?: string,
  source?: "keyboard" | "blur",
) => void;

export type TimelineCollectionKeyDown = (
  event: ReactKeyboardEvent<HTMLInputElement>,
  rowKey: string,
  fieldKey: CollectionFieldKey,
  draftKey: CollectionDraftKey,
) => void;

export type RenderTimelineCollectionInput = (
  row: WorkbookRow,
  binding: TimelineCollectionBinding,
) => ReactNode;

export type RenderTimelineGridEditor = (
  row: WorkbookRow,
  binding: TimelineScalarBinding,
  closeGridEditor?: ((commit: boolean, draftValue: string) => void) | undefined,
  focusTargetRef?:
    | ((element: GridEditorFocusTarget | null) => void)
    | undefined,
  controlledDraftValue?: string | undefined,
  onControlledDraftChange?: ((value: string) => void) | undefined,
) => ReactNode;

export type RenderTimelineScalarCell = (
  row: WorkbookRow,
  binding: TimelineScalarBinding,
) => ReactNode;

export type TimelineWorkbookRenderers = {
  readonly renderTimelineCollectionInput: RenderTimelineCollectionInput;
  readonly renderTimelineGridEditor: RenderTimelineGridEditor;
  readonly renderTimelineInspectorEditor: (
    row: WorkbookRow,
    binding: TimelineScalarBinding,
  ) => ReactNode;
  readonly timelineBindingLabel: (fieldKey: string) => string;
  readonly timelineColumns: readonly GridColumn<WorkbookRow>[];
};
