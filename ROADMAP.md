# QuizArena — Ijro Rejasi (ROADMAP)

> Bu **ijro hujjati** (PLAN.md = spetsifikatsiya). Bu yerda: hozirgi holat, qolgan
> ishlar **ketma-ketligi**, ishlov **workflow**i va **agent qoidalari**. PLAN.md §-larga
> havola qiladi. Til: **o'zbekcha** (kod/izoh/commit bilan bir xil).

---

## 1. Hozirgi holat (2026-06)

### ✅ Tugagan (barchasi commit qilingan — 64 Go test, web build toza)
| Bosqich | Nima | Holat |
|---|---|---|
| B0 | Poydevor: config, Docker, sqlc/goose quvuri, `state.Store`, Hub | ✅ |
| B1 | Auth (mehmon+akkaunt+JWT), host-led xona, classic, server-authoritative, reconnect, `deadlineTs`, persist, i18n, English provider | ✅ |
| B2 | math/general/**programming** providerlar, admin CRUD, **10 savol turi** (mcq, true_false, numeric, type_answer, fill_blank, multi_select*, **match, ordering, categorize, cloze, anagram**) — `multi_select`'dan boshqasi end-to-end o'ynaladigan | ✅ |
| B3 | Rejimlar: classic, survival, **time_attack** (per-player), **team** (jamoa reytingi) | ✅ |
| B4 | SRS (practice, SM-2), Mastery (assessment, EMA) | ✅ |
| B5 | **BotPlayer** (qiyilik darajalari, sinxron + time_attack), **Matchmaking 1v1 + ELO** (`user_rating`, duel, bot-fallback), **Turnirlar** (asinxron, server-authoritative ball, leaderboard) | ✅ |
| B6 | **Global leaderboard** + **reyting ko'rinishi** UI (Reyting tab) | ✅ |
| — | `answers_log` auditi; **xavfsizlik** (rate-limit, secure headers, CORS); **frontend sayqal** | ✅ |
| Arxitektura | **`packages/core` ajratildi** (platforma-agnostik: `configureCore`, `@core/*`) | ✅ |
| Platforma | Telegram Mini App (auth + bot `/start`) + **RN skeleton** (`apps/native`, Expo) | ✅ / ⚠️ RN test qilinmagan |

### ⏳ Ochiq (faqat ops + §3.5/§3.6 qoldiqlar)
- **Migratsiyalar qo'llanilmagan**: `00006_rating`, `00007_tournaments` — `make migrate-up` kerak (goose o'rnatilmagan, DB ulanmagan). **Aks holda ELO/turnir ishlamaydi.** → §3.6
- **Branch** `b2-b3-rejimlar` `main`ga merge + **push qilinmagan**. → §3.6
- `CLAUDE.md`, `PLAN.md`, `.github/` — hali untracked. → §3.6
- Qolgan reja punktlari: **§3.5** (spelling, multi_select UI, `/stats`, bulk import, profil UI, ui-web) va **Tier 4** (media, kod, AI, scaling, CI/CD).
- Ma'lum cheklovlar: **§3.6** (RN test, startDuel TOCTOU, N+1, hangman/word_search, eski TODO'lar).

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
1. ✅ **Matchmaking stage commit** (xavfsizlik alohida commit; HEAD toza quriladi). ⏳ `main`ga merge + push — **ops** (§3.6).
2. ⏳ **Migratsiya qo'llash** (`make migrate-up` → `00006`, `00007`) — **ops, DB kerak** (§3.6). Migratsiya FAYLLARI yozildi ✅.
3. ✅/⏳ `ROADMAP.md` commit qilindi; `CLAUDE.md`/`PLAN.md` hali untracked (§3.6).

### Tier 1 — Boshlangan ishni yopish (kichik, ko'rinadigan qiymat)
4. ✅ **Reyting + Global leaderboard**: `/api/leaderboard/global` + reyting/leaderboard ko'rinishi (Reyting tab).
5. ✅ **Frontend sayqal**: `alert()`→inline, jim `.catch()`lar, TS `any`/non-null, yangi turlar reveal'ida to'g'ri javob.

### Tier 2 — B5/B3 qoldig'i
6. ✅ **Turnirlar** (B5): migratsiya + CRUD + asinxron + UI (server-authoritative ball, anti-cheat).
7. ✅ **`anagram`** (+ `type_answer`/`fill_blank` o'ynaladigan). ⏳ `hangman`/`word_search` — engine'ga to'g'ri kelmaydi (§3.6).
8. ✅ **Bot kengaytmasi**: qiyilik darajalari (oson/o'rta/qiyin) + `time_attack` bot.

### Tier 3 — Arxitektura & platformalar
9. ✅ **`packages/core` ajratish** — bajarildi (core endi platforma-agnostik: `configureCore`, `@core/*` alias).
10. ✅ **`apps/telegram`** (web build reuse, hujjat) + ✅ **React Native skeleton** (`apps/native`, Expo) — *jonli build Expo talab qiladi, sandbox'da test qilinmagan*.

### Tier 4 — B6 production
11. **Kod turlari** (`code_*`) → **Media turlari** (`image_choice/hotspot/audio` + MinIO/S3).
12. **Xavfsizlik/CI-CD** (security.go yakuni, CORS, monitoring, GitHub Actions).
13. **Haqiqiy AI** (ixtiyoriy) → **Scaling** (Redis, faqat yuk testidan keyin).

> **Bog'liqliklar:** 10 ⟸ 9; 11 (media) ⟸ obyekt-saqlash infra; 13 (scaling) ⟸ yuk talabi.
> **Tavsiya:** 4 → 5 → 6 ketma-ket; 9 ni RN'dan oldin albatta.

---

## 3.5 Qo'shimcha — PLAN'da bor, lekin yuqoridagi Tier'larda tushirib qoldirilgan

Bular PLAN.md'da yozilgan reja punktlari (Tier ro'yxatiga keyin qo'shilsin):

| Punkt | Rejada | Holat |
|---|---|---|
| **`spelling`** turi | §5 2-guruh | yo'q (`type_answer` patternida qilsa bo'ladi) |
| **`multi_select`** client UI | §5 1-guruh, B2 | server validatsiyasi bor, PlayPage chizmaydi |
| Bot **`/stats`** komandasi | Bosqich 2 | hozir faqat `/start` |
| **Bulk import** (admin) | §8, Bosqich 6 | `POST /api/admin/questions` bittalab; ommaviy yo'q |
| **Profil / o'yin tarixi UI** | Bosqich 6 | `GET /api/me/history` endpoint bor, sahifa yo'q |
| **`packages/ui-web`** | §10 | shadcn umumiy komponentlar paketi — bo'sh placeholder |

## 3.6 Reja TASHQARISI — ops amallari va ma'lum cheklovlar

Bular **spetsifikatsiya emas** — ish davomida chiqqan (ops yoki review). Roadmap punkti emas, lekin kuzatib borish kerak.

### Ops amallari (kod emas — foydalanuvchi bajaradi)
- **Migratsiyalarni qo'llash**: `make tools` (goose) → `DATABASE_URL=... make migrate-up` — `00006_rating`, `00007_tournaments` (aks holda ELO/turnir ishlamaydi). *DB kerak.*
- **`make seed`** — yangi namuna savollar.
- **Branch** `b2-b3-rejimlar` → `main` merge + **push** (push qilinmagan).
- `.github/`, `CLAUDE.md`, `PLAN.md` ni commit qilish (hali untracked).

### Ma'lum cheklovlar / review-buglar (past daraja)
- **RN** ilovasi jonli build/test qilinmagan (Expo muhiti yo'q).
- **`startDuel` disconnect TOCTOU** — juftlash oynasida o'yinchi uzilsa fantom duel (kam UX nuqsoni).
- **`GetQuestionByID` N+1** — assess/turnir submit'da har javobga alohida DB so'rovi (perf, `WHERE id = ANY(...)` bilan jamlasa bo'ladi).
- **`hangman`/`word_search`** — bir-javob-har-savol engine'iga to'g'ri kelmaydi (interaktiv per-harf/grid protokol = engine kengaytmasi kerak).
- **Eski TODO'lar**: Telegram `auth_date` replay himoyasi (`telegram.go`); engine early-advance (hamma javob bersa deadline'siz keyingisiga o'tish).
- **`assess.go`** variantlarni aralashtirmaydi (turnir'da tuzatildi; eski kod, `ordering` UI yo'qligi sabab ta'siri nol).

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
