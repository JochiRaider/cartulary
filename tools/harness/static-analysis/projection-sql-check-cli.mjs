import { inspectProjectionSQL } from "./projection-sql-check.mjs";
try {
  const failures = inspectProjectionSQL(process.cwd());
  if (failures.length) { console.error(failures.join("\n")); process.exitCode = 1; }
  else console.log("projection SQL declaration ownership verified");
} catch (error) { console.error(error.message); process.exitCode = 1; }
