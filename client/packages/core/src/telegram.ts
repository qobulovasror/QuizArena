// Telegram Mini App SDK ko'prigi (telegram-web-app.js global obyekti).

interface TgUser {
  id: number;
  first_name?: string;
  username?: string;
}

export interface TgWebApp {
  initData: string;
  initDataUnsafe?: { user?: TgUser };
  ready: () => void;
  expand?: () => void;
}

// getTelegram — Telegram ichida ishlamasa null (yoki initData bo'sh).
// `typeof window` tekshiruvi — React Native muhitida (window yo'q) xavfsiz.
export function getTelegram(): TgWebApp | null {
  if (typeof window === "undefined") return null;
  return (window as unknown as { Telegram?: { WebApp?: TgWebApp } }).Telegram?.WebApp ?? null;
}
