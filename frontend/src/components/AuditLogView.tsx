import { useCallback, useEffect, useState } from "react";
import { api } from "@/api/client";
import type { components } from "@/api/types.gen";

type AuditEntry = components["schemas"]["AuditEntry"];

export function AuditLogView({ sessionID }: { sessionID: string }) {
  const [rows, setRows] = useState<AuditEntry[]>([]);
  const [err, setErr] = useState<string | null>(null);

  const refresh = useCallback(() => {
    let alive = true;
    void api
      .audit(sessionID)
      .then((r) => {
        if (alive) setRows(r ?? []);
      })
      .catch((e) => {
        if (alive) setErr(String(e));
      });
    return () => {
      alive = false;
    };
  }, [sessionID]);

  useEffect(() => refresh(), [refresh]);

  if (err) return <div role="alert">Audit error: {err}</div>;

  return (
    <details onToggle={(e) => e.currentTarget.open && refresh()}>
      <summary>Audit log ({rows.length})</summary>
      <table style={{ fontSize: 12, width: "100%" }}>
        <thead>
          <tr>
            <th>time</th>
            <th>kind</th>
            <th>run_id</th>
            <th>payload</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r) => (
            <tr key={r.id}>
              <td>{new Date(r.occurred_at).toLocaleTimeString()}</td>
              <td>{r.kind}</td>
              <td>{r.run_id}</td>
              <td>
                <code>{JSON.stringify(r.payload).slice(0, 80)}…</code>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </details>
  );
}
