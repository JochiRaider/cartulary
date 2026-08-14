import { TimelineCollaborationBoundary } from "../collaboration/TimelineCollaborationBoundary";
import { useTimelineWorkbookComposition } from "../composition/useTimelineWorkbookComposition";
import type { TimelineWorkbookSurfaceRuntime } from "../models/timelineWorkbookSurfaceRuntime";
import { TimelineWorkbookView } from "../presentation/TimelineWorkbookView";
import { useTimelineWorkbookPresentation } from "../presentation/useTimelineWorkbookPresentation";

export type TimelineWorkbookProps = {
  readonly runtime: TimelineWorkbookSurfaceRuntime;
};

function TimelineWorkbookContent({
  runtime,
}: {
  readonly runtime: TimelineWorkbookSurfaceRuntime;
}) {
  const composition = useTimelineWorkbookComposition({ runtime });
  const presentation = useTimelineWorkbookPresentation({
    composition: composition.presentation,
    runtime: {
      entities: {
        hosts: runtime.entities.hosts,
        identities: runtime.entities.identities,
        index: runtime.entities.index,
      },
      indicatorWorkflow: runtime.indicatorWorkflow,
      layout: runtime.layout,
      queryControls: {
        renderInlineControls: runtime.query.renderInlineControls,
        savedViewSelector: runtime.query.savedViewSelector,
        viewBarQueryControls: runtime.query.viewBarQueryControls,
      },
    },
  });

  return <TimelineWorkbookView model={presentation} />;
}

export function TimelineWorkbook({ runtime }: TimelineWorkbookProps) {
  return (
    <TimelineCollaborationBoundary
      apiBase={runtime.incident.apiBase}
      attachSession={runtime.attachCollaborationSession}
      incidentId={runtime.incident.id}
      projection={runtime.collaborationProjection}
      sheetRef={runtime.incident.sheetRef}
    >
      <TimelineWorkbookContent key={runtime.incident.id} runtime={runtime} />
    </TimelineCollaborationBoundary>
  );
}
