// One authored migration-owner allocation shared by generation and validation.
export const ownersByVersion = new Map([
  [1, "database_migrations"], [2, "auth"], [3, "incidents"], [4, "recovery"],
  [5, "deployment_admin"], [6, "platform_jobs"], [7, "records"], [8, "revisions"],
  [9, "parties"], [10, "timeline"], [11, "entities"], [12, "indicators"],
  [13, "assessments"], [14, "links"], [15, "tasksdecisions"], [16, "artifacts"],
  [17, "evidence"], [18, "savedviews"], [19, "imports"], [20, "networkflow"],
  [21, "projections"], [22, "graphprojection"], [23, "reporting"],
  [24, "reportcomposition"], [25, "incidentbundles"], [26, "reference_data"],
  [27, "extensions"], [28, "audit"], [29, "collaboration"], [30, "networkflow"], [31, "imports"], [32, "imports"], [33, "imports"],
]);
