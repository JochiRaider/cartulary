export function readBrowserState(name: string): string | null;
export function writeBrowserState(name: string, contents: string): void;
export function withBrowserSetupLock<T>(callback: () => Promise<T>): Promise<T>;
