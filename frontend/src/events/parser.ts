import type { AgentEvent as Generated } from "./types.gen";

// Re-export under a friendlier name. The generated type is a discriminated
// union on `type`, which is exactly what callers want.
export type AgentEvent = Generated;

/**
 * Parses an SSE stream where every event is `data: <json>\n\n`. Calls `onEvent`
 * for each successfully-parsed canonical event. Skips comment/heartbeat lines.
 */
export async function parseSSE(
  stream: ReadableStream<Uint8Array>,
  onEvent: (ev: AgentEvent) => void,
): Promise<void> {
  const reader = stream.getReader();
  const decoder = new TextDecoder();
  let buf = "";

  for (;;) {
    const { value, done } = await reader.read();
    if (done) break;
    buf += decoder.decode(value, { stream: true });

    let sep: number;
    while ((sep = buf.indexOf("\n\n")) >= 0) {
      const block = buf.slice(0, sep);
      buf = buf.slice(sep + 2);
      const data = block
        .split("\n")
        .filter((l) => l.startsWith("data:"))
        .map((l) => l.slice(5).trimStart())
        .join("");
      if (!data) continue;
      try {
        onEvent(JSON.parse(data) as AgentEvent);
      } catch {
        // ignore unparseable frames; backend validates server-side
      }
    }
  }
}
