declare const workbookContinuityTokenBrand: unique symbol;

export type WorkbookContinuityToken = string & {
  readonly [workbookContinuityTokenBrand]: true;
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
  readonly focus: (anchor: WorkbookContinuityAnchor) => boolean;
  readonly restore: (
    anchor: WorkbookContinuityAnchor | null,
    driverSnapshot: unknown,
  ) => boolean;
  readonly select: (anchor: WorkbookContinuityAnchor | null) => void;
};

export type WorkbookContinuityPort = {
  readonly capture: (
    anchor?: WorkbookContinuityAnchor | null,
  ) => WorkbookContinuityToken;
  readonly focus: (anchor: WorkbookContinuityAnchor) => boolean;
  readonly select: (anchor: WorkbookContinuityAnchor | null) => void;
  readonly clear: () => void;
  readonly restore: (token: WorkbookContinuityToken) => boolean;
  readonly snapshot: () => WorkbookContinuitySnapshot;
  readonly dispose: () => void;
};

export function createWorkbookContinuityPort(
  driver: WorkbookContinuityDriver,
): WorkbookContinuityPort {
  let disposed = false;
  let nextToken = 1;
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
    selectedAnchor = anchor;
    driver.select(anchor);
  };

  return {
    capture: (anchor = selectedAnchor) => {
      if (disposed) {
        throw new Error("Workbook continuity port is disposed.");
      }
      const token =
        `workbook-continuity-${nextToken}` as WorkbookContinuityToken;
      nextToken += 1;
      captures.clear();
      captures.set(token, {
        anchor,
        driverSnapshot: driver.capture(anchor),
      });
      return token;
    },
    focus: (anchor) => !disposed && driver.focus(anchor),
    select,
    clear: () => {
      if (disposed) {
        return;
      }
      captures.clear();
      selectedAnchor = null;
      driver.select(null);
    },
    restore: (token) => {
      if (disposed) {
        return false;
      }
      const capture = captures.get(token);
      if (capture === undefined) {
        return false;
      }
      captures.delete(token);
      const restored = driver.restore(capture.anchor, capture.driverSnapshot);
      if (restored && capture.anchor !== null) {
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
