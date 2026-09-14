export type WorkbookContinuityToken = {
  readonly sequence: number;
};

export type WorkbookContinuityAnchor = {
  readonly viewSchemaId: string;
  readonly recordId: string;
  readonly fieldKey: string;
};

export type WorkbookContinuitySnapshot = {
  readonly anchor: WorkbookContinuityAnchor | null;
};

type WorkbookContinuityDriver = {
  readonly capture: (anchor: WorkbookContinuityAnchor | null) => unknown;
  readonly focus: (
    anchor: WorkbookContinuityAnchor,
    signal: AbortSignal,
  ) => Promise<boolean>;
  readonly restore: (
    anchor: WorkbookContinuityAnchor | null,
    driverSnapshot: unknown,
    signal: AbortSignal,
  ) => Promise<boolean>;
  readonly select: (anchor: WorkbookContinuityAnchor | null) => void;
};

export type WorkbookContinuityPort = {
  readonly capture: (
    anchor?: WorkbookContinuityAnchor | null,
  ) => WorkbookContinuityToken;
  readonly focus: (anchor: WorkbookContinuityAnchor) => Promise<boolean>;
  readonly select: (anchor: WorkbookContinuityAnchor | null) => void;
  readonly clear: () => void;
  readonly restore: (token: WorkbookContinuityToken) => Promise<boolean>;
  readonly snapshot: () => WorkbookContinuitySnapshot;
  readonly dispose: () => void;
};

export function createWorkbookContinuityPort(
  driver: WorkbookContinuityDriver,
): WorkbookContinuityPort {
  let disposed = false;
  let nextToken = 1;
  let pending: {
    controller: AbortController;
    anchor: WorkbookContinuityAnchor | null;
  } | null = null;
  const cancel = () => {
    pending?.controller.abort();
    pending = null;
  };
  const begin = (anchor: WorkbookContinuityAnchor | null) => {
    cancel();
    const request = { controller: new AbortController(), anchor };
    pending = request;
    return request;
  };
  let selectedAnchor: WorkbookContinuityAnchor | null = null;
  const captures = new Map<
    WorkbookContinuityToken,
    {
      readonly anchor: WorkbookContinuityAnchor | null;
      readonly driverSnapshot: unknown;
    }
  >();

  const select = (anchor: WorkbookContinuityAnchor | null) => {
    if (disposed || continuityAnchorsEqual(selectedAnchor, anchor)) {
      return;
    }
    if (pending !== null && !continuityAnchorsEqual(pending.anchor, anchor))
      cancel();
    selectedAnchor = anchor;
    driver.select(anchor);
  };

  return {
    capture: (anchor = selectedAnchor) => {
      if (disposed) {
        throw new Error("Workbook continuity port is disposed.");
      }
      cancel();
      const token: WorkbookContinuityToken = { sequence: nextToken };
      nextToken += 1;
      captures.clear();
      captures.set(token, {
        anchor,
        driverSnapshot: driver.capture(anchor),
      });
      return token;
    },
    focus: async (anchor) => {
      if (disposed) return false;
      const request = begin(anchor);
      const focused = await driver.focus(anchor, request.controller.signal);
      const current = pending === request && !request.controller.signal.aborted;
      if (pending === request) pending = null;
      return current && focused;
    },
    select,
    clear: () => {
      if (disposed) {
        return;
      }
      cancel();
      captures.clear();
      selectedAnchor = null;
      driver.select(null);
    },
    restore: async (token) => {
      if (disposed) {
        return false;
      }
      const capture = captures.get(token);
      if (capture === undefined) {
        return false;
      }
      captures.delete(token);
      const request = begin(capture.anchor);
      const restored = await driver.restore(
        capture.anchor,
        capture.driverSnapshot,
        request.controller.signal,
      );
      if (pending !== request || request.controller.signal.aborted)
        return false;
      pending = null;
      if (!disposed && restored && capture.anchor !== null) {
        select(capture.anchor);
      }
      return restored;
    },
    snapshot: () => ({ anchor: selectedAnchor }),
    dispose: () => {
      if (disposed) {
        return;
      }
      disposed = true;
      cancel();
      captures.clear();
      selectedAnchor = null;
      driver.select(null);
    },
  };
}

function continuityAnchorsEqual(
  left: WorkbookContinuityAnchor | null,
  right: WorkbookContinuityAnchor | null,
): boolean {
  return (
    left === right ||
    (left !== null &&
      right !== null &&
      left.viewSchemaId === right.viewSchemaId &&
      left.recordId === right.recordId &&
      left.fieldKey === right.fieldKey)
  );
}
