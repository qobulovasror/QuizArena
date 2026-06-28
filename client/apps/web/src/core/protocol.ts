// WebSocket protokol tiplari (manba: shared/protocol/messages.ts).
// B1 web uchun kerakli qism. Keyin packages/core'ga ko'chiriladi / codegen.

export interface Envelope<T = unknown> {
  type: string;
  data?: T;
  id?: string;
}

export type RoomStatus = "lobby" | "running" | "finished";

export interface Player {
  userId: string;
  name: string;
  score: number;
  connected: boolean;
  isBot?: boolean;
  eliminated?: boolean;
  team?: string;
}

export interface RoomConfig {
  subjectId: string;
  mode: string;
  questionCount: number;
  timePerQ: number;
}

export interface Option {
  id: string;
  text: string;
}

export interface LeaderboardEntry {
  userId: string;
  name: string;
  score: number;
  correctCnt: number;
  rank: number;
  eliminated?: boolean;
  team?: string;
}

export interface TeamStanding {
  team: string;
  score: number;
  correctCnt: number;
  rank: number;
}

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
  type: string;
  prompt: string;
  options?: Option[];
  deadlineTs: number;
}

export interface QuestionRevealData {
  index: number;
  correct: { optionId?: string } | unknown;
  explanation?: string;
  leaderboard: LeaderboardEntry[];
  teams?: TeamStanding[];
}

export interface GameOverData {
  finalLeaderboard: LeaderboardEntry[];
  teams?: TeamStanding[];
}

export interface ErrorData {
  code: string;
  message: string;
}
