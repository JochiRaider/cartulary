import type {
  GridNavigationIntent,
  GridNavigationKey,
} from "@cartulary/grid-adapter";
import {
  decideWorkbookApplicationShortcut,
  type WorkbookApplicationShortcutEvent,
} from "../../policies/workbookApplicationShortcuts";
import type {
  RowValues,
  TimelineScalarEditorSurface,
} from "./timelineFieldRegistry";
import type { WorkbookRow } from "./timelineRowModel";

export type TimelineEditorKeyboardIntent =
  | {
      readonly kind: "none";
      readonly preventDefault: boolean;
      readonly stopPropagation: boolean;
    }
  | {
      readonly kind: "restore_prior_grid_focus";
      readonly preventDefault: true;
      readonly stopPropagation: true;
    }
  | {
      readonly kind: "close_inspector";
      readonly preventDefault: true;
      readonly stopPropagation: true;
    }
  | {
      readonly kind: "navigate";
      readonly navigation: GridNavigationIntent;
      readonly preventDefault: true;
      readonly stopPropagation: true;
    }
  | {
      readonly kind: "save";
      readonly navigateAfterSave: GridNavigationIntent | null;
      readonly preserveInputFocus: boolean;
      readonly preventDefault: true;
      readonly recordBlankRowTiming: boolean;
      readonly stopPropagation: true;
    };

export type TimelineKeyboardEvent = WorkbookApplicationShortcutEvent;

const timelineEditorNavigationKeys = new Set<string>([
  "ArrowDown",
  "ArrowLeft",
  "ArrowRight",
  "ArrowUp",
  "Enter",
  "Tab",
]);

const noTimelineEditorIntent = {
  kind: "none",
  preventDefault: false,
  stopPropagation: false,
} as const;

function timelineEditorNavigationIntent(
  event: TimelineKeyboardEvent,
): GridNavigationIntent | null {
  if (
    event.altKey === true ||
    event.ctrlKey === true ||
    event.metaKey === true ||
    !isTimelineEditorNavigationKey(event.key)
  ) {
    return null;
  }
  return { key: event.key, shiftKey: event.shiftKey === true };
}

function isTimelineEditorNavigationKey(key: string): key is GridNavigationKey {
  return timelineEditorNavigationKeys.has(key);
}

export function mapTimelineScalarEditorIntent({
  event,
  focusField,
  hasCommittedAnchor,
  inspectorCanClose,
  priorTimelineGridAnchor,
  surface,
}: {
  readonly event: TimelineKeyboardEvent;
  readonly focusField: keyof RowValues;
  readonly hasCommittedAnchor: boolean;
  readonly inspectorCanClose: boolean;
  readonly priorTimelineGridAnchor: boolean;
  readonly surface: TimelineScalarEditorSurface;
}): TimelineEditorKeyboardIntent {
  if (
    surface === "inspector" &&
    event.key === "Escape" &&
    priorTimelineGridAnchor
  ) {
    return {
      kind: "restore_prior_grid_focus",
      preventDefault: true,
      stopPropagation: true,
    };
  }
  const applicationShortcut = decideWorkbookApplicationShortcut(event, {
    capabilities: {
      closeInspector: hasCommittedAnchor && inspectorCanClose,
      history: false,
      linkedEvidence: false,
      quickLink: false,
    },
    focusOwner: "editor",
    previewableEvidenceCount: 0,
    rowKind: hasCommittedAnchor ? "committed" : "draft",
    selectionIdentity: hasCommittedAnchor ? "timeline-editor-row" : null,
  });
  if (applicationShortcut.kind === "close_inspector") {
    return {
      kind: "close_inspector",
      preventDefault: true,
      stopPropagation: true,
    };
  }
  const navigation = timelineEditorNavigationIntent(event);
  if (navigation === null) return noTimelineEditorIntent;
  if (
    surface === "grid" &&
    navigation.shiftKey &&
    navigation.key.startsWith("Arrow")
  ) {
    return {
      kind: "none",
      preventDefault: true,
      stopPropagation: true,
    };
  }
  if (!hasCommittedAnchor) {
    return navigation.key === "Enter" || navigation.key === "Tab"
      ? {
          kind: "save",
          navigateAfterSave: null,
          preserveInputFocus: true,
          preventDefault: true,
          recordBlankRowTiming:
            navigation.key === "Enter" &&
            focusField === "activitySynopsisText" &&
            surface === "grid",
          stopPropagation: true,
        }
      : {
          kind: "none",
          preventDefault: true,
          stopPropagation: true,
        };
  }
  if (surface === "grid" && navigation.key === "Tab") {
    return {
      kind: "save",
      navigateAfterSave: null,
      preserveInputFocus: false,
      preventDefault: true,
      recordBlankRowTiming: false,
      stopPropagation: true,
    };
  }
  return navigation.key === "Enter" || navigation.key === "Tab"
    ? {
        kind: "save",
        navigateAfterSave: navigation,
        preserveInputFocus: false,
        preventDefault: true,
        recordBlankRowTiming: false,
        stopPropagation: true,
      }
    : {
        kind: "navigate",
        navigation,
        preventDefault: true,
        stopPropagation: true,
      };
}

export function mapTimelineCollectionEditorIntent({
  event,
  hasCommittedAnchor,
  inspectorCanClose,
}: {
  readonly event: TimelineKeyboardEvent;
  readonly hasCommittedAnchor: boolean;
  readonly inspectorCanClose: boolean;
}): TimelineEditorKeyboardIntent {
  const applicationShortcut = decideWorkbookApplicationShortcut(event, {
    capabilities: {
      closeInspector: hasCommittedAnchor && inspectorCanClose,
      history: false,
      linkedEvidence: false,
      quickLink: false,
    },
    focusOwner: "editor",
    previewableEvidenceCount: 0,
    rowKind: hasCommittedAnchor ? "committed" : "draft",
    selectionIdentity: hasCommittedAnchor ? "timeline-editor-row" : null,
  });
  if (applicationShortcut.kind === "close_inspector") {
    return {
      kind: "close_inspector",
      preventDefault: true,
      stopPropagation: true,
    };
  }
  const navigation = timelineEditorNavigationIntent(event);
  if (navigation === null) return noTimelineEditorIntent;
  if (navigation.shiftKey && navigation.key.startsWith("Arrow")) {
    return {
      kind: "none",
      preventDefault: true,
      stopPropagation: true,
    };
  }
  if (!hasCommittedAnchor) {
    return navigation.key === "Enter" || navigation.key === "Tab"
      ? {
          kind: "save",
          navigateAfterSave: null,
          preserveInputFocus: false,
          preventDefault: true,
          recordBlankRowTiming: false,
          stopPropagation: true,
        }
      : {
          kind: "none",
          preventDefault: true,
          stopPropagation: true,
        };
  }
  if (navigation.key === "Tab") {
    return {
      kind: "save",
      navigateAfterSave: null,
      preserveInputFocus: false,
      preventDefault: true,
      recordBlankRowTiming: false,
      stopPropagation: true,
    };
  }
  return navigation.key === "Enter"
    ? {
        kind: "save",
        navigateAfterSave: navigation,
        preserveInputFocus: false,
        preventDefault: true,
        recordBlankRowTiming: false,
        stopPropagation: true,
      }
    : {
        kind: "navigate",
        navigation,
        preventDefault: true,
        stopPropagation: true,
      };
}

type TimelineKeyboardRow = {
  readonly collectionValues: {
    readonly hostRefs: readonly {
      readonly itemKind: string;
      readonly itemRef: string;
    }[];
    readonly identityRefs: readonly {
      readonly itemKind: string;
      readonly itemRef: string;
    }[];
  };
  readonly recordId: WorkbookRow["recordId"];
  readonly rowVersion: WorkbookRow["rowVersion"];
};

export type TimelineWorkAreaInspectorIntent<Row extends TimelineKeyboardRow> =
  | {
      readonly kind: "none";
      readonly preventDefault: false;
      readonly stopPropagation: false;
    }
  | {
      readonly kind: "open_panel";
      readonly panelId: "evidence" | "history";
      readonly preventDefault: true;
      readonly row: Row;
      readonly stopPropagation: true;
    }
  | {
      readonly itemRef: string | null;
      readonly kind: "quick_link";
      readonly preventDefault: true;
      readonly row: Row;
      readonly stopPropagation: true;
    };

export function mapTimelineWorkAreaInspectorIntent<
  Row extends TimelineKeyboardRow,
>({
  event,
  fieldKey,
  row,
}: {
  readonly event: TimelineKeyboardEvent;
  readonly fieldKey: string | undefined;
  readonly row: Row | null;
}): TimelineWorkAreaInspectorIntent<Row> {
  const recordId = row?.recordId ?? null;
  const command = decideWorkbookApplicationShortcut(event, {
    capabilities: {
      closeInspector: false,
      history: true,
      linkedEvidence: true,
      quickLink:
        fieldKey === "timeline.host_refs" ||
        fieldKey === "timeline.identity_refs",
    },
    focusOwner: "grid_navigation",
    previewableEvidenceCount: 0,
    rowKind: recordId === null ? "none" : "committed",
    selectionIdentity: recordId,
  });
  if (row === null || recordId === null) {
    return {
      kind: "none",
      preventDefault: false,
      stopPropagation: false,
    };
  }
  if (command.kind === "open_history") {
    return {
      kind: "open_panel",
      panelId: "history",
      preventDefault: true,
      row,
      stopPropagation: true,
    };
  }
  if (command.kind === "preview_linked_evidence") {
    return {
      kind: "open_panel",
      panelId: "evidence",
      preventDefault: true,
      row,
      stopPropagation: true,
    };
  }
  if (command.kind !== "quick_link") {
    return {
      kind: "none",
      preventDefault: false,
      stopPropagation: false,
    };
  }
  const mention = [
    ...row.collectionValues.hostRefs,
    ...row.collectionValues.identityRefs,
  ].find((item) => item.itemKind !== "resolved_ref");
  return {
    itemRef: mention?.itemRef ?? null,
    kind: "quick_link",
    preventDefault: true,
    row,
    stopPropagation: true,
  };
}
