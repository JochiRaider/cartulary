import { createContext } from "react";
import type { WorkbookEvidenceAttachmentOwner } from "./WorkbookEvidenceAttachmentOwner";
export const EvidenceAttachmentContext =
  createContext<WorkbookEvidenceAttachmentOwner | null>(null);

import type { WorkbookTimelineFileOwner } from "./WorkbookTimelineFileOwner";
export const TimelineFileContext =
  createContext<WorkbookTimelineFileOwner | null>(null);
