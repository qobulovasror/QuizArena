# apps/telegram — Telegram Mini App

> **Alohida kod yo'q** — Telegram Mini App `apps/web` ning **aynan o'sha build**ini
> Telegram WebView ichida ishlatadi (PLAN.md §1.5: "aynan o'sha React ilova"). Bu yerda
> faqat sozlash hujjati. Kod dublikati ataylab qilinmaydi.

## Nima allaqachon tayyor (apps/web ichida)

- **SDK**: `apps/web/index.html` da `telegram-web-app.js` ulangan (oddiy brauzerda zararsiz).
- **Avtomatik kirish**: `App.tsx` Telegram ichida ochilsa `tg.ready()`+`expand()` qiladi va
  `initData` bo'lsa `store.telegramLogin()` orqali **parolsiz** kiradi (`@core/telegram`, `@core/store`).
- **Server tekshiruvi**: `initData` HMAC-SHA256 bilan serverda tasdiqlanadi
  (`server/internal/auth/telegram.go`).
- **Bot**: `server/internal/telegram/bot.go` — `/start` → "O'ynash" tugmasi Mini App'ni ochadi
  (`MINI_APP_URL` env).

## Mini App'ni ishga tushirish (deploy)

1. **Web build**: `cd apps/web && npm run build` → `dist/` ni HTTPS hostingga joylashtir
   (Telegram faqat HTTPS qabul qiladi).
2. **BotFather**: bot yarat, `TELEGRAM_BOT_TOKEN` ni serverga ber. BotFather'da Mini App
   (Web App) URL'ini hosting manziliga qo'y.
3. **Server env**: `TELEGRAM_BOT_TOKEN` + `MINI_APP_URL=https://<hosting>` (bot `/start`
   tugmasi shu URL'ni ochadi).
4. Telegram'da botni och → "O'ynash" → Mini App o'sha web ilovani ko'rsatadi, foydalanuvchi
   avtomatik kiradi.

## Kelajak (ixtiyoriy)
- `@telegram-apps/sdk-react` bilan boyroq integratsiya (haptik, tema, MainButton).
- PWA manifest (`vite-plugin-pwa`) — o'rnatiladigan web ilova uchun.
- `auth_date` yangiligi tekshiruvi (replay himoyasi) — `telegram.go` dagi TODO.
