import { useState } from "react";
import { api } from "@/api/client";
import { useAuthStore } from "@/store/auth";

export function LoginForm() {
  const setLoggedIn = useAuthStore((s) => s.setLoggedIn);
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr(null);
    setBusy(true);
    try {
      await api.login({ username, password });
      setLoggedIn(username);
    } catch (e) {
      setErr(e instanceof Error ? e.message : "login failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <form
      onSubmit={submit}
      aria-label="login form"
      style={{ display: "grid", gap: 8, maxWidth: 320 }}
    >
      <label>
        Username
        <input value={username} onChange={(e) => setUsername(e.target.value)} />
      </label>
      <label>
        Password
        <input type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
      </label>
      <button type="submit" disabled={busy}>
        {busy ? "Signing in..." : "Sign in"}
      </button>
      {err && (
        <div role="alert" style={{ color: "crimson" }}>
          {err}
        </div>
      )}
    </form>
  );
}
