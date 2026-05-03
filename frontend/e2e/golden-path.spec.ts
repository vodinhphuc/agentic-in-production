import { test, expect } from "@playwright/test";

test("login → create session → send message → see tool-call card → audit", async ({ page }) => {
  await page.goto("/");

  // 1. Login (admin password from env, set by `make admin-password` before running).
  const password = process.env.AIP_E2E_ADMIN_PASSWORD;
  test.skip(!password, "AIP_E2E_ADMIN_PASSWORD not set");
  await page.getByLabel(/password/i).fill(password!);
  await page.getByRole("button", { name: /sign in/i }).click();

  // 2. Create a session against the mock agent.
  await page.getByRole("button", { name: /new session: mock-trino-flavored/i }).click();
  await expect(page.getByPlaceholder(/ask the agent/i)).toBeVisible();

  // 3. Send a message that triggers the trino-investigation scenario.
  await page.getByPlaceholder(/ask the agent/i).fill("investigate WIN-WS-014");
  await page.getByRole("button", { name: /^send$/i }).click();

  // 4. Tool-call card appears.
  const toolBtn = page.getByRole("button", { name: /describe_table/ }).first();
  await expect(toolBtn).toBeVisible({ timeout: 10_000 });

  // 5. Final assistant text mentions the suspicious PowerShell.
  await expect(page.getByText(/suspicious PowerShell|03:14/i)).toBeVisible({ timeout: 10_000 });

  // 6. Audit log expands and shows entries.
  await page.getByText(/audit log/i).click();
  await expect(page.locator("table tbody tr").first()).toBeVisible();
});
