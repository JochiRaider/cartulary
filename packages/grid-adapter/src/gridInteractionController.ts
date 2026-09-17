import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import type { GridCellAnchor, GridCellRange } from "./core";
import type { ActiveEditorSession } from "./editorSessionPolicy";
import {
  type GridSemanticCoordinateModel,
  retainGridCellRange,
  sameGridCellAnchor,
  sameGridCellRange,
  semanticPresentationContainsAnchor,
} from "./semanticPresentation";
import { extendSemanticCellRange } from "./semanticSelectionPolicy";

const limits = cartularyDesignPresentation.gridCellRangeSelection;

export type GridPointerPosition = { readonly x: number; readonly y: number };
export type GridPointerInput = GridPointerPosition & {
  readonly pointerId: number;
  readonly shiftKey: boolean;
};

export type GridInteractionSnapshot = {
  readonly active: GridCellAnchor | null;
  readonly editor: ActiveEditorSession | null;
  readonly model: GridSemanticCoordinateModel;
  readonly range: GridCellRange | null;
  readonly enabled: boolean;
  readonly available: boolean;
  readonly scopeKey: string;
  readonly authorityKey: string;
};

export type GridInteractionDriver = {
  readonly read: () => GridInteractionSnapshot;
  readonly accept: (
    range: GridCellRange,
    edit: boolean,
    inspect: boolean,
  ) => void;
  readonly changeRange: (range: GridCellRange | null) => void;
  readonly preview: (range: GridCellRange | null) => void;
  readonly announce: (message: string) => void;
  readonly stopped: () => void;
};

type Gesture = {
  readonly origin: GridCellAnchor;
  readonly start: GridCellAnchor;
  readonly input: GridPointerInput;
  readonly snapshot: GridInteractionSnapshot;
  end: GridCellAnchor;
  moved: boolean;
  extending: boolean;
  pending: boolean;
  released: boolean;
};

/** Owns intentions only. Editor drafts and authoritative write outcomes stay source-owned. */
export function createGridInteractionController(driver: GridInteractionDriver) {
  let gesture: Gesture | null = null;
  let previous: GridInteractionSnapshot | null = null;

  const isRange = (current: Gesture) =>
    current.snapshot.enabled && (current.extending || current.moved);
  const rangeOf = (current: Gesture): GridCellRange => ({
    start: current.start,
    end: current.end,
  });
  const clear = () => {
    gesture = null;
    driver.stopped();
    driver.preview(null);
    driver.announce("");
  };
  const cancel = (announce = false) => {
    const current = gesture;
    if (current === null) return false;

    clear();
    if (announce && isRange(current)) driver.announce("Selection canceled.");
    return true;
  };
  const valid = (current: Gesture) => {
    const latest = driver.read();
    const captured = current.snapshot.model;
    const firstRow = captured.rowIdentities[0];
    const lastRow = captured.rowIdentities.at(-1);
    const firstField = captured.fieldKeys[0];
    const lastField = captured.fieldKeys.at(-1);
    if (
      !latest.available ||
      latest.enabled !== current.snapshot.enabled ||
      !sameGridCellRange(latest.range, current.snapshot.range) ||
      latest.scopeKey !== current.snapshot.scopeKey ||
      latest.authorityKey !== current.snapshot.authorityKey ||
      !firstRow ||
      !lastRow ||
      !firstField ||
      !lastField
    )
      return false;
    // The loaded membership is frozen for this gesture; appends outside it are safe.
    return (
      retainGridCellRange(captured, latest.model, {
        start: {
          surface: captured.surface,
          rowIdentity: firstRow,
          fieldKey: firstField,
        },
        end: {
          surface: captured.surface,
          rowIdentity: lastRow,
          fieldKey: lastField,
        },
      }) !== null &&
      semanticPresentationContainsAnchor(latest.model, current.end)
    );
  };
  const publish = (current: Gesture) => {
    if (gesture !== current || current.pending || !current.released) return;
    if (!valid(current)) {
      cancel();
      return;
    }
    const range = rangeOf(current);
    const selecting = isRange(current);
    const edit = !selecting && !current.moved;
    clear();
    if (selecting || edit) driver.accept(range, edit, !selecting);
  };
  const showPreview = (current: Gesture) => {
    driver.preview(isRange(current) ? rangeOf(current) : null);
  };

  return {
    cancel,
    enterNativeEditor() {
      cancel();
      driver.changeRange(null);
    },
    get pointerId() {
      return gesture?.input.pointerId ?? null;
    },
    get active() {
      return gesture !== null;
    },
    get scrolling() {
      return (
        gesture !== null &&
        isRange(gesture) &&
        gesture.moved &&
        !gesture.pending &&
        !gesture.released
      );
    },
    get capturedModel() {
      return gesture?.snapshot.model ?? null;
    },
    begin(origin: GridCellAnchor, input: GridPointerInput): boolean {
      cancel();
      const snapshot = driver.read();
      if (
        !snapshot.available ||
        !semanticPresentationContainsAnchor(snapshot.model, origin)
      )
        return false;
      if (input.shiftKey && !snapshot.enabled) return false;
      if (snapshot.editor && sameGridCellAnchor(snapshot.editor.target, origin))
        return false;
      const range = input.shiftKey
        ? extendSemanticCellRange(
            snapshot.model,
            snapshot.active,
            snapshot.range,
            origin,
          )
        : { start: origin, end: origin };
      if (range === null) return false;
      const current: Gesture = {
        origin,
        start: range.start,
        end: origin,
        input,
        snapshot: {
          ...snapshot,
          model: {
            ...snapshot.model,
            fieldKeys: [...snapshot.model.fieldKeys],
            rowIdentities: [...snapshot.model.rowIdentities],
          },
        },
        moved: false,
        extending: input.shiftKey,
        pending: snapshot.editor !== null,
        released: false,
      };
      gesture = current;
      showPreview(current);
      if (snapshot.editor !== null) {
        if (isRange(current)) driver.announce("Waiting for the edit to save.");
        const session = snapshot.editor;
        void session.requestCommit().then((accepted) => {
          if (gesture !== current) return;
          if (!accepted) {
            cancel();
            // A source may detach on authority loss; never focus that old attachment.
            const latest = driver.read();
            if (
              latest.available &&
              latest.editor &&
              sameGridCellAnchor(latest.editor.target, session.target)
            )
              latest.editor.focus();
            return;
          }
          current.pending = false;
          driver.announce("");
          publish(current);
        });
      }
      return true;
    },
    track(point: GridPointerPosition) {
      const current = gesture;
      if (current === null || current.released) return;
      if (crossesGridDragThreshold(current.input, point) && !current.moved) {
        current.moved = true;
        if (current.pending && !current.extending && current.snapshot.enabled)
          driver.announce("Waiting for the edit to save.");
        showPreview(current);
      }
    },
    move(point: GridPointerPosition, destination: GridCellAnchor | null) {
      this.track(point);
      const current = gesture;
      if (current === null || current.released) return;
      if (!valid(current)) {
        cancel();
        return;
      }
      if (
        destination !== null &&
        isRange(current) &&
        semanticPresentationContainsAnchor(current.snapshot.model, destination)
      )
        current.end = destination;
      showPreview(current);
    },
    release(point: GridPointerPosition, destination: GridCellAnchor | null) {
      this.move(point, destination);
      const current = gesture;
      if (current === null) return;
      current.released = true;
      driver.stopped();
      if (!current.moved && !sameGridCellAnchor(current.origin, destination)) {
        cancel();
        return;
      }
      publish(current);
    },
    reconcile() {
      const latest = driver.read();
      if (gesture !== null && !valid(gesture)) cancel();
      const before = previous ?? latest;
      const retained =
        latest.scopeKey === before.scopeKey &&
        latest.authorityKey === before.authorityKey &&
        latest.available
          ? retainGridCellRange(before.model, latest.model, latest.range)
          : null;
      if (retained !== latest.range) driver.changeRange(retained);
      previous = latest;
    },
    dispose() {
      cancel();
    },
  };
}

export function crossesGridDragThreshold(
  origin: GridPointerPosition,
  point: GridPointerPosition,
) {
  return (
    Math.abs(point.x - origin.x) > limits.stationaryTolerancePx ||
    Math.abs(point.y - origin.y) > limits.stationaryTolerancePx
  );
}

export function gridEdgeScrollDelta(
  position: number,
  minimum: number,
  maximum: number,
  elapsedMs: number,
) {
  if (maximum <= minimum || !Number.isFinite(elapsedMs)) return 0;
  const band = Math.min(limits.edgeBandPx, (maximum - minimum) / 2);
  const proximity =
    position < minimum + band
      ? -Math.min(1, (minimum + band - position) / band)
      : position > maximum - band
        ? Math.min(1, (position - maximum + band) / band)
        : 0;
  return (
    (proximity *
      limits.maximumScrollPxPerSecond *
      Math.max(0, Math.min(limits.maximumFrameMs, elapsedMs))) /
    1000
  );
}
