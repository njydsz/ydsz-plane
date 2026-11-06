"""Ydsz Plane Python SDK — 类型化 API 客户端。

安装：
    pip install ydsz-plane

快速上手：
    from ydsz_plane import PlaneClient

    plane = PlaneClient(base_url="https://your-host", token="ydz_xxx")
    items = plane.work_items.list(project_id=1, type="task")
"""
from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any, Dict, List, Optional, TypeVar, cast
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode
from urllib.request import Request, urlopen

from .resources import (
    AIAgentsApi,
    CustomersApi,
    CyclesApi,
    InitiativesApi,
    MembersApi,
    ModulesApi,
    PagesApi,
    ProjectsApi,
    SearchApi,
    TeamspacesApi,
    VersionsApi,
    ViewsApi,
    WorkItemsApi,
    WorkspacesApi,
)

T = TypeVar("T")


@dataclass
class PlaneClientOptions:
    """客户端配置。"""

    base_url: str
    token: str
    timeout: float = 30.0


class PlaneError(Exception):
    """API 错误。"""

    def __init__(
        self,
        message: str,
        status: int,
        code: Optional[str] = None,
        details: Any = None,
    ) -> None:
        super().__init__(message)
        self.status = status
        self.code = code
        self.details = details


class PlaneClient:
    """Ydsz Plane 客户端。"""

    workspaces: WorkspacesApi
    projects: ProjectsApi
    work_items: WorkItemsApi
    cycles: CyclesApi
    modules: ModulesApi
    versions: VersionsApi
    pages: PagesApi
    members: MembersApi
    search: SearchApi
    ai_agents: AIAgentsApi
    views: ViewsApi
    initiatives: InitiativesApi
    teamspaces: TeamspacesApi
    customers: CustomersApi

    def __init__(self, base_url: str, token: str, timeout: float = 30.0) -> None:
        self.base_url = base_url.rstrip("/")
        self.token = token
        self.timeout = timeout

        self.workspaces = WorkspacesApi(self)
        self.projects = ProjectsApi(self)
        self.work_items = WorkItemsApi(self)
        self.cycles = CyclesApi(self)
        self.modules = ModulesApi(self)
        self.versions = VersionsApi(self)
        self.pages = PagesApi(self)
        self.members = MembersApi(self)
        self.search = SearchApi(self)
        self.ai_agents = AIAgentsApi(self)
        self.views = ViewsApi(self)
        self.initiatives = InitiativesApi(self)
        self.teamspaces = TeamspacesApi(self)
        self.customers = CustomersApi(self)

    def request(
        self,
        method: str,
        path: str,
        *,
        query: Optional[Dict[str, Any]] = None,
        body: Any = None,
    ) -> Any:
        """发送 HTTP 请求。"""
        url = f"{self.base_url}/api/v1{path}"
        if query:
            filtered = {k: v for k, v in query.items() if v not in (None, "")}
            if filtered:
                url = f"{url}?{urlencode(filtered)}"

        data = json.dumps(body).encode("utf-8") if body is not None else None
        headers = {
            "Content-Type": "application/json",
            "Authorization": f"Bearer {self.token}",
        }
        req = Request(url, data=data, headers=headers, method=method)

        try:
            with urlopen(req, timeout=self.timeout) as resp:
                raw = resp.read().decode("utf-8")
        except HTTPError as exc:
            raw = exc.read().decode("utf-8", errors="replace")
            payload = _safe_json(raw)
            raise PlaneError(
                (payload or {}).get("error") or f"HTTP {exc.code}",
                exc.code,
                (payload or {}).get("code"),
                (payload or {}).get("details"),
            ) from None
        except URLError as exc:
            raise PlaneError(f"请求失败：{exc.reason}", 0, "NETWORK_ERROR") from None

        return _safe_json(raw)


def _safe_json(raw: str) -> Any:
    if not raw:
        return None
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return raw
