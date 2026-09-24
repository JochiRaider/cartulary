import {
  type FocusEvent,
  type KeyboardEvent,
  type RefCallback,
  type RefObject,
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
} from "react";

type OverlayItem = HTMLElement & { readonly disabled?: boolean | undefined };

export type RegisteredOverlayNavigation<Key extends string> = {
  readonly activeKey: Key | null;
  readonly close: (options: { readonly restoreTriggerFocus: boolean }) => void;
  readonly focusItem: (itemKey: Key) => boolean;
  readonly onItemFocus: (itemKey: Key) => void;
  readonly onOverlayBlur: (event: FocusEvent<HTMLElement>) => void;
  readonly onOverlayFocus: (event: FocusEvent<HTMLElement>) => void;
  readonly onOverlayKeyDown: (event: KeyboardEvent<HTMLElement>) => void;
  readonly onItemKeyDown: (
    event: KeyboardEvent<HTMLElement>,
    itemKey: Key,
  ) => void;
  readonly prepareOpen: (preferredKey?: Key | null) => void;
  readonly registerItem: (itemKey: Key) => RefCallback<OverlayItem>;
  readonly tabIndexFor: (itemKey: Key) => 0 | -1;
};

export function useRegisteredOverlayNavigation<Key extends string>({
  fallbackFocusRef,
  initialItemKey,
  isOpen,
  itemKeys,
  keyboardMode = "menu",
  onRequestClose,
  preferredReturnFocusRef,
  reconcileItemKey,
  reconcileItems = false,
  restoreFocusOnSubjectChange = true,
  restoreFocusOnUnmount = true,
  onRestoreFocus,
  subjectKey,
  trapTab = false,
  triggerRef,
}: {
  readonly fallbackFocusRef?: RefObject<HTMLElement | null> | undefined;
  readonly initialItemKey: Key | null;
  readonly isOpen: boolean;
  readonly itemKeys: readonly Key[];
  readonly keyboardMode?: "form" | "menu";
  readonly onRequestClose: () => void;
  readonly preferredReturnFocusRef?: RefObject<HTMLElement | null> | undefined;
  /** A consumer may choose a semantic successor when its focused item retires. */
  readonly reconcileItemKey?:
    | ((
        itemKey: Key,
        previousKeys: readonly Key[],
        eligibleKeys: readonly Key[],
      ) => Key | null)
    | undefined;
  readonly reconcileItems?: boolean;
  readonly restoreFocusOnSubjectChange?: boolean;
  readonly restoreFocusOnUnmount?: boolean;
  readonly subjectKey: string;
  readonly trapTab?: boolean | undefined;
  readonly triggerRef?: RefObject<HTMLElement | null>;
  /** Semantic consumers resolve their current invoking target at dismissal. */
  readonly onRestoreFocus?: () => void;
}): RegisteredOverlayNavigation<Key> {
  const [activeKey, setActiveKey] = useState<Key | null>(null);
  const itemRefs = useRef(new Map<Key, OverlayItem>());
  const focusedItemRef = useRef<OverlayItem | null>(null);
  const pendingInitialKeyRef = useRef<Key | null>(null);
  const pendingTriggerRestoreRef = useRef(false);
  const wasOpenRef = useRef(false);
  const previousSubjectKeyRef = useRef(subjectKey);
  const previousEligibleKeysRef = useRef<readonly Key[]>([]);

  const eligibleKeys = useCallback(() => {
    return itemKeys.filter((key) => {
      const item = itemRefs.current.get(key);
      return (
        item?.isConnected === true &&
        item.disabled !== true &&
        item.getAttribute("aria-disabled") !== "true"
      );
    });
  }, [itemKeys]);

  const focusItem = useCallback((itemKey: Key) => {
    const item = itemRefs.current.get(itemKey);
    if (
      item === undefined ||
      !item.isConnected ||
      item.disabled === true ||
      item.getAttribute("aria-disabled") === "true"
    ) {
      return false;
    }
    setActiveKey(itemKey);
    focusedItemRef.current = item;
    item.focus({ preventScroll: true });
    return true;
  }, []);

  const restoreTriggerFocus = useCallback(() => {
    if (onRestoreFocus) {
      onRestoreFocus();
      return;
    }
    const preferred = preferredReturnFocusRef?.current;
    if (
      preferred?.isConnected &&
      !preferred.hasAttribute("disabled") &&
      preferred.getAttribute("aria-disabled") !== "true"
    ) {
      preferred.focus({ preventScroll: true });
      if (document.activeElement === preferred) return;
    }
    const trigger = triggerRef?.current;
    if (trigger?.isConnected) {
      trigger.focus({ preventScroll: true });
      if (document.activeElement === trigger) return;
    }
    const fallback = fallbackFocusRef?.current;
    if (fallback?.isConnected) {
      fallback.focus({ preventScroll: true });
    }
  }, [fallbackFocusRef, preferredReturnFocusRef, triggerRef, onRestoreFocus]);

  const close = useCallback(
    (options: { readonly restoreTriggerFocus: boolean }) => {
      pendingInitialKeyRef.current = null;
      pendingTriggerRestoreRef.current = options.restoreTriggerFocus;
      setActiveKey(null);
      onRequestClose();
    },
    [onRequestClose],
  );

  // A registered control can become disabled without changing its key or ref.
  useLayoutEffect(() => {
    const subjectChanged = previousSubjectKeyRef.current !== subjectKey;
    previousSubjectKeyRef.current = subjectKey;
    if (subjectChanged && isOpen && wasOpenRef.current) {
      pendingInitialKeyRef.current = null;
      pendingTriggerRestoreRef.current = false;
      setActiveKey(null);
      onRequestClose();
      if (restoreFocusOnSubjectChange) restoreTriggerFocus();
      wasOpenRef.current = false;
      return;
    }
    if (!isOpen) {
      wasOpenRef.current = false;
      setActiveKey(null);
      focusedItemRef.current = null;
      if (pendingTriggerRestoreRef.current) {
        pendingTriggerRestoreRef.current = false;
        restoreTriggerFocus();
      }
      return;
    }
    const keys = eligibleKeys();
    const previousKeys = previousEligibleKeysRef.current;
    previousEligibleKeysRef.current = keys;
    if (wasOpenRef.current) {
      const focusedItem = focusedItemRef.current;
      const focusIsOwned =
        focusedItem !== null &&
        (document.activeElement === focusedItem ||
          (document.activeElement === document.body &&
            (!focusedItem.isConnected ||
              focusedItem.disabled === true ||
              focusedItem.getAttribute("aria-disabled") === "true")));
      if (reconcileItems && activeKey !== null && !keys.includes(activeKey)) {
        const index = Math.max(0, previousKeys.indexOf(activeKey));
        const preferredKey = reconcileItemKey?.(activeKey, previousKeys, keys);
        const nextKey =
          preferredKey !== undefined &&
          preferredKey !== null &&
          keys.includes(preferredKey)
            ? preferredKey
            : keys[Math.min(index, keys.length - 1)];
        setActiveKey(null);
        if (nextKey !== undefined && focusIsOwned) focusItem(nextKey);
        else if (nextKey === undefined && focusIsOwned) {
          // The menu has no destination after its last control retires.
          close({ restoreTriggerFocus: true });
        }
      } else if (
        reconcileItems &&
        activeKey !== null &&
        keys.includes(activeKey) &&
        focusedItem !== null &&
        !focusedItem.isConnected &&
        document.activeElement === document.body
      ) {
        // A remount with the same semantic key may replace the focused element.
        focusItem(activeKey);
      }
      return;
    }
    wasOpenRef.current = true;
    const requestedKey = pendingInitialKeyRef.current ?? initialItemKey;
    pendingInitialKeyRef.current = null;
    if (requestedKey !== null && focusItem(requestedKey)) return;
    const firstKey = eligibleKeys()[0];
    if (firstKey !== undefined) {
      focusItem(firstKey);
    }
  });

  useLayoutEffect(
    () => () => {
      if (!pendingTriggerRestoreRef.current) return;
      pendingTriggerRestoreRef.current = false;
      if (restoreFocusOnUnmount) restoreTriggerFocus();
    },
    [restoreFocusOnUnmount, restoreTriggerFocus],
  );

  const onItemKeyDown = (event: KeyboardEvent<HTMLElement>, itemKey: Key) => {
    if (
      event.defaultPrevented ||
      event.nativeEvent.isComposing ||
      event.altKey ||
      event.ctrlKey ||
      event.metaKey
    )
      return;
    const decision = registeredOverlayKeyDecision(
      event.key,
      event.shiftKey,
      itemKey,
      eligibleKeys(),
      trapTab,
      keyboardMode,
    );
    if (decision.kind === "none") return;
    event.preventDefault();
    event.stopPropagation();
    if (decision.kind === "close") {
      close({ restoreTriggerFocus: true });
      return;
    }
    focusItem(decision.itemKey);
  };

  const registeredKeyFor = (target: EventTarget | null) =>
    [...itemRefs.current].find(([, item]) => item === target)?.[0] ?? null;

  return {
    activeKey,
    close,
    focusItem,
    onItemFocus: (itemKey) => {
      focusedItemRef.current = itemRefs.current.get(itemKey) ?? null;
      setActiveKey(itemKey);
    },
    onOverlayBlur: (event) => {
      const nextFocus = event.relatedTarget;
      if (
        (nextFocus !== null && nextFocus === triggerRef?.current) ||
        (nextFocus instanceof Node &&
          event.currentTarget.contains(nextFocus)) ||
        [...itemRefs.current.values()].some((item) => item === nextFocus)
      ) {
        return;
      }
      focusedItemRef.current = null;
      close({ restoreTriggerFocus: false });
    },
    onOverlayFocus: (event) => {
      const key = registeredKeyFor(event.target);
      focusedItemRef.current =
        key === null ? null : (itemRefs.current.get(key) ?? null);
      setActiveKey(key);
    },
    onOverlayKeyDown: (event) => {
      onItemKeyDown(event, registeredKeyFor(event.target) ?? ("" as Key));
    },
    onItemKeyDown,
    prepareOpen: (preferredKey = null) => {
      pendingInitialKeyRef.current = preferredKey;
    },
    registerItem: (itemKey) => (item) => {
      if (item === null) {
        itemRefs.current.delete(itemKey);
        return;
      }
      itemRefs.current.set(itemKey, item);
    },
    tabIndexFor: (itemKey) => (activeKey === itemKey ? 0 : -1),
  };
}

type RegisteredOverlayKeyDecision<Key extends string> =
  | { readonly kind: "none" }
  | { readonly kind: "close" }
  | { readonly kind: "focus"; readonly itemKey: Key };

function registeredOverlayKeyDecision<Key extends string>(
  key: string,
  shiftKey: boolean,
  itemKey: Key,
  eligibleKeys: readonly Key[],
  trapTab: boolean,
  keyboardMode: "form" | "menu",
): RegisteredOverlayKeyDecision<Key> {
  if (key === "Escape") return { kind: "close" };
  if (keyboardMode === "form") {
    if (key !== "Tab" || !trapTab || eligibleKeys.length === 0) {
      return { kind: "none" };
    }
    const currentIndex = eligibleKeys.indexOf(itemKey);
    const target = shiftKey
      ? currentIndex <= 0
        ? eligibleKeys[eligibleKeys.length - 1]
        : undefined
      : currentIndex < 0 || currentIndex === eligibleKeys.length - 1
        ? eligibleKeys[0]
        : undefined;
    return target === undefined
      ? { kind: "none" }
      : { kind: "focus", itemKey: target };
  }
  if (key === "Tab" && trapTab && eligibleKeys.length > 0) {
    const currentIndex = eligibleKeys.indexOf(itemKey);
    const targetIndex = shiftKey
      ? currentIndex <= 0
        ? eligibleKeys.length - 1
        : currentIndex - 1
      : currentIndex < 0 || currentIndex === eligibleKeys.length - 1
        ? 0
        : currentIndex + 1;
    const target = eligibleKeys[targetIndex];
    return target === undefined
      ? { kind: "none" }
      : { kind: "focus", itemKey: target };
  }
  if (!isOverlayNavigationKey(key) || eligibleKeys.length === 0) {
    return { kind: "none" };
  }
  const currentIndex = eligibleKeys.indexOf(itemKey);
  const targetIndex = overlayNavigationTargetIndex(
    key,
    currentIndex,
    eligibleKeys.length,
  );
  const target = eligibleKeys[targetIndex];
  return target === undefined
    ? { kind: "none" }
    : { kind: "focus", itemKey: target };
}

function isOverlayNavigationKey(
  key: string,
): key is "ArrowDown" | "ArrowUp" | "End" | "Home" {
  return (
    key === "ArrowDown" || key === "ArrowUp" || key === "End" || key === "Home"
  );
}

function overlayNavigationTargetIndex(
  key: "ArrowDown" | "ArrowUp" | "End" | "Home",
  currentIndex: number,
  itemCount: number,
): number {
  if (key === "End") return itemCount - 1;
  if (key === "Home") return 0;
  if (key === "ArrowUp") {
    return currentIndex <= 0 ? itemCount - 1 : currentIndex - 1;
  }
  return currentIndex < 0 || currentIndex === itemCount - 1
    ? 0
    : currentIndex + 1;
}
