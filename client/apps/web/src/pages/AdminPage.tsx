import { useCallback, useEffect, useState } from "react";
import { useGame } from "../core/store";
import { api } from "../core/api";
import type { CategoryInfo, AdminQuestion } from "../core/api";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Card } from "../components/ui/card";

const OPT_IDS = ["a", "b", "c", "d"];

const QTYPES = ["mcq", "true_false", "numeric", "ordering", "cloze", "match", "categorize"];
const BULK_TYPES = ["ordering", "cloze", "match", "categorize"];

// bo'sh bo'lmagan, trim qilingan qatorlar
const lines = (s: string) => s.split("\n").map((x) => x.trim()).filter(Boolean);

const bulkHint: Record<string, string> = {
  ordering: "Elementlar TO'G'RI tartibda, har biri yangi qatorda.",
  cloze: "Savol matniga ___ qo'ying. Bu yerga har bo'shliq javobi (sinonim = vergul), yangi qatorda.",
  match: "Har qatorda: chap = o'ng (masalan: cat = mushuk).",
  categorize: "Har qatorda: element = Toifa (masalan: olma = Meva).",
};

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
  const [bulk, setBulk] = useState(""); // ordering/cloze/match/categorize uchun matn kiritish
  const [expl, setExpl] = useState("");

  // yangi kategoriya
  const [newCat, setNewCat] = useState("");

  // turnir formi
  const [tnTitle, setTnTitle] = useState("");
  const [tnSubject, setTnSubject] = useState("");
  const [tnCount, setTnCount] = useState("5");
  const [tnStart, setTnStart] = useState("");
  const [tnEnd, setTnEnd] = useState("");
  const [tnMsg, setTnMsg] = useState("");

  useEffect(() => {
    loadSubjects();
  }, [loadSubjects]);

  useEffect(() => {
    if (subjects.length && !subjectId) setSubjectId(subjects[0].id);
    if (subjects.length && !tnSubject) setTnSubject(subjects[0].slug);
  }, [subjects, subjectId, tnSubject]);

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
    } else if (type === "ordering") {
      const items = lines(bulk);
      body.options = items.map((tx, i) => ({ id: `o${i + 1}`, text: tx }));
      body.correct = { order: items.map((_, i) => `o${i + 1}`) };
    } else if (type === "cloze") {
      body.correct = { blanks: lines(bulk).map((l) => ({ accepted: l.split(",").map((x) => x.trim()).filter(Boolean) })) };
    } else if (type === "match") {
      const rows = lines(bulk).map((l) => l.split("="));
      body.options = rows.map((r, i) => ({ id: `l${i + 1}`, text: (r[0] ?? "").trim() }));
      body.targets = rows.map((r, i) => ({ id: `r${i + 1}`, text: (r[1] ?? "").trim() }));
      body.correct = { pairs: Object.fromEntries(rows.map((_, i) => [`l${i + 1}`, `r${i + 1}`])) };
    } else if (type === "categorize") {
      const rows = lines(bulk).map((l) => l.split("="));
      const catNames = [...new Set(rows.map((r) => (r[1] ?? "").trim()))];
      body.options = rows.map((r, i) => ({ id: `i${i + 1}`, text: (r[0] ?? "").trim() }));
      body.targets = catNames.map((n, i) => ({ id: `c${i + 1}`, text: n }));
      body.correct = { assign: Object.fromEntries(rows.map((r, i) => [`i${i + 1}`, `c${catNames.indexOf((r[1] ?? "").trim()) + 1}`])) };
    }
    try {
      await api.adminCreateQuestion(body, token);
      setMsg("✓ savol qo'shildi");
      setPrompt("");
      setOpts(["", "", "", ""]);
      setNumValue("");
      setBulk("");
      setExpl("");
      loadQuestions();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "xato");
    }
    setTimeout(() => setMsg(""), 2500);
  }

  async function addCategory() {
    if (!subjectId || !newCat) return;
    try {
      await api.adminCreateCategory({ subjectId, slug: newCat.toLowerCase().replace(/\s+/g, "-"), name: newCat }, token);
      setNewCat("");
      loadCats();
    } catch (e) {
      setMsg(e instanceof Error ? e.message : "xato");
      setTimeout(() => setMsg(""), 2500);
    }
  }

  async function addTournament() {
    if (!tnTitle || !tnSubject || !tnStart || !tnEnd) return;
    try {
      await api.adminCreateTournament(
        {
          title: tnTitle,
          subjectSlug: tnSubject,
          questionCount: Number(tnCount),
          startsAt: new Date(tnStart).toISOString(),
          endsAt: new Date(tnEnd).toISOString(),
        },
        token,
      );
      setTnMsg("✓ turnir yaratildi");
      setTnTitle("");
      setTnStart("");
      setTnEnd("");
    } catch (e) {
      setTnMsg(e instanceof Error ? e.message : "xato");
    }
    setTimeout(() => setTnMsg(""), 2500);
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
        <div className="flex flex-wrap gap-2 text-xs">
          {QTYPES.map((t) => (
            <button
              key={t}
              onClick={() => setType(t)}
              className={"rounded-lg px-2.5 py-1.5 " + (type === t ? "bg-indigo-100 text-indigo-700" : "bg-slate-100 text-slate-500")}
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

        {BULK_TYPES.includes(type) && (
          <div className="space-y-1">
            <textarea
              value={bulk}
              onChange={(e) => setBulk(e.target.value)}
              rows={4}
              placeholder={bulkHint[type]}
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
            />
            <p className="text-xs text-slate-400">{bulkHint[type]}</p>
          </div>
        )}

        <Input value={expl} onChange={(e) => setExpl(e.target.value)} placeholder="Izoh (ixtiyoriy)" />
        <Button className="w-full" onClick={addQuestion} disabled={!categoryId || !prompt || (BULK_TYPES.includes(type) && !bulk.trim())}>
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
                try {
                  await api.adminDeleteQuestion(q.id, token);
                  loadQuestions();
                } catch (e) {
                  setMsg(e instanceof Error ? e.message : "xato");
                  setTimeout(() => setMsg(""), 2500);
                }
              }}
              className="shrink-0 text-red-500 hover:text-red-700"
            >
              ✕
            </button>
          </div>
        ))}
        {questions.length === 0 && <p className="text-sm text-slate-400">Savol yo'q.</p>}
      </Card>

      <Card className="space-y-3">
        <h3 className="font-medium">🏆 Yangi turnir</h3>
        <Input value={tnTitle} onChange={(e) => setTnTitle(e.target.value)} placeholder="Turnir nomi" />
        <select className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm" value={tnSubject} onChange={(e) => setTnSubject(e.target.value)}>
          {subjects.map((s) => (
            <option key={s.id} value={s.slug}>
              {s.icon} {s.name}
            </option>
          ))}
        </select>
        <div className="space-y-1">
          <label className="text-xs text-slate-400">Savol soni</label>
          <Input type="number" value={tnCount} onChange={(e) => setTnCount(e.target.value)} placeholder="Savol soni" />
        </div>
        <div className="space-y-1">
          <label className="text-xs text-slate-400">Boshlanish</label>
          <Input type="datetime-local" value={tnStart} onChange={(e) => setTnStart(e.target.value)} />
        </div>
        <div className="space-y-1">
          <label className="text-xs text-slate-400">Tugash</label>
          <Input type="datetime-local" value={tnEnd} onChange={(e) => setTnEnd(e.target.value)} />
        </div>
        <Button className="w-full" onClick={addTournament} disabled={!tnTitle || !tnSubject || !tnStart || !tnEnd}>
          Turnir yaratish
        </Button>
        {tnMsg && <p className="text-center text-sm text-indigo-600">{tnMsg}</p>}
      </Card>
    </div>
  );
}
