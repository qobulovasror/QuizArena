import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useGame } from "../core/store";
import { api } from "../core/api";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Card } from "../components/ui/card";

type Mode = "guest" | "login" | "register";

export function AuthPage() {
  const { t } = useTranslation();
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
      <p className="mb-6 text-center text-sm text-slate-500">{t("auth.tagline")}</p>

      <Card className="space-y-4">
        <div>
          <label className="mb-1 block text-sm font-medium">{t("auth.nameLabel")}</label>
          <Input
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            placeholder={t("auth.namePlaceholder")}
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
              {t(`auth.${m}`)}
            </button>
          ))}
        </div>

        {mode === "register" && (
          <Input value={username} onChange={(e) => setUsername(e.target.value)} placeholder={t("auth.username")} />
        )}
        {mode !== "guest" && (
          <>
            <Input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder={t("auth.email")}
            />
            <Input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder={t("auth.password")}
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
          {busy ? "..." : mode === "guest" ? t("auth.guestBtn") : mode === "login" ? t("auth.loginBtn") : t("auth.registerBtn")}
        </Button>
      </Card>
    </div>
  );
}
