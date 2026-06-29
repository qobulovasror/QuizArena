import { useTranslation } from "react-i18next";
import { useGame } from "@core/store";
import { Button } from "../components/ui/button";
import { Card } from "../components/ui/card";
import { cn } from "../lib/cn";

const medals = ["🥇", "🥈", "🥉"];

export function ResultPage() {
  const { t } = useTranslation();
  const { gameOver, selfUserId, newGame } = useGame();
  const board = gameOver?.finalLeaderboard ?? [];
  const teams = gameOver?.teams ?? [];

  return (
    <div className="mx-auto max-w-md space-y-4 p-4">
      <h1 className="text-center text-2xl font-bold">{t("result.gameOver")}</h1>

      {teams.length > 0 && (
        <Card className="space-y-2">
          <p className="text-sm font-medium text-slate-500">{t("result.teams")}</p>
          {teams.map((s) => (
            <div
              key={s.team}
              className="flex items-center justify-between rounded-xl bg-slate-50 px-4 py-3"
            >
              <span className="flex items-center gap-2">
                <span className="w-6 text-center">{medals[s.rank - 1] ?? s.rank}</span>
                <span className="font-medium">{t("result.team", { team: s.team })}</span>
              </span>
              <span className="text-indigo-600">{Math.round(s.score)}</span>
            </div>
          ))}
        </Card>
      )}

      <Card className="space-y-2">
        {board.map((e) => (
          <div
            key={e.userId}
            className={cn(
              "flex items-center justify-between rounded-xl px-4 py-3",
              e.eliminated ? "bg-slate-50 text-slate-400" : e.userId === selfUserId ? "bg-indigo-50 font-semibold" : "bg-slate-50",
            )}
          >
            <span className="flex items-center gap-2">
              <span className="w-6 text-center">{e.eliminated ? "💀" : (medals[e.rank - 1] ?? e.rank)}</span>
              <span className={cn(e.eliminated && "line-through")}>{e.name}</span>
              {e.team && (
                <span className="rounded-full bg-slate-200 px-2 py-0.5 text-xs font-medium text-slate-600">
                  {e.team}
                </span>
              )}
            </span>
            <span className="text-indigo-600">{Math.round(e.score)}</span>
          </div>
        ))}
      </Card>
      <Button className="w-full" onClick={newGame}>
        {t("result.newGame")}
      </Button>
    </div>
  );
}
