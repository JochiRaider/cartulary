/** Retain the active case before cleanup, preserving the original failure. */
export async function retainAutosaveFeedbackCase<T extends { phase: string }>(
  observation: T,
  run: () => Promise<void>,
  attach: (observation: T) => Promise<void>,
  cleanup: () => Promise<void>,
): Promise<void> {
  let failed = false;
  let failure: unknown;
  const remember = (error: unknown) => {
    if (!failed) {
      failed = true;
      failure = error;
    }
  };
  try {
    await run();
  } catch (error) {
    remember(error);
  }
  try {
    await attach(observation);
  } catch (error) {
    remember(error);
  }
  try {
    await cleanup();
  } catch (error) {
    remember(error);
    // Cleanup has a separate observation; the pre-cleanup case keeps its phase.
    try {
      await attach({ ...observation, phase: "cleanup-failed" });
    } catch (error) {
      remember(error);
    }
  }
  if (failed) throw failure;
}
