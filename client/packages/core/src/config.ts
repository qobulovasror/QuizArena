// Platforma sozlamasi — umumiy mantiq (web/RN/Telegram) uchun.
// Web: bo'sh (relative URL + location'dan WS). React Native: absolute o'rnatiladi.

let apiBase = ""; // REST asosi (mas. "http://10.0.2.2:8080"); bo'sh → relative
let wsBase = ""; //  WS asosi (mas. "ws://10.0.2.2:8080"); bo'sh → location'dan (web)

// configureCore — platforma kirish nuqtasida (RN App) chaqiriladi.
export function configureCore(opts: { apiBase?: string; wsBase?: string }): void {
  if (opts.apiBase !== undefined) apiBase = opts.apiBase;
  if (opts.wsBase !== undefined) wsBase = opts.wsBase;
}

export function getApiBase(): string {
  return apiBase;
}

export function getWsBase(): string {
  return wsBase;
}
