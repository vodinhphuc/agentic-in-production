import { describe, it, expect } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { ToolCallCard } from "./ToolCallCard";

describe("ToolCallCard", () => {
  it("renders collapsed by default and expands on click", () => {
    render(
      <ToolCallCard
        tool="execute_query"
        status="ok"
        args={{ sql: "SELECT 1" }}
        resultPreview="1 row"
      />,
    );
    const btn = screen.getByRole("button", { expanded: false });
    expect(screen.queryByText(/SELECT 1/)).toBeNull();
    fireEvent.click(btn);
    expect(screen.getByText(/SELECT 1/)).toBeInTheDocument();
    expect(screen.getByText(/1 row/)).toBeInTheDocument();
  });
});
