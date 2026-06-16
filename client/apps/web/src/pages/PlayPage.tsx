import { useEffect, useState } from "react";
import { useGame } from "../core/store";
import { Card } from "../components/ui/card";
import { cn } from "../lib/cn";

export function PlayPage() {
  const { countdown, question, reveal, answeredIndex, answer, room } = useGame();

  if (countdown !== null) {
    return (
      <div className="flex min-h-full items-center justify-center">
        <div className="text-center">
          <p className="text-slate-500">Boshlanmoqda…</p>
          <div className="text-7xl font-bold text-indigo-600">{countdown}</div>
        </div>
      </div>
    );
  }

  if (!question) {
    return <div className="flex min-h-full items-center justify-center text-slate-400">Yuklanmoqda…</div>;
  }

  const revealed = !!reveal && reveal.index === question.index;
  const correctId = revealed ? (reveal!.correct as { optionId?: string }).optionId : undefined;
  const answered = answeredIndex === question.index;
  const timePerQ = room?.config.timePerQ ?? 15;

  return (
    <div className="mx-auto max-w-2xl space-y-4 p-4">
      <div className="flex items-center justify-between text-sm text-slate-500">
        <span>
          Savol {question.index + 1} / {question.total}
        </span>
      </div>

      {!revealed && <TimerBar deadlineTs={question.deadlineTs} totalMs={timePerQ * 1000} />}

      <Card>
        <h2 className="mb-5 text-center text-xl font-semibold">{question.prompt}</h2>
        <div className="grid gap-3">
          {question.options?.map((o) => {
            const isCorrect = correctId === o.id;
            return (
              <button
                key={o.id}
                disabled={answered || revealed}
                onClick={() => answer(o.id)}
                className={cn(
                  "rounded-xl border px-4 py-3 text-left text-sm font-medium transition",
                  revealed && isCorrect && "border-green-500 bg-green-50 text-green-700",
                  revealed && !isCorrect && "border-slate-200 opacity-60",
                  !revealed && "border-slate-300 hover:border-indigo-400 hover:bg-indigo-50",
                  answered && !revealed && "opacity-70",
                )}
              >
                {o.text}
              </button>
            );
          })}
        </div>
        {answered && !revealed && (
          <p className="mt-4 text-center text-sm text-indigo-600">Javob qabul qilindi ✓</p>
        )}
      </Card>

      {revealed && (
        <Card className="space-y-2">
          {reveal!.explanation && <p className="text-sm text-slate-600">{reveal!.explanation}</p>}
          <Leaderboard entries={reveal!.leaderboard} selfId={useGame.getState().selfUserId} />
        </Card>
      )}
    </div>
  );
}

function TimerBar({ deadlineTs, totalMs }: { deadlineTs: number; totalMs: number }) {
  const [now, setNow] = useState(Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 100);
    return () => clearInterval(id);
  }, []);
  const remaining = Math.max(0, deadlineTs - now);
  const pct = Math.min(100, (remaining / totalMs) * 100);
  return (
    <div className="h-2 w-full overflow-hidden rounded-full bg-slate-200">
      <div
        className={cn("h-full transition-all duration-100", pct < 30 ? "bg-red-500" : "bg-indigo-500")}
        style={{ width: `${pct}%` }}
      />
    </div>
  );
}

function Leaderboard({
  entries,
  selfId,
}: {
  entries: { userId: string; name: string; score: number; rank: number }[];
  selfId: string | null;
}) {
  return (
    <div className="space-y-1">
      {entries.map((e) => (
        <div
          key={e.userId}
          className={cn(
            "flex items-center justify-between rounded-lg px-3 py-1.5 text-sm",
            e.userId === selfId ? "bg-indigo-50 font-semibold" : "bg-slate-50",
          )}
        >
          <span>
            {e.rank}. {e.name}
          </span>
          <span>{Math.round(e.score)}</span>
        </div>
      ))}
    </div>
  );
}
