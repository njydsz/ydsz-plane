interface DashboardData {
  widgets: { id: number }[];
  snapshots: Record<string, any>;
  alerts: { id: number }[];
}
declare const dashboardData: { value: DashboardData | null };
dashboardData.value = { ...dashboardData.value };
