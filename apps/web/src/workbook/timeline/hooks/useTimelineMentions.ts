import { useState } from "react";
import type {
  AutoResolutionNotice,
  DismissedMention,
} from "../models/workbookMentionChips";

export function useTimelineMentions() {
  const [selectedMentionRef, setSelectedMentionRef] = useState<string | null>(
    null,
  );
  const [selectedResolveTargetId, setSelectedResolveTargetId] = useState("");
  const [dismissedMentionsByRow, setDismissedMentionsByRow] = useState<
    Record<string, DismissedMention[]>
  >({});
  const [autoResolutionNotices, setAutoResolutionNotices] = useState<
    AutoResolutionNotice[]
  >([]);

  return {
    commands: {
      setAutoResolutionNotices,
      setDismissedMentionsByRow,
      setSelectedMentionRef,
      setSelectedResolveTargetId,
    },
    snapshot: {
      autoResolutionNotices,
      dismissedMentionsByRow,
      selectedMentionRef,
      selectedResolveTargetId,
    },
  };
}
