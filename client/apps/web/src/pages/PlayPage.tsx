import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useGame } from "../core/store";
import { Card } from "../components/ui/card";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { cn } from "../lib/cn";
import type { QuestionShowData, QuestionRevealData } from "../core/protocol";

type Choice = { optionId?: string; value?: number | boolean };

export function PlayPage() {
  const { t } = useTranslation();
  const { countdown, question, reveal, answeredIndex, myAnswer, answer, room, selfUserId, eliminated } = useGame();

  if (countdown !== null) {
    return (
      <div className="flex min-h-[70vh] items-center justify-center">
        <div className="text-center">
          <p className="text-slate-500">{t("play.starting")}</p>
          <div className="text-7xl font-bold text-indigo-600">{countdown}</div>
        </div>
      </div>
    );
  }

  if (!question) {
    return <div className="flex min-h-[70vh] items-center justify-center text-slate-400">{t("common.loading")}</div>;
  }

  const revealed = !!reveal && reveal.index === question.index;
  const answered = answeredIndex === question.index;
  const timePerQ = room?.config.timePerQ ?? 15;
  // time_attack — yagona vaqt byudjeti (timePerQ × savol soni); aks holda har savol vaqti.
  const totalMs =
    (room?.config.mode === "time_attack" ? timePerQ * (room?.config.questionCount ?? 1) : timePerQ) * 1000;
  const myChoice: Choice | undefined = myAnswer?.index === question.index ? myAnswer.choice : undefined;
  const iWasRight = revealed ? isMine(question.type, myChoice, reveal!.correct) : false;

  return (
    <div className="mx-auto max-w-2xl space-y-4 p-4">
      <div className="text-sm text-slate-500">{t("play.question", { n: question.index + 1, total: question.total })}</div>

      {eliminated && (
        <div className="rounded-lg bg-red-50 px-4 py-2 text-center text-sm font-medium text-red-600">{t("play.eliminated")}</div>
      )}

      {!revealed && <TimerBar deadlineTs={question.deadlineTs} totalMs={totalMs} />}

      <Card>
        <h2 className="mb-5 text-center text-xl font-semibold">{question.prompt}</h2>
        <QuestionBody question={question} reveal={revealed ? reveal! : null} myChoice={myChoice} disabled={answered || revealed || eliminated} onAnswer={answer} />
        {answered && !revealed && <p className="mt-4 text-center text-sm text-indigo-600">{t("play.accepted")}</p>}
        {revealed && (
          <p className={cn("mt-4 text-center text-sm font-semibold", iWasRight ? "text-green-600" : "text-red-600")}>
            {myChoice ? (iWasRight ? t("play.correct") : t("play.wrong")) : t("play.noAnswer")}
          </p>
        )}
      </Card>

      {revealed && (
        <Card className="space-y-2">
          {reveal!.explanation && <p className="text-sm text-slate-600">{reveal!.explanation}</p>}
          <Leaderboard entries={reveal!.leaderboard} selfId={selfUserId} />
        </Card>
      )}
    </div>
  );
}

function isMine(type: string, mine: Choice | undefined, correct: unknown): boolean {
  if (!mine) return false;
  const c = correct as { optionId?: string; value?: number | boolean };
  if (type === "numeric") return Number(mine.value) === Number(c.value);
  if (type === "true_false") return mine.value === c.value;
  return mine.optionId === c.optionId;
}

function QuestionBody({
  question,
  reveal,
  myChoice,
  disabled,
  onAnswer,
}: {
  question: QuestionShowData;
  reveal: QuestionRevealData | null;
  myChoice: Choice | undefined;
  disabled: boolean;
  onAnswer: (choice: Choice) => void;
}) {
  const { t } = useTranslation();
  const correct = (reveal?.correct ?? {}) as Choice;

  if (question.type === "true_false") {
    return (
      <div className="grid grid-cols-2 gap-3">
        {[
          { key: "play.tfTrue", val: true },
          { key: "play.tfFalse", val: false },
        ].map((b) => (
          <button
            key={b.key}
            disabled={disabled}
            onClick={() => onAnswer({ value: b.val })}
            className={cn(
              "rounded-xl border px-4 py-4 text-sm font-medium transition",
              reveal && correct.value === b.val && "border-green-500 bg-green-50 text-green-700",
              reveal && myChoice?.value === b.val && correct.value !== b.val && "border-red-400 bg-red-50 text-red-700",
              !reveal && "border-slate-300 hover:border-indigo-400 hover:bg-indigo-50",
            )}
          >
            {t(b.key)}
          </button>
        ))}
      </div>
    );
  }

  if (question.type === "numeric") {
    return <NumericBody reveal={reveal} myChoice={myChoice} disabled={disabled} onAnswer={onAnswer} />;
  }

  return (
    <div className="grid gap-3">
      {question.options?.map((o) => {
        const isCorrect = correct.optionId === o.id;
        const isMyWrong = myChoice?.optionId === o.id && !isCorrect;
        return (
          <button
            key={o.id}
            disabled={disabled}
            onClick={() => onAnswer({ optionId: o.id })}
            className={cn(
              "rounded-xl border px-4 py-3 text-left text-sm font-medium transition",
              reveal && isCorrect && "border-green-500 bg-green-50 text-green-700",
              reveal && isMyWrong && "border-red-400 bg-red-50 text-red-700",
              reveal && !isCorrect && !isMyWrong && "border-slate-200 opacity-60",
              !reveal && "border-slate-300 hover:border-indigo-400 hover:bg-indigo-50",
            )}
          >
            {o.text}
          </button>
        );
      })}
    </div>
  );
}

function NumericBody({
  reveal,
  myChoice,
  disabled,
  onAnswer,
}: {
  reveal: QuestionRevealData | null;
  myChoice: Choice | undefined;
  disabled: boolean;
  onAnswer: (choice: Choice) => void;
}) {
  const { t } = useTranslation();
  const [val, setVal] = useState("");
  const correct = (reveal?.correct ?? {}) as { value?: number };
  return (
    <div className="space-y-3">
      <Input
        type="number"
        value={reveal && myChoice ? String(myChoice.value) : val}
        disabled={disabled}
        onChange={(e) => setVal(e.target.value)}
        placeholder={t("play.enterAnswer")}
        className="text-center text-lg"
      />
      {reveal ? (
        <p className="text-center text-sm text-green-700">{t("play.correctAnswer", { val: correct.value })}</p>
      ) : (
        <Button className="w-full" disabled={disabled || val === ""} onClick={() => onAnswer({ value: Number(val) })}>
          {t("play.submitAnswer")}
        </Button>
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
  entries: { userId: string; name: string; score: number; rank: number; eliminated?: boolean }[];
  selfId: string | null;
}) {
  return (
    <div className="space-y-1">
      {entries.map((e) => (
        <div
          key={e.userId}
          className={cn(
            "flex items-center justify-between rounded-lg px-3 py-1.5 text-sm",
            e.eliminated ? "bg-slate-50 text-slate-400 line-through" : e.userId === selfId ? "bg-indigo-50 font-semibold" : "bg-slate-50",
          )}
        >
          <span>
            {e.rank}. {e.name}
            {e.eliminated && <span className="ml-1 no-underline">💀</span>}
          </span>
          <span>{Math.round(e.score)}</span>
        </div>
      ))}
    </div>
  );
}
