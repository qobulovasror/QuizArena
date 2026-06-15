# QuizArena

Ko'p-sohali, ko'p-o'yinchi **real-time quiz platforma** — uch niyat (🏆 raqobat,
📚 o'rganish, 📊 baholash) bitta o'yin engine'i ustida.

> To'liq texnik reja: [`../PLAN.md`](../PLAN.md). Bu papka — **Bosqich 0 skeleti**
> (poydevor; logika hali yo'q).

## Stack
- **Backend:** Go — `chi` (HTTP), `gorilla/websocket` + Hub, `pgx`+`sqlc` (Postgres), `goose` (migration)
- **Frontend:** React + Vite + TypeScript + Tailwind + shadcn/ui (web/telegram), React Native + Expo (mobile)
- **Umumiy:** TS `core` (Zustand, WS/API client, tiplar)
- **DB:** PostgreSQL (jonli holat — in-memory `StateStore`)
- **Real-time:** sof WebSocket + JSON protokol (Socket.IO emas)

## Struktura (3 qism)
```
quizarena/
├── server/     # Go backend (cmd + internal/*)
├── client/     # TS monorepo: packages/{core,ui-web} + apps/{web,telegram,native}
└── shared/     # Kontrakt: protocol/ (WS) + openapi/ (REST) — manba haqiqat
```

## Boshlash (Bosqich 0)
```bash
cp .env.example .env          # sozlamalarni to'ldiring
docker compose up -d postgres # faqat Postgres
make run                      # Go serverni ishga tushirish (skelet)
```

## Yo'l xaritasi (qisqacha — to'liq §11 `../PLAN.md`)
- **B0** Poydevor (shu skelet) · **B1** 🏆 Raqobat MVP (host-led) + Web
- **B2** Telegram + ko'p soha · **B3** Ko'p metod + so'z o'yinlari + Mobile (RN)
- **B4** 📚 SRS + 📊 mastery · **B5** bot/matchmaking/turnir · **B6** admin + production + scaling
