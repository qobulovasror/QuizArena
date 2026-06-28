# QuizArena — Ijro Rejasi (ROADMAP)

> Bu **ijro hujjati** (PLAN.md = spetsifikatsiya). Bu yerda: hozirgi holat, qolgan
> ishlar **ketma-ketligi**, ishlov **workflow**i va **agent qoidalari**. PLAN.md §-larga
> havola qiladi. Til: **o'zbekcha** (kod/izoh/commit bilan bir xil).

---

## 1. Hozirgi holat (2026-06)

### ✅ Tugagan (commit qilingan yoki shu sessiyada)
| Bosqich | Nima | Holat |
|---|---|---|
| B0 | Poydevor: config, Docker, sqlc/goose quvuri, `state.Store`, Hub | ✅ |
| B1 | Auth (mehmon+akkaunt+JWT), host-led xona, classic, server-authoritative, reconnect, `deadlineTs`, persist, i18n, English provider | ✅ |
| B2 | math/general/**programming** providerlar, admin CRUD, **9 savol turi** (mcq, true_false, numeric, type_answer, multi_select, **match, ordering, categorize, cloze**) — barchasi end-to-end o'ynaladigan | ✅ |
| B3 | Rejimlar: classic, survival, **time_attack** (per-player), **team** (jamoa reytingi) | ✅ |
| B4 | SRS (practice, SM-2), Mastery (assessment, EMA) | ✅ |
| B5 | **BotPlayer** (sinxron rejimlar), **Matchmaking 1v1 + ELO** (`user_rating`, duel, bot-fallback) | ✅ (commit kutmoqda) |
| — | `answers_log` auditi (DB-savollar uchun) | ✅ |
| Telegram | Mini App auth + minimal bot (`/start`) | ✅ |

### ⏳ Ochiq texnik qarz / qo'llanilmagan
- **Matchmaking stage commit qilinmagan** (`?? matchmaking.go, rating/, 00006_rating.sql, ...`).
- **Migratsiya 00006 qo'llanilmagan** — `make migrate-up` kerak (goose **o'rnatilmagan**, DB ulanmagan).
- **Branch**: `b2-b3-rejimlar` 4 commit + ishchi o'zgarishlar — nomidan oshib ketgan, `main`ga merge kutmoqda.
- **Begona ishchi fayllar** (men yozmaganman): `security.go`, `security_test.go`, `config.go`, `router.go`, `.env.example` — rate-limit/xavfsizlik ishi davom etmoqda. **Tegmaslik / egasidan so'rash**.
- `CLAUDE.md`, `PLAN.md`, `.github/` — **hech qachon commit qilinmagan** (`??`).
- Reyting raqami UI'da ko'rinmaydi (`GET /api/me/rating` bor, ko'rinish yo'q).
- Yangi 4 turning reveal'da to'g'ri javobi inline ko'rsatilmaydi.

---

## 2. Qolgan ishlar — reja qayta tahlili (PLAN bo'yicha)

### Bosqich 3 qoldig'i
- **So'z o'yinlari turlari** (§5 3-guruh): `anagram`, `hangman` (oson), `word_search` (murakkab UI).
- **React Native (Expo)** ilova (`apps/native`) — `core`ga bog'liq (3.6-bandga qarang).

### Bosqich 5 qoldig'i
- **Turnirlar** (§4, §8): `tournaments` + `tournament_entries` migratsiyalari, CRUD, asinxron natija.
- **Bot kengaytmasi**: qiyinlik darajalari (oson/o'rta/qiyin), `time_attack` uchun bot.

### Bosqich 6
- **Global reyting** `GET /api/leaderboard/global` (§8) — mavjud `game_results`dan, oson.
- **Profil / statistika / reyting ko'rinishi** UI.
- **Media turlari** (§5 4-guruh): `image_choice`, `hotspot`, `audio` + obyekt-saqlash (MinIO/S3).
- **Kod turlari** (§5 5-guruh): `code_output`, `code_debug`, `code_fill`.
- **Xavfsizlik**: rate-limit kengaytirish (davom etayotgan `security.go`), CORS, monitoring, CI/CD.
- **Haqiqiy AI** (ixtiyoriy): savol/izoh generatsiya, adaptiv qiyinlik, tutor.
- **Scaling** (faqat yuk talab qilganda): Redis `state.Store` + pub/sub, N instans, sticky LB.

### Arxitektura qarzi (kesib o'tuvchi)
- **`packages/core` ajratish**: `apps/web/src/core` → `packages/core` (RN/Telegram ulashishi uchun) — §1.5.
- **`apps/telegram`** wrapper (web build, Mini App manifest).

---

## 3. Maqsadli ketma-ketlik (qaysidan boshlash)

Tartib **qiymat ÷ xavf ÷ bog'liqlik** bo'yicha. Har band — bitta branch + bitta stage.

### Tier 0 — Gigiyena (avval)
1. **Matchmaking stage commit** + `main`ga merge strategiyasini hal qilish.
2. **Migratsiya 00006 qo'llash** (`make tools` → goose, `make migrate-up`) — ELO saqlanishi uchun. *Ops qadam, DB kerak.*
3. `CLAUDE.md`/`PLAN.md`/`ROADMAP.md` ni commit qilish.

### Tier 1 — Boshlangan ishni yopish (kichik, ko'rinadigan qiymat)
4. **Reyting + Global leaderboard**: `/api/leaderboard/global` (`game_results`dan) + reyting/leaderboard ko'rinishi (profil yoki home). ELO halqasini ko'rinadigan qiladi.
5. **Frontend sayqal**: `AssessPage` `alert()`→toast, jim `.catch()`lar, TS `any`/non-null; yangi turlar reveal'ida to'g'ri javob inline.

### Tier 2 — B5/B3 qoldig'i
6. **Turnirlar** (B5): migratsiyalar + CRUD + asinxron natija + UI. Raqobat ustunini yakunlaydi.
7. **So'z o'yinlari** (B3): `anagram`, `hangman` (qtype + client UI + seed); `word_search` keyin.
8. **Bot kengaytmasi**: qiyinlik darajalari + `time_attack` bot.

### Tier 3 — Arxitektura & platformalar
9. **`packages/core` ajratish** (arxitektura qarzi) — RN/Telegram'ni ochadi.
10. **`apps/telegram`** wrapper → **React Native** ilova (`apps/native`).

### Tier 4 — B6 production
11. **Kod turlari** (`code_*`) → **Media turlari** (`image_choice/hotspot/audio` + MinIO/S3).
12. **Xavfsizlik/CI-CD** (security.go yakuni, CORS, monitoring, GitHub Actions).
13. **Haqiqiy AI** (ixtiyoriy) → **Scaling** (Redis, faqat yuk testidan keyin).

> **Bog'liqliklar:** 10 ⟸ 9; 11 (media) ⟸ obyekt-saqlash infra; 13 (scaling) ⟸ yuk talabi.
> **Tavsiya:** 4 → 5 → 6 ketma-ket; 9 ni RN'dan oldin albatta.

---

## 4. Ishlov Workflow (har bir stage uchun)

```
1. BRANCH    — main'dan yangi branch (pattern: bosqich/qisqa-nom, mas. b5-turnir)
2. O'QISH    — PLAN.md tegishli § + atrofdagi kod (uslub, nomlash, mavjud strategiya)
3. SHARTNOMA — protokol o'zgarsa: shared/protocol/README.md AVVAL → keyin 3 mirror:
               server/internal/ws/protocol.go, shared/protocol/messages.ts,
               client/apps/web/src/core/protocol.ts
4. DB        — sxema o'zgarsa: migrations/ + queries/ → `make sqlc` (PATH'ga GOPATH/bin)
               → generated store.Queries. store/*.sql.go NI QO'LDA TAHRIRLAMA.
5. KOD       — atrofdagi uslubda; o'zbekcha izoh; ortiqcha abstraksiyasiz; faqat so'ralgan ish
6. TEST      — sof mantiq → unit; engine/WS → integratsiya (ws test harness, dialShared)
7. GATE      — server: `go build ./... && go vet ./... && go test ./...`
               client: `cd client/apps/web && npm run build`  (tsc+vite)
8. HISOBOT   — halol: nima sinaldi, nima yo'q, qaysi migratsiya qo'llanishi kerak
9. COMMIT    — FAQAT foydalanuvchi aytganda; main'da bo'lsa avval branch; o'zbekcha xabar,
               oldingi pattern; Claude co-author QO'SHMA (so'ralmasa); begona fayllarni qo'shma
10. OPS      — migrate/push/deploy = tashqi ta'sir → AVVAL tasdiqlat
```

### Maxsus quvurlar
- **sqlc:** `export PATH="$(go env GOPATH)/bin:$PATH"; cd server; sqlc generate` (sqlc o'rnatilgan).
- **goose (migratsiya):** `make tools` (goose o'rnatadi) → `DATABASE_URL=... make migrate-up`. **DB kerak.**
- **Yangi savol turi:** `qtype/qtype.go` strategiya + `For()` qator; render kerak bo'lsa `targets` (match/categorize); client `PlayPage` + `AdminPage` + seed.
- **Yangi rejim:** `modes/modes.go` strategiya + `For()`; per-player bo'lsa engine `run` dispatch.
- **Yangi soha:** provider yozish/ulanish, `cmd/server/main.go`da `registry.Register`.

---

## 5. Agent ishlash qoidalari (subagent/AI uchun)

### Kod invariантlari (BUZILMASIN)
1. **Server-authoritative:** to'g'ri javob (`correct`) hech qachon client payload'iga kirmaydi; baholash serverda (`qtype.Validate`). `options`/`targets` opaque `id`, server aralashtiradi.
2. **Holat ajratimi:** jonli (`state.Store`, vaqtinchalik) ≠ doimiy (Postgres). Aralashtirma.
3. **Protokol manba haqiqat** = `shared/protocol/README.md`. 3 mirror qo'lda sinxron.
4. **Plugin uslubi:** rejim/tur/provider = strategiya; asosiy engine o'zgarmaydi.
5. **Redis ataylab yo'q** — jonli holat `state.Store` interfeysi ortida qolsin.
6. **Generated kod** (`store/*.sql.go`) qo'lda tahrirlanmaydi — `sqlc`.

### Jarayon qoidalari
7. **Til:** izoh va commit **o'zbekcha**.
8. **O'qib keyin yoz:** mavjud uslub/strukturani o'qi, moslash; o'chirishdan oldin tushun.
9. **Minimal:** faqat so'ralgan ish; "kelajak uchun" kod va ortiqcha abstraksiya yo'q.
10. **Noaniqlikda** taxmin qilma — bitta aniq savol ber (effort/qiymat farq qilsa, variantlar bilan).
11. **Test gate** o'tmasdan "tayyor" dema; o'tmagan testni chiqishi bilan ayt.
12. **Buyruqlar:** build/vet/test = tekshiruv (ruxsat etilgan). migrate/push/deploy = tasdiqlat.
13. **Commit:** o'zboshimchalik bilan emas; co-author qo'shma; begona ishchi fayllarni (security.go, config.go, ...) qo'shma.
14. **Maxfiy** ma'lumot (token/parol/kalit) log/kod/commit'ga yozilmaydi.

### Parallel agentlar (fan-out)
- Faqat **mustaqil, faqat-o'qish** tahlil/qidiruvda parallel agent ishlat (mas. audit, kod xaritasi).
- Fayl yozadigan parallel agentlar **konflikt** bermasligi uchun alohida fayl/papkada ishlasin (yoki worktree).
- Har agentga **aniq ko'lam + fayl:satr formatida hisobot** talab qil.

---

## 6. Tezkor ma'lumotnoma
- Server testlari: `cd server && go test ./...`  (hozir **58 test**, build+vet toza)
- Client build: `cd client/apps/web && npm run build`
- Lokal port: server `PORT=8099 make run` (Vite proxy 8099'ga ketadi)
- Admin: `make admin EMAIL=x@y.com`
