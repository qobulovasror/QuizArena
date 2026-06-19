import { useState } from "react";
import { useGame } from "../core/store";
import { LobbyPage } from "./LobbyPage";
import { PracticePage } from "./PracticePage";
import { AssessPage } from "./AssessPage";
import { AdminPage } from "./AdminPage";
import { cn } from "../lib/cn";

type Tab = "compete" | "practice" | "assess" | "admin";

const baseTabs: { id: Tab; label: string }[] = [
  { id: "compete", label: "🏆 O'ynash" },
  { id: "practice", label: "📚 O'rganish" },
  { id: "assess", label: "📊 Baholash" },
];

export function HomePage() {
  const isAdmin = useGame((s) => s.user?.role === "admin");
  const [tab, setTab] = useState<Tab>("compete");

  const tabs = isAdmin ? [...baseTabs, { id: "admin" as Tab, label: "🛠 Admin" }] : baseTabs;

  return (
    <div>
      <div className="mx-auto max-w-md px-4 pt-4">
        <div className="flex gap-1 rounded-xl bg-slate-100 p-1">
          {tabs.map((t) => (
            <button
              key={t.id}
              onClick={() => setTab(t.id)}
              className={cn(
                "flex-1 rounded-lg py-2 text-xs font-medium transition",
                tab === t.id ? "bg-white text-indigo-600 shadow" : "text-slate-500",
              )}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>
      {tab === "compete" && <LobbyPage />}
      {tab === "practice" && <PracticePage />}
      {tab === "assess" && <AssessPage />}
      {tab === "admin" && <AdminPage />}
    </div>
  );
}
