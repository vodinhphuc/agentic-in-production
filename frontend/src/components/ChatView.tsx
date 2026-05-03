import { useState } from "react";
import { sendMessage } from "@/api/client";
import { parseSSE, type AgentEvent } from "@/events/parser";
import { MessageBubble } from "./MessageBubble";
import { ToolCallCard } from "./ToolCallCard";
import type { components } from "@/api/types.gen";

type Session = components["schemas"]["Session"];

type Item =
  | { kind: "text"; role: "user" | "assistant"; text: string }
  | {
      kind: "tool";
      call_id: string;
      tool: string;
      args?: unknown;
      status: "pending" | "ok" | "error";
      resultPreview?: string;
      errorMessage?: string;
    };

export function ChatView({ session }: { session: Session }) {
  const [items, setItems] = useState<Item[]>([]);
  const [input, setInput] = useState("");
  const [busy, setBusy] = useState(false);

  function appendText(role: "user" | "assistant", text: string) {
    setItems((xs) => [...xs, { kind: "text", role, text }]);
  }

  function appendOrUpdateAssistantDelta(delta: string) {
    setItems((xs) => {
      const last = xs[xs.length - 1];
      if (last && last.kind === "text" && last.role === "assistant") {
        return [...xs.slice(0, -1), { ...last, text: last.text + delta }];
      }
      return [...xs, { kind: "text", role: "assistant", text: delta }];
    });
  }

  function applyEvent(ev: AgentEvent) {
    switch (ev.type) {
      case "run_started":
        /* noop in UI */ break;
      case "text_delta":
        appendOrUpdateAssistantDelta(ev.text);
        break;
      case "tool_call_start":
        setItems((xs) => [
          ...xs,
          { kind: "tool", call_id: ev.call_id, tool: ev.tool, args: ev.args, status: "pending" },
        ]);
        break;
      case "tool_call_end":
        setItems((xs) =>
          xs.map((it) =>
            it.kind === "tool" && it.call_id === ev.call_id
              ? {
                  ...it,
                  status: ev.ok ? "ok" : "error",
                  resultPreview: ev.result_preview,
                  errorMessage: ev.error_message,
                }
              : it,
          ),
        );
        break;
      case "state_update":
        /* could surface in a side panel later */ break;
      case "error":
        appendText("assistant", `(error: ${ev.code}) ${ev.message}`);
        break;
      case "run_finished":
        /* end of stream */ break;
    }
  }

  async function send() {
    if (!input.trim() || busy) return;
    const text = input;
    setInput("");
    setBusy(true);
    appendText("user", text);
    const ctrl = new AbortController();
    try {
      const res = await sendMessage(session.id, text, ctrl.signal);
      if (!res.ok || !res.body) throw new Error(`HTTP ${res.status}`);
      await parseSSE(res.body, applyEvent);
    } catch (e) {
      appendText("assistant", `(stream error: ${(e as Error).message})`);
    } finally {
      setBusy(false);
    }
  }

  return (
    <section style={{ display: "flex", flexDirection: "column", height: "calc(100vh - 32px)" }}>
      <h2>
        {session.agent_name} — {session.id}
      </h2>
      <div style={{ flex: 1, overflow: "auto", padding: 4 }} aria-label="message list">
        {items.map((it, i) =>
          it.kind === "text" ? (
            <MessageBubble key={i} role={it.role} text={it.text} />
          ) : (
            <ToolCallCard
              key={i}
              tool={it.tool}
              args={it.args}
              status={it.status}
              resultPreview={it.resultPreview}
              errorMessage={it.errorMessage}
            />
          ),
        )}
      </div>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          void send();
        }}
        style={{ display: "flex", gap: 8 }}
      >
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          placeholder="Ask the agent..."
          disabled={busy}
          style={{ flex: 1 }}
        />
        <button type="submit" disabled={busy || !input.trim()}>
          {busy ? "..." : "Send"}
        </button>
      </form>
    </section>
  );
}
