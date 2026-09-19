import { useCallback } from "react";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type { TimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import type {
  CollectionFieldKey,
  TimelineCollectionBinding,
  TimelineScalarEditorSurface,
} from "../models/timelineFieldRegistry";
import type { WorkbookRow } from "../models/timelineRowModel";
import { TimelineCollectionCell } from "./TimelineCollectionCell";
import type {
  RegisterTimelineInput,
  TimelineCollectionKeyDown,
  TimelineCollectionSave,
  TimelineEntityIndex,
} from "./TimelineWorkbookRendererTypes";

export function useTimelineCollectionRenderer({
  entityIndex,
  editorDraftRegistry,
  elementRegistry,
  handleInspectCollection,
  handleCollectionKeyDown,
  handleSelectRow,
  queueCollectionSave,
  readOnly,
  registerInput,
  timelineBindingLabel,
  updateTimelineSurfaceFocusAnchor,
}: {
  readonly entityIndex: TimelineEntityIndex;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly elementRegistry: TimelineInspectorElementRegistry;
  readonly handleInspectCollection: (
    recordId: string,
    fieldKey: CollectionFieldKey,
    itemRef: string,
  ) => void;
  readonly handleCollectionKeyDown: TimelineCollectionKeyDown;
  readonly handleSelectRow: (recordId: string) => void;
  readonly queueCollectionSave: TimelineCollectionSave;
  readonly readOnly: boolean;
  readonly registerInput: RegisterTimelineInput;
  readonly timelineBindingLabel: (fieldKey: string) => string;
  readonly updateTimelineSurfaceFocusAnchor: (
    recordId: string | null,
    fieldKey: string,
  ) => void;
}) {
  return useCallback(
    (
      row: WorkbookRow,
      binding: TimelineCollectionBinding,
      focusTargetRef?: (element: HTMLInputElement | null) => void,
      surface: TimelineScalarEditorSurface = "grid",
    ) => (
      <TimelineCollectionCell
        surface={surface}
        isInspectionControlTarget={elementRegistry.isInspectionControlTarget}
        registerCollectionItem={elementRegistry.registerCollectionItem}
        registerTrigger={elementRegistry.registerCollectionTrigger}
        rememberReturnFocus={elementRegistry.rememberCollectionReturnFocus}
        handleInspectCollection={handleInspectCollection}
        editorDraftRegistry={editorDraftRegistry}
        binding={binding}
        entityIndex={entityIndex}
        {...(focusTargetRef === undefined ? {} : { focusTargetRef })}
        handleCollectionKeyDown={handleCollectionKeyDown}
        handleSelectRow={handleSelectRow}
        label={timelineBindingLabel(binding.fieldKey)}
        queueCollectionSave={queueCollectionSave}
        readOnly={readOnly}
        registerInput={registerInput}
        row={row}
        updateTimelineSurfaceFocusAnchor={updateTimelineSurfaceFocusAnchor}
      />
    ),
    [
      entityIndex,
      editorDraftRegistry,
      elementRegistry,
      handleInspectCollection,
      handleCollectionKeyDown,
      handleSelectRow,
      queueCollectionSave,
      readOnly,
      registerInput,
      timelineBindingLabel,
      updateTimelineSurfaceFocusAnchor,
    ],
  );
}
