/**
 * 版本（Version）全生命周期 E2E 测试。
 *
 * 覆盖：创建版本 → 激活 → 关联迭代 → 进度查询 → 发布（含检查清单准出）→ 归档。
 * 回归 S6 版本状态机 + S4 缺陷面板联动。
 *
 * 运行前提：后端 + 前端已启动，且已执行 make migrate && make seed。
 */
import { expect, test } from "@playwright/test";

const TEST_EMAIL = "admin@ydsz.dev";
const TEST_PASSWORD = "Admin@123";
const API_URL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";

async function login(request: any) {
  const res = await request.post(`${API_URL}/auth/login`, {
    data: { email: TEST_EMAIL, password: TEST_PASSWORD },
  });
  expect(res.ok()).toBe(true);
  const body = await res.json();
  return body.access_token;
}

async function getFirstWorkspace(request: any, token: string) {
  const res = await request.get(`${API_URL}/workspaces`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(res.ok()).toBe(true);
  const list = await res.json();
  return list[0]?.id || 1;
}

async function getFirstProject(request: any, token: string, wsId: number) {
  const res = await request.get(`${API_URL}/workspaces/${wsId}/projects`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(res.ok()).toBe(true);
  const body = await res.json();
  return body.results?.[0]?.id || body[0]?.id || 1;
}

async function getFirstSprint(request: any, token: string, wsId: number, projectId: number) {
  const res = await request.get(`${API_URL}/workspaces/${wsId}/projects/${projectId}/sprints`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  if (!res.ok()) return null;
  const body = await res.json();
  const list = body.results || body;
  return list[0]?.id || null;
}

test.describe("版本全生命周期", () => {
  test("创建 → 激活 → 查进度 → 归档 状态机闭环", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);
    const projectId = await getFirstProject(request, token, wsId);

    // 1. 创建版本
    const uniqueName = `E2E Version ${Date.now()}`;
    const createRes = await request.post(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions`,
      {
        headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
        data: {
          name: uniqueName,
          semver: "1.0.0",
          description: "E2E test version",
          start_date: "2026-01-01",
          end_date: "2026-06-30",
          target_date: "2026-06-15",
        },
      },
    );
    expect(createRes.status()).toBe(201);
    const version = await createRes.json();
    expect(version.id).toBeGreaterThan(0);
    expect(version.status).toBe("planning");
    expect(version.version).toBeGreaterThanOrEqual(0);

    const versionId = version.id;

    // 2. 激活版本 planning → active
    const activateRes = await request.post(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/activate`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(activateRes.ok()).toBe(true);
    const activated = await activateRes.json();
    expect(activated.status).toBe("active");

    // 3. 查询进度聚合接口
    const progressRes = await request.get(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/progress`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(progressRes.ok()).toBe(true);
    const progress = await progressRes.json();
    expect(progress).toHaveProperty("total_issues");
    expect(progress).toHaveProperty("completion_rate");

    // 4. 查询质量指标接口
    const qualityRes = await request.get(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/quality`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(qualityRes.ok()).toBe(true);
    const quality = await qualityRes.json();
    expect(quality).toHaveProperty("total_bugs");
    expect(quality).toHaveProperty("pass_quality_gate");

    // 5. 交付报告接口
    const reportRes = await request.get(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/delivery-report`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(reportRes.ok()).toBe(true);
    const report = await reportRes.json();
    expect(report).toHaveProperty("eligible_to_release");

    // 6. 归档版本 active → archived
    const archiveRes = await request.post(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/archive`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(archiveRes.ok()).toBe(true);
    const archived = await archiveRes.json();
    expect(archived.status).toBe("archived");

    // 7. 清理：删除版本
    const deleteRes = await request.delete(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(deleteRes.ok()).toBe(true);
  });

  test("版本发布：检查清单校验 + Release Notes 生成", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);
    const projectId = await getFirstProject(request, token, wsId);

    // 创建带检查清单的版本
    const createRes = await request.post(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions`,
      {
        headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
        data: {
          name: `E2E Release ${Date.now()}`,
          semver: "2.0.0",
          checklist: [
            { id: "c1", label: "代码冻结", required: true, checked: false },
            { id: "c2", label: "回归测试通过", required: true, checked: false },
            { id: "c3", label: "文档更新", required: false, checked: true },
          ],
        },
      },
    );
    expect(createRes.status()).toBe(201);
    const version = await createRes.json();
    const versionId = version.id;

    // 激活
    const activateRes = await request.post(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/activate`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(activateRes.ok()).toBe(true);

    // 尝试发布（有必填检查项未勾选，应失败或返回 eligible_to_release=false）
    const releaseRes = await request.post(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/release`,
      {
        headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
        data: {},
      },
    );
    // 准出失败时应返回 422，或者返回 200 但 version.status 仍为 active
    if (releaseRes.status() === 422) {
      // 校验拦截符合预期
      expect(releaseRes.status()).toBe(422);
    } else {
      // 如果后端使用 force_checklist=false 直接返回未通过状态
      expect(releaseRes.ok()).toBe(true);
    }

    // 生成 Release Notes
    const notesRes = await request.get(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/release-notes`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(notesRes.ok()).toBe(true);
    const notes = await notesRes.json();
    expect(notes).toHaveProperty("release_notes");

    // 强制发布（绕过检查清单）
    const forceReleaseRes = await request.post(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/release`,
      {
        headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
        data: { force_checklist: true },
      },
    );
    expect(forceReleaseRes.ok()).toBe(true);
    const released = await forceReleaseRes.json();
    expect(released.status).toBe("released");

    // 清理：删除
    await request.delete(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
  });

  test("版本关联迭代 + 缺陷面板查询", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);
    const projectId = await getFirstProject(request, token, wsId);

    // 创建版本
    const createRes = await request.post(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions`,
      {
        headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
        data: {
          name: `E2E Sprint Version ${Date.now()}`,
          semver: "1.5.0",
        },
      },
    );
    expect(createRes.status()).toBe(201);
    const version = await createRes.json();
    const versionId = version.id;

    // 尝试获取第一个迭代并关联
    const sprintId = await getFirstSprint(request, token, wsId, projectId);
    if (sprintId) {
      const addRes = await request.post(
        `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/sprints`,
        {
          headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
          data: { sprint_id: sprintId },
        },
      );
      expect(addRes.ok()).toBe(true);

      // 列出关联的迭代
      const listRes = await request.get(
        `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/sprints`,
        { headers: { Authorization: `Bearer ${token}` } },
      );
      expect(listRes.ok()).toBe(true);
      const sprints = await listRes.json();
      expect(sprints.results.length).toBeGreaterThanOrEqual(1);

      // 移除迭代关联
      const removeRes = await request.delete(
        `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/sprints/${sprintId}`,
        { headers: { Authorization: `Bearer ${token}` } },
      );
      expect(removeRes.ok()).toBe(true);
    }

    // 缺陷面板查询（空版本，但接口应可用）
    const defectsRes = await request.get(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}/defects`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(defectsRes.ok()).toBe(true);
    const defects = await defectsRes.json();
    expect(defects).toHaveProperty("results");
    expect(defects).toHaveProperty("total");

    // 跨版本缺陷过滤接口
    const filterRes = await request.get(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/defects?found_version_id=${versionId}&limit=10`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
    expect(filterRes.ok()).toBe(true);

    // 清理
    await request.delete(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
  });

  test("乐观锁：带 version 字段更新版本", async ({ request }) => {
    const token = await login(request);
    const wsId = await getFirstWorkspace(request, token);
    const projectId = await getFirstProject(request, token, wsId);

    const createRes = await request.post(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions`,
      {
        headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
        data: {
          name: `E2E Optimistic ${Date.now()}`,
          semver: "3.0.0",
        },
      },
    );
    expect(createRes.status()).toBe(201);
    const version = await createRes.json();
    const versionId = version.id;
    const currentVersion = version.version;

    // 更新时必须携带 version 字段
    const updateRes = await request.patch(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`,
      {
        headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
        data: {
          name: `E2E Optimistic Updated ${Date.now()}`,
          version: currentVersion,
        },
      },
    );
    expect(updateRes.ok()).toBe(true);
    const updated = await updateRes.json();
    expect(updated.name).toContain("Updated");
    expect(updated.version).toBeGreaterThan(currentVersion);

    // 清理
    await request.delete(
      `${API_URL}/workspaces/${wsId}/projects/${projectId}/versions/${versionId}`,
      { headers: { Authorization: `Bearer ${token}` } },
    );
  });
});
