import { Login } from "@/pages/Login";
import { Sessions } from "@/pages/Sessions";
import { useAuthStore } from "@/store/auth";

export function App() {
  const loggedIn = useAuthStore((s) => s.loggedIn);
  return loggedIn ? <Sessions /> : <Login />;
}
export default App;
