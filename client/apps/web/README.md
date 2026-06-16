# QuizArena — Web (Vite + React + TS + Tailwind)

Bosqich 1 web ilova. Logika `src/core/` da (keyin `packages/core`ga ko'chiriladi).

```
src/
├── core/          # PLATFORMADAN MUSTAQIL mantiq
│   ├── protocol.ts  WS xabar tiplari (shared/protocol bilan mos)
│   ├── api.ts       REST (auth)
│   └── store.ts     Zustand: auth + WS + o'yin holati
├── components/ui/ # shadcn-uslub (Button, Input, Card)
├── pages/         # Auth → Lobby → Play → Result
└── App.tsx        # holatga qarab ekran tanlash
```

## Ishga tushirish (to'liq stack)

1. **Backend + Postgres** (boshqa terminalda — `server/` ildizidan yoki repo ildizidan):
   ```bash
   docker compose up -d postgres      # yoki lokal Postgres
   cd server && make migrate-up && make seed   # (ildizdan: make migrate-up / make seed)
   PORT=8099 DATABASE_URL=... go run ./cmd/server
   ```
2. **Frontend**:
   ```bash
   cd client/apps/web
   npm install
   npm run dev            # http://localhost:5173
   ```
   Vite `/api` va `/ws` ni `localhost:8099` (backend) ga proxy qiladi.

## Build
```bash
npm run build            # tsc (strict) + vite → dist/
```
