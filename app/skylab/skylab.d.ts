// These values correspond to the stages present inside the stage_enum table in the database
export enum Stage {
  Application = "application",
  Submission = "submission",
  Evaluation = "evaluation",
  Feedback = "feedback",
  StageNull = "",
}

// These values correspond to the milestones present inside the milestone_enum table in the database
export enum Milestone {
  Milestone1 = "milestone1",
  Milestone2 = "milestone2",
  Milestone3 = "milestone3",
  MilestoneNull = "",
}

// These values correspond to the milestones present inside the project_level_enum table in the database
export enum ProjectLevel {
  Vostok = "vostok",
  Gemini = "gemini",
  Apollo = "apollo",
}

// These values correspond to the milestones present inside the applications_status_enum table in the database
export enum ApplicationStatus {
  Pending = "pending",
  Accepted = "accepted",
  Deleted = "deleted",
}
