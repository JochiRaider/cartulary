import {
  type RefCallback,
  type RefObject,
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
} from "react";

type PendingRemoval = {
  readonly removedId: string;
  readonly before: readonly string[];
  readonly expected: readonly string[];
  readonly source: HTMLButtonElement;
  readonly scopeKey: string;
  readonly listeners: AbortController;
};

function sameIds(left: readonly string[], right: readonly string[]) {
  return (
    left.length === right.length &&
    left.every((id, index) => id === right[index])
  );
}

/** Reveal only inside the nearest owned scrollport; never scroll the page or grid. */
function revealLocalFocus(target: HTMLElement) {
  for (
    let parent = target.parentElement;
    parent && parent !== document.body && parent !== document.documentElement;
    parent = parent.parentElement
  ) {
    if (parent.getAttribute("role") === "grid") return;
    const style = getComputedStyle(parent);
    if (!/(auto|scroll)/.test(`${style.overflowY} ${style.overflowX}`))
      continue;
    if (
      parent.scrollHeight <= parent.clientHeight &&
      parent.scrollWidth <= parent.clientWidth
    )
      continue;
    const control = target.getBoundingClientRect();
    const viewport = parent.getBoundingClientRect();
    if (control.top < viewport.top)
      parent.scrollTop += control.top - viewport.top;
    else if (control.bottom > viewport.bottom)
      parent.scrollTop += control.bottom - viewport.bottom;
    if (control.left < viewport.left)
      parent.scrollLeft += control.left - viewport.left;
    else if (control.right > viewport.right)
      parent.scrollLeft += control.right - viewport.right;
    return;
  }
}

/** A presentation-only intention; the caller remains the selection authority. */
export function useSelectedReferenceRemovalFocus({
  ids,
  scopeKey,
  disabled,
  fallback,
  groupRef,
}: {
  readonly ids: readonly string[];
  readonly scopeKey: string;
  readonly disabled: boolean;
  readonly fallback: () => HTMLElement | null;
  readonly groupRef: RefObject<HTMLElement | null>;
}) {
  const buttons = useRef(new Map<string, HTMLButtonElement>());
  const pending = useRef<PendingRemoval | null>(null);
  const priorScope = useRef(scopeKey);
  const cancel = useCallback(() => {
    pending.current?.listeners.abort();
    pending.current = null;
  }, []);

  useEffect(() => () => cancel(), [cancel]);
  useLayoutEffect(() => {
    if (priorScope.current !== scopeKey || disabled) {
      priorScope.current = scopeKey;
      cancel();
      return;
    }
    const intent = pending.current;
    if (!intent) return;
    if (intent.scopeKey !== scopeKey) {
      cancel();
      return;
    }
    if (ids.includes(intent.removedId)) {
      if (!sameIds(ids, intent.before)) cancel();
      return;
    }
    if (
      !sameIds(ids, intent.expected) ||
      (document.activeElement !== intent.source &&
        document.activeElement !== document.body)
    ) {
      cancel();
      return;
    }
    const index = intent.before.indexOf(intent.removedId);
    const candidates = [
      ...intent.before.slice(index + 1),
      ...intent.before.slice(0, index).reverse(),
    ];
    const next = candidates
      .filter((id) => ids.includes(id))
      .map((id) => buttons.current.get(id))
      .find((button) => button?.isConnected && !button.disabled);
    const selector = fallback();
    const target =
      next ??
      (selector?.isConnected && !selector.matches(":disabled")
        ? selector
        : groupRef.current);
    cancel();
    if (target?.isConnected) {
      target.focus({ preventScroll: true });
      if (document.activeElement === target) revealLocalFocus(target);
    }
  });

  const buttonRef =
    (id: string): RefCallback<HTMLButtonElement> =>
    (element) => {
      if (element) buttons.current.set(id, element);
      else buttons.current.delete(id);
    };
  const remove = (id: string, source: HTMLButtonElement, apply: () => void) => {
    cancel();
    if (!disabled && document.activeElement === source && ids.includes(id)) {
      const listeners = new AbortController();
      const retire = () => cancel();
      document.addEventListener(
        "focusin",
        (event) => {
          if (event.target !== source && event.target !== document.body)
            retire();
        },
        { capture: true, signal: listeners.signal },
      );
      document.addEventListener("pointerdown", retire, {
        capture: true,
        signal: listeners.signal,
      });
      document.addEventListener("keydown", retire, {
        capture: true,
        signal: listeners.signal,
      });
      pending.current = {
        removedId: id,
        before: [...ids],
        expected: ids.filter((item) => item !== id),
        source,
        scopeKey,
        listeners,
      };
    }
    try {
      apply();
    } catch (error) {
      cancel();
      throw error;
    }
  };
  return { buttonRef, remove };
}
