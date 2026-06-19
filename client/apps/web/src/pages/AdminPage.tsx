import { useCallback, useEffect, useState } from "react";
import { useGame } from "../core/store";
import { api } from "../core/api";
import type { CategoryInfo, AdminQuestion } from "../core/api";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Card } from "../components/ui/card";

const OPT_IDS = ["a", "b", "c", "d"];

export function AdminPage() {
  const token = useGame((s) => s.token)!;
  const subjects = useGame((s) => s.subjects);
  const loadSubjects = useGame((s) => s.loadSubjects);

  const [subjectId, setSubjectId] = useState("");
  const [cats, setCats] = useState<CategoryInfo[]>([]);
  const [categoryId, setCategoryId] = useState("");
  const [questions, setQuestions] = useState<AdminQuestion[]>([]);
  const [msg, setMsg] = useState("");

  // savol formi
  const [type, setType] = useState("mcq");
  const [prompt, setPrompt] = useState("");
  const [opts, setOpts] = useState(["", "", "", ""]);
  const [correctIdx, setCorrectIdx] = useState(0);
  const [tfValue, setTfValue] = useState(true);
  const [numValue, setNumValue] = useState("");
  const [expl, setExpl] = useState("");

  // yangi kategoriya
  const [newCat, setNewCat] = useState("");

  useEffect(() => {
    loadSubjects();
  }, [loadSubjects]);

  useEffect(() => {
    if (subjects.length && !subjectId) setSubjectId(subjects[0].id);
  }, [subjects, subjectId]);

  const loadCats = useCallback(async () => {
    if (!subjectId) return;
    const c = await api.categories(subjectId);
    setCats(c);
    setCategoryId(c[0]?.id ?? "");
  }, [subjectId]);

  useEffect(() => {
    loadCats();
  }, [loadCats]);

  const loadQuestions = useCallback(async () => {
    if (!categoryId) {
      setQuestions([]);
      return;
    }
    try {
      setQuestions(await api.adminListQuestions(categoryId, token));
    } catch {
      setQuestions([]);
    }
  }, [categoryId, token]);

  useEffect(() => {
    loadQuestions();
  }, [loadQuestions]);

  async function addQuestion() {
    if (!categoryId || !prompt) return;
    const body: Record<string, unknown> = { categoryId, type, prompt, explanation: expl };
    if (type === "mcq") {
      body.options = opts.map((t, i) => ({ id: OPT_IDS[i], text: t }));
      body.correct = { optionId: OPT_IDS[correctIdx] };
    } else if (type === "true_false") {
      body.correct = { value: tfValue };
    } else if (type === "numeric") {
      body.correct = { value: Number(numValue), tolerance: 0 };
    }
    try {
      await api.adminCreateQuestion(body, token);
      setMsg("✓ savol qo'shildi");
      setPrompt("");
      setOpts(["", "", "", ""]);
      setNumValue("");
      setExpl("");
      loadQuestions();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "xato");
    }
    setTimeout(() => setMsg(""), 2500);
  }

  async function addCategory() {
    if (!subjectId || !newCat) return;
    await api.adminCreateCategory({ subjectId, slug: newCat.toLowerCase().replace(/\s+/g, "-"), name: newCat }, token).catch(() => {});
    setNewCat("");
    loadCats();
  }

  return (
    <div className="mx-auto max-w-md space-y-4 p-4">
      <h2 className="text-lg font-semibold">🛠 Admin — savol bankini boshqarish</h2>

      <Card className="space-y-3">
        <select className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" value={subjectId} onChange={(e) => setSubjectId(e.target.value)}>
          {subjects.map((s) => (
            <option key={s.id} value={s.id}>
              {s.icon} {s.name}
            </option>
          ))}
        </select>
        <div className="flex gap-2">
          <select className="flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm" value={categoryId} onChange={(e) => setCategoryId(e.target.value)}>
            {cats.map((c) => (
              <option key={c.id} value={c.id}>
                {c.name}
              </option>
            ))}
            {cats.length === 0 && <option value="">— kategoriya yo'q —</option>}
          </select>
        </div>
        <div className="flex gap-2">
          <Input value={newCat} onChange={(e) => setNewCat(e.target.value)} placeholder="Yangi kategoriya nomi" />
          <Button variant="outline" onClick={addCategory} disabled={!newCat}>
            +
          </Button>
        </div>
      </Card>

      <Card className="space-y-3">
        <h3 className="font-medium">Yangi savol</h3>
        <div className="flex gap-2 text-sm">
          {["mcq", "true_false", "numeric"].map((t) => (
            <button
              key={t}
              onClick={() => setType(t)}
              className={"flex-1 rounded-lg py-1.5 " + (type === t ? "bg-indigo-100 text-indigo-700" : "bg-slate-100 text-slate-500")}
            >
              {t}
            </button>
          ))}
        </div>
        <Input value={prompt} onChange={(e) => setPrompt(e.target.value)} placeholder="Savol matni" />

        {type === "mcq" &&
          opts.map((o, i) => (
            <div key={i} className="flex items-center gap-2">
              <input type="radio" checked={correctIdx === i} onChange={() => setCorrectIdx(i)} title="to'g'ri" />
              <Input value={o} onChange={(e) => setOpts(opts.map((x, j) => (j === i ? e.target.value : x)))} placeholder={`Variant ${OPT_IDS[i]}`} />
            </div>
          ))}
        {type === "true_false" && (
          <div className="flex gap-2 text-sm">
            <button onClick={() => setTfValue(true)} className={"flex-1 rounded-lg py-2 " + (tfValue ? "bg-green-100 text-green-700" : "bg-slate-100")}>
              To'g'ri
            </button>
            <button onClick={() => setTfValue(false)} className={"flex-1 rounded-lg py-2 " + (!tfValue ? "bg-green-100 text-green-700" : "bg-slate-100")}>
              Noto'g'ri
            </button>
          </div>
        )}
        {type === "numeric" && <Input type="number" value={numValue} onChange={(e) => setNumValue(e.target.value)} placeholder="To'g'ri javob (raqam)" />}

        <Input value={expl} onChange={(e) => setExpl(e.target.value)} placeholder="Izoh (ixtiyoriy)" />
        <Button className="w-full" onClick={addQuestion} disabled={!categoryId || !prompt}>
          Qo'shish
        </Button>
        {msg && <p className="text-center text-sm text-indigo-600">{msg}</p>}
      </Card>

      <Card className="space-y-2">
        <h3 className="font-medium">Savollar ({questions.length})</h3>
        {questions.map((q) => (
          <div key={q.id} className="flex items-center justify-between gap-2 rounded-lg bg-slate-50 px-3 py-2 text-sm">
            <span className="truncate">
              <span className="text-slate-400">[{q.type}]</span> {q.prompt}
            </span>
            <button
              onClick={async () => {
                await api.adminDeleteQuestion(q.id, token).catch(() => {});
                loadQuestions();
              }}
              className="shrink-0 text-red-500 hover:text-red-700"
            >
              ✕
            </button>
          </div>
        ))}
        {questions.length === 0 && <p className="text-sm text-slate-400">Savol yo'q.</p>}
      </Card>
    </div>
  );
}
