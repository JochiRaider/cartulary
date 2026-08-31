import { WorkbookGridControls } from "../../components/WorkbookGridControls";
import { WorkbookViewBar } from "../../components/WorkbookViewBar";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";
import type { TimelineWorkbookPresentationModel } from "./useTimelineWorkbookPresentation";

const bulkActionFieldsetStyle = {
  border: 0,
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  margin: 0,
  padding: 0,
};

export function TimelineWorkbookViewBarRegion({
  model,
}: {
  readonly model: TimelineWorkbookPresentationModel["viewBar"];
}) {
  const bulk = model.bulk;
  return (
    <WorkbookViewBar
      addRowDisabled={model.addRowDisabled}
      chromeMode={model.chromeMode}
      queryControls={
        model.inlineQuery === null ? (
          model.viewBarQueryControls
        ) : (
          <WorkbookGridControls {...model.inlineQuery} />
        )
      }
      savedViewControls={model.savedViewControls}
      supplementalControls={
        bulk === null ? undefined : (
          <fieldset style={bulkActionFieldsetStyle}>
            <legend style={visuallyHiddenStyle}>
              Timeline bulk record actions
            </legend>
            <span aria-live="polite">{bulk.selectedCount} selected</span>
            <input
              aria-label="Tag for selected Timeline records"
              disabled={!bulk.canAssign || bulk.selectedCount === 0}
              placeholder="Tag selected"
              type="text"
              value={bulk.tagName}
              onChange={(event) => {
                bulk.onTagNameChange(event.target.value);
              }}
            />
            <button
              disabled={!bulk.canSubmit}
              type="button"
              onClick={() => {
                void bulk.onAssign();
              }}
            >
              Assign tag
            </button>
            {bulk.message === null ? null : (
              <span
                aria-live={
                  bulk.message.kind === "error" ? "assertive" : "polite"
                }
                role={bulk.message.kind === "error" ? "alert" : "status"}
              >
                {bulk.message.message}
              </span>
            )}
          </fieldset>
        )
      }
      onAddRow={model.onAddRow}
      onInspectorToggle={model.onInspectorToggle}
      surface={model.surface}
    />
  );
}
