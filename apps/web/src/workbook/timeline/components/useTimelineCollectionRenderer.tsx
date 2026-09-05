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
  activateCollectionInput,
  activeCollectionInputKey,
  deactivateCollectionInput,
  entityIndex,
  editorDraftRegistry,
  elementRegistry,
  handleInspectCollection,
  handleCollectionInputChange,
  handleCollectionKeyDown,
  handleSelectRow,
  queueCollectionSave,
  readOnly,
  registerInput,
  timelineBindingLabel,
  updateTimelineSurfaceFocusAnchor,
}: {
  readonly activateCollectionInput: (focusKey: string) => void;
  readonly activeCollectionInputKey: string | null;
  readonly deactivateCollectionInput: (focusKey: string) => void;
  readonly entityIndex: TimelineEntityIndex;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
  readonly elementRegistry: TimelineInspectorElementRegistry;
  readonly handleInspectCollection: (
    recordId: string,
    fieldKey: CollectionFieldKey,
    itemRef: string,
  ) => void;
  readonly handleCollectionInputChange: (
    focusKey: string,
    value: string,
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
        retainedDraft={editorDraftRegistry.draftValue({
          rowKey: row.key,
          field: binding.draftKey,
          surface: "grid",
        })}
        retainDraft={(value) =>
          editorDraftRegistry.setDraft(
            { rowKey: row.key, field: binding.draftKey, surface: "grid" },
            value,
          )
        }
        activateCollectionInput={activateCollectionInput}
        activeCollectionInputKey={activeCollectionInputKey}
        binding={binding}
        deactivateCollectionInput={deactivateCollectionInput}
        entityIndex={entityIndex}
        {...(focusTargetRef === undefined ? {} : { focusTargetRef })}
        handleCollectionInputChange={handleCollectionInputChange}
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
      activateCollectionInput,
      activeCollectionInputKey,
      deactivateCollectionInput,
      entityIndex,
      editorDraftRegistry,
      elementRegistry,
      handleInspectCollection,
      handleCollectionInputChange,
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
