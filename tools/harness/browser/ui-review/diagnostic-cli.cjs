// The bundled CLI and every daemon it forks inherit private file permissions.
process.umask(0o077);
process.argv.splice(1, 1);
require(process.argv[1]);
