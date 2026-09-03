import { useCallback } from "react";
import type { TimelineCollectionBinding } from "../models/timelineFieldRegistry";
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
  handleCollectionInputChange,
  handleCollectionKeyDown,
  handleSelectMention,
  handleSelectRow,
  openInspectorForRow,
  queueCollectionSave,
  readOnly,
  registerInput,
  resolveInputElement,
  timelineBindingLabel,
  updateTimelineSurfaceFocusAnchor,
}: {
  readonly activateCollectionInput: (focusKey: string) => void;
  readonly activeCollectionInputKey: string | null;
  readonly deactivateCollectionInput: (focusKey: string) => void;
  readonly entityIndex: TimelineEntityIndex;
  readonly handleCollectionInputChange: (
    focusKey: string,
    value: string,
  ) => void;
  readonly handleCollectionKeyDown: TimelineCollectionKeyDown;
  readonly handleSelectMention: (recordId: string, itemRef: string) => void;
  readonly handleSelectRow: (recordId: string) => void;
  readonly openInspectorForRow: (recordId: string) => void;
  readonly queueCollectionSave: TimelineCollectionSave;
  readonly readOnly: boolean;
  readonly registerInput: RegisterTimelineInput;
  readonly resolveInputElement: (
    focusKey: string,
  ) => HTMLInputElement | HTMLTextAreaElement | null;
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
    ) => (
      <TimelineCollectionCell
        activateCollectionInput={activateCollectionInput}
        activeCollectionInputKey={activeCollectionInputKey}
        binding={binding}
        deactivateCollectionInput={deactivateCollectionInput}
        entityIndex={entityIndex}
        {...(focusTargetRef === undefined ? {} : { focusTargetRef })}
        handleCollectionInputChange={handleCollectionInputChange}
        handleCollectionKeyDown={handleCollectionKeyDown}
        handleSelectMention={handleSelectMention}
        handleSelectRow={handleSelectRow}
        label={timelineBindingLabel(binding.fieldKey)}
        openInspectorForRow={openInspectorForRow}
        queueCollectionSave={queueCollectionSave}
        readOnly={readOnly}
        registerInput={registerInput}
        resolveInputElement={resolveInputElement}
        row={row}
        updateTimelineSurfaceFocusAnchor={updateTimelineSurfaceFocusAnchor}
      />
    ),
    [
      activateCollectionInput,
      activeCollectionInputKey,
      deactivateCollectionInput,
      entityIndex,
      handleCollectionInputChange,
      handleCollectionKeyDown,
      handleSelectMention,
      handleSelectRow,
      openInspectorForRow,
      queueCollectionSave,
      readOnly,
      registerInput,
      resolveInputElement,
      timelineBindingLabel,
      updateTimelineSurfaceFocusAnchor,
    ],
  );
}
