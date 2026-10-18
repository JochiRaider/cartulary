// Inspection executes inside the exact owned container. Only closed booleans
// cross the adapter boundary; endpoints, socket tables and PIDs stay private.
export const rendererQuiescenceScript = String.raw`
const fs = require('node:fs');
const processes = fs.readdirSync('/proc').filter(x => /^[0-9]+$/.test(x));
let server = false, descendants = false;
for (const pid of processes) {
  if (+pid === process.pid || +pid === 1) continue;
  let args;
  try { args = fs.readFileSync('/proc/'+pid+'/cmdline','utf8').split('\0'); }
  catch (e) { if (e.code === 'ENOENT') continue; throw e; }
  if (args.includes('/home/pwuser/playwright/cli.js') && args.includes('run-server')) server = true;
  else descendants = true;
}
let clients = false, forwarding = false, listening = false;
for (const file of ['/proc/net/tcp','/proc/net/tcp6']) {
  for (const row of fs.readFileSync(file,'utf8').trim().split('\n').slice(1)) {
    const fields = row.trim().split(/\s+/);
    const port = parseInt(fields[1].split(':').at(-1),16), state = fields[3];
    if (state === '0A') { if (port === 3000) listening = true; else forwarding = true; }
    else if (state !== '06') clients = true;
  }
}
process.stdout.write(JSON.stringify({server_alive:server && listening,client_connections_absent:!clients,browser_processes_absent:!descendants,forwarding_listeners_absent:!forwarding}));
`;

export async function settleRenderer(containerID, run, timeoutMs = 30000) {
  const deadline = Date.now() + timeoutMs;
  do {
    const value = JSON.parse(run(["exec", containerID, "node", "-e", rendererQuiescenceScript], { timeout: Math.max(1, Math.min(3000, deadline - Date.now())) }));
    const keys = ["browser_processes_absent", "client_connections_absent", "forwarding_listeners_absent", "server_alive"];
    if (Object.keys(value).sort().join() !== keys.join() || keys.some(key => typeof value[key] !== "boolean")) throw new Error("invalid renderer settlement");
    if (!value.server_alive) throw new Error("renderer server exited");
    if (keys.every(key => value[key])) return value;
    await new Promise(resolve => setTimeout(resolve, 100));
  } while (Date.now() < deadline);
  throw new Error("renderer did not become quiescent within drain deadline");
}
