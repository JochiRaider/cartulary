import { WorkbookStatusStrip } from "../../components/WorkbookStatusStrip";
import { WorkbookSurfaceLayout } from "../../layout/WorkbookSurfaceLayout";
import { TimelineWorkbookGrid } from "../components/TimelineWorkbookGrid";
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
      statusStrip={<WorkbookStatusStrip {...model.status} />}
      viewBar={<TimelineWorkbookViewBarRegion model={model.viewBar} />}
      workAreaOverlays={
        <TimelineWorkbookOverlayRegion model={model.overlays} />
      }
    />
  );
}
