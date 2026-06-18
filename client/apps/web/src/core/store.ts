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
import { api } from "./api";
import type { User, SubjectInfo } from "./api";
import { getTelegram } from "./telegram";

let socket: WebSocket | null = null;
let intentional = false; // ataylab yopilganda reconnect qilinmaydi
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

const AUTH_KEY = "quizarena_auth";

function saveAuth(token: string, user: User, displayName: string) {
  localStorage.setItem(AUTH_KEY, JSON.stringify({ token, user, displayName }));
}
function loadAuth(): { token: string; user: User; displayName: string } | null {
  try {
    return JSON.parse(localStorage.getItem(AUTH_KEY) || "null");
  } catch {
    return null;
  }
}
function clearAuth() {
  localStorage.removeItem(AUTH_KEY);
}

export type ConnStatus = "offline" | "connecting" | "online" | "reconnecting";

interface CreateOpts {
  subjectId: string;
  mode: string;
  questionCount: number;
  timePerQ: number;
}

interface MyAnswer {
  index: number;
  choice: { optionId?: string; value?: number | boolean };
}

interface GameStore {
  token: string | null;
  user: User | null;
  displayName: string;
  status: ConnStatus;
  error: string | null;

  selfUserId: string | null;
  sessionId: string | null;
  resumeToken: string | null;

  room: RoomStateData | null;
  countdown: number | null;
  question: QuestionShowData | null;
  answeredIndex: number | null;
  myAnswer: MyAnswer | null;
  reveal: QuestionRevealData | null;
  gameOver: GameOverData | null;
  eliminated: boolean;
  subjects: SubjectInfo[];

  setDisplayName: (n: string) => void;
  setAuth: (token: string, user: User) => void;
  telegramLogin: () => Promise<void>;
  connect: () => void;
  loadSubjects: () => Promise<void>;
  createRoom: (opts: CreateOpts) => void;
  joinRoom: (code: string) => void;
  start: () => void;
  answer: (choice: MyAnswer["choice"]) => void;
  leaveRoom: () => void;
  logout: () => void;
  clearError: () => void;
  newGame: () => void;
}

function send(type: string, data: unknown) {
  if (socket && socket.readyState === WebSocket.OPEN) {
    socket.send(JSON.stringify({ type, data }));
  }
}

const persisted = loadAuth();

export const useGame = create<GameStore>((set, get) => ({
  token: persisted?.token ?? null,
  user: persisted?.user ?? null,
  displayName: persisted?.displayName ?? "",
  status: "offline",
  error: null,

  selfUserId: null,
  sessionId: null,
  resumeToken: null,

  room: null,
  countdown: null,
  question: null,
  answeredIndex: null,
  myAnswer: null,
  reveal: null,
  gameOver: null,
  eliminated: false,
  subjects: [],

  setDisplayName: (n) => set({ displayName: n }),

  setAuth: (token, user) => {
    saveAuth(token, user, get().displayName);
    set({ token, user });
  },

  telegramLogin: async () => {
    const tg = getTelegram();
    if (!tg?.initData) return;
    try {
      const res = await api.telegram(tg.initData);
      const name = tg.initDataUnsafe?.user?.first_name || get().displayName || "O'yinchi";
      set({ displayName: name });
      get().setAuth(res.token, res.user); // token → App avtomatik ulanadi
    } catch {
      /* Telegram auth ishlamadi — oddiy auth ekraniga tushadi */
    }
  },

  loadSubjects: async () => {
    try {
      set({ subjects: await api.subjects() });
    } catch {
      /* jim o'tamiz */
    }
  },

  connect: () => {
    const token = get().token;
    if (!token) return;
    if (socket && (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING)) {
      return;
    }
    intentional = false;
    open(set, get, token);
  },

  createRoom: ({ subjectId, mode, questionCount, timePerQ }) =>
    send("room:create", {
      subjectId,
      mode,
      questionCount,
      timePerQ,
      displayName: get().displayName || "O'yinchi",
    }),

  joinRoom: (code) =>
    send("room:join", { code: code.toUpperCase(), displayName: get().displayName || "O'yinchi" }),

  start: () => send("game:start", {}),

  answer: (choice) => {
    const q = get().question;
    if (!q || get().answeredIndex === q.index) return;
    send("answer:submit", { questionIndex: q.index, choice });
    set({ answeredIndex: q.index, myAnswer: { index: q.index, choice } });
  },

  leaveRoom: () => {
    send("room:leave", {});
    set({ room: null, sessionId: null, resumeToken: null, question: null, reveal: null, gameOver: null, countdown: null, answeredIndex: null, myAnswer: null, eliminated: false });
  },

  logout: () => {
    intentional = true;
    if (reconnectTimer) clearTimeout(reconnectTimer);
    socket?.close();
    socket = null;
    clearAuth();
    set({
      token: null, user: null, status: "offline",
      room: null, sessionId: null, resumeToken: null,
      question: null, reveal: null, gameOver: null, countdown: null, answeredIndex: null, myAnswer: null, eliminated: false,
    });
  },

  clearError: () => set({ error: null }),

  newGame: () =>
    set({ room: null, sessionId: null, resumeToken: null, question: null, reveal: null, gameOver: null, countdown: null, answeredIndex: null, myAnswer: null, eliminated: false }),
}));

function open(set: (p: Partial<GameStore>) => void, get: () => GameStore, token: string) {
  const first = get().status === "offline";
  set({ status: first ? "connecting" : "reconnecting" });

  const scheme = location.protocol === "https:" ? "wss" : "ws";
  const ws = new WebSocket(`${scheme}://${location.host}/ws?token=${encodeURIComponent(token)}`);
  socket = ws;

  ws.onopen = () => {
    set({ status: "online" });
    // O'yin o'rtasida uzilgan bo'lsak — qaytib qo'shilamiz.
    const s = get();
    if (s.sessionId && s.resumeToken) {
      send("room:resume", { sessionId: s.sessionId, resumeToken: s.resumeToken });
    }
  };

  ws.onclose = () => {
    socket = null;
    if (intentional) {
      set({ status: "offline" });
      return;
    }
    set({ status: "reconnecting" });
    reconnectTimer = setTimeout(() => open(set, get, token), 1000);
  };

  ws.onmessage = (ev) => {
    let env: Envelope;
    try {
      env = JSON.parse(ev.data as string) as Envelope;
    } catch {
      return;
    }
    handle(env, set, get);
  };
}

function handle(env: Envelope, set: (p: Partial<GameStore>) => void, get: () => GameStore) {
  const amIOut = (board: { userId: string; eliminated?: boolean }[]) =>
    board.some((e) => e.userId === get().selfUserId && e.eliminated);
  switch (env.type) {
    case "room:joined": {
      const d = env.data as RoomJoinedData;
      set({ selfUserId: d.userId, sessionId: d.sessionId, resumeToken: d.resumeToken });
      break;
    }
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
        myAnswer: null,
        reveal: null,
      });
      break;
    case "question:reveal": {
      const d = env.data as QuestionRevealData;
      set({ reveal: d });
      if (amIOut(d.leaderboard)) set({ eliminated: true });
      break;
    }
    case "game:over": {
      const d = env.data as GameOverData;
      set({ gameOver: d });
      if (amIOut(d.finalLeaderboard)) set({ eliminated: true });
      break;
    }
    case "error": {
      const e = env.data as ErrorData;
      if (e.code === "ROOM_NOT_FOUND") {
        // Sessiya yo'qolgan (o'yin tugagan) — lobbyga qaytamiz, jim.
        set({ room: null, sessionId: null, resumeToken: null, question: null, reveal: null, gameOver: null, countdown: null });
      } else {
        set({ error: e.message });
      }
      break;
    }
  }
}
