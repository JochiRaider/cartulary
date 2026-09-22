import type { Page } from "@playwright/test";

// The browser probe observes only this public capability, keeping application
// and CSS modules out of Playwright’s Node-side module graph.
type GridFocusTarget =
  | {
      kind: "cell";
      anchor: {
        rowIdentity: { kind: string; recordId?: string };
        fieldKey: string;
      };
    }
  | { kind: "draft"; fieldKey: string }
  | { kind: "root" };
type GridHandle = {
  requestFocus: (
    target: GridFocusTarget,
    options?: { signal?: AbortSignal; preserveSelection?: boolean },
  ) => Promise<string>;
  getScrollElement: () => HTMLElement | null;
};

type FocusObservation = {
  target: GridFocusTarget;
  result: string | null;
  signal: AbortSignal | undefined;
};
type ContinuityProbe = {
  observed: FocusObservation[];
  handles: number;
  destination: Element | null;
  hold: (recordId: string, fieldKey: string) => void;
  release: () => void;
};
declare global {
  interface Window {
    timelineContinuityProbe: ContinuityProbe;
  }
}

/** Browser-only observation and a connected-target gate; no runtime test branch. */
export async function installDeferredContinuityProbe(page: Page) {
  await page.addInitScript(() => {
    const observed: FocusObservation[] = [];
    const refs = new WeakSet<object>();
    const handles = new WeakSet<object>();
    let held: { recordId: string; fieldKey: string } | null = null;
    const connected = Object.getOwnPropertyDescriptor(
      Node.prototype,
      "isConnected",
    )?.get;
    if (!connected) throw new Error("Missing native connection getter");
    // Delay only the semantic target's connected-readiness observation. Removing
    // React-owned DOM would corrupt the renderer rather than delay registration.
    Object.defineProperty(HTMLElement.prototype, "isConnected", {
      configurable: true,
      get(this: HTMLElement) {
        const target = held;
        if (
          target &&
          this.matches("[role='gridcell']") &&
          this.closest<HTMLElement>("[data-grid-record-id]")?.dataset
            .gridRecordId === target.recordId &&
          (this.dataset.gridFieldKey === target.fieldKey ||
            [
              ...this.querySelectorAll<HTMLElement>("[data-grid-field-key]"),
            ].some((field) => field.dataset.gridFieldKey === target.fieldKey))
        )
          return false;
        return connected.call(this);
      },
    });
    window.timelineContinuityProbe = {
      observed,
      handles: 0,
      destination: null,
      hold: (recordId, fieldKey) => {
        held = { recordId, fieldKey };
      },
      release: () => {
        held = null;
      },
    };
    const wrap = (handle: GridHandle | null) => {
      if (!handle || handles.has(handle)) return;
      handles.add(handle);
      window.timelineContinuityProbe.handles++;
      const request = handle.requestFocus;
      Object.defineProperty(handle, "requestFocus", {
        configurable: true,
        value: (
          target: GridFocusTarget,
          options?: Parameters<GridHandle["requestFocus"]>[1],
        ) => {
          const observation: FocusObservation = {
            target,
            signal: options?.signal,
            result: null,
          };
          observed.push(observation);
          const promise = request(target, options);
          void promise.then((result) => {
            observation.result = result;
          });
          return promise;
        },
      });
    };
    type Hook = { memoizedState?: unknown; next?: Hook };
    type Fiber = { memoizedState?: Hook; child?: Fiber; sibling?: Fiber };
    // Observe public semantic capabilities through React's development hook,
    // without component-name selectors or invoking feature/mutation internals.
    const hooked = window as unknown as {
      __REACT_DEVTOOLS_GLOBAL_HOOK__: unknown;
    };
    hooked.__REACT_DEVTOOLS_GLOBAL_HOOK__ = {
      supportsFiber: true,
      inject: () => 1,
      onCommitFiberUnmount: () => {},
      onCommitFiberRoot: (_renderer: number, root: { current: Fiber }) => {
        const visit = (fiber?: Fiber) => {
          if (!fiber) return;
          let hook = fiber.memoizedState;
          while (hook) {
            const ref = hook.memoizedState as {
              current?: GridHandle | null;
            } | null;
            if (
              ref &&
              typeof ref === "object" &&
              ref.current &&
              typeof ref.current.requestFocus === "function" &&
              typeof ref.current.getScrollElement === "function" &&
              !refs.has(ref)
            ) {
              refs.add(ref);
              let current: GridHandle | null = ref.current;
              wrap(current);
              Object.defineProperty(ref, "current", {
                configurable: true,
                get: () => current,
                set: (value: GridHandle | null) => {
                  current = value;
                  wrap(value);
                },
              });
            }
            hook = hook.next;
          }
          visit(fiber.child);
          visit(fiber.sibling);
        };
        visit(root.current);
      },
    };
  });
}

export async function pendingContinuity(page: Page, recordId: string) {
  return page.evaluate(
    (id) =>
      window.timelineContinuityProbe.observed
        .filter(
          (item) =>
            item.target.kind === "cell" &&
            item.target.anchor.rowIdentity.kind === "core_record" &&
            item.target.anchor.rowIdentity.recordId === id &&
            item.signal,
        )
        .map((item) => ({
          result: item.result,
          aborted: item.signal?.aborted,
        })),
    recordId,
  );
}
