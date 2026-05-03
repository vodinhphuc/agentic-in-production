import { SessionList } from "@/components/SessionList";
import { useSessionStore } from "@/store/session";
import { ChatView } from "@/components/ChatView";
import { AuditLogView } from "@/components/AuditLogView";

export function Sessions() {
  const current = useSessionStore((s) => s.current);
  return (
    <main
      style={{ display: "grid", gridTemplateColumns: "320px 1fr", gap: 16, padding: 16 }}
    >
      <SessionList />
      {current ? (
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          <ChatView session={current} />
          <AuditLogView sessionID={current.id} />
        </div>
      ) : (
        <p>Select or create a session.</p>
      )}
    </main>
  );
}
