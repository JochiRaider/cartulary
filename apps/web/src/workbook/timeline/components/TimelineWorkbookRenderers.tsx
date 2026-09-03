import type { ViewContract } from "@cartulary/view-contracts";
import { useCallback } from "react";
import type { PresenceRecord } from "../../utils/workbookPresence";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type {
  FocusFieldKey,
  TimelineScalarEditorSurface,
} from "../models/timelineFieldRegistry";
import type {
  TimelineCollectionKeyDown,
  TimelineCollectionSave,
  TimelineEntityIndex,
  TimelineScalarBlurCommit,
  TimelineScalarGridCommit,
  TimelineScalarKeyCommit,
  TimelineScalarPasteCommit,
  TimelineWorkbookRenderers,
} from "./TimelineWorkbookRendererTypes";
import { useTimelineCollectionRenderer } from "./useTimelineCollectionRenderer";
import { useTimelineColumnAssembly } from "./useTimelineColumnAssembly";
import { useTimelineScalarRenderers } from "./useTimelineScalarRenderers";

export type { TimelineWorkbookRenderers } from "./TimelineWorkbookRendererTypes";

export function useTimelineWorkbookRenderers({
  activateCollectionInput,
  activateConflictCell,
  activeCollectionInputKey,
  conflictQueue,
  commitScalarGridEdit,
  editorDraftRegistry,
  editingPresenceForCell,
  entityIndex,
  gridShellWidth,
  handleBlur,
  handleCollectionInputChange,
  handleCollectionKeyDown,
  handleEditModePresence,
  handleKeyDown,
  handlePaste,
  handleSelectMention,
  handleSelectRow,
  openInspectorForRow,
  queueCollectionSave,
  readOnly,
  rowGutterWidth,
  deactivateCollectionInput,
  timelineContract,
  updateTimelineSurfaceFocusAnchor,
}: {
  readonly activateCollectionInput: (focusKey: string) => void;
  readonly activateConflictCell: (key: string | null) => void;
  readonly activeCollectionInputKey: string | null;
  readonly conflictQueue: Record<string, { readonly key: string }>;
  readonly commitScalarGridEdit: TimelineScalarGridCommit;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly editingPresenceForCell: (
    recordId: string | null,
    fieldKey: string,
  ) => readonly PresenceRecord[];
  readonly entityIndex: TimelineEntityIndex;
  readonly gridShellWidth: number;
  readonly handleBlur: TimelineScalarBlurCommit;
  readonly handleCollectionInputChange: (
    focusKey: string,
    value: string,
  ) => void;
  readonly handleCollectionKeyDown: TimelineCollectionKeyDown;
  readonly handleEditModePresence: (
    recordId: string | null,
    fieldKey: string,
    editing: boolean,
  ) => void;
  readonly handleKeyDown: TimelineScalarKeyCommit;
  readonly handlePaste: TimelineScalarPasteCommit;
  readonly handleSelectMention: (recordId: string, itemRef: string) => void;
  readonly handleSelectRow: (recordId: string) => void;
  readonly openInspectorForRow: (recordId: string) => void;
  readonly queueCollectionSave: TimelineCollectionSave;
  readonly readOnly: boolean;
  readonly rowGutterWidth: number;
  readonly deactivateCollectionInput: (focusKey: string) => void;
  readonly timelineContract: ViewContract;
  readonly updateTimelineSurfaceFocusAnchor: (
    recordId: string | null,
    fieldKey: string,
  ) => void;
}): TimelineWorkbookRenderers {
  const registerInput = useCallback(
    (
      rowKey: string,
      field: FocusFieldKey,
      surface: TimelineScalarEditorSurface,
      element: HTMLInputElement | HTMLTextAreaElement | null,
    ) => {
      editorDraftRegistry.registerInput({ field, rowKey, surface }, element);
    },
    [editorDraftRegistry],
  );
  const timelineBindingLabel = useCallback(
    (fieldKey: string) =>
      timelineContract.fieldMap[fieldKey]?.label ?? fieldKey,
    [timelineContract],
  );

  const {
    renderTimelineGridEditor,
    renderTimelineInspectorEditor,
    renderTimelineScalarCell,
  } = useTimelineScalarRenderers({
    conflictQueue,
    editingPresenceForCell,
    editorDraftRegistry,
    handleBlur,
    handleEditModePresence,
    handleKeyDown,
    handlePaste,
    handleSelectRow,
    readOnly,
    registerInput,
    setActiveConflictKey: activateConflictCell,
    timelineBindingLabel,
    updateTimelineSurfaceFocusAnchor,
  });
  const renderTimelineCollectionInput = useTimelineCollectionRenderer({
    activateCollectionInput,
    activeCollectionInputKey,
    entityIndex,
    handleCollectionInputChange,
    handleCollectionKeyDown,
    handleSelectMention,
    handleSelectRow,
    openInspectorForRow,
    queueCollectionSave,
    readOnly,
    registerInput,
    resolveInputElement: editorDraftRegistry.inputElementForFocusKey,
    deactivateCollectionInput,
    timelineBindingLabel,
    updateTimelineSurfaceFocusAnchor,
  });
  const timelineColumns = useTimelineColumnAssembly({
    commitScalarGridEdit,
    editorDraftRegistry,
    gridShellWidth,
    renderTimelineCollectionInput,
    renderTimelineGridEditor,
    renderTimelineScalarCell,
    rowGutterWidth,
    timelineBindingLabel,
    timelineContract,
  });

  return {
    renderTimelineCollectionInput,
    renderTimelineGridEditor,
    renderTimelineInspectorEditor,
    timelineBindingLabel,
    timelineColumns,
  };
}
