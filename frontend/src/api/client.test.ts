import { describe, it, expect, vi } from "vitest";
import { api } from "./client";

describe("api.login", () => {
  it("posts JSON to /api/auth/login", async () => {
    const json = vi.fn().mockResolvedValue({ token: "tok" });
    const fetchMock = vi.fn().mockResolvedValue({ ok: true, json });
    vi.stubGlobal("fetch", fetchMock);
    const res = await api.login({ username: "admin", password: "x" });
    expect(res).toEqual({ token: "tok" });
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining("/api/auth/login"),
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("throws on non-2xx", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        ok: false,
        status: 401,
        statusText: "Unauthorized",
        text: () => Promise.resolve("bad"),
      }),
    );
    await expect(api.login({ username: "x", password: "x" })).rejects.toThrow(/401/);
  });
});
