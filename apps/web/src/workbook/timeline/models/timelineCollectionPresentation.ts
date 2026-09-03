import type { WorkbookRelationshipChipPresentation } from "../../models/workbookRelationshipChip";
import type { TimelineCollectionBinding } from "./timelineFieldRegistry";
import type { TagCollectionItem, WorkbookRow } from "./timelineRowModel";

type TimelineCollectionEntityIndex = Readonly<
  Record<string, { readonly label: string }>
>;

import {
  type CollectionItem,
  timelineRelationshipChipPresentation,
} from "./workbookMentionChips";

export type TimelineRelationshipPresentationItem = {
  readonly kind: "relationship";
  readonly itemRef: string;
  readonly chip: WorkbookRelationshipChipPresentation;
};

export type TimelineTagPresentationItem = {
  readonly kind: "tag";
  readonly itemRef: string;
  readonly displayText: string;
};

export type TimelineCollectionPresentation =
  | {
      readonly kind: "relationship";
      readonly visibleItems: readonly TimelineRelationshipPresentationItem[];
      readonly hiddenItemCount: number;
      readonly hiddenLabels: readonly string[];
      readonly firstHiddenItemRef: string | null;
      readonly overflowRecordId: string | null;
    }
  | {
      readonly kind: "tag";
      readonly visibleItems: readonly TimelineTagPresentationItem[];
      readonly hiddenItemCount: number;
      readonly hiddenLabels: readonly string[];
      readonly firstHiddenItemRef: null;
      readonly overflowRecordId: null;
    };

function relationshipPresentationItem(
  item: CollectionItem,
  entityIndex: TimelineCollectionEntityIndex,
): TimelineRelationshipPresentationItem {
  return {
    kind: "relationship",
    itemRef: item.itemRef,
    chip: timelineRelationshipChipPresentation({ entityIndex, item }),
  };
}

function tagPresentationItem(
  item: TagCollectionItem,
): TimelineTagPresentationItem {
  return {
    kind: "tag",
    itemRef: item.itemRef,
    displayText: item.displayText,
  };
}

export function projectTimelineCollectionPresentation({
  binding,
  entityIndex,
  row,
}: {
  readonly binding: TimelineCollectionBinding;
  readonly entityIndex: TimelineCollectionEntityIndex;
  readonly row: WorkbookRow;
}): TimelineCollectionPresentation {
  if (binding.collectionKind === "relationship") {
    const items = row.collectionValues[binding.draftKey].map((item) =>
      relationshipPresentationItem(item, entityIndex),
    );
    return {
      kind: "relationship",
      visibleItems: items.slice(0, 1),
      hiddenItemCount: Math.max(0, items.length - 1),
      hiddenLabels: items.slice(1).map((item) => item.chip.label),
      firstHiddenItemRef: items[1]?.itemRef ?? null,
      overflowRecordId: row.recordId,
    };
  }
  const items = row.collectionValues.tags.map(tagPresentationItem);
  return {
    kind: "tag",
    visibleItems: items.slice(0, 1),
    hiddenItemCount: Math.max(0, items.length - 1),
    hiddenLabels: items.slice(1).map((item) => item.displayText),
    firstHiddenItemRef: null,
    overflowRecordId: null,
  };
}
