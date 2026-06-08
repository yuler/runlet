export type RunletStatus = "draft" | "active" | "paused" | "archived";

export type RunStatus =
  | "queued"
  | "running"
  | "succeeded"
  | "failed"
  | "canceled";

export type RunletTrigger = "manual" | "schedule" | "api";

export interface Runlet {
  id: string;
  name: string;
  slug: string;
  description: string | null;
  command: string;
  status: RunletStatus;
  createdAt: string;
  updatedAt: string;
}

export interface Run {
  id: string;
  runletId: string;
  status: RunStatus;
  trigger: RunletTrigger;
  exitCode: number | null;
  startedAt: string | null;
  finishedAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface RunEvent {
  id: string;
  runId: string;
  sequence: number;
  level: "debug" | "info" | "warn" | "error";
  message: string;
  metadata: Record<string, unknown>;
  createdAt: string;
}

export interface Schedule {
  id: string;
  runletId: string;
  name: string;
  rrule: string;
  enabled: boolean;
  nextRunAt: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface Page<T> {
  data: T[];
  nextCursor: string | null;
}
