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
    let warningActive = false;
    const warnBeforeUnload = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    const updateWarning = () => {
      const state = runtime.getSnapshot();
      const required =
        state.queuedCount + state.inFlightCount > 0 ||
        state.unresolvedConflictCount > 0;
      if (required === warningActive) return;
      warningActive = required;
      if (required) window.addEventListener("beforeunload", warnBeforeUnload);
      else window.removeEventListener("beforeunload", warnBeforeUnload);
    };
    const unsubscribe = runtime.subscribe(updateWarning);
    updateWarning();
    return () => {
      unsubscribe();
      window.removeEventListener("beforeunload", warnBeforeUnload);
    };
  }, [runtime]);
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
