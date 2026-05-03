import { defineConfig } from "@playwright/test";

export default defineConfig({
  testDir: "./e2e",
  timeout: 30_000,
  fullyParallel: false,
  reporter: [["list"]],
  use: {
    baseURL: "http://localhost:5173",
    trace: "retain-on-failure",
  },
  // CI runs the dev server externally (Postgres + backend + Vite dev) — see Task 27.
  // For local dev you can `pnpm dev` in another terminal and just run `pnpm playwright test`.
});
