import { useState } from "react";
import { useGame } from "../core/store";
import { api } from "../core/api";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Card } from "../components/ui/card";

type Mode = "guest" | "login" | "register";

export function AuthPage() {
  const { displayName, setDisplayName, setAuth, connect } = useGame();
  const [mode, setMode] = useState<Mode>("guest");
  const [email, setEmail] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  async function go(action: () => Promise<{ token: string; user: any }>) {
    setBusy(true);
    setErr(null);
    try {
      const res = await action();
      setAuth(res.token, res.user);
      connect();
    } catch (e) {
      setErr(e instanceof Error ? e.message : "xato");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto flex min-h-full max-w-md flex-col justify-center p-4">
      <h1 className="mb-1 text-center text-3xl font-bold text-indigo-600">QuizArena</h1>
      <p className="mb-6 text-center text-sm text-slate-500">Real-time bilim musobaqasi</p>

      <Card className="space-y-4">
        <div>
          <label className="mb-1 block text-sm font-medium">Ismingiz (o'yinda)</label>
          <Input
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            placeholder="Masalan: Ali"
          />
        </div>

        <div className="flex gap-2 text-sm">
          {(["guest", "login", "register"] as Mode[]).map((m) => (
            <button
              key={m}
              onClick={() => setMode(m)}
              className={
                "flex-1 rounded-lg py-1.5 font-medium " +
                (mode === m ? "bg-indigo-100 text-indigo-700" : "text-slate-500 hover:bg-slate-100")
              }
            >
              {m === "guest" ? "Mehmon" : m === "login" ? "Kirish" : "Ro'yxat"}
            </button>
          ))}
        </div>

        {mode === "register" && (
          <Input value={username} onChange={(e) => setUsername(e.target.value)} placeholder="Username" />
        )}
        {mode !== "guest" && (
          <>
            <Input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="Email"
            />
            <Input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Parol"
            />
          </>
        )}

        {err && <p className="text-sm text-red-600">{err}</p>}

        <Button
          className="w-full"
          disabled={busy || (mode !== "guest" && (!email || !password))}
          onClick={() => {
            if (mode === "guest") go(() => api.guest());
            else if (mode === "login") go(() => api.login(email, password));
            else go(() => api.register(username, email, password));
          }}
        >
          {busy ? "..." : mode === "guest" ? "Mehmon sifatida o'ynash" : mode === "login" ? "Kirish" : "Ro'yxatdan o'tish"}
        </Button>
      </Card>
    </div>
  );
}
