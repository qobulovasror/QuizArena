import { useEffect } from "react";
import { useTranslation } from "react-i18next";
import { useGame } from "./core/store";
import { setLang } from "./core/i18n";
import { getTelegram } from "./core/telegram";
import { AuthPage } from "./pages/AuthPage";
import { LobbyPage } from "./pages/LobbyPage";
import { HomePage } from "./pages/HomePage";
import { PlayPage } from "./pages/PlayPage";
import { ResultPage } from "./pages/ResultPage";

export default function App() {
  const { t, i18n } = useTranslation();
  const token = useGame((s) => s.token);
  const status = useGame((s) => s.status);
  const displayName = useGame((s) => s.displayName);
  const room = useGame((s) => s.room);
  const gameOver = useGame((s) => s.gameOver);
  const inGame = useGame((s) => s.room?.status === "running" || s.countdown !== null || s.question !== null);
  const error = useGame((s) => s.error);
  const connect = useGame((s) => s.connect);
  const telegramLogin = useGame((s) => s.telegramLogin);
  const clearError = useGame((s) => s.clearError);
  const leaveRoom = useGame((s) => s.leaveRoom);
  const logout = useGame((s) => s.logout);

  // Telegram Mini App ichida ochilsa — avtomatik kirish (token bo'lmasa).
  useEffect(() => {
    const tg = getTelegram();
    if (tg) {
      tg.ready();
      tg.expand?.();
      if (tg.initData && !useGame.getState().token) telegramLogin();
    }
  }, [telegramLogin]);

  // Saqlangan token bo'lsa — sahifa ochilishida avtomatik ulanamiz.
  useEffect(() => {
    if (token) connect();
  }, [token, connect]);

  // Xatolarni 4 soniyada avto-yopish.
  useEffect(() => {
    if (!error) return;
    const id = setTimeout(clearError, 4000);
    return () => clearTimeout(id);
  }, [error, clearError]);

  let screen: JSX.Element;
  if (!token) screen = <AuthPage />;
  else if (gameOver) screen = <ResultPage />;
  else if (inGame) screen = <PlayPage />;
  else if (room) screen = <LobbyPage />; // xona lobby (kutish)
  else screen = <HomePage />; // tab: O'ynash / O'rganish

  const inRoom = !!room || inGame || !!gameOver;

  return (
    <div className="min-h-full">
      <header className="flex items-center justify-between border-b border-slate-200 bg-white px-4 py-2 text-sm">
        <span className="font-semibold text-indigo-600">QuizArena</span>
        <div className="flex items-center gap-3">
          <LangToggle current={i18n.language} />
          {token && <span className="text-slate-500">{displayName || t("common.player")}</span>}
          {token &&
            (inRoom ? (
              <button onClick={leaveRoom} className="rounded-md px-2 py-1 text-slate-500 hover:bg-slate-100">
                {t("common.leave")}
              </button>
            ) : (
              <button onClick={logout} className="rounded-md px-2 py-1 text-slate-500 hover:bg-slate-100">
                {t("common.logout")}
              </button>
            ))}
        </div>
      </header>

      {token && status !== "online" && status !== "offline" && (
        <div className="bg-amber-100 py-1 text-center text-xs text-amber-700">
          {status === "connecting" ? t("common.connecting") : t("common.reconnecting")}
        </div>
      )}

      {screen}

      {error && (
        <div
          className="fixed bottom-4 left-1/2 -translate-x-1/2 cursor-pointer rounded-lg bg-red-600 px-4 py-2 text-sm text-white shadow-lg"
          onClick={clearError}
        >
          {error} ✕
        </div>
      )}
    </div>
  );
}

function LangToggle({ current }: { current: string }) {
  const lang = current.startsWith("en") ? "en" : "uz";
  return (
    <div className="flex overflow-hidden rounded-md border border-slate-200 text-xs">
      {(["uz", "en"] as const).map((l) => (
        <button
          key={l}
          onClick={() => setLang(l)}
          className={lang === l ? "bg-indigo-600 px-2 py-0.5 text-white" : "px-2 py-0.5 text-slate-500"}
        >
          {l.toUpperCase()}
        </button>
      ))}
    </div>
  );
}
