import { readLocalFile, listLocalDirectory } from "../runtime/secure-local-files.mjs";
import path from "node:path";
import { parseStrictJSON } from "../contract/index.mjs";

function producerFromFilename(filename) {
  if (!filename.endsWith(".ndjson")) return "";
  return filename.slice(0, -".ndjson".length);
}

function readProducerJournal(file, producerId) {
  const raw = readLocalFile(file, { maximum: 64 * 1024 * 1024 }).toString("utf8");
  const complete = raw.endsWith("\n");
  const lines = raw.split(/\n/u);
  if (!complete) lines.pop();
  const events = [];
  for (const line of lines) {
    if (!line.trim()) continue;
    let event;
    try {
      event = parseStrictJSON(line);
    } catch (error) {
      throw new Error(`malformed completed service journal record in ${file}: ${error}`);
    }
    const expectedSequence = events.length + 1;
    if (
      event.schema_id !== "cartulary.test_services.journal_event.v1" ||
      event.producer_id !== producerId ||
      event.seq !== expectedSequence ||
      !event.type ||
      !event.timestamp ||
      !Number.isInteger(event.pid) ||
      event.pid < 1
    ) {
      throw new Error(
        `service journal identity/sequence mismatch in ${file}: expected ${producerId}/${expectedSequence}`,
      );
    }
    events.push(event);
  }
  return events;
}

export function loadServiceJournalEvents({ resultsRoot, runId }) {
  const servicesRoot = path.join(resultsRoot, runId, "_shared", "test-services");
  const records = [];
  for (const suiteEntry of listLocalDirectory(servicesRoot, { optional: true })
    .filter((entry) => entry.isDirectory())
    .sort((left, right) => left.name.localeCompare(right.name))) {
    const suiteRoot = path.join(servicesRoot, suiteEntry.name);
    const journalsRoot = path.join(suiteRoot, "journals");
    for (const journalEntry of listLocalDirectory(journalsRoot, { optional: true })
      .filter((entry) => entry.isFile() || entry.isSymbolicLink())
      .sort((left, right) => left.name.localeCompare(right.name))) {
      const producerId = producerFromFilename(journalEntry.name);
      if (!producerId) continue;
      const journalPath = path.join(journalsRoot, journalEntry.name);
      for (const event of readProducerJournal(journalPath, producerId)) {
        records.push({ event, suiteRoot, journalPath });
      }
    }
  }
  records.sort(
    (left, right) =>
      left.event.timestamp.localeCompare(right.event.timestamp) ||
      left.event.producer_id.localeCompare(right.event.producer_id) ||
      left.event.seq - right.event.seq,
  );
  return records;
}
