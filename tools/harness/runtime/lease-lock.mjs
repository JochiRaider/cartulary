import { transactAdmission } from "./host-admission.mjs";

// Own the whole transaction while flock holds the inherited descriptor. Never
// wait for parent acknowledgements: scheduler fixture work may block its loop.
try {
  const input = process.env.CARTULARY_HOST_ADMISSION_REQUEST;
  if (!input || input.length > 8192) throw new Error("invalid request");
  process.stdout.write(`${JSON.stringify({ ok: true, value: transactAdmission(JSON.parse(input)) })}\n`);
} catch { process.stdout.write('{"ok":false}\n'); process.exitCode = 1; }
