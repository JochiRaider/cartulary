import {
  type RefCallback,
  type RefObject,
  useCallback,
  useLayoutEffect,
  useRef,
} from "react";
import {
  type WorkbookSessionNavigation,
  type WorkbookSessionPin,
  workbookPinIdentity,
} from "./WorkbookSessionNavigation";

type Removal = {
  readonly session: WorkbookSessionNavigation;
  readonly attemptId: number;
  readonly before: readonly string[];
  readonly expected: readonly string[];
  readonly identity: string;
  readonly source: HTMLButtonElement;
  readonly dock: HTMLElement;
  readonly dispose: () => void;
};

const samePins = (left: readonly string[], right: readonly string[]) =>
  left.length === right.length &&
  left.every((id, index) => id === right[index]);

/** Only the Work scrollport may move to reveal a recovered control and its ring. */
function reveal(dock: HTMLElement, target: HTMLButtonElement) {
  const control = target.getBoundingClientRect();
  const bounds = dock.getBoundingClientRect();
  const style = getComputedStyle(dock);
  const scale = dock.offsetHeight ? bounds.height / dock.offsetHeight : 1;
  const top =
    bounds.top +
    (dock.clientTop + Number.parseFloat(style.paddingTop || "0")) * scale;
  const bottom =
    bounds.top +
    (dock.clientTop +
      dock.clientHeight -
      Number.parseFloat(style.paddingBottom || "0")) *
      scale;
  if (control.top < top) dock.scrollTop += (control.top - top) / scale;
  else if (control.bottom > bottom)
    dock.scrollTop += (control.bottom - bottom) / scale;
}

/** Work owns only this short-lived focus intention; the session owns all pins. */
export function useWorkPinRemovalFocus(
  session: WorkbookSessionNavigation | undefined,
  dockRef: RefObject<HTMLElement | null>,
  pinViewRef: RefObject<HTMLButtonElement | null>,
) {
  const buttons = useRef(new Map<string, HTMLButtonElement>());
  const pending = useRef<Removal | null>(null);
  const clearIndicator = useRef<(() => void) | null>(null);
  const cancel = useCallback(() => {
    pending.current?.dispose();
    pending.current = null;
  }, []);

  // biome-ignore lint/correctness/useExhaustiveDependencies: Replacing the session retires this attachment's focus intention and indicator.
  useLayoutEffect(
    () => () => {
      cancel();
      clearIndicator.current?.();
    },
    [session, cancel],
  );

  useLayoutEffect(() => {
    const intent = pending.current;
    if (!intent) return;
    const snapshot = session?.getSnapshot();
    const ids = snapshot?.pins.map(workbookPinIdentity) ?? [];
    if (
      session !== intent.session ||
      !snapshot?.readable ||
      snapshot.attemptId !== intent.attemptId ||
      dockRef.current !== intent.dock ||
      !intent.dock.isConnected ||
      (document.activeElement !== intent.source &&
        document.activeElement !== document.body)
    ) {
      cancel();
      return;
    }
    if (samePins(ids, intent.before)) return;
    if (!samePins(ids, intent.expected)) {
      cancel();
      return;
    }
    const index = intent.before.indexOf(intent.identity);
    const candidates = [
      ...intent.before.slice(index + 1),
      ...intent.before.slice(0, index).reverse(),
    ];
    const eligible = (button: HTMLButtonElement | null | undefined) =>
      button?.isConnected && !button.disabled && intent.dock.contains(button);
    const next = candidates.map((id) => buttons.current.get(id)).find(eligible);
    const target =
      next ?? (eligible(pinViewRef.current) ? pinViewRef.current : null);
    cancel();
    if (!target) return;
    clearIndicator.current?.();
    target.setAttribute("data-work-pin-removal-focus", "");
    const clear = () => {
      target.removeAttribute("data-work-pin-removal-focus");
      target.removeEventListener("blur", clear);
      clearIndicator.current = null;
    };
    clearIndicator.current = clear;
    target.addEventListener("blur", clear, { once: true });
    target.focus({ preventScroll: true });
    if (document.activeElement === target) reveal(intent.dock, target);
    else clear();
  });

  const buttonRef =
    (identity: string): RefCallback<HTMLButtonElement> =>
    (button) => {
      if (button) buttons.current.set(identity, button);
      else buttons.current.delete(identity);
    };
  const remove = (pin: WorkbookSessionPin, source: HTMLButtonElement) => {
    cancel();
    if (!session) return;
    const snapshot = session.getSnapshot();
    const dock = dockRef.current;
    const identity = workbookPinIdentity(pin);
    const before = snapshot.pins.map(workbookPinIdentity);
    if (
      snapshot.readable &&
      document.activeElement === source &&
      dock?.contains(source) &&
      before.includes(identity)
    ) {
      const listeners = new AbortController();
      const expected = before.filter((id) => id !== identity);
      const unsubscribe = session.subscribe(() => {
        const current = session.getSnapshot();
        const ids = current.pins.map(workbookPinIdentity);
        if (
          !current.readable ||
          current.attemptId !== snapshot.attemptId ||
          (!samePins(ids, before) && !samePins(ids, expected))
        )
          cancel();
      });
      pending.current = {
        session,
        attemptId: snapshot.attemptId,
        before,
        expected,
        identity,
        source,
        dock,
        dispose: () => {
          listeners.abort();
          unsubscribe();
        },
      };
      document.addEventListener(
        "focusin",
        (event) => {
          if (event.target !== source && event.target !== document.body)
            cancel();
        },
        { capture: true, signal: listeners.signal },
      );
      document.addEventListener("pointerdown", cancel, {
        capture: true,
        signal: listeners.signal,
      });
      document.addEventListener("keydown", cancel, {
        capture: true,
        signal: listeners.signal,
      });
    }
    try {
      session.remove(pin);
    } catch (error) {
      cancel();
      throw error;
    }
  };
  return { buttonRef, remove };
}
