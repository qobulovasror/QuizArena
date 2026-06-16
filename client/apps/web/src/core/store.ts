import { create } from "zustand";
import type {
  Envelope,
  RoomStateData,
  RoomJoinedData,
  CountdownData,
  QuestionShowData,
  QuestionRevealData,
  GameOverData,
  ErrorData,
} from "./protocol";
import type { User } from "./api";

let socket: WebSocket | null = null;

interface CreateOpts {
  questionCount: number;
  timePerQ: number;
}

interface GameStore {
  // auth
  token: string | null;
  user: User | null;
  displayName: string;
  // ulanish
  connected: boolean;
  error: string | null;
  // o'yin holati
  selfUserId: string | null;
  room: RoomStateData | null;
  countdown: number | null;
  question: QuestionShowData | null;
  answeredIndex: number | null;
  reveal: QuestionRevealData | null;
  gameOver: GameOverData | null;

  // amallar
  setDisplayName: (n: string) => void;
  setAuth: (token: string, user: User) => void;
  connect: () => void;
  createRoom: (opts: CreateOpts) => void;
  joinRoom: (code: string) => void;
  start: () => void;
  answer: (optionId: string) => void;
  clearError: () => void;
  newGame: () => void;
}

function send(type: string, data: unknown) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type, data }));
  }
}

export const useGame = create<GameStore>((set, get) => ({
  token: null,
  user: null,
  displayName: "",
  connected: false,
  error: null,
  selfUserId: null,
  room: null,
  countdown: null,
  question: null,
  answeredIndex: null,
  reveal: null,
  gameOver: null,

  setDisplayName: (n) => set({ displayName: n }),

  setAuth: (token, user) => set({ token, user }),

  connect: () => {
    const token = get().token;
    if (socket || !token) return;
    const scheme = location.protocol === "https:" ? "wss" : "ws";
    socket = new WebSocket(`${scheme}://${location.host}/ws?token=${encodeURIComponent(token)}`);

    socket.onopen = () => set({ connected: true });
    socket.onclose = () => {
      set({ connected: false });
      socket = null;
    };
    socket.onmessage = (ev) => {
      let env: Envelope;
      try {
        env = JSON.parse(ev.data as string) as Envelope;
      } catch {
        return;
      }
      handle(env, set);
    };
  },

  createRoom: ({ questionCount, timePerQ }) =>
    send("room:create", {
      subjectId: "english",
      mode: "classic",
      questionCount,
      timePerQ,
      displayName: get().displayName || "O'yinchi",
    }),

  joinRoom: (code) =>
    send("room:join", { code: code.toUpperCase(), displayName: get().displayName || "O'yinchi" }),

  start: () => send("game:start", {}),

  answer: (optionId) => {
    const q = get().question;
    if (!q || get().answeredIndex === q.index) return;
    send("answer:submit", { questionIndex: q.index, choice: { optionId } });
    set({ answeredIndex: q.index });
  },

  clearError: () => set({ error: null }),

  newGame: () => set({ room: null, question: null, reveal: null, gameOver: null, countdown: null, answeredIndex: null }),
}));

function handle(env: Envelope, set: (p: Partial<GameStore>) => void) {
  switch (env.type) {
    case "room:joined":
      set({ selfUserId: (env.data as RoomJoinedData).userId });
      break;
    case "room:state": {
      const room = env.data as RoomStateData;
      if (room.status === "lobby") set({ room, gameOver: null });
      else set({ room });
      break;
    }
    case "game:countdown":
      set({ countdown: (env.data as CountdownData).secondsLeft });
      break;
    case "question:show":
      set({
        question: env.data as QuestionShowData,
        countdown: null,
        answeredIndex: null,
        reveal: null,
      });
      break;
    case "question:reveal":
      set({ reveal: env.data as QuestionRevealData });
      break;
    case "game:over":
      set({ gameOver: env.data as GameOverData });
      break;
    case "error":
      set({ error: (env.data as ErrorData).message });
      break;
  }
}
