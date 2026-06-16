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

export const api = {
  guest: () => post<AuthResp>("/api/auth/guest", {}),
  register: (username: string, email: string, password: string) =>
    post<AuthResp>("/api/auth/register", { username, email, password }),
  login: (email: string, password: string) =>
    post<AuthResp>("/api/auth/login", { email, password }),
  subjects: () => get<SubjectInfo[]>("/api/subjects"),
};
