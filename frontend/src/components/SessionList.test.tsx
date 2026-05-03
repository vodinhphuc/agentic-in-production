import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { SessionList } from "./SessionList";

describe("SessionList", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation((url: string, init?: RequestInit) => {
        if (url.endsWith("/api/agents")) {
          return Promise.resolve({
            ok: true,
            json: () =>
              Promise.resolve([{ name: "mock-trino-flavored", version: "0.1.0", enabled: true }]),
          });
        }
        if (url.endsWith("/api/sessions") && init?.method === "POST") {
          return Promise.resolve({
            ok: true,
            json: () =>
              Promise.resolve({
                id: "sess_1",
                agent_name: "mock-trino-flavored",
                created_at: new Date().toISOString(),
              }),
          });
        }
        return Promise.reject(new Error("unexpected " + url));
      }),
    );
  });

  it("creates a new session when an agent button is clicked", async () => {
    render(<SessionList />);
    const btn = await screen.findByRole("button", { name: /new session: mock/i });
    fireEvent.click(btn);
    await waitFor(() => expect(screen.getByText(/sess_1/)).toBeInTheDocument());
  });
});
