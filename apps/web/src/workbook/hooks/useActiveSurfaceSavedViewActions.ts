import { useCallback } from "react";
import type {
  SavedViewIntent,
  SavedViewSubject,
} from "../savedviews/savedViewOperationModel";
import type { WorkbookSavedViewController } from "../savedviews/WorkbookSavedViewController";

export type SavedViewActionIntent = SavedViewIntent;

/** Presentation captures its subject; the operation owner admits or rejects it synchronously. */
export function useActiveSurfaceSavedViewActions(
  controller: WorkbookSavedViewController,
  subject: SavedViewSubject,
) {
  const runAction = useCallback(
    (intent: SavedViewIntent) => controller.run(intent, subject),
    [controller, subject],
  );
  return { runAction };
}
