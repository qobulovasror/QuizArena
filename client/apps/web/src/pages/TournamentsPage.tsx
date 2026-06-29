import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useGame } from "@core/store";
import { api } from "@core/api";
import type { TournamentInfo, TournamentQ, TournamentLeaderRow, TournamentStatus, AssessAnswer } from "@core/api";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Card } from "../components/ui/card";
import { cn } from "../lib/cn";

type View = "list" | "quiz" | "result" | "leaderboard";

const statusStyle: Record<TournamentStatus, string> = {
  upcoming: "bg-slate-100 text-slate-500",
  active: "bg-green-100 text-green-700",
  finished: "bg-red-100 text-red-600",
};

const medals = ["🥇", "🥈", "🥉"];

function fmtRange(startsAt: string, endsAt: string): string {
  const f = (s: string) => new Date(s).toLocaleString();
  return `${f(startsAt)} — ${f(endsAt)}`;
}

export function TournamentsPage() {
  const { t } = useTranslation();
  const token = useGame((s) => s.token);

  const [view, setView] = useState<View>("list");
  const [items, setItems] = useState<TournamentInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [err, setErr] = useState("");

  // quiz oqimi
  const [active, setActive] = useState<TournamentInfo | null>(null);
  const [questions, setQuestions] = useState<TournamentQ[]>([]);
  const [idx, setIdx] = useState(0);
  const [answers, setAnswers] = useState<AssessAnswer[]>([]);
  const [score, setScore] = useState({ correct: 0, total: 0 });

  // reyting
  const [board, setBoard] = useState<TournamentLeaderRow[]>([]);

  const load = useCallback(async () => {
    if (!token) return;
    setLoading(true);
    setErr("");
    try {
      setItems(await api.tournaments(token));
    } catch (e) {
      setErr(e instanceof Error ? e.message : "xato");
    } finally {
      setLoading(false);
    }
  }, [token]);

  useEffect(() => {
    load();
  }, [load]);

  async function play(tn: TournamentInfo) {
    if (!token) return;
    setErr("");
    try {
      const qs = await api.tournamentPlay(tn.id, token);
      if (qs.length === 0) {
        setErr(t("tournament.noQuestions"));
        return;
      }
      setActive(tn);
      setQuestions(qs);
      setIdx(0);
      setAnswers([]);
      setView("quiz");
    } catch (e) {
      setErr(e instanceof Error ? e.message : "xato");
    }
  }

  async function answer(choice: unknown) {
    const q = questions[idx];
    const next = [...answers, { questionId: q.questionId, choice }];
    setAnswers(next);
    if (idx + 1 < questions.length) {
      setIdx(idx + 1);
    } else if (token && active) {
      try {
        const res = await api.tournamentSubmit(active.id, next, token);
        setScore(res);
        setView("result");
      } catch (e) {
        setErr(e instanceof Error ? e.message : "xato");
        setView("list");
      }
    }
  }

  async function showLeaderboard(tn: TournamentInfo) {
    if (!token) return;
    setErr("");
    setActive(tn);
    setBoard([]);
    setView("leaderboard");
    try {
      setBoard(await api.tournamentLeaderboard(tn.id, token));
    } catch (e) {
      setErr(e instanceof Error ? e.message : "xato");
    }
  }

  if (view === "quiz") {
    const q = questions[idx];
    return (
      <div className="mx-auto max-w-md space-y-4 p-4">
        <p className="text-center text-xs text-slate-400">{t("tournament.progress", { n: idx + 1, total: questions.length })}</p>
        <Card>
          <h2 className="mb-5 text-center text-xl font-semibold">{q.prompt}</h2>
          <AnswerInput q={q} onAnswer={answer} />
        </Card>
      </div>
    );
  }

  if (view === "result") {
    const pct = Math.round((score.correct / Math.max(1, score.total)) * 100);
    return (
      <div className="mx-auto max-w-md space-y-4 p-4">
        <Card className="space-y-3 text-center">
          <h2 className="text-2xl font-bold">{t("tournament.result")}</h2>
          <div className="text-4xl font-bold text-indigo-600">
            {score.correct} / {score.total}
          </div>
          <p className="text-sm text-slate-500">{t("tournament.percentCorrect", { pct })}</p>
          <Button className="w-full" onClick={() => active && showLeaderboard(active)}>
            {t("tournament.leaderboard")}
          </Button>
          <Button variant="outline" className="w-full" onClick={() => setView("list")}>
            {t("tournament.back")}
          </Button>
        </Card>
      </div>
    );
  }

  if (view === "leaderboard") {
    return (
      <div className="mx-auto max-w-md space-y-4 p-4">
        <Card className="space-y-2">
          <p className="text-sm font-medium text-slate-500">
            {t("tournament.leaderboard")}{active ? ` · ${active.title}` : ""}
          </p>
          {err && <p className="py-2 text-center text-sm text-red-600">{err}</p>}
          {board.length === 0 && !err ? (
            <p className="py-6 text-center text-sm text-slate-400">{t("tournament.noScores")}</p>
          ) : (
            board.map((row, i) => (
              <div key={row.username + i} className="flex items-center justify-between rounded-xl bg-slate-50 px-4 py-3">
                <span className="flex items-center gap-2">
                  <span className="w-6 text-center">{medals[i] ?? i + 1}</span>
                  <span>{row.username}</span>
                </span>
                <span className="font-semibold text-indigo-600">{row.score}</span>
              </div>
            ))
          )}
          <Button variant="outline" className="w-full" onClick={() => setView("list")}>
            {t("tournament.back")}
          </Button>
        </Card>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-md space-y-4 p-4">
      {err && <p className="text-center text-sm text-red-600">{err}</p>}
      {loading ? (
        <p className="py-6 text-center text-slate-400">{t("common.loading")}</p>
      ) : items.length === 0 ? (
        <Card>
          <p className="py-6 text-center text-sm text-slate-400">{t("tournament.empty")}</p>
        </Card>
      ) : (
        items.map((tn) => (
          <Card key={tn.id} className="space-y-3">
            <div className="flex items-start justify-between gap-2">
              <div>
                <h3 className="font-semibold">{tn.title}</h3>
                <p className="text-xs text-slate-400">
                  {tn.subjectName} · {t("tournament.questions", { count: tn.questionCount })}
                </p>
              </div>
              <span className={cn("shrink-0 rounded-full px-2.5 py-1 text-xs font-medium", statusStyle[tn.status])}>
                {t(`tournament.status.${tn.status}`)}
              </span>
            </div>
            <p className="text-xs text-slate-400">{fmtRange(tn.startsAt, tn.endsAt)}</p>
            <div className="flex gap-2">
              {tn.status === "active" && (
                <Button className="flex-1" onClick={() => play(tn)}>
                  {t("tournament.join")}
                </Button>
              )}
              <Button variant="outline" className="flex-1" onClick={() => showLeaderboard(tn)}>
                {t("tournament.leaderboard")}
              </Button>
            </div>
          </Card>
        ))
      )}
    </div>
  );
}

function AnswerInput({ q, onAnswer }: { q: TournamentQ; onAnswer: (choice: unknown) => void }) {
  const { t } = useTranslation();
  const [val, setVal] = useState("");

  if (q.type === "true_false") {
    return (
      <div className="grid grid-cols-2 gap-3">
        <button className="rounded-xl border border-slate-300 py-4 text-sm font-medium hover:bg-indigo-50" onClick={() => onAnswer({ value: true })}>
          {t("play.tfTrue")}
        </button>
        <button className="rounded-xl border border-slate-300 py-4 text-sm font-medium hover:bg-indigo-50" onClick={() => onAnswer({ value: false })}>
          {t("play.tfFalse")}
        </button>
      </div>
    );
  }

  if (q.type === "numeric") {
    return (
      <div className="space-y-3">
        <Input type="number" value={val} onChange={(e) => setVal(e.target.value)} placeholder={t("play.enterAnswer")} className="text-center text-lg" />
        <Button className="w-full" disabled={val === ""} onClick={() => onAnswer({ value: Number(val) })}>
          {t("tournament.next")}
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
