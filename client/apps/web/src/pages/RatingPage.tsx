import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useGame } from "@core/store";
import { api } from "@core/api";
import type { RatingItem, LeaderboardRow } from "@core/api";
import { Card } from "../components/ui/card";
import { cn } from "../lib/cn";

const medals = ["🥇", "🥈", "🥉"];

export function RatingPage() {
  const { t } = useTranslation();
  const token = useGame((s) => s.token);
  const selfName = useGame((s) => s.user?.username);

  const [mine, setMine] = useState<RatingItem[]>([]);
  const [global, setGlobal] = useState<LeaderboardRow[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    Promise.all([
      token ? api.myRating(token).catch(() => []) : Promise.resolve([]),
      api.globalLeaderboard().catch(() => []),
    ])
      .then(([r, g]) => {
        setMine(r);
        setGlobal(g);
      })
      .finally(() => setLoading(false));
  }, [token]);

  return (
    <div className="mx-auto max-w-md space-y-4 p-4">
      {/* Mening reytingim — soha bo'yicha ELO */}
      <Card className="space-y-2">
        <p className="text-sm font-medium text-slate-500">{t("rating.myRating")}</p>
        {loading ? (
          <p className="py-6 text-center text-slate-400">{t("common.loading")}</p>
        ) : mine.length === 0 ? (
          <p className="py-6 text-center text-sm text-slate-400">{t("rating.noRating")}</p>
        ) : (
          mine.map((r) => (
            <div
              key={r.subject}
              className="flex items-center justify-between rounded-xl bg-slate-50 px-4 py-3"
            >
              <span className="font-medium">{r.subjectName}</span>
              <span className="flex items-center gap-3">
                <span className="text-xs text-slate-400">{t("rating.games", { count: r.games })}</span>
                <span className="font-semibold text-indigo-600">{r.rating}</span>
              </span>
            </div>
          ))
        )}
      </Card>

      {/* Global reyting — top 20 */}
      <Card className="space-y-2">
        <p className="text-sm font-medium text-slate-500">{t("rating.global")}</p>
        {loading ? (
          <p className="py-6 text-center text-slate-400">{t("common.loading")}</p>
        ) : (
          global.map((row, i) => (
            <div
              key={row.username}
              className={cn(
                "flex items-center justify-between rounded-xl px-4 py-3",
                row.username === selfName ? "bg-indigo-50 font-semibold" : "bg-slate-50",
              )}
            >
              <span className="flex items-center gap-2">
                <span className="w-6 text-center">{medals[i] ?? i + 1}</span>
                <span>{row.username}</span>
              </span>
              <span className="text-indigo-600">{Math.round(row.totalScore)}</span>
            </div>
          ))
        )}
      </Card>
    </div>
  );
}
