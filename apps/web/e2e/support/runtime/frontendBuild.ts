import { startFrontendPublication } from "../../../../../tools/harness/browser/frontend-publication.mjs";
import type { OwnedBrowserScope } from "./cleanup";

/** Own the independent builds until their browser dependencies have settled. */
export class OwnedFrontendPublications {
  readonly #publications: ReturnType<typeof startFrontendPublication>[] = [];
  #completion: Promise<void> | undefined;

  constructor(
    private readonly browserScope: Pick<OwnedBrowserScope, "dispose">,
    private readonly startPublication = startFrontendPublication,
  ) {}

  start(measurement: boolean) {
    if (this.#completion !== undefined)
      throw new Error("Frontend publication disposal has begun");
    const publication = this.startPublication(measurement);
    this.#publications.push(publication);
    return publication;
  }

  dispose(): Promise<void> {
    this.#completion ??= this.#dispose();
    return this.#completion;
  }

  async #dispose() {
    const failures: string[] = [];
    try {
      await this.browserScope.dispose();
    } catch {
      failures.push("browser.dispose");
    }
    const results = await Promise.allSettled(
      this.#publications.map(async (publication) => publication.stop()),
    );
    const failedBuilds = results.filter(
      (result) => result.status === "rejected",
    ).length;
    if (failedBuilds) failures.push(`publication.stop=${failedBuilds}`);
    if (failures.length)
      throw new Error(
        `Frontend publication cleanup failed: ${failures.join(", ")}`,
      );
  }
}
