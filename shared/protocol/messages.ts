// QuizArena — WebSocket JSON protokol tiplari (v1).
// MANBA SHARTNOMA: ./README.md. Go tomoni: server/internal/ws/protocol.go.
// Ikkala fayl qo'lda sinxron saqlanadi (keyin codegen bilan almashtiriladi).

// ---- Konvert ----
export interface Envelope<T = unknown> {
  type: MsgType;
  data?: T;
  id?: string; // ixtiyoriy client→server korrelyatsiya
}

export type MsgType = ClientMsgType | ServerMsgType;

export type ClientMsgType =
  | "room:create"
  | "room:join"
  | "room:resume"
  | "room:leave"
  | "game:start"
  | "answer:submit";

export type ServerMsgType =
  | "room:state"
  | "room:joined"
  | "game:countdown"
  | "question:show"
  | "answer:ack"
  | "question:reveal"
  | "player:scored"
  | "game:over"
  | "error";

// ---- Umumiy tuzilmalar ----
export type RoomStatus = "lobby" | "running" | "finished";
export type GameMode =
  | "classic" | "survival" | "time_attack" | "team" | "practice" | "assessment";
export type OpponentKind = "human" | "bot" | "mixed";

export interface Player {
  userId: string;
  name: string;
  score: number;
  connected: boolean;
  isBot?: boolean;
}

export interface RoomConfig {
  subjectId: string;
  mode: GameMode;
  questionCount: number;
  timePerQ: number; // soniya
}

export interface Option {
  id: string; // opaque — server shuffle qiladi
  text: string;
}

export interface LeaderboardEntry {
  userId: string;
  name: string;
  score: number;
  correctCnt: number;
  rank: number;
}

// `choice` / `correct` savol turiga qarab o'zgaradi (README §6).
export type AnswerChoice =
  | { optionId: string }            // mcq
  | { value: boolean }              // true_false
  | { optionIds: string[] }         // multi_select
  | { text: string }                // type_answer / fill_blank
  | { value: number };              // numeric

// ---- Client → Server yuklar ----
export interface RoomCreateData {
  subjectId: string;
  categoryId?: string;
  mode: GameMode;
  opponent?: OpponentKind;
  questionCount: number;
  timePerQ: number;
  displayName: string; // host ham o'yinchi
}
export interface RoomJoinData {
  code: string;
  displayName: string;
}
export interface RoomResumeData {
  sessionId: string;
  resumeToken: string;
}
export interface AnswerSubmitData {
  questionIndex: number;
  choice: AnswerChoice;
}

// ---- Server → Client yuklar ----
export interface RoomStateData {
  sessionId: string;
  code: string;
  host: string;
  status: RoomStatus;
  config: RoomConfig;
  players: Player[];
}
export interface RoomJoinedData {
  sessionId: string;
  userId: string;
  resumeToken: string;
}
export interface CountdownData {
  secondsLeft: number;
}
export interface QuestionShowData {
  index: number;
  total: number;
  type: string; // §5 katalog turi
  prompt: string;
  options?: Option[];
  deadlineTs: number; // server epoch ms
}
export interface AnswerAckData {
  index: number;
  accepted: boolean; // to'g'rilikni OSHKOR QILMAYDI
}
export interface QuestionRevealData {
  index: number;
  correct: unknown; // tur bo'yicha shakl (README §6)
  explanation?: string;
  leaderboard: LeaderboardEntry[];
}
export interface PlayerScoredData {
  leaderboard: LeaderboardEntry[];
}
export interface GameOverData {
  finalLeaderboard: LeaderboardEntry[];
}

export type ErrorCode =
  | "BAD_REQUEST"
  | "INVALID_MESSAGE"
  | "ROOM_NOT_FOUND"
  | "ROOM_FULL"
  | "NOT_HOST"
  | "GAME_ALREADY_STARTED"
  | "ALREADY_ANSWERED"
  | "DEADLINE_PASSED"
  | "UNAUTHENTICATED"
  | "INTERNAL";

export interface ErrorData {
  code: ErrorCode;
  message: string;
}

// ---- Diskriminatsiyalangan (tur ↔ yuk) xaritasi ----
export interface ClientMessageMap {
  "room:create": RoomCreateData;
  "room:join": RoomJoinData;
  "room:resume": RoomResumeData;
  "room:leave": Record<string, never>;
  "game:start": Record<string, never>;
  "answer:submit": AnswerSubmitData;
}
export interface ServerMessageMap {
  "room:state": RoomStateData;
  "room:joined": RoomJoinedData;
  "game:countdown": CountdownData;
  "question:show": QuestionShowData;
  "answer:ack": AnswerAckData;
  "question:reveal": QuestionRevealData;
  "player:scored": PlayerScoredData;
  "game:over": GameOverData;
  error: ErrorData;
}

// Yordamchilar: tip-xavfsiz konvert qurish/o'qish.
export function msg<K extends ClientMsgType>(
  type: K,
  data: ClientMessageMap[K],
  id?: string,
): Envelope<ClientMessageMap[K]> {
  return { type, data, id };
}
