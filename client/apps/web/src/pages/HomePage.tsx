import { useState } from "react";
import { LobbyPage } from "./LobbyPage";
import { PracticePage } from "./PracticePage";
import { cn } from "../lib/cn";

const tabs = [
  { id: "compete", label: "🏆 O'ynash" },
  { id: "practice", label: "📚 O'rganish" },
] as const;

export function HomePage() {
  const [tab, setTab] = useState<"compete" | "practice">("compete");
  return (
    <div>
      <div className="mx-auto max-w-md px-4 pt-4">
        <div className="flex gap-1 rounded-xl bg-slate-100 p-1">
          {tabs.map((t) => (
            <button
              key={t.id}
              onClick={() => setTab(t.id)}
              className={cn(
                "flex-1 rounded-lg py-2 text-sm font-medium transition",
                tab === t.id ? "bg-white text-indigo-600 shadow" : "text-slate-500",
              )}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>
      {tab === "compete" ? <LobbyPage /> : <PracticePage />}
    </div>
  );
}
