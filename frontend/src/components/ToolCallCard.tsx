import { useState } from "react";

type Props = {
  tool: string;
  args?: unknown;
  status: "pending" | "ok" | "error";
  resultPreview?: string;
  errorMessage?: string;
};

export function ToolCallCard({ tool, args, status, resultPreview, errorMessage }: Props) {
  const [open, setOpen] = useState(false);
  const colour = status === "pending" ? "#f59e0b" : status === "ok" ? "#10b981" : "#ef4444";
  return (
    <div
      style={{
        border: `1px solid ${colour}`,
        borderRadius: 6,
        margin: "6px 0",
        padding: "6px 10px",
        background: "#fafafa",
      }}
    >
      <button
        onClick={() => setOpen(!open)}
        aria-expanded={open}
        style={{
          background: "none",
          border: "none",
          padding: 0,
          fontWeight: 600,
          cursor: "pointer",
        }}
      >
        {open ? "▼" : "▶"} {tool} <span style={{ color: colour }}>· {status}</span>
      </button>
      {open && (
        <div style={{ marginTop: 6 }}>
          {args !== undefined && (
            <details open>
              <summary>args</summary>
              <pre style={{ fontSize: 12 }}>{JSON.stringify(args, null, 2)}</pre>
            </details>
          )}
          {resultPreview && (
            <details open>
              <summary>result</summary>
              <pre style={{ fontSize: 12 }}>{resultPreview}</pre>
            </details>
          )}
          {errorMessage && <div style={{ color: "crimson" }}>{errorMessage}</div>}
        </div>
      )}
    </div>
  );
}
