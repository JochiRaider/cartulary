import { useEffect, useState } from "react";
import type {
  WorkbookMutationRuntime,
  WorkbookSaveAnnouncement,
} from "../runtime/WorkbookMutationRuntime";
import { visuallyHiddenStyle } from "../utils/workbookStyles";

/** One shell-lifetime host, separate from visible status and inert surface content. */
export function WorkbookSaveAnnouncements({
  runtime,
}: {
  readonly runtime: WorkbookMutationRuntime;
}) {
  const [delivered, setDelivered] = useState<{
    readonly runtime: WorkbookMutationRuntime;
    readonly event: WorkbookSaveAnnouncement;
  } | null>(null);
  const announcement = delivered?.runtime === runtime ? delivered.event : null;
  useEffect(() => {
    const announce = () => {
      const event = runtime.takeSaveAnnouncement();
      if (event !== null) setDelivered({ runtime, event });
    };
    const unsubscribe = runtime.subscribe(announce);
    announce();
    return unsubscribe;
  }, [runtime]);
  return (
    <div style={visuallyHiddenStyle}>
      <span
        aria-label="Workbook save updates"
        aria-live="polite"
        aria-atomic="true"
        role="status"
      >
        {announcement?.priority === "polite" ? announcement.message : ""}
      </span>
      <span
        aria-label="Workbook save conflicts"
        aria-live="assertive"
        aria-atomic="true"
        role="alert"
      >
        {announcement?.priority === "assertive" ? announcement.message : ""}
      </span>
    </div>
  );
}
