import { useEffect, useState } from "react";
import { api } from "@/api/client";
import { useSessionStore } from "@/store/session";
import type { components } from "@/api/types.gen";

type Agent = components["schemas"]["Agent"];

export function SessionList() {
  const { sessions, setSessions, setCurrent } = useSessionStore();
  const [agents, setAgents] = useState<Agent[]>([]);
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    void api
      .agents()
      .then(setAgents)
      .catch(() => setAgents([]));
  }, []);

  async function newSession(name: string) {
    setBusy(true);
    try {
      const s = await api.createSession(name);
      setSessions([s, ...sessions]);
      setCurrent(s);
    } finally {
      setBusy(false);
    }
  }

  return (
    <section>
      <h2>Sessions</h2>
      <div style={{ display: "flex", gap: 8 }}>
        {agents.map((a) => (
          <button key={a.name} onClick={() => newSession(a.name)} disabled={busy || !a.enabled}>
            New session: {a.name}
          </button>
        ))}
      </div>
      <ul>
        {sessions.map((s) => (
          <li key={s.id}>
            <button onClick={() => setCurrent(s)}>
              {s.id} · {s.agent_name} · {new Date(s.created_at).toLocaleString()}
            </button>
          </li>
        ))}
      </ul>
    </section>
  );
}
