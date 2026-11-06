"""Ydsz Plane Python SDK — 资源模块。"""
from __future__ import annotations

from typing import Any, Dict, List, Optional, TYPE_CHECKING

if TYPE_CHECKING:  # pragma: no cover
    from .client import PlaneClient


class _Resource:
    """资源基类。"""

    def __init__(self, client: "PlaneClient") -> None:
        self._client = client


class WorkspacesApi(_Resource):
    """工作空间。"""

    def list(self) -> Dict[str, Any]:
        return self._client.request("GET", "/workspaces")

    def create(self, name: str, slug: Optional[str] = None, logo_url: Optional[str] = None) -> Dict[str, Any]:
        return self._client.request("POST", "/workspaces", body={"name": name, "slug": slug, "logo_url": logo_url})

    def get(self, workspace_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}")

    def get_by_slug(self, slug: str) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/slug/{slug}")


class ProjectsApi(_Resource):
    """项目。"""

    def list(self, workspace_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/projects")

    def create(self, workspace_id: Any, name: str, **kwargs: Any) -> Dict[str, Any]:
        body = {"name": name, **kwargs}
        return self._client.request("POST", f"/workspaces/{workspace_id}/projects", body=body)

    def get(self, workspace_id: Any, project_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/projects/{project_id}")

    def archive(self, workspace_id: Any, project_id: Any) -> Dict[str, Any]:
        return self._client.request("POST", f"/workspaces/{workspace_id}/projects/{project_id}/archive")


class WorkItemsApi(_Resource):
    """工作项。"""

    def list(
        self,
        project_id: Any,
        *,
        type: Optional[str] = None,
        state: Optional[str] = None,
        assignee: Optional[str] = None,
        limit: Optional[int] = None,
        offset: Optional[int] = None,
        workspace_id: Any = None,
    ) -> Dict[str, Any]:
        query = {"type": type, "state": state, "assignee": assignee, "limit": limit, "offset": offset}
        return self._client.request(
            "GET",
            f"/workspaces/{workspace_id or ''}/projects/{project_id}/issues",
            query=query,
        )

    def get(self, workspace_id: Any, project_id: Any, id: Any, *, type: Optional[str] = None) -> Dict[str, Any]:
        return self._client.request(
            "GET",
            f"/workspaces/{workspace_id}/projects/{project_id}/issues/{id}",
            query={"type": type},
        )

    def create(
        self,
        workspace_id: Any,
        project_id: Any,
        type: str,
        title: str,
        **kwargs: Any,
    ) -> Dict[str, Any]:
        body = {"type": type, "title": title, **kwargs}
        return self._client.request("POST", f"/workspaces/{workspace_id}/projects/{project_id}/issues", body=body)

    def update(
        self,
        workspace_id: Any,
        project_id: Any,
        id: Any,
        **kwargs: Any,
    ) -> Dict[str, Any]:
        return self._client.request(
            "PATCH",
            f"/workspaces/{workspace_id}/projects/{project_id}/issues/{id}",
            body=kwargs,
        )

    def remove(self, workspace_id: Any, project_id: Any, id: Any) -> None:
        self._client.request("DELETE", f"/workspaces/{workspace_id}/projects/{project_id}/issues/{id}")

    def add_comment(self, workspace_id: Any, project_id: Any, id: Any, content: str) -> Dict[str, Any]:
        return self._client.request(
            "POST",
            f"/workspaces/{workspace_id}/projects/{project_id}/issues/{id}/comments",
            body={"content": content},
        )


class CyclesApi(_Resource):
    """迭代。"""

    def list(self, workspace_id: Any, project_id: Any, status: Optional[str] = None) -> Dict[str, Any]:
        return self._client.request(
            "GET",
            f"/workspaces/{workspace_id}/projects/{project_id}/sprints",
            query={"status": status},
        )

    def create(self, workspace_id: Any, project_id: Any, name: str, **kwargs: Any) -> Dict[str, Any]:
        return self._client.request(
            "POST",
            f"/workspaces/{workspace_id}/projects/{project_id}/sprints",
            body={"name": name, **kwargs},
        )

    def start(self, workspace_id: Any, project_id: Any, cycle_id: Any) -> Dict[str, Any]:
        return self._client.request(
            "POST",
            f"/workspaces/{workspace_id}/projects/{project_id}/sprints/{cycle_id}/start",
        )

    def complete(self, workspace_id: Any, project_id: Any, cycle_id: Any) -> Dict[str, Any]:
        return self._client.request(
            "POST",
            f"/workspaces/{workspace_id}/projects/{project_id}/sprints/{cycle_id}/complete",
        )

    def analytics(self, workspace_id: Any, project_id: Any, cycle_id: Any) -> Dict[str, Any]:
        return self._client.request(
            "GET",
            f"/workspaces/{workspace_id}/projects/{project_id}/sprints/{cycle_id}/analytics",
        )


class ModulesApi(_Resource):
    """模块。"""

    def list(self, workspace_id: Any, project_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/projects/{project_id}/modules")

    def create(self, workspace_id: Any, project_id: Any, name: str, **kwargs: Any) -> Dict[str, Any]:
        return self._client.request(
            "POST",
            f"/workspaces/{workspace_id}/projects/{project_id}/modules",
            body={"name": name, **kwargs},
        )

    def update(self, workspace_id: Any, project_id: Any, module_id: Any, **kwargs: Any) -> Dict[str, Any]:
        return self._client.request(
            "PATCH",
            f"/workspaces/{workspace_id}/projects/{project_id}/modules/{module_id}",
            body=kwargs,
        )

    def remove(self, workspace_id: Any, project_id: Any, module_id: Any) -> None:
        self._client.request("DELETE", f"/workspaces/{workspace_id}/projects/{project_id}/modules/{module_id}")


class VersionsApi(_Resource):
    """版本。"""

    def list(self, workspace_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/versions")

    def create(self, workspace_id: Any, name: str, **kwargs: Any) -> Dict[str, Any]:
        return self._client.request("POST", f"/workspaces/{workspace_id}/versions", body={"name": name, **kwargs})

    def release(self, workspace_id: Any, version_id: Any) -> Dict[str, Any]:
        return self._client.request("POST", f"/workspaces/{workspace_id}/versions/{version_id}/release")


class PagesApi(_Resource):
    """文档。"""

    def list(self, workspace_id: Any, project_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/projects/{project_id}/pages")

    def get(self, workspace_id: Any, project_id: Any, page_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/projects/{project_id}/pages/{page_id}")

    def create(self, workspace_id: Any, project_id: Any, title: str, **kwargs: Any) -> Dict[str, Any]:
        return self._client.request(
            "POST",
            f"/workspaces/{workspace_id}/projects/{project_id}/pages",
            body={"title": title, **kwargs},
        )


class MembersApi(_Resource):
    """成员。"""

    def list(self, workspace_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/members")

    def update_role(self, workspace_id: Any, member_id: Any, role: str) -> Dict[str, Any]:
        return self._client.request("PATCH", f"/workspaces/{workspace_id}/members/{member_id}", body={"role": role})

    def remove(self, workspace_id: Any, member_id: Any) -> None:
        self._client.request("DELETE", f"/workspaces/{workspace_id}/members/{member_id}")


class SearchApi(_Resource):
    """搜索。"""

    def query(self, workspace_id: Any, q: str, limit: int = 20) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/search", query={"q": q, "limit": limit})


class AIAgentsApi(_Resource):
    """AI Agent。"""

    def list(self, workspace_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/ai-agents")

    def create(self, workspace_id: Any, **kwargs: Any) -> Dict[str, Any]:
        return self._client.request("POST", f"/workspaces/{workspace_id}/ai-agents", body=kwargs)

    def remove(self, workspace_id: Any, agent_id: Any) -> None:
        self._client.request("DELETE", f"/workspaces/{workspace_id}/ai-agents/{agent_id}")

    def chat(
        self,
        workspace_id: Any,
        conversation_id: Any,
        content: str,
        mode: Optional[str] = None,
    ) -> Dict[str, Any]:
        return self._client.request(
            "POST",
            f"/workspaces/{workspace_id}/ai/conversations/{conversation_id}/messages",
            body={"content": content, "mode": mode},
        )

    def messages(self, workspace_id: Any, conversation_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/ai/conversations/{conversation_id}")


class ViewsApi(_Resource):
    """工作空间视图。"""

    def list(self, workspace_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/views")

    def create(self, workspace_id: Any, **kwargs: Any) -> Dict[str, Any]:
        return self._client.request("POST", f"/workspaces/{workspace_id}/views", body=kwargs)

    def duplicate(self, workspace_id: Any, view_id: Any) -> Dict[str, Any]:
        return self._client.request("POST", f"/workspaces/{workspace_id}/views/{view_id}/duplicate")

    def toggle_favorite(self, workspace_id: Any, view_id: Any) -> Dict[str, Any]:
        return self._client.request("POST", f"/workspaces/{workspace_id}/views/{view_id}/favorite")


class InitiativesApi(_Resource):
    """倡议。"""

    def list(self, workspace_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/initiatives")

    def create(self, workspace_id: Any, title: str, **kwargs: Any) -> Dict[str, Any]:
        return self._client.request(
            "POST",
            f"/workspaces/{workspace_id}/initiatives",
            body={"title": title, **kwargs},
        )

    def progress(self, workspace_id: Any, initiative_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/initiatives/{initiative_id}/progress")


class TeamspacesApi(_Resource):
    """团队空间。"""

    def list(self, workspace_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/teamspaces")

    def create(self, workspace_id: Any, name: str, **kwargs: Any) -> Dict[str, Any]:
        return self._client.request("POST", f"/workspaces/{workspace_id}/teamspaces", body={"name": name, **kwargs})

    def progress(self, workspace_id: Any, teamspace_id: Any) -> Dict[str, Any]:
        return self._client.request("GET", f"/workspaces/{workspace_id}/teamspaces/{teamspace_id}/progress")


class CustomersApi(_Resource):
    """客户。"""

    def list(
        self,
        workspace_id: Any,
        *,
        stage: Optional[str] = None,
        tier: Optional[str] = None,
        q: Optional[str] = None,
    ) -> Dict[str, Any]:
        return self._client.request(
            "GET",
            f"/workspaces/{workspace_id}/customers",
            query={"stage": stage, "tier": tier, "q": q},
        )

    def create(self, workspace_id: Any, name: str, **kwargs: Any) -> Dict[str, Any]:
        return self._client.request("POST", f"/workspaces/{workspace_id}/customers", body={"name": name, **kwargs})
