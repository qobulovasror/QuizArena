# QuizArena — client (frontend monorepo)

pnpm workspaces. **`core` faqat mantiq** (UI yo'q) — uchchala platformada bir xil.

```
client/
├── packages/
│   ├── core/      # FAQAT MANTIQ: WS/API client, Zustand store, tiplar, i18n (UI yo'q!)
│   └── ui-web/    # shadcn/ui komponentlar (web + telegram uchun umumiy)
└── apps/
    ├── web/       # Vite + React + Tailwind        ┐ core + ui-web
    ├── telegram/  # Telegram Mini App (@telegram-apps/sdk-react) ┘
    └── native/    # React Native (Expo) + NativeWind — faqat core'ni import qiladi
```

> Skelet bosqichi: ilovalar hali bo'sh. Web ilova **Bosqich 1** da to'ldiriladi
> (qarang `../../PLAN.md` §11). Native — **Bosqich 3**.
