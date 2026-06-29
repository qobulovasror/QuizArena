import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";
import { useGame } from "@core/store";
import { Button } from "../components/ui/button";
import { Input } from "../components/ui/input";
import { Card } from "../components/ui/card";
import { cn } from "../lib/cn";

// API bo'sh bo'lsa (seed qilinmagan) zaxira ro'yxat.
const fallback = [
  { slug: "english", name: "Ingliz tili", icon: "📘" },
  { slug: "math", name: "Matematika", icon: "🔢" },
  { slug: "general", name: "Umumiy bilim", icon: "🌍" },
  { slug: "programming", name: "Dasturlash", icon: "💻" },
];

const modeIds = ["classic", "survival", "time_attack", "team"];

export function LobbyPage() {
  const { t } = useTranslation();
  const { room, selfUserId, createRoom, joinRoom, queueMatch, cancelMatch, start, status, matchSearching } = useGame();
  const online = status === "online";

  if (matchSearching) {
    return (
      <div className="mx-auto max-w-md p-4">
        <Card className="space-y-4 text-center">
          <div className="animate-pulse text-5xl">🔍</div>
          <p className="font-medium">{t("match.searching")}</p>
          <p className="text-xs text-slate-400">{t("match.searchingHint")}</p>
          <Button variant="outline" className="w-full" onClick={cancelMatch}>
            {t("match.cancel")}
          </Button>
        </Card>
      </div>
    );
  }

  if (room && room.status === "lobby") {
    const isHost = selfUserId === room.host;
    return (
      <div className="mx-auto max-w-md p-4">
        <Card className="space-y-4 text-center">
          <p className="text-sm text-slate-500">{t("lobby.roomCode")}</p>
          <div className="text-4xl font-bold tracking-widest text-indigo-600">{room.code}</div>
          <p className="text-xs text-slate-400">{t("lobby.shareCode")}</p>
          <span className="inline-block rounded-full bg-slate-100 px-3 py-0.5 text-xs font-medium text-slate-600">
            {t("lobby.mode", { mode: room.config.mode })}
          </span>

          <div className="space-y-1 text-left">
            <p className="text-sm font-medium">{t("lobby.players", { count: room.players.length })}</p>
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
              {t("lobby.startGame")}
            </Button>
          ) : (
            <p className="text-sm text-slate-500">{t("lobby.waitHost")}</p>
          )}
        </Card>
      </div>
    );
  }

  return <CreateOrJoin onCreate={createRoom} onJoin={joinRoom} onQueue={queueMatch} />;
}

function CreateOrJoin({
  onCreate,
  onJoin,
  onQueue,
}: {
  onCreate: (o: { subjectId: string; mode: string; questionCount: number; timePerQ: number; opponent: string; botDifficulty: string }) => void;
  onJoin: (code: string) => void;
  onQueue: (subjectId: string) => void;
}) {
  const { t } = useTranslation();
  const { subjects, loadSubjects, status } = useGame();
  const online = status === "online";
  const [subjectId, setSubjectId] = useState("english");
  const [mode, setMode] = useState("classic");
  const [opponent, setOpponent] = useState("human");
  const [botDifficulty, setBotDifficulty] = useState("medium");
  const [count, setCount] = useState(5);
  const [time, setTime] = useState(15);
  const [code, setCode] = useState("");

  useEffect(() => {
    loadSubjects();
  }, [loadSubjects]);

  const list = subjects.length > 0 ? subjects : fallback;

  return (
    <div className="mx-auto max-w-md space-y-4 p-4">
      <h2 className="text-center text-xl font-semibold">{t("lobby.newRoom")}</h2>
      <Card className="space-y-3">
        <p className="text-sm text-slate-500">{t("lobby.pickSubject")}</p>
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
        <p className="text-sm text-slate-500">{t("lobby.modeLabel")}</p>
        <div className="grid grid-cols-2 gap-2">
          {modeIds.map((id) => (
            <button
              key={id}
              onClick={() => setMode(id)}
              className={cn(
                "rounded-xl border px-3 py-2 text-left text-sm transition",
                mode === id ? "border-indigo-500 bg-indigo-50 text-indigo-700" : "border-slate-200 hover:bg-slate-50",
              )}
            >
              <div className="font-medium">{t(`lobby.${id}`)}</div>
              <div className="text-xs text-slate-400">{t(`lobby.${id}Desc`)}</div>
            </button>
          ))}
        </div>
        <p className="text-sm text-slate-500">{t("lobby.opponent")}</p>
        <div className="grid grid-cols-2 gap-2">
          {["human", "bot"].map((o) => (
            <button
              key={o}
              onClick={() => setOpponent(o)}
              className={cn(
                "rounded-xl border px-3 py-2 text-sm transition",
                opponent === o ? "border-indigo-500 bg-indigo-50 text-indigo-700" : "border-slate-200 hover:bg-slate-50",
              )}
            >
              {t(`lobby.${o}`)}
            </button>
          ))}
        </div>
        {opponent === "bot" && (
          <div className="grid grid-cols-3 gap-2">
            {["easy", "medium", "hard"].map((d) => (
              <button
                key={d}
                onClick={() => setBotDifficulty(d)}
                className={cn(
                  "rounded-lg border px-2 py-1.5 text-xs transition",
                  botDifficulty === d ? "border-indigo-500 bg-indigo-50 text-indigo-700" : "border-slate-200 hover:bg-slate-50",
                )}
              >
                {t(`lobby.diff_${d}`)}
              </button>
            ))}
          </div>
        )}
        <div className="grid grid-cols-2 gap-3">
          <label className="text-sm">
            {t("lobby.questionCount")}
            <Input type="number" min={1} max={20} value={count} onChange={(e) => setCount(+e.target.value)} />
          </label>
          <label className="text-sm">
            {t("lobby.timeSec")}
            <Input type="number" min={5} max={60} value={time} onChange={(e) => setTime(+e.target.value)} />
          </label>
        </div>
        <Button
          className="w-full"
          disabled={!online}
          onClick={() => onCreate({ subjectId, mode, questionCount: count, timePerQ: time, opponent, botDifficulty })}
        >
          {t("lobby.createRoom")}
        </Button>
      </Card>

      <Card className="space-y-2">
        <h2 className="font-semibold">{t("match.duel")}</h2>
        <p className="text-xs text-slate-400">{t("match.duelHint")}</p>
        <Button variant="outline" className="w-full" disabled={!online} onClick={() => onQueue(subjectId)}>
          ⚔️ {t("match.find")}
        </Button>
      </Card>

      <div className="text-center text-sm text-slate-400">{t("lobby.or")}</div>

      <Card className="space-y-3">
        <h2 className="font-semibold">{t("lobby.joinByCode")}</h2>
        <Input
          value={code}
          onChange={(e) => setCode(e.target.value.toUpperCase())}
          placeholder={t("lobby.joinPlaceholder")}
          maxLength={6}
          className="text-center tracking-widest"
        />
        <Button variant="outline" className="w-full" disabled={code.length < 4 || !online} onClick={() => onJoin(code)}>
          {t("lobby.join")}
        </Button>
      </Card>
    </div>
  );
}
