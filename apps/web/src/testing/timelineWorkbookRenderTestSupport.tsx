import { type RenderResult, render } from "@testing-library/react";
import type { ComponentProps } from "react";
import { TimelineWorkbookRuntimeFixture } from "./TimelineWorkbookRuntimeFixture";

type TimelineWorkbookProps = ComponentProps<
  typeof TimelineWorkbookRuntimeFixture
>;

export function renderTimelineWorkbook(
  props: Partial<TimelineWorkbookProps> = {},
): RenderResult {
  return render(<TimelineWorkbookRuntimeFixture {...props} />);
}
