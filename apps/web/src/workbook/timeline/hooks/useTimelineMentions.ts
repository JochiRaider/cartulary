import { useMemo, useState, useSyncExternalStore } from "react";
import type { WorkbookTimelineMentionOperationOwner } from "../actions/WorkbookTimelineMentionOperationOwner";
import type { DismissedMention } from "../models/workbookMentionChips";

export function useTimelineMentions(
  owner: WorkbookTimelineMentionOperationOwner,
) {
  const [selectedMentionRef, setSelectedMentionRef] = useState<string | null>(
    null,
  );
  const [selectedResolveTargetId, setSelectedResolveTargetId] = useState("");
  const retained = useSyncExternalStore(
    owner.subscribe,
    owner.getMentionsSnapshot,
  );
  const dismissedMentionsByRow = useMemo(() => {
    const result: Record<string, DismissedMention[]> = {};
    for (const subject of retained) {
      if (subject.state !== "dismissed") continue;
      result[subject.sourceRecordId] ??= [];
      result[subject.sourceRecordId]?.push({
        entityMentionId: subject.mentionId,
        rowRecordId: subject.sourceRecordId,
        fieldKey: subject.sourceFieldKey,
        entityType: subject.entityType,
        itemRef: subject.itemRef,
        rawText: subject.rawText,
        resolvedRecordId: null,
        mentionRowVersion: subject.mentionRowVersion,
        resolutionMethod: null,
        autoResolved: false,
      });
    }
    return result;
  }, [retained]);
  return {
    commands: {
      setSelectedMentionRef,
      setSelectedResolveTargetId,
    },
    snapshot: {
      dismissedMentionsByRow,
      observedMentions: retained,
      selectedMentionRef,
      selectedResolveTargetId,
    },
  };
}
