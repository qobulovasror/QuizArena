import { useGame } from "../core/store";
import { Button } from "../components/ui/button";
import { Card } from "../components/ui/card";
import { cn } from "../lib/cn";

const medals = ["🥇", "🥈", "🥉"];

export function ResultPage() {
  const { gameOver, selfUserId, newGame } = useGame();
  const board = gameOver?.finalLeaderboard ?? [];

  return (
    <div className="mx-auto max-w-md space-y-4 p-4">
      <h1 className="text-center text-2xl font-bold">O'yin tugadi 🎉</h1>
      <Card className="space-y-2">
        {board.map((e) => (
          <div
            key={e.userId}
            className={cn(
              "flex items-center justify-between rounded-xl px-4 py-3",
              e.userId === selfUserId ? "bg-indigo-50 font-semibold" : "bg-slate-50",
            )}
          >
            <span className="flex items-center gap-2">
              <span className="w-6 text-center">{medals[e.rank - 1] ?? e.rank}</span>
              {e.name}
            </span>
            <span className="text-indigo-600">{Math.round(e.score)}</span>
          </div>
        ))}
      </Card>
      <Button className="w-full" onClick={newGame}>
        Yangi o'yin
      </Button>
    </div>
  );
}
