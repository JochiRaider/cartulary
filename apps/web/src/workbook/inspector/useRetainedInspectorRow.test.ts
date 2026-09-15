import { renderHook } from "@testing-library/react";
import { expect, it } from "vitest";
import { useRetainedInspectorRow } from "./useRetainedInspectorRow";

it("retains one inspector source across window eviction without treating absence as deletion", () => {
  const row = { recordId: "original", version: 4 };
  const hook = renderHook(
    ({
      current,
      recordId,
      scope,
      readable,
    }: {
      current: typeof row | null;
      recordId: string | null;
      scope: string;
      readable: boolean;
    }) =>
      useRetainedInspectorRow({
        recordId,
        row: current,
        rowVersion: (value) => value.version,
        scope,
        readable,
      }),
    {
      initialProps: {
        current: row as typeof row | null,
        recordId: "original" as string | null,
        scope: "account-one",
        readable: true,
      },
    },
  );
  hook.rerender({
    current: null,
    recordId: "original",
    scope: "account-one",
    readable: true,
  });
  expect(hook.result.current).toBe(row);
  hook.rerender({
    current: { recordId: "original", version: 2 },
    recordId: "original",
    scope: "account-one",
    readable: true,
  });
  expect(hook.result.current).toBe(row);
  hook.rerender({
    current: null,
    recordId: "different",
    scope: "account-one",
    readable: true,
  });
  expect(hook.result.current).toBeNull();
  hook.rerender({
    current: row,
    recordId: "original",
    scope: "account-one",
    readable: true,
  });
  hook.rerender({
    current: null,
    recordId: "original",
    scope: "account-two",
    readable: true,
  });
  expect(hook.result.current).toBeNull();
  hook.rerender({
    current: row,
    recordId: "original",
    scope: "account-two",
    readable: false,
  });
  expect(hook.result.current).toBeNull();
  hook.unmount();
});
