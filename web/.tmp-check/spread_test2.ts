interface DashboardData {
  widgets: { id: number }[];
  snapshots: Record<string, any>;
  alerts: { id: number }[];
}
declare const d: { value: DashboardData | null };
if (d.value) {
  d.value = { ...d.value };
}
