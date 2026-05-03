import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { ChatView } from "./ChatView";
import type { components } from "@/api/types.gen";

type Session = components["schemas"]["Session"];

function makeStream(events: object[]): Response {
  const body = events.map((e) => `data: ${JSON.stringify(e)}\n\n`).join("");
  const enc = new TextEncoder();
  const stream = new ReadableStream({
    start(c) {
      c.enqueue(enc.encode(body));
      c.close();
    },
  });
  return new Response(stream, { status: 200 });
}

describe("ChatView", () => {
  it("renders user message, tool call card, and assistant text from a fixture stream", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockImplementation(() =>
        Promise.resolve(
          makeStream([
            { type: "run_started", run_id: "r1" },
            { type: "text_delta", text: "Looking at " },
            {
              type: "tool_call_start",
              call_id: "c1",
              tool: "execute_query",
              args: { sql: "SELECT 1" },
            },
            { type: "tool_call_end", call_id: "c1", ok: true, result_preview: "1 row" },
            { type: "text_delta", text: "the orders table." },
            { type: "run_finished", reason: "done" },
          ]),
        ),
      ),
    );

    const session: Session = {
      id: "s1",
      agent_name: "mock-trino-flavored",
      created_at: new Date().toISOString(),
    };
    render(<ChatView session={session} />);
    fireEvent.change(screen.getByPlaceholderText(/ask the agent/i), { target: { value: "go" } });
    fireEvent.click(screen.getByRole("button", { name: /send/i }));

    await waitFor(() =>
      expect(screen.getByText(/the orders table/)).toBeInTheDocument(),
    );
    expect(screen.getByRole("button", { name: /execute_query/ })).toBeInTheDocument();
  });
});
