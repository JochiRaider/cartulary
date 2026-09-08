type DepartureOwner = {
  hasWork: () => boolean;
  requestLeave: () => Promise<boolean>;
};

/** Fixed App-owned participants. Each explicit choice applies only to its reviewed owner. */
export async function reviewAppDeparture(options: {
  memberships: DepartureOwner;
  metadata: DepartureOwner;
  lifecycle: DepartureOwner;
  deploymentUsers: DepartureOwner;
  isCurrent: () => boolean;
}): Promise<boolean> {
  const owners = [
    options.memberships,
    options.metadata,
    options.lifecycle,
    options.deploymentUsers,
  ];
  for (const owner of owners) {
    if (!options.isCurrent()) return false;
    if (owner.hasWork() && !(await owner.requestLeave())) return false;
  }
  return options.isCurrent() && owners.every((owner) => !owner.hasWork());
}
