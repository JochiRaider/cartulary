export function browserStageCompleteRuntimeCommand({
  browserSessionScript,
  env,
  leaseFile,
  shouldStopSession,
  target,
  testOutputCommand,
}) {
  const commands = [
    `${testOutputCommand} target-summary ${JSON.stringify(target)} pass --quiet-success`,
    "summary_status=$?",
  ];
  if (shouldStopSession) {
    commands.push(
      `${JSON.stringify(browserSessionScript)} --session-stop --lease-file ${JSON.stringify(leaseFile)}`,
      "stop_status=$?",
    );
  } else {
    commands.push("stop_status=0");
  }
  commands.push(
    'if [[ "$summary_status" -ne 0 ]]; then exit "$summary_status"; fi',
    'exit "$stop_status"',
  );
  return {
    command: "bash",
    args: ["-c", commands.join("; ")],
    env,
  };
}
