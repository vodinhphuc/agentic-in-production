type Props = { role: "user" | "assistant"; text: string };

export function MessageBubble({ role, text }: Props) {
  const align = role === "user" ? "flex-end" : "flex-start";
  const bg = role === "user" ? "#dbeafe" : "#f3f4f6";
  return (
    <div style={{ display: "flex", justifyContent: align, margin: "4px 0" }}>
      <div style={{ background: bg, padding: "8px 12px", borderRadius: 8, maxWidth: "70%" }}>
        {text}
      </div>
    </div>
  );
}
