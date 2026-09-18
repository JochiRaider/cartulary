import { Eraser } from "lucide-react";
import { WorkbookViewBar } from "../../components/WorkbookViewBar";
import { WorkbookFindControl } from "../../find/WorkbookFindControl";
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
            style={{
              borderRadius: "var(--ct-rounded-xs)",
              border: "var(--ct-border-hairline)",
              background: "var(--ct-colors-surface-1)",
              color: "var(--ct-colors-ink)",
              font: "inherit",
              minHeight: "1.75rem",
              padding: "0.22rem 0.45rem",
              whiteSpace: "nowrap",
              cursor: "pointer",
              display: "inline-flex",
              alignItems: "center",
              justifyContent: "center",
              gap: "0.35rem",
            }}
            onClick={model.onClearContents}
          >
            <Eraser size={16} aria-hidden="true" />
            {model.chromeMode === "base" ? "Clear" : null}
          </button>
        </>
      }
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
      workingSet={model.workingSet ?? undefined}
    />
  );
}
