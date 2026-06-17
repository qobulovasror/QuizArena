import { useEffect, useState } from "react";
import { useGame } from "../core/store";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Card } from "../components/ui/card";
import { cn } from "../lib/cn";

// API bo'sh bo'lsa (seed qilinmagan) zaxira ro'yxat.
const fallback = [
  { slug: "english", name: "Ingliz tili", icon: "📘" },
  { slug: "math", name: "Matematika", icon: "🔢" },
  { slug: "general", name: "Umumiy bilim", icon: "🌍" },
];

export function LobbyPage() {
  const { room, selfUserId, createRoom, joinRoom, start, status } = useGame();
  const online = status === "online";

  if (room && room.status === "lobby") {
    const isHost = selfUserId === room.host;
    return (
      <div className="mx-auto max-w-md p-4">
        <Card className="space-y-4 text-center">
          <p className="text-sm text-slate-500">Xona kodi</p>
          <div className="text-4xl font-bold tracking-widest text-indigo-600">{room.code}</div>
          <p className="text-xs text-slate-400">Do'stlaringizga kodni ulashing</p>

          <div className="space-y-1 text-left">
            <p className="text-sm font-medium">O'yinchilar ({room.players.length})</p>
            {room.players.map((p) => (
              <div key={p.userId} className="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2 text-sm">
                <span>
                  {p.name}
                  {p.userId === room.host ? " 👑" : ""}
                </span>
                <span className={p.connected ? "text-green-500" : "text-slate-300"}>●</span>
              </div>
            ))}
          </div>

          {isHost ? (
            <Button className="w-full" onClick={start} disabled={!online}>
              O'yinni boshlash
            </Button>
          ) : (
            <p className="text-sm text-slate-500">Host boshlashini kuting…</p>
          )}
        </Card>
      </div>
    );
  }

  return <CreateOrJoin onCreate={createRoom} onJoin={joinRoom} />;
}

function CreateOrJoin({
  onCreate,
  onJoin,
}: {
  onCreate: (o: { subjectId: string; questionCount: number; timePerQ: number }) => void;
  onJoin: (code: string) => void;
}) {
  const { subjects, loadSubjects, status } = useGame();
  const online = status === "online";
  const [subjectId, setSubjectId] = useState("english");
  const [count, setCount] = useState(5);
  const [time, setTime] = useState(15);
  const [code, setCode] = useState("");

  useEffect(() => {
    loadSubjects();
  }, [loadSubjects]);

  const list = subjects.length > 0 ? subjects : fallback;

  return (
    <div className="mx-auto max-w-md space-y-4 p-4">
      <h2 className="text-center text-xl font-semibold">Yangi xona</h2>
      <Card className="space-y-3">
        <p className="text-sm text-slate-500">Sohani tanlang</p>
        <div className="grid grid-cols-3 gap-2">
          {list.map((s) => (
            <button
              key={s.slug}
              onClick={() => setSubjectId(s.slug)}
              className={cn(
                "rounded-xl border px-2 py-3 text-center text-sm transition",
                subjectId === s.slug
                  ? "border-indigo-500 bg-indigo-50 font-medium text-indigo-700"
                  : "border-slate-200 hover:bg-slate-50",
              )}
            >
              <div className="text-xl">{s.icon}</div>
              {s.name}
            </button>
          ))}
        </div>
        <div className="grid grid-cols-2 gap-3">
          <label className="text-sm">
            Savol soni
            <Input type="number" min={1} max={20} value={count} onChange={(e) => setCount(+e.target.value)} />
          </label>
          <label className="text-sm">
            Vaqt (soniya)
            <Input type="number" min={5} max={60} value={time} onChange={(e) => setTime(+e.target.value)} />
          </label>
        </div>
        <Button
          className="w-full"
          disabled={!online}
          onClick={() => onCreate({ subjectId, questionCount: count, timePerQ: time })}
        >
          Xona yaratish (classic)
        </Button>
      </Card>

      <div className="text-center text-sm text-slate-400">— yoki —</div>

      <Card className="space-y-3">
        <h2 className="font-semibold">Kod bilan qo'shilish</h2>
        <Input
          value={code}
          onChange={(e) => setCode(e.target.value.toUpperCase())}
          placeholder="Xona kodi"
          maxLength={6}
          className="text-center tracking-widest"
        />
        <Button variant="outline" className="w-full" disabled={code.length < 4 || !online} onClick={() => onJoin(code)}>
          Qo'shilish
        </Button>
      </Card>
    </div>
  );
}
