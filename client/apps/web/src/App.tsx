import { useGame } from "./core/store";
import { AuthPage } from "./pages/AuthPage";
import { LobbyPage } from "./pages/LobbyPage";
import { PlayPage } from "./pages/PlayPage";
import { ResultPage } from "./pages/ResultPage";

export default function App() {
  const token = useGame((s) => s.token);
  const gameOver = useGame((s) => s.gameOver);
  const inGame = useGame(
    (s) => s.room?.status === "running" || s.countdown !== null || s.question !== null,
  );
  const error = useGame((s) => s.error);
  const clearError = useGame((s) => s.clearError);

  let screen: JSX.Element;
  if (!token) screen = <AuthPage />;
  else if (gameOver) screen = <ResultPage />;
  else if (inGame) screen = <PlayPage />;
  else screen = <LobbyPage />;

  return (
    <div className="min-h-full">
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
