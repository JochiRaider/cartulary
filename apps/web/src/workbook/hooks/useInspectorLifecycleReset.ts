import { useLayoutEffect, useRef } from "react";

function normalizeInspectorResetKey(
  resetKey: string | undefined,
): string | undefined {
  return resetKey === "" ? undefined : resetKey;
}

export function useInspectorLifecycleReset(
  resetKey: string | undefined,
  reset: () => void,
): void {
  const normalizedResetKey = normalizeInspectorResetKey(resetKey);
  const previousResetKeyRef = useRef(normalizedResetKey);
  const resetRef = useRef(reset);

  useLayoutEffect(() => {
    resetRef.current = reset;
  }, [reset]);

  useLayoutEffect(() => {
    if (previousResetKeyRef.current === normalizedResetKey) {
      return;
    }
    previousResetKeyRef.current = normalizedResetKey;
    if (normalizedResetKey === undefined) {
      return;
    }
    resetRef.current();
  }, [normalizedResetKey]);
}
