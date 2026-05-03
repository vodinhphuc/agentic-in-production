import { describe, it, expect, vi } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import { AuditLogView } from "./AuditLogView";

describe("AuditLogView", () => {
  it("loads and renders audit entries", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: true,
        json: () =>
          Promise.resolve([
            {
              id: 1,
              run_id: "r1",
              kind: "tool_call_start",
              occurred_at: new Date().toISOString(),
              payload: { x: 1 },
            },
          ]),
      }),
    );
    render(<AuditLogView sessionID="s1" />);
    await waitFor(() => expect(screen.getByText("tool_call_start")).toBeInTheDocument());
  });
});
