import { test, expect } from "@playwright/test";

test("diag read-all status", async ({ request }) => {
  const apiURL = process.env.API_URL || "http://127.0.0.1:8080/api/v1";
  const loginRes = await request.post(`${apiURL}/auth/login`, {
    data: { email: "admin@njydsz.com", password: "Admin@1020" },
  });
  console.log("LOGIN STATUS", loginRes.status());
  const tok = (await loginRes.json()) as any;
  console.log("TOKEN PRESENT", !!tok.access_token, "KEYS", Object.keys(tok));
  const wsRes = await request.get(`${apiURL}/workspaces`, {
    headers: { Authorization: `Bearer ${tok.access_token}` },
  });
  console.log("WS STATUS", wsRes.status());
  const wsList = (await wsRes.json()) as any[];
  const wsId = wsList[0]?.id || 1;
  console.log("WSID", wsId);
  const markAllRes = await request.put(
    `${apiURL}/workspaces/${wsId}/notifications/read-all`,
    { headers: { Authorization: `Bearer ${tok.access_token}` } },
  );
  console.log("READALL STATUS", markAllRes.status());
  const body = await markAllRes.text();
  console.log("READALL BODY", body.slice(0, 300));
});
