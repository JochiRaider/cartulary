import { WorkbookObservedStatusStrip } from "../../components/WorkbookStatusStrip";
import {
  WorkbookSurfaceLayout,
  workbookSurfaceFeedbackStyle,
} from "../../layout/WorkbookSurfaceLayout";
import { TimelineBulkTagControl } from "../components/TimelineBulkTagControl";
import { TimelineWorkbookGrid } from "../components/TimelineWorkbookGrid";
import { TimelineWorkbookNotices } from "../components/TimelineWorkbookNotices";
import { TimelineWorkbookInspectorRegion } from "./TimelineWorkbookInspectorRegion";
import { TimelineWorkbookOverlayRegion } from "./TimelineWorkbookOverlayRegion";
import { TimelineWorkbookViewBarRegion } from "./TimelineWorkbookViewBarRegion";
import type { TimelineWorkbookPresentationModel } from "./useTimelineWorkbookPresentation";

export function TimelineWorkbookView({
  model,
}: {
  readonly model: TimelineWorkbookPresentationModel;
}) {
  return (
    <WorkbookSurfaceLayout
      {...model.layout}
      inspector={
        model.inspector === null ? undefined : (
          <TimelineWorkbookInspectorRegion model={model.inspector} />
        )
      }
      primaryGrid={<TimelineWorkbookGrid {...model.grid} />}
      statusStrip={<WorkbookObservedStatusStrip {...model.status} />}
      viewBar={<TimelineWorkbookViewBarRegion model={model.viewBar} />}
      workAreaFeedback={
        <div style={workbookSurfaceFeedbackStyle}>
          <TimelineBulkTagControl binding={model.bulkTag} />
          <TimelineWorkbookNotices {...model.notices} />
        </div>
      }
      workAreaOverlays={
        <TimelineWorkbookOverlayRegion model={model.overlays} />
      }
    />
  );
}
