import type { components } from "./types.gen";

type LoginRequest = components["schemas"]["LoginRequest"];
type LoginResponse = components["schemas"]["LoginResponse"];
type Agent = components["schemas"]["Agent"];
type Session = components["schemas"]["Session"];
type Message = components["schemas"]["Message"];
type AuditEntry = components["schemas"]["AuditEntry"];

const BASE = import.meta.env.VITE_API_BASE_URL ?? "";

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    credentials: "include",
    headers: { "content-type": "application/json", ...(init?.headers ?? {}) },
    ...init,
  });
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}: ${await res.text()}`);
  return res.json() as Promise<T>;
}

export const api = {
  healthz: () => req<{ ok: boolean; version?: string }>("/api/healthz"),
  login: (body: LoginRequest) =>
    req<LoginResponse>("/api/auth/login", { method: "POST", body: JSON.stringify(body) }),
  agents: () => req<Agent[]>("/api/agents"),
  createSession: (agent_name: string) =>
    req<Session>("/api/sessions", { method: "POST", body: JSON.stringify({ agent_name }) }),
  listMessages: (id: string) => req<Message[]>(`/api/sessions/${id}/messages`),
  audit: (id: string) => req<AuditEntry[]>(`/api/sessions/${id}/audit`),
};

// SSE message-send returns a ReadableStream of decoded text lines.
export function sendMessage(sessionID: string, text: string, signal: AbortSignal) {
  return fetch(`${BASE}/api/sessions/${sessionID}/messages`, {
    method: "POST",
    credentials: "include",
    headers: { "content-type": "application/json", accept: "text/event-stream" },
    body: JSON.stringify({ text }),
    signal,
  });
}
