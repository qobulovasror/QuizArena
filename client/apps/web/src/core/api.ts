// REST API client (auth). Dev'da Vite proxy orqali Go backend'ga.

export interface User {
  id: string;
  username: string | null;
  email: string | null;
  isGuest: boolean;
  role: string;
}

export interface AuthResp {
  token: string;
  user: User;
}

export interface SubjectInfo {
  id: string;
  slug: string;
  name: string;
  icon: string | null;
}

async function post<T>(path: string, body: unknown): Promise<T> {
  const r = await fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!r.ok) {
    const e = (await r.json().catch(() => ({ error: r.statusText }))) as { error?: string };
    throw new Error(e.error || "so'rov xatosi");
  }
  return r.json() as Promise<T>;
}

async function get<T>(path: string): Promise<T> {
  const r = await fetch(path);
  if (!r.ok) throw new Error("so'rov xatosi");
  return r.json() as Promise<T>;
}

async function authGet<T>(path: string, token: string): Promise<T> {
  const r = await fetch(path, { headers: { Authorization: `Bearer ${token}` } });
  if (!r.ok) throw new Error("so'rov xatosi");
  return r.json() as Promise<T>;
}

async function authPost<T>(path: string, body: unknown, token: string): Promise<T> {
  const r = await fetch(path, {
    method: "POST",
    headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
    body: JSON.stringify(body),
  });
  if (!r.ok) throw new Error("so'rov xatosi");
  return r.json() as Promise<T>;
}

export interface SrsCard {
  questionId: string;
  type: string;
  prompt: string;
  answer: string;
  explanation?: string;
}

export interface MasteryItem {
  subject: string;
  category: string;
  mastery: number;
  attempts: number;
}

export interface AssessQuestion {
  questionId: string;
  type: string;
  prompt: string;
  options?: { id: string; text: string }[];
}

export interface AssessAnswer {
  questionId: string;
  choice: unknown;
}

export const api = {
  guest: () => post<AuthResp>("/api/auth/guest", {}),
  register: (username: string, email: string, password: string) =>
    post<AuthResp>("/api/auth/register", { username, email, password }),
  login: (email: string, password: string) =>
    post<AuthResp>("/api/auth/login", { email, password }),
  telegram: (initData: string) => post<AuthResp>("/api/auth/telegram", { initData }),
  subjects: () => get<SubjectInfo[]>("/api/subjects"),
  srsDue: (subject: string, token: string) =>
    authGet<SrsCard[]>(`/api/me/srs/due?subject=${encodeURIComponent(subject)}`, token),
  srsReview: (questionId: string, grade: number, token: string) =>
    authPost<{ ok: boolean }>("/api/srs/review", { questionId, grade }, token),
  mastery: (token: string) => authGet<MasteryItem[]>("/api/me/mastery", token),
  assessQuestions: (subject: string, token: string) =>
    authGet<AssessQuestion[]>(`/api/me/assessment?subject=${encodeURIComponent(subject)}`, token),
  assessSubmit: (answers: AssessAnswer[], token: string) =>
    authPost<{ correct: number; total: number }>("/api/me/assessment/submit", { answers }, token),
};
