# QuizArena — shared (kontrakt / manba haqiqat)

Backend (Go) va frontend (TS) **bir xil shartnomaga** amal qilishi uchun
til-mustaqil kontrakt shu yerda saqlanadi.

```
shared/
├── protocol/   # WebSocket JSON xabar tiplari (PLAN.md §8) — { type, data }
└── openapi/    # REST API shartnomasi (OpenAPI) — auth, subjects, history, ...
```

Maqsad: bu yerdan **ikkala tomon uchun tiplar generatsiya** qilish
(Go struct + TS interface), shunda event/payload nomlari hech qachon ajralib qolmaydi.

> Skelet bosqichi: bo'sh. WebSocket protokoli **Bosqich 1** dan oldin tip-tip
> aniqlanadi (PLAN.md §8 asosida).
