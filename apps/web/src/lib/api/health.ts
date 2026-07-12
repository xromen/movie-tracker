import { fetchApiResponse } from "./client";

export type HealthStatus = "ok" | "degraded" | "unavailable";

export interface DependencyStatus {
  status: "ok" | "unavailable";
  latency?: string;
  error?: string;
}

export interface PostgresStatus extends DependencyStatus {
  total_conns?: number;
  acquired_conns?: number;
  idle_conns?: number;
  empty_acquire_count?: number;
  empty_acquire_wait_time?: string;
}

export interface ReadyResponse {
  status: Extract<HealthStatus, "ok" | "degraded">;
  version: string;
  uptime: string;
  goroutines: number;
  dependencies: {
    postgres: PostgresStatus;
    redis: DependencyStatus;
  };
}

export interface HealthOverview {
  liveOk: boolean;
  ready?: ReadyResponse;
}

export const getHealthOverview = async (): Promise<HealthOverview> => {
  const [liveResponse, readyResponse] = await Promise.all([
    fetchApiResponse("/health/live"),
    fetchApiResponse("/health/ready"),
  ]);

  const ready =
    readyResponse.ok || readyResponse.status === 503
      ? ((await readyResponse.json()) as ReadyResponse)
      : undefined;

  return {
    liveOk: liveResponse.ok,
    ready,
  };
};
