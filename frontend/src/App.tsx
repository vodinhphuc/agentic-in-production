import { Login } from "@/pages/Login";
import { useAuthStore } from "@/store/auth";

export function App() {
  const loggedIn = useAuthStore((s) => s.loggedIn);
  if (!loggedIn) return <Login />;
  return (
    <main style={{ padding: 24 }}>
      <h1>agentic-in-production</h1>
      <p>(home — implemented in Task 22)</p>
    </main>
  );
}
export default App;
