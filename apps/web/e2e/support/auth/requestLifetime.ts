import type { APIRequestContext } from "@playwright/test";

/** Own one transient authentication transport through operation and retirement. */
export async function withOwnedRequestContext<T>(
  context: APIRequestContext,
  operation: (context: APIRequestContext) => Promise<T>,
  retire?: (context: APIRequestContext) => Promise<void>,
): Promise<T> {
  let result: T | undefined;
  let primary: unknown;
  let failed = false;
  try {
    result = await operation(context);
  } catch (error) {
    failed = true;
    primary = error;
  }
  const cleanup: unknown[] = [];
  for (const release of [
    ...(retire ? [() => retire(context)] : []),
    () => context.dispose(),
  ]) {
    try {
      await release();
    } catch (error) {
      cleanup.push(error);
    }
  }
  if (cleanup.length) {
    if (failed) {
      throw new AggregateError(
        [primary, ...cleanup],
        "Authentication operation and request cleanup failed",
        { cause: primary },
      );
    }
    throw new AggregateError(cleanup, "Authentication request cleanup failed");
  }
  if (failed) throw primary;
  return result as T;
}
