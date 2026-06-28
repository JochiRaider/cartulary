import { type RenderResult, render } from "@testing-library/react";
import type { ComponentProps } from "react";
import { TimelineWorkbook } from "../workbook/timeline/components/TimelineWorkbook";

type TimelineWorkbookProps = ComponentProps<typeof TimelineWorkbook>;

export function renderTimelineWorkbook(
  props: Partial<TimelineWorkbookProps> = {},
): RenderResult {
  return render(<TimelineWorkbook incidentId="incident-1" {...props} />);
}
