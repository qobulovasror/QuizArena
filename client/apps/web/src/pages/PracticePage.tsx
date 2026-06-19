import { useCallback, useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useGame } from "../core/store";
import { api } from "../core/api";
import type { SrsCard } from "../core/api";
import { Button } from "../components/ui/button";
import { Card } from "../components/ui/card";
import { cn } from "../lib/cn";

const grades = [
  { g: 0, key: "practice.again", color: "bg-red-500 hover:bg-red-600" },
  { g: 1, key: "practice.good", color: "bg-indigo-600 hover:bg-indigo-700" },
  { g: 2, key: "practice.easy", color: "bg-green-600 hover:bg-green-700" },
];

export function PracticePage() {
  const { t } = useTranslation();
  const token = useGame((s) => s.token);
  const subjects = useGame((s) => s.subjects);
  const loadSubjects = useGame((s) => s.loadSubjects);

  const [subject, setSubject] = useState("general");
  const [cards, setCards] = useState<SrsCard[]>([]);
  const [idx, setIdx] = useState(0);
  const [revealed, setRevealed] = useState(false);
  const [loading, setLoading] = useState(false);

  const load = useCallback(
    async (subj: string) => {
      if (!token) return;
      setLoading(true);
      setRevealed(false);
      setIdx(0);
      try {
        setCards(await api.srsDue(subj, token));
      } catch {
        setCards([]);
      } finally {
        setLoading(false);
      }
    },
    [token],
  );

  useEffect(() => {
    loadSubjects();
  }, [loadSubjects]);

  useEffect(() => {
    load(subject);
  }, [subject, load]);

  const card = cards[idx];
  const done = !loading && (cards.length === 0 || idx >= cards.length);

  async function grade(g: number) {
    if (!token || !card) return;
    api.srsReview(card.questionId, g, token).catch(() => {});
    setRevealed(false);
    setIdx((i) => i + 1);
  }

  const list = subjects.length > 0 ? subjects : [{ slug: "general", name: "Umumiy bilim", icon: "🌍" }];

  return (
    <div className="mx-auto max-w-md space-y-4 p-4">
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

      {loading && <p className="py-10 text-center text-slate-400">{t("common.loading")}</p>}

      {!loading && done && (
        <Card className="space-y-3 text-center">
          <p className="text-slate-600">{cards.length === 0 ? t("practice.noCards") : t("practice.sessionDone")}</p>
          <Button className="w-full" onClick={() => load(subject)}>
            {t("practice.refresh")}
          </Button>
        </Card>
      )}

      {!loading && card && (
        <>
          <p className="text-center text-xs text-slate-400">
            {idx + 1} / {cards.length}
          </p>
          <Card className="space-y-5">
            <h2 className="text-center text-xl font-semibold">{card.prompt}</h2>

            {!revealed ? (
              <Button className="w-full" onClick={() => setRevealed(true)}>
                {t("practice.showAnswer")}
              </Button>
            ) : (
              <div className="space-y-4">
                <div className="rounded-xl bg-green-50 px-4 py-3 text-center">
                  <div className="text-lg font-semibold text-green-700">{card.answer}</div>
                  {card.explanation && <div className="mt-1 text-sm text-slate-500">{card.explanation}</div>}
                </div>
                <p className="text-center text-xs text-slate-400">{t("practice.howWell")}</p>
                <div className="grid grid-cols-3 gap-2">
                  {grades.map((gr) => (
                    <button
                      key={gr.g}
                      onClick={() => grade(gr.g)}
                      className={cn("rounded-lg py-2 text-sm font-medium text-white", gr.color)}
                    >
                      {t(gr.key)}
                    </button>
                  ))}
                </div>
              </div>
            )}
          </Card>
        </>
      )}
    </div>
  );
}
