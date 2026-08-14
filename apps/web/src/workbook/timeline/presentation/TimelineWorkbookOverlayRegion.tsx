import { TimelineRowContextMenu } from "../components/TimelineRowActions";
import { TimelineWorkbookNotices } from "../components/TimelineWorkbookNotices";
import type { TimelineWorkbookPresentationModel } from "./useTimelineWorkbookPresentation";

export function TimelineWorkbookOverlayRegion({
  model,
}: {
  readonly model: TimelineWorkbookPresentationModel["overlays"];
}) {
  return (
    <>
      <TimelineWorkbookNotices {...model.notices} />
      {model.contextMenu === null ? null : (
        <TimelineRowContextMenu {...model.contextMenu} />
      )}
    </>
  );
}
