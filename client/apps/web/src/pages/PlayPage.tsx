import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useGame } from "../core/store";
import type { AnswerChoice } from "../core/store";
import { Card } from "../components/ui/card";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { cn } from "../lib/cn";
import type { QuestionShowData, QuestionRevealData, Option } from "../core/protocol";

type Choice = AnswerChoice;

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
  const iWasRight = revealed && reveal ? isMine(question.type, myChoice, reveal.correct) : false;

  return (
    <div className="mx-auto max-w-2xl space-y-4 p-4">
      <div className="text-sm text-slate-500">{t("play.question", { n: question.index + 1, total: question.total })}</div>

      {eliminated && (
        <div className="rounded-lg bg-red-50 px-4 py-2 text-center text-sm font-medium text-red-600">{t("play.eliminated")}</div>
      )}

      {!revealed && <TimerBar deadlineTs={question.deadlineTs} totalMs={totalMs} />}

      <Card>
        {question.type !== "cloze" && <h2 className="mb-5 text-center text-xl font-semibold">{question.prompt}</h2>}
        <QuestionBody key={question.index} question={question} reveal={revealed && reveal ? reveal : null} myChoice={myChoice} disabled={answered || revealed || eliminated} onAnswer={answer} />
        {answered && !revealed && <p className="mt-4 text-center text-sm text-indigo-600">{t("play.accepted")}</p>}
        {revealed && (
          <p className={cn("mt-4 text-center text-sm font-semibold", iWasRight ? "text-green-600" : "text-red-600")}>
            {myChoice ? (iWasRight ? t("play.correct") : t("play.wrong")) : t("play.noAnswer")}
          </p>
        )}
      </Card>

      {revealed && reveal && (
        <Card className="space-y-2">
          {["ordering", "cloze", "match", "categorize"].includes(question.type) && (
            <CorrectAnswer question={question} correct={reveal.correct} />
          )}
          {reveal.explanation && <p className="text-sm text-slate-600">{reveal.explanation}</p>}
          <Leaderboard entries={reveal.leaderboard} selfId={selfUserId} />
        </Card>
      )}
    </div>
  );
}

function isMine(type: string, mine: Choice | undefined, correct: unknown): boolean {
  if (!mine) return false;
  const c = (correct ?? {}) as {
    optionId?: string;
    value?: number | boolean;
    order?: string[];
    pairs?: Record<string, string>;
    assign?: Record<string, string>;
    blanks?: { accepted: string[] }[];
    accepted?: string[];
  };
  const norm = (s: string) => s.trim().toLowerCase();
  switch (type) {
    case "numeric":
      return Number(mine.value) === Number(c.value);
    case "type_answer":
    case "fill_blank":
    case "anagram":
      return (c.accepted ?? []).some((x) => norm(x) === norm(mine.text ?? ""));
    case "true_false":
      return mine.value === c.value;
    case "ordering": {
      const a = mine.order ?? [];
      const b = c.order ?? [];
      return a.length === b.length && a.every((x, i) => x === b[i]);
    }
    case "match": {
      const a = mine.pairs ?? {};
      const b = c.pairs ?? {};
      const bk = Object.keys(b);
      return Object.keys(a).length === bk.length && bk.every((k) => a[k] === b[k]);
    }
    case "categorize": {
      const a = mine.assign ?? {};
      const b = c.assign ?? {};
      const bk = Object.keys(b);
      return Object.keys(a).length === bk.length && bk.every((k) => a[k] === b[k]);
    }
    case "cloze": {
      const a = mine.blanks ?? [];
      const b = c.blanks ?? [];
      return a.length === b.length && b.every((bl, i) => bl.accepted.some((x) => norm(x) === norm(a[i] ?? "")));
    }
    default:
      return mine.optionId === c.optionId;
  }
}

// CorrectAnswer — ordering/cloze/match/categorize uchun to'g'ri javobni matn ko'rinishida ko'rsatadi.
function CorrectAnswer({ question, correct }: { question: QuestionShowData; correct: unknown }) {
  const { t } = useTranslation();
  const c = (correct ?? {}) as {
    order?: string[];
    pairs?: Record<string, string>;
    assign?: Record<string, string>;
    blanks?: { accepted: string[] }[];
  };
  const optText = (id: string) => question.options?.find((o) => o.id === id)?.text ?? id;
  const tgtText = (id: string) => question.targets?.find((o) => o.id === id)?.text ?? id;

  let rows: string[] = [];
  if (question.type === "ordering" && c.order) {
    rows = c.order.map((id, i) => `${i + 1}. ${optText(id)}`);
  } else if (question.type === "cloze" && c.blanks) {
    rows = c.blanks.map((b, i) => `${i + 1}. ${b.accepted.join(" / ")}`);
  } else if (question.type === "match" && c.pairs) {
    rows = Object.entries(c.pairs).map(([l, r]) => `${optText(l)} → ${tgtText(r)}`);
  } else if (question.type === "categorize" && c.assign) {
    rows = Object.entries(c.assign).map(([item, cat]) => `${optText(item)} → ${tgtText(cat)}`);
  }
  if (rows.length === 0) return null;

  return (
    <div className="rounded-xl bg-green-50 px-4 py-3">
      <p className="mb-1 text-sm font-semibold text-green-700">{t("play.correctLabel")}</p>
      <ul className="space-y-0.5 text-sm text-slate-700">
        {rows.map((r, i) => (
          <li key={i}>{r}</li>
        ))}
      </ul>
    </div>
  );
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

  // type_answer/fill_blank/anagram — matn yoziladi (anagram: harflar prompt'da).
  if (["type_answer", "fill_blank", "anagram"].includes(question.type)) {
    return <TextBody reveal={reveal} myChoice={myChoice} disabled={disabled} onAnswer={onAnswer} />;
  }

  if (question.type === "ordering") {
    return <OrderingBody items={question.options ?? []} disabled={disabled} onAnswer={onAnswer} />;
  }
  if (question.type === "cloze") {
    return <ClozeBody prompt={question.prompt} disabled={disabled} onAnswer={onAnswer} />;
  }
  if (question.type === "match") {
    return <AssignBody items={question.options ?? []} targets={question.targets ?? []} answerKey="pairs" disabled={disabled} onAnswer={onAnswer} />;
  }
  if (question.type === "categorize") {
    return <AssignBody items={question.options ?? []} targets={question.targets ?? []} answerKey="assign" disabled={disabled} onAnswer={onAnswer} />;
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

// TextBody — type_answer/fill_blank/anagram uchun matnli javob (NumericBody uslubida).
function TextBody({
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
  const correct = (reveal?.correct ?? {}) as { accepted?: string[] };
  return (
    <div className="space-y-3">
      <Input
        value={reveal && myChoice ? String(myChoice.text ?? "") : val}
        disabled={disabled}
        onChange={(e) => setVal(e.target.value)}
        placeholder={t("play.enterAnswer")}
        className="text-center text-lg"
      />
      {reveal ? (
        <p className="text-center text-sm text-green-700">{t("play.correctAnswer", { val: (correct.accepted ?? []).join(" / ") })}</p>
      ) : (
        <Button className="w-full" disabled={disabled || val.trim() === ""} onClick={() => onAnswer({ text: val })}>
          {t("play.submitAnswer")}
        </Button>
      )}
    </div>
  );
}

function OrderingBody({ items, disabled, onAnswer }: { items: Option[]; disabled: boolean; onAnswer: (c: Choice) => void }) {
  const { t } = useTranslation();
  const [order, setOrder] = useState<Option[]>(items);
  const move = (i: number, dir: -1 | 1) => {
    const j = i + dir;
    if (j < 0 || j >= order.length) return;
    const next = order.slice();
    [next[i], next[j]] = [next[j], next[i]];
    setOrder(next);
  };
  return (
    <div className="space-y-2">
      {order.map((o, i) => (
        <div key={o.id} className="flex items-center gap-2 rounded-xl border border-slate-300 px-3 py-2 text-sm">
          <span className="w-5 text-slate-400">{i + 1}.</span>
          <span className="flex-1">{o.text}</span>
          <button disabled={disabled || i === 0} onClick={() => move(i, -1)} className="px-1.5 text-slate-500 disabled:opacity-30">▲</button>
          <button disabled={disabled || i === order.length - 1} onClick={() => move(i, 1)} className="px-1.5 text-slate-500 disabled:opacity-30">▼</button>
        </div>
      ))}
      <Button className="w-full" disabled={disabled} onClick={() => onAnswer({ order: order.map((o) => o.id) })}>
        {t("play.submitAnswer")}
      </Button>
    </div>
  );
}

function ClozeBody({ prompt, disabled, onAnswer }: { prompt: string; disabled: boolean; onAnswer: (c: Choice) => void }) {
  const { t } = useTranslation();
  const parts = prompt.split("___");
  const blankCount = parts.length - 1;
  const [vals, setVals] = useState<string[]>(Array(blankCount).fill(""));
  return (
    <div className="space-y-4">
      <p className="text-center text-lg leading-9">
        {parts.map((p, i) => (
          <span key={i}>
            {p}
            {i < blankCount && (
              <input
                disabled={disabled}
                value={vals[i]}
                onChange={(e) => setVals(vals.map((v, j) => (j === i ? e.target.value : v)))}
                className="mx-1 w-24 border-b-2 border-indigo-400 text-center outline-none disabled:opacity-60"
              />
            )}
          </span>
        ))}
      </p>
      <Button className="w-full" disabled={disabled || vals.some((v) => v.trim() === "")} onClick={() => onAnswer({ blanks: vals })}>
        {t("play.submitAnswer")}
      </Button>
    </div>
  );
}

// AssignBody — match (chap→o'ng) va categorize (element→toifa) uchun umumiy.
function AssignBody({
  items,
  targets,
  answerKey,
  disabled,
  onAnswer,
}: {
  items: Option[];
  targets: Option[];
  answerKey: "pairs" | "assign";
  disabled: boolean;
  onAnswer: (c: Choice) => void;
}) {
  const { t } = useTranslation();
  const [map, setMap] = useState<Record<string, string>>({});
  const complete = items.length > 0 && items.every((it) => map[it.id]);
  return (
    <div className="space-y-2">
      {items.map((it) => (
        <div key={it.id} className="flex items-center gap-2">
          <span className="flex-1 rounded-lg bg-slate-50 px-3 py-2 text-sm">{it.text}</span>
          <span className="text-slate-400">→</span>
          <select
            disabled={disabled}
            value={map[it.id] ?? ""}
            onChange={(e) => setMap({ ...map, [it.id]: e.target.value })}
            className="flex-1 rounded-lg border border-slate-300 px-2 py-2 text-sm disabled:opacity-60"
          >
            <option value="">—</option>
            {targets.map((tg) => (
              <option key={tg.id} value={tg.id}>
                {tg.text}
              </option>
            ))}
          </select>
        </div>
      ))}
      <Button className="w-full" disabled={disabled || !complete} onClick={() => onAnswer({ [answerKey]: map })}>
        {t("play.submitAnswer")}
      </Button>
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
