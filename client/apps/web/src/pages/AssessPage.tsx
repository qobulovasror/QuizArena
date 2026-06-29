import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useGame } from "../core/store";
import { api } from "../core/api";
import type { MasteryItem, AssessQuestion, AssessAnswer } from "../core/api";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Card } from "../components/ui/card";
import { cn } from "../lib/cn";

type View = "overview" | "quiz" | "result";

export function AssessPage() {
  const { t } = useTranslation();
  const token = useGame((s) => s.token);
  const subjects = useGame((s) => s.subjects);
  const loadSubjects = useGame((s) => s.loadSubjects);

  const [subject, setSubject] = useState("general");
  const [view, setView] = useState<View>("overview");
  const [items, setItems] = useState<MasteryItem[]>([]);
  const [questions, setQuestions] = useState<AssessQuestion[]>([]);
  const [idx, setIdx] = useState(0);
  const [answers, setAnswers] = useState<AssessAnswer[]>([]);
  const [score, setScore] = useState({ correct: 0, total: 0 });
  const [msg, setMsg] = useState("");

  const loadMastery = useCallback(async () => {
    if (!token) return;
    try {
      setItems(await api.mastery(token));
    } catch {
      setItems([]);
    }
  }, [token]);

  useEffect(() => {
    loadSubjects();
    loadMastery();
  }, [loadSubjects, loadMastery]);

  async function start() {
    if (!token) return;
    setMsg("");
    try {
      const qs = await api.assessQuestions(subject, token);
      if (qs.length === 0) {
        setMsg(t("assess.noQuestions"));
        return;
      }
      setQuestions(qs);
      setIdx(0);
      setAnswers([]);
      setView("quiz");
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "xato");
    }
  }

  async function answer(choice: unknown) {
    const q = questions[idx];
    const next = [...answers, { questionId: q.questionId, choice }];
    setAnswers(next);
    if (idx + 1 < questions.length) {
      setIdx(idx + 1);
    } else if (token) {
      try {
        const res = await api.assessSubmit(next, token);
        setScore(res);
        setView("result");
        loadMastery();
      } catch (e) {
        setMsg(e instanceof Error ? e.message : "xato");
      }
    }
  }

  const list = subjects.length > 0 ? subjects : [{ slug: "general", name: "Umumiy bilim", icon: "🌍" }];

  if (view === "quiz") {
    const q = questions[idx];
    return (
      <div className="mx-auto max-w-md space-y-4 p-4">
        <p className="text-center text-xs text-slate-400">{t("assess.testProgress", { n: idx + 1, total: questions.length })}</p>
        <Card>
          <h2 className="mb-5 text-center text-xl font-semibold">{q.prompt}</h2>
          <AnswerInput q={q} onAnswer={answer} />
          {msg && <p className="mt-4 text-center text-sm text-red-600">{msg}</p>}
        </Card>
      </div>
    );
  }

  if (view === "result") {
    const pct = Math.round((score.correct / Math.max(1, score.total)) * 100);
    return (
      <div className="mx-auto max-w-md space-y-4 p-4">
        <Card className="space-y-3 text-center">
          <h2 className="text-2xl font-bold">{t("assess.result")}</h2>
          <div className="text-4xl font-bold text-indigo-600">
            {score.correct} / {score.total}
          </div>
          <p className="text-sm text-slate-500">{t("assess.percentCorrect", { pct })}</p>
          <Button className="w-full" onClick={() => setView("overview")}>
            {t("assess.back")}
          </Button>
        </Card>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-md space-y-4 p-4">
      <Card className="space-y-3">
        <h2 className="font-semibold">{t("assess.yourLevel")}</h2>
        {items.length === 0 ? (
          <p className="text-sm text-slate-400">{t("assess.noTests")}</p>
        ) : (
          items.map((m) => (
            <div key={m.subject + m.category} className="space-y-1">
              <div className="flex justify-between text-sm">
                <span>
                  {m.subject} · {m.category}
                </span>
                <span className="font-medium">{Math.round(m.mastery)}%</span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-slate-200">
                <div
                  className={cn("h-full", m.mastery >= 70 ? "bg-green-500" : m.mastery >= 40 ? "bg-amber-500" : "bg-red-500")}
                  style={{ width: `${m.mastery}%` }}
                />
              </div>
            </div>
          ))
        )}
      </Card>

      <Card className="space-y-3">
        <h2 className="font-semibold">{t("assess.takeTest")}</h2>
        <div className="flex flex-wrap gap-2">
          {list.map((s) => (
            <button
              key={s.slug}
              onClick={() => setSubject(s.slug)}
              className={cn(
                "rounded-full border px-3 py-1 text-sm transition",
                subject === s.slug ? "border-indigo-500 bg-indigo-50 text-indigo-700" : "border-slate-200 hover:bg-slate-50",
              )}
            >
              {s.icon} {s.name}
            </button>
          ))}
        </div>
        <Button className="w-full" onClick={start}>
          {t("assess.start")}
        </Button>
        {msg && <p className="text-center text-sm text-red-600">{msg}</p>}
      </Card>
    </div>
  );
}

function AnswerInput({ q, onAnswer }: { q: AssessQuestion; onAnswer: (choice: unknown) => void }) {
  const [val, setVal] = useState("");

  if (q.type === "true_false") {
    return (
      <div className="grid grid-cols-2 gap-3">
        <button className="rounded-xl border border-slate-300 py-4 text-sm font-medium hover:bg-indigo-50" onClick={() => onAnswer({ value: true })}>
          To'g'ri ✓
        </button>
        <button className="rounded-xl border border-slate-300 py-4 text-sm font-medium hover:bg-indigo-50" onClick={() => onAnswer({ value: false })}>
          Noto'g'ri ✕
        </button>
      </div>
    );
  }

  if (q.type === "numeric") {
    return (
      <div className="space-y-3">
        <Input type="number" value={val} onChange={(e) => setVal(e.target.value)} placeholder="Javob" className="text-center text-lg" />
        <Button className="w-full" disabled={val === ""} onClick={() => onAnswer({ value: Number(val) })}>
          Keyingi
        </Button>
      </div>
    );
  }

  return (
    <div className="grid gap-3">
      {q.options?.map((o) => (
        <button
          key={o.id}
          onClick={() => onAnswer({ optionId: o.id })}
          className="rounded-xl border border-slate-300 px-4 py-3 text-left text-sm font-medium hover:border-indigo-400 hover:bg-indigo-50"
        >
          {o.text}
        </button>
      ))}
    </div>
  );
}
