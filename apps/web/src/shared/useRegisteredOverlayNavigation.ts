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
  onRequestClose,
  preferredReturnFocusRef,
  reconcileItems = false,
  restoreFocusOnSubjectChange = true,
  restoreFocusOnUnmount = true,
  subjectKey,
  trapTab = false,
  triggerRef,
}: {
  readonly fallbackFocusRef?: RefObject<HTMLElement | null> | undefined;
  readonly initialItemKey: Key | null;
  readonly isOpen: boolean;
  readonly itemKeys: readonly Key[];
  readonly onRequestClose: () => void;
  readonly preferredReturnFocusRef?: RefObject<HTMLElement | null> | undefined;
  readonly reconcileItems?: boolean;
  readonly restoreFocusOnSubjectChange?: boolean;
  readonly restoreFocusOnUnmount?: boolean;
  readonly subjectKey: string;
  readonly trapTab?: boolean | undefined;
  readonly triggerRef: RefObject<HTMLElement | null>;
}): RegisteredOverlayNavigation<Key> {
  const [activeKey, setActiveKey] = useState<Key | null>(null);
  const itemRefs = useRef(new Map<Key, OverlayItem>());
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
    item.focus({ preventScroll: true });
    return true;
  }, []);

  const restoreTriggerFocus = useCallback(() => {
    const preferred = preferredReturnFocusRef?.current;
    if (
      preferred?.isConnected &&
      !preferred.hasAttribute("disabled") &&
      preferred.getAttribute("aria-disabled") !== "true"
    ) {
      preferred.focus({ preventScroll: true });
      if (document.activeElement === preferred) return;
    }
    const trigger = triggerRef.current;
    if (trigger?.isConnected) {
      trigger.focus({ preventScroll: true });
      if (document.activeElement === trigger) return;
    }
    const fallback = fallbackFocusRef?.current;
    if (fallback?.isConnected) {
      fallback.focus({ preventScroll: true });
    }
  }, [fallbackFocusRef, preferredReturnFocusRef, triggerRef]);

  const close = useCallback(
    (options: { readonly restoreTriggerFocus: boolean }) => {
      pendingInitialKeyRef.current = null;
      pendingTriggerRestoreRef.current = options.restoreTriggerFocus;
      setActiveKey(null);
      onRequestClose();
    },
    [onRequestClose],
  );

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
      if (reconcileItems && activeKey !== null && !keys.includes(activeKey)) {
        const index = Math.max(0, previousKeys.indexOf(activeKey));
        const nextKey = keys[Math.min(index, keys.length - 1)];
        setActiveKey(null);
        // Removal may move focus to body. Never override an external owner.
        if (
          nextKey !== undefined &&
          (document.activeElement === document.body ||
            document.activeElement === itemRefs.current.get(activeKey))
        ) {
          focusItem(nextKey);
        }
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
  }, [
    activeKey,
    eligibleKeys,
    focusItem,
    initialItemKey,
    isOpen,
    onRequestClose,
    restoreTriggerFocus,
    restoreFocusOnSubjectChange,
    reconcileItems,
    subjectKey,
  ]);

  useLayoutEffect(
    () => () => {
      if (!pendingTriggerRestoreRef.current) return;
      pendingTriggerRestoreRef.current = false;
      if (restoreFocusOnUnmount) restoreTriggerFocus();
    },
    [restoreFocusOnUnmount, restoreTriggerFocus],
  );

  return {
    activeKey,
    close,
    focusItem,
    onItemFocus: (itemKey) => setActiveKey(itemKey),
    onOverlayBlur: (event) => {
      const nextFocus = event.relatedTarget;
      if (
        nextFocus === triggerRef.current ||
        (nextFocus instanceof Node &&
          event.currentTarget.contains(nextFocus)) ||
        [...itemRefs.current.values()].some((item) => item === nextFocus)
      ) {
        return;
      }
      close({ restoreTriggerFocus: false });
    },
    onItemKeyDown: (event, itemKey) => {
      if (
        event.defaultPrevented ||
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
      );
      if (decision.kind === "none") return;
      event.preventDefault();
      event.stopPropagation();
      if (decision.kind === "close") {
        close({ restoreTriggerFocus: true });
        return;
      }
      focusItem(decision.itemKey);
    },
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
): RegisteredOverlayKeyDecision<Key> {
  if (key === "Escape") return { kind: "close" };
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
