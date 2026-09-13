import { createContext } from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookTimelineRelatedEvidenceOwner } from "./WorkbookTimelineRelatedEvidenceOwner";

export const TimelineRelatedEvidenceContext = createContext<{
  readonly owner: WorkbookTimelineRelatedEvidenceOwner;
  readonly sheetRef: SheetRef;
} | null>(null);
