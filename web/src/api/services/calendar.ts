/**
 * Calendar 日历事件 API。
 *
 * 后端路由：/api/v1/workspaces/:wsId/calendar/events
 * 权限：读 workspace:read；写 project:update。
 */
import { apiClient } from "../client";

/** 事件类型（与后端 EventType 枚举对齐）。 */
export type CalendarEventType = "meeting" | "milestone" | "release" | "deadline" | "other";

/** 日历事件实体。 */
export interface CalendarEvent {
  id: number;
  tenant_id: number;
  workspace_id: number;
  project_id?: number;
  title: string;
  description?: string;
  /** RFC3339 时间戳。 */
  start_time: string;
  end_time: string;
  is_all_day: boolean;
  location?: string;
  event_type: CalendarEventType;
  source_type?: string;
  source_id?: number;
  idempotency_key?: string;
  organizer_id?: number;
  status: "active" | "archived";
  version: number;
  created_by: number;
  created_at: string;
  updated_at: string;
}

/** 创建事件入参。 */
export interface CreateCalendarEventInput {
  title: string;
  description?: string;
  /** RFC3339 时间戳。 */
  start_time: string;
  end_time: string;
  is_all_day?: boolean;
  location?: string;
  event_type: CalendarEventType;
  source_type?: string;
  source_id?: number;
  idempotency_key?: string;
  organizer_id?: number;
}

/** 更新事件入参（乐观锁，需携带当前 version）。 */
export interface UpdateCalendarEventInput extends Partial<CreateCalendarEventInput> {
  version: number;
}

/** 列表查询参数。 */
export interface ListCalendarEventsParams {
  /** RFC3339 时间范围起点（含）。 */
  start?: string;
  /** RFC3339 时间范围终点（含）。 */
  end?: string;
  event_type?: CalendarEventType;
  limit?: number;
  offset?: number;
}

export const calendarApi = {
  /** 事件列表。 */
  async list(wsId: number, params: ListCalendarEventsParams = {}): Promise<CalendarEvent[]> {
    const query = new URLSearchParams();
    if (params.start) query.set("start", params.start);
    if (params.end) query.set("end", params.end);
    if (params.event_type) query.set("event_type", params.event_type);
    if (params.limit) query.set("limit", String(params.limit));
    if (params.offset) query.set("offset", String(params.offset));

    const qs = query.toString();
    const { data } = await apiClient.get<{ results: CalendarEvent[] }>(
      `/workspaces/${wsId}/calendar/events${qs ? `?${qs}` : ""}`,
    );
    return data.results ?? [];
  },

  /** 事件详情。 */
  async get(wsId: number, eventId: number): Promise<CalendarEvent> {
    const { data } = await apiClient.get<CalendarEvent>(
      `/workspaces/${wsId}/calendar/events/${eventId}`,
    );
    return data;
  },

  /** 创建事件（幂等键去重）。 */
  async create(wsId: number, input: CreateCalendarEventInput): Promise<CalendarEvent> {
    const { data } = await apiClient.post<CalendarEvent>(
      `/workspaces/${wsId}/calendar/events`,
      input,
    );
    return data;
  },

  /** 更新事件（version 不匹配返回 409）。 */
  async update(
    wsId: number,
    eventId: number,
    input: UpdateCalendarEventInput,
  ): Promise<CalendarEvent> {
    const { data } = await apiClient.patch<CalendarEvent>(
      `/workspaces/${wsId}/calendar/events/${eventId}`,
      input,
    );
    return data;
  },

  /** 归档事件（软删除）。 */
  async archive(wsId: number, eventId: number): Promise<void> {
    await apiClient.delete(`/workspaces/${wsId}/calendar/events/${eventId}`);
  },
};
