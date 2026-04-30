import type { BrowserContext } from "@playwright/test";

export type StorageState = Awaited<ReturnType<BrowserContext["storageState"]>>;
