import { SessionList } from "@/components/SessionList";
import { useSessionStore } from "@/store/session";
import { ChatView } from "@/components/ChatView";

export function Sessions() {
  const current = useSessionStore((s) => s.current);
  return (
    <main
      style={{ display: "grid", gridTemplateColumns: "320px 1fr", gap: 16, padding: 16 }}
    >
      <SessionList />
      {current ? <ChatView session={current} /> : <p>Select or create a session.</p>}
    </main>
  );
}
