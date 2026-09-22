import { Eraser } from "lucide-react";
import { WorkbookViewBar } from "../../components/WorkbookViewBar";
import { workbookQuietCommandStyle } from "../../components/workbookFormStyles";
import { WorkbookFindControl } from "../../find/WorkbookFindControl";
import type { TimelineWorkbookPresentationModel } from "./useTimelineWorkbookPresentation";

export function TimelineWorkbookViewBarRegion({
  model,
}: {
  readonly model: TimelineWorkbookPresentationModel["viewBar"];
}) {
  return (
    <WorkbookViewBar
      addRowDisabled={model.addRowDisabled}
      chromeMode={model.chromeMode}
      iconOnlyActions={model.chromeMode === "narrow_desktop"}
      findControls={
        <>
          <WorkbookFindControl
            binding={model.find}
            chromeMode={model.chromeMode}
          />
          <button
            type="button"
            data-grid-editor-external-action="true"
            aria-label="Clear contents"
            title="Clear selected cell contents"
            style={workbookQuietCommandStyle}
            onClick={model.onClearContents}
          >
            <Eraser size={16} aria-hidden="true" />
            {model.chromeMode === "base" ? "Clear" : null}
          </button>
        </>
      }
      onAddRow={model.onAddRow}
      onInspectorToggle={model.onInspectorToggle}
      surface={model.surface}
      workingSet={model.workingSet ?? undefined}
    />
  );
}
