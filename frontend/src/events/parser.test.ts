import { describe, it, expect } from "vitest";
import { parseSSE, type AgentEvent } from "./parser";

function streamFrom(chunks: string[]): ReadableStream<Uint8Array> {
  const enc = new TextEncoder();
  return new ReadableStream({
    start(controller) {
      for (const c of chunks) controller.enqueue(enc.encode(c));
      controller.close();
    },
  });
}

describe("parseSSE", () => {
  it("parses two events split across chunks", async () => {
    const got: AgentEvent[] = [];
    await parseSSE(
      streamFrom([
        'data: {"type":"run_started","run_id":"r1"}\n\n',
        'data: {"type":"text_delta","text":"hi"}\n\n',
      ]),
      (ev) => got.push(ev),
    );
    expect(got).toHaveLength(2);
    expect(got[0]).toMatchObject({ type: "run_started", run_id: "r1" });
    expect(got[1]).toMatchObject({ type: "text_delta", text: "hi" });
  });

  it("handles a JSON message split across chunk boundaries", async () => {
    const got: AgentEvent[] = [];
    await parseSSE(
      streamFrom(['data: {"type":"run_st', 'arted","run_id":"r2"}\n\n']),
      (ev) => got.push(ev),
    );
    expect(got).toEqual([{ type: "run_started", run_id: "r2" }]);
  });
});
