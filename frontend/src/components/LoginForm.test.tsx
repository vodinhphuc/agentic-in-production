import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { LoginForm } from "./LoginForm";

describe("LoginForm", () => {
  it("calls /api/auth/login on submit", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ token: "tok" }),
    });
    vi.stubGlobal("fetch", fetchMock);
    render(<LoginForm />);
    fireEvent.change(screen.getByLabelText(/password/i), { target: { value: "hunter2" } });
    fireEvent.click(screen.getByRole("button", { name: /sign in/i }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalled());
  });
});
