import { useEffect, useState } from "react";
import {
  type WorkbookInspectorNotice,
  workbookInspectorNoticeIdentity,
} from "./workbookInspectorErrorModel";

/** The owner admits emission; local state only drives the live-region transition. */
export function useWorkbookInspectorNotice(
  notice: WorkbookInspectorNotice,
  consume: (notice: WorkbookInspectorNotice) => boolean,
) {
  const [emission, setEmission] = useState<WorkbookInspectorNotice | null>(
    null,
  );
  useEffect(() => {
    if (consume(notice)) setEmission(notice);
  }, [notice, consume]);
  return emission &&
    notice.announcement !== "none" &&
    workbookInspectorNoticeIdentity(emission) ===
      workbookInspectorNoticeIdentity(notice)
    ? emission
    : null;
}
