# QuizArena — Texnik Reja (Multi-subject, Multi-user Real-time Quiz Platforma)

> Mavjud `word-find-game` (Kahoot uslubidagi quiz) asosida qurilgan, lekin
> **ko'p sohali testlar**, **ko'p o'yin metodlari** va **minglab bir vaqtli
> foydalanuvchi**ga dosh beradigan kengaytirilgan platforma.

Status: **Reja (v2)** — kod yozishdan oldingi spetsifikatsiya.
Sana: 2026-06-15.

> **Backend:** Go (noldan yangi repo). Mavjud `word-find-game` kodi ko'chirilmaydi —
> faqat **ma'lumot** (`irregularVerb.json`) va ba'zi **mantiq g'oyalari** (distractor
> tanlash, shuffle) qayta yoziladi. Frontend: React (web) → Telegram Mini App → PWA.

---

## 1. Maqsadlar va printsiplar

1. **Masshtablashga tayyor, lekin hozir bitta instans.** Boshlang'ich versiya **bitta server instansi**da ishlaydi — Go va goroutine modeli bunda bir necha **mingdan** ortiq bir vaqtli WebSocket ulanishini bemalol ko'taradi. Horizontal scaling (bir nechta instans + **Redis pub/sub**) **ataylab kechiktiriladi**: hozir u ortiqcha infra xarajati. Real-time (Hub) va jonli-holat qatlami **interfeys orqali** yoziladi, shunda kerak bo'lganda kod tubdan o'zgarmasdan Redis qo'shiladi (qarang §2, §11).
2. **Ko'p platforma (bitta backend).** Web, Telegram va Mobile — **bir xil REST + WebSocket protokol**ga ulanadi. Backend (Go) platformadan mustaqil; UI qatlamlari almashinadi (qarang §1.5).
3. **Kengaytiriladigan soha modeli** — ingliz tili, matematika, dasturlash/IT, umumiy bilim; admin panel orqali yangi soha/savol qo'shish.
4. **Ko'p o'yin metodlari** — klassik quiz, survival, vaqtga poyga, jamoaviy, amaliyot (flashcard) rejimi.
5. **Server-authoritative** — to'g'ri javob hech qachon client'ga oldindan yuborilmaydi; ball serverda hisoblanadi (anti-cheat).
6. **Barqarorlik** — server restart o'yinni o'ldirmaydi; reconnect qo'llab-quvvatlanadi.

### Dizayn printsiplari
- Jonli (ephemeral) holat ≠ doimiy (persistent) ma'lumot. Ikkalasi alohida saqlanadi.
- Client "ahmoq", server "aqlli": client faqat ko'rsatadi va xom javob yuboradi.
- Soha/savol turlari **plugin** sifatida qo'shiladi, asosiy o'yin mexanikasi o'zgarmaydi.

---

## 1.5 Maqsadli platformalar va auth

**Maqsad: 3 ta platforma, 1 ta backend.** Backend (Go: REST + WebSocket) platformadan
mustaqil — har platforma faqat **UI qatlami**. Bu takroriy ishni minimallashtiradi.

| Platforma | Bosqich | Texnologiya | Izoh |
|---|---|---|---|
| **Web (brauzer)** | 1 (avval) | React SPA (hozirgi kod asosida) | Desktop + mobil brauzerni ham qamraydi. Asosiy ishlab chiqish maydoni. |
| **Telegram** | 2 | **Telegram Mini App (WebApp)** — *aynan o'sha React ilova* Telegram WebView ichida | Pure chat-bot emas, **Mini App** tavsiya: quiz UI (timer, variantlar, reyting) chatda yomon ko'rinadi. O'zbekistonda eng keng auditoriya. Auth — Telegram `initData` (parolsiz). |
| **Mobile** | 3 | **React Native (Expo)** | To'liq native ilova (store, push). UI **RN'da qayta yoziladi** — faqat mantiq ulashiladi (pastdagi izohga qarang). |
| ~~Desktop~~ | — | — | Kerak emas (web brauzer yetarli). |

> **Arxitektura natijasi (RN tanlovi sababli muhim):** "client" **uch qatlam**:
> (a) **`core` — faqat mantiq** (WS protokol client, API client, Zustand store, tiplar) — **uchchala platforma uchun bir xil**;
> (b) **web UI** (React DOM + Tailwind + **shadcn/ui**) — brauzer **va Telegram Mini App** *bir xil build*'ni ishlatadi;
> (c) **native UI** (React Native komponentlari + **NativeWind/Tamagui**) — alohida yoziladi, lekin (a) `core`'ni import qiladi.
> Ya'ni shadcn faqat (b) uchun; RN o'z UI kutubxonasini ishlatadi, biznes-mantiqni esa qayta yozmaydi.

### Auth — 3 ta usul, bitta `users` jadvali
| Usul | Bosqich | Mexanizm |
|---|---|---|
| **Mehmon** | 1 | Faqat ism → vaqtinchalik `users` yozuvi (`is_guest=true`). Past to'siq, tez kirish. |
| **Akkaunt** | 1 | Email + parol (**bcrypt**, `golang.org/x/crypto`), JWT sessiya. Tarix/statistika saqlanadi. |
| **Telegram** | 2 | Telegram `initData` ni serverda **HMAC-SHA256** bilan bot-token orqali tekshirish → `users` ga `telegram_id` bilan bog'lash. Parolsiz, avtomatik. |

Auth qatlami **provider interfeysi** (Go `interface`): `GuestProvider`, `PasswordProvider`,
`TelegramProvider` — yangi usul (Google, Apple) qo'shilsa yadro o'zgarmaydi.

### Auth siyosati matritsasi (gibrid: o'ynash ochiq, saqlash akkaunt talab qiladi)
| Amal | Mehmon | Akkaunt / Telegram |
|---|---|---|
| Xonaga qo'shilib **o'ynash** | ✅ | ✅ |
| Jonli reyting / o'yin natijasini ko'rish | ✅ | ✅ |
| **Xona ochish (host)** | ⚙️ sozlamaga bog'liq (default: ✅) | ✅ |
| O'yin **tarixi** saqlanishi | ❌ (o'yindan keyin yo'qoladi) | ✅ |
| **Global reyting**da ishtirok | ❌ | ✅ |
| Profil / statistika / yutuqlar | ❌ | ✅ |
| Admin panel | ❌ | ✅ (faqat `role=admin`) |

> Mehmonga o'yin oxirida "natijangizni saqlash uchun akkaunt oching" deb **yumshoq
> taklif** (progressive upgrade) ko'rsatiladi — `is_guest` yozuvni akkauntga bog'lash.

---

## 1.6 Texnologiya stacki (Go backend)

Tanlov printsipi: **idiomatik, yengil, standart kutubxonaga yaqin** — Go o'rganish
maqsadiga ham mos (sehr kam, nazorat ko'p). Og'ir ORM/framework'lardan qochamiz.

### Backend (Go)
| Vazifa | Tanlov | Sabab / muqobil |
|---|---|---|
| Til | **Go** (so'nggi stabil, 1.2x) | `log/slog`, generics, standart `net/http` yetuk. |
| HTTP router | **`chi`** | `net/http`-bilan to'liq mos, yengil, middleware oson. *Muqobil:* `gin`, `echo`. |
| WebSocket | **`gorilla/websocket`** + o'zimizning **Hub** (xonalar, broadcast) | Kanonik, yetuk. Hub = goroutine + channel pattern (Socket.IO "rooms" o'rnini bosadi). *Tez muqobil:* `melody` (gorilla ustida sodda API). *Modern:* `coder/websocket`. |
| Real-time scaling (keyin) | **Redis pub/sub** (`go-redis`) yoki **Centrifugo** | Bitta instansda kerak emas; ko'p instansda Hub'lar Redis orqali xabar almashadi. |
| Postgres drayveri | **`pgx`** (`jackc/pgx`) | Eng tezkor, idiomatik Postgres drayveri. |
| DB-kod | **`sqlc`** (SQL'dan tip-xavfsiz Go generatsiya) | ORM "sehri"siz, sof SQL + tiplar. Go'ni o'rganishga ideal. *Muqobil:* `GORM` (ORM), sof `pgx`. |
| Migratsiya | **`goose`** yoki **`golang-migrate`** | SQL migratsiya fayllari, versiyalash. |
| Validatsiya | **`go-playground/validator`** | Struct-teg validatsiya (Joi o'rnini bosadi). |
| JWT | **`golang-jwt/jwt`** | Sessiya tokenlari. |
| Parol hash | **`x/crypto/bcrypt`** | Standart, ishonchli. (Argon2 ham `x/crypto`da.) |
| UUID | **`google/uuid`** | PK'lar uchun. |
| Telegram bot | **`telego`** (Bosqich 2) | Go Bot API; `/start`→Mini App tugmasi, `/stats`. *Muqobil:* `go-telegram-bot-api`. |
| Config | **`env`** (`caarlos0/env`) + `.env` | Struct'ga env o'qish, sodda. |
| Log | **`log/slog`** (standart kutubxona) | Strukturalangan log, qo'shimcha lib shart emas. |
| Test | standart **`testing`** + **`testify`** | Assert'lar uchun testify. |
| Jonli holat | **In-memory** Go struct + `sync.RWMutex`, **`StateStore` interfeysi** orqasida | Hozir xotirada; keyin Redis implementatsiyasi drop-in. |
| Savol taymeri | **`time.Timer` + goroutine + `context`** | Tashqi lib shart emas; markazlashgan scheduler. Keyin scaling'da Redis/stream'ga. |

### Frontend (backend tilidan mustaqil)
| Vazifa | Tanlov | Qaysi platforma |
|---|---|---|
| Web framework | **React** + **Vite** (CRA emas) | web + Telegram |
| Til | **TypeScript** (hamma qatlamda) | hammasi |
| Styling/UI (web) | **Tailwind CSS + shadcn/ui** | web + Telegram |
| Native framework | **React Native (Expo)** | mobile |
| Styling/UI (native) | **NativeWind** (Tailwind-RN) yoki **Tamagui** | mobile |
| Client-state | **Zustand** (`core`da — web va RN bir xil) | hammasi |
| Server-state | **TanStack Query** (REST: reyting, tarix) | hammasi |
| WebSocket | Native `WebSocket` + ingichka wrapper (`core`da; reconnect, JSON protokol) — *socket.io-client emas* | hammasi |
| Telegram | **@telegram-apps/sdk-react** (Mini App) | Telegram |
| i18n | **react-i18next** (`core`da; **uzbek default**, ingliz qo'shimcha) | hammasi |
| PWA (ixtiyoriy) | **vite-plugin-pwa** (web o'rnatiladigan bo'lsa) | web |

> `core` paketi (Zustand store, WS/API client, tiplar) **uchchala platformada bir xil**;
> faqat UI qatlami platformaga xos (web=shadcn, native=NativeWind). Qarang §1.5 izoh.

> **Diqqat — Socket.IO yo'q.** Go tomonda Socket.IO protokolining yetuk serveri yo'q.
> Shuning uchun **sof WebSocket + o'zimizning JSON xabar protokoli** (§8) ishlatamiz.
> Socket.IO bepul bergan ikki narsani o'zimiz qilamiz: **xonalar** (server Hub) va
> **reconnect** (client wrapper + sessiya re-join). Buning evaziga to'liq nazorat.

### Infra
| Vazifa | Tanlov |
|---|---|
| Konteyner | **Docker** (multi-stage Go build → kichik image) + **Docker Compose** (postgres + app) |
| Reverse proxy (keyin) | nginx / Caddy — WS upgrade + sticky session (scaling bosqichida) |

---

## 1.7 Mahsulot ko'rinishi — foydalanuvchi nima qiladi

Platforma **uchta niyat**ga xizmat qiladi, hammasi **bitta o'yin engine**'i ustida.
Foydalanuvchi bosh ekranda shu uch ustundan birini tanlaydi.

### 🏆 Ustun 1 — Raqobat (Compete)
Boshqalar (yoki kompyuter) bilan musobaqa; tezlik + to'g'rilik = ball.

| Format | Tavsif | Bosqich |
|---|---|---|
| **Host-led xona** | Bir kishi xona ochadi, kod ulashadi, hamma sinxron o'ynaydi (Kahoot). | **1** (avval) |
| **1v1 / Matchmaking** | Tizim raqib topadi (yoki do'st bilan). Subject bo'yicha **rating (ELO)**. | keyin |
| **Kompyuter/AI raqib** | Raqib topilmasa yoki yolg'iz — **simulyatsion bot** (qiyinlik: oson/o'rta/qiyin → to'g'ri javob ehtimoli va tezligi). Matchmaking hech qachon "bo'sh" qolmaydi. | keyin |
| **Ochiq turnir / reyting** | Vaqt oralig'ida asinxron musobaqa; umumiy reytingda kim yuqori. | keyin |

> **AI bosqichlari:** (a) **bot** — LLM'siz, statistik (avval); (b) **haqiqiy AI** (ixtiyoriy, keyin) — savol/izoh generatsiya, adaptiv qiyinlik, "tutor".

### 📚 Ustun 2 — O'rganish (Learn)
Yakka, bosimsiz; maqsad — yangi materialni **uzoq muddat eslab qolish**.

- **Spaced Repetition (SRS, Anki uslubi)** — har savol foydalanuvchi javobiga qarab **optimal vaqtda** qayta beriladi (SM-2 algoritmi: `ease`, `interval`, `due_at`).
- Flashcard + **izoh** + to'g'ri javob ko'rsatiladi (raqobatsiz).
- "Bugun takrorlash kerak" navbati (due savollar).

### 📊 Ustun 3 — Baholash (Assess)
Bilim darajasini o'lchash va kuzatish.

- **Soha/kategoriya bo'yicha "mastery" daraja** (masalan *Algebra 72%*) — vaqt bo'yicha o'sadi/tushadi.
- Kuchli / zaif tomonlar tahlili, progress grafigi.
- Placement/self-test → boshlang'ich daraja aniqlash.

### Foydalanuvchi rollari va amallari
| Rol | Amallar |
|---|---|
| **Mehmon** | Kod bilan jonli o'yinga qo'shilish, o'ynash (natija saqlanmaydi), bot bilan o'ynash |
| **Akkaunt / Telegram** | + tarix, statistika, mastery, SRS progress, global reyting, **xona ochish (host)**, turnir, matchmaking |
| **Host** | Xona yaratish, sozlash (soha, mode, savol soni, vaqt), boshlash, o'yinchi chiqarish |
| **Admin** | Soha/kategoriya/savol boshqaruvi, bulk import, analitika (`/admin`, `role=admin`) |

> **Modes (§6)** — bu ustunlarning texnik amalga oshirilishi. Raqobat → `classic/survival/time_attack/team`; O'rganish → `practice` (SRS bilan); Baholash → `assessment`.

---

## 2. Infratuzilma tavsiyasi (maslahat)

**Hozirgi bosqich tavsiyasi: bitta instans + PostgreSQL. Redis/scaling kechiktiriladi.**

Masshtablash hozir kerak emas (§1.1) — shuning uchun Redis va ko'p-instans
**1-bosqichdan chiqariladi**. Ularning vazifasini bitta Go instansida soddaroq
vositalar (goroutine, channel, in-memory map) bajaradi, lekin **interfeys orqasida**
— minglab o'yinchida almashtirish oson.

| Komponent | Hozir (1 instans) | Keyin (minglab o'yinchi) | Sabab |
|---|---|---|---|
| Doimiy DB | **PostgreSQL** (`pgx`+`sqlc`) | PostgreSQL (o'zgarmaydi) | Akkaunt, savol banki, natijalar tarixi, statistika. Relyatsion model ierarxiyaga ideal. |
| Jonli holat | **In-memory `StateStore`** (Go map + `RWMutex`) | **Redis** (`go-redis`, bir xil interfeys) | Faol sessiya, o'yinchilar, deadline, ball. Hozir xotirada — eng tez va infrasiz. Interfeys bir xil bo'lgani uchun keyin Redis drop-in. |
| Real-time | **Go Hub** (goroutine + channel, adaptersiz) | Hub + **Redis pub/sub** yoki Centrifugo | Bitta instansda Hub yetarli. Ko'p instansda Hub'lar Redis orqali bog'lanadi. |
| Savol taymeri | **`time.Timer` + goroutine** (`deadlineTs` bilan) | Redis stream / taqsimlangan scheduler | Markazlashgan boshqaruvchi. Restart'ga chidamlilik keyingi bosqich. |
| Media saqlash (rasm/audio) | — (keyin) | **Obyekt-saqlash**: dev'da **MinIO**, prod'da S3/R2 + CDN | `image_choice`, `hotspot`, `audio` turlari uchun (§5). Faqat shu turlar kiritilganda. |

**Nima uchun in-memory SQLite emas (hozirgi kod):** doimiy ma'lumot (akkaunt,
savol banki, tarix) restartdan omon qolishi shart — buning uchun **PostgreSQL**.
Lekin **jonli o'yin holati** (vaqtinchalik) xotirada bo'lishi mutlaqo normal —
faqat **doimiy ≠ jonli** ajratilsa. Hozirgi kodning xatosi: ikkalasini ham bitta
SQLite'da aralashtirib, xona tugagach o'chirib yuborishi.

> **Qaror:** Boshlovni **Docker Compose** (postgres + app, **Redis'siz**) bilan
> qilamiz — bitta `docker compose up` dev muhitini beradi. Redis xizmatini keyin
> compose'ga qo'shish bir qator o'zgartirish (kod tayyor turadi).

---

## 3. Yuqori darajadagi arxitektura

**Hozirgi bosqich (bitta instans):**

```
   Web (React SPA) ────┐
   Telegram Mini App ──┼──  WebSocket / REST  ──►  ┌──────────────────────┐
   Mobile (PWA) ───────┘                           │   Go App (1 instans)  │
                                                    │  net/http + chi (REST)│
                                                    │  gorilla/ws + Hub     │
                                                    │  ┌─────────────────┐  │
                                                    │  │ In-memory       │  │
                                                    │  │ StateStore      │  │  ← jonli holat
                                                    │  │ (room:{id} map) │  │    (vaqtinchalik)
                                                    │  │ + time.Timer    │  │
                                                    │  └─────────────────┘  │
                                                    └──────────┬───────────┘
                                                               │ (faqat doimiy yozuvlar)
                                                               ▼
                                          ┌─────────────────────────────────────┐
                                          │       PostgreSQL (pgx + sqlc)        │
                                          │  users, subjects, categories,        │
                                          │  questions, game_sessions,           │
                                          │  game_results, answers_log           │
                                          └─────────────────────────────────────┘
```

> **Kelajakdagi scaling (kechiktirilgan, §11 Bosqich 5):** `StateStore` → Redis,
> Hub'lar orasiga Redis pub/sub, oldiga sticky-session LB, N instans. Kod interfeys
> orqali yozilgani uchun bu qatlam **almashtiriladi**, qayta yozilmaydi.

**Jonli o'yin oqimi qisqacha (bitta instans):**
1. O'yinchi socket orqali xonaga qo'shiladi → holat `StateStore`da `room:{id}` (hozir xotirada).
2. Owner "start" → server savollarni DB'dan tanlaydi, har biriga **deadline** qo'yadi (`time.Timer` + goroutine).
3. Server savolni Hub orqali broadcast qiladi **lekin to'g'ri javobsiz**. To'g'ri javob faqat serverda (`StateStore`da) qoladi.
4. O'yinchi xom javob yuboradi → server vaqt va to'g'rilikni tekshiradi → ball `StateStore`da.
5. Deadline (timer) kelganda goroutine keyingi savolga o'tadi.
6. O'yin oxirida natija PostgreSQL'ga ko'chiriladi (tarix), jonli holat tugash bilan tozalanadi.

> Scaling bosqichida 1–6 mantiq o'zgarmaydi — faqat `StateStore` Redis'ga,
> Hub Redis pub/sub'ga o'tadi (qadam 3 va 5 "instans qaysi bo'lishidan qat'i nazar" bo'ladi).

---

## 4. Ma'lumotlar modeli (PostgreSQL sxema)

```sql
-- Foydalanuvchilar (akkaunt + mehmon)
users (
  id            UUID PK,
  username      VARCHAR UNIQUE,        -- akkaunt uchun global; mehmon uchun NULL
  email         VARCHAR UNIQUE NULL,
  password_hash VARCHAR NULL,          -- mehmon/telegram uchun NULL
  telegram_id   BIGINT UNIQUE NULL,    -- Telegram orqali kirgan uchun (Bosqich 2)
  is_guest      BOOLEAN DEFAULT false,
  created_at    TIMESTAMPTZ
)

-- Soha (ingliz tili, matematika, IT, umumiy bilim)
subjects (
  id    UUID PK,
  slug  VARCHAR UNIQUE,   -- 'english', 'math', 'programming', 'general'
  name  VARCHAR,
  icon  VARCHAR
)

-- Kategoriya (soha ichida: 'irregular-verbs', 'algebra', 'sql', ...)
categories (
  id          UUID PK,
  subject_id  UUID FK -> subjects,
  slug        VARCHAR,
  name        VARCHAR,
  UNIQUE(subject_id, slug)
)

-- Savol banki (statik savollar)
questions (
  id           UUID PK,
  category_id  UUID FK -> categories,
  type         VARCHAR,   -- §5 katalog: mcq|multi_select|true_false|match|ordering|
                          --   categorize|fill_blank|type_answer|numeric|cloze|spelling|
                          --   anagram|word_search|hangman|image_choice|hotspot|audio|
                          --   code_output|code_debug|code_fill
  prompt       TEXT,
  options      JSONB,     -- variantlar / strukturani turga qarab (mcq, match, ordering...)
  correct      JSONB,     -- TO'G'RI JAVOB — client'ga hech qachon yuborilmaydi
  accept       JSONB NULL,-- kiritish turlari uchun qabul-ro'yxati / tolerance (type_answer, numeric)
  media_url    TEXT NULL, -- rasm/audio turlari uchun (obyekt-saqlashdagi havola)
  explanation  TEXT NULL, -- o'yin tugagach ko'rsatiladigan izoh
  difficulty   SMALLINT,  -- 1..5
  meta         JSONB      -- tarjima, manba, tag'lar
)

-- O'yin sessiyasi (xona) — tugagach tarix uchun saqlanadi
game_sessions (
  id            UUID PK,
  code          VARCHAR UNIQUE,        -- join code
  host_user_id  UUID FK -> users,
  subject_id    UUID FK,
  category_id   UUID FK NULL,          -- NULL = aralash
  mode          VARCHAR,               -- 'classic'|'survival'|'time_attack'|'team'|'practice'|'assessment'
  opponent      VARCHAR,               -- 'human'|'bot'|'mixed' (raqobat formati)
  question_count INT,
  time_per_q    INT,
  status        VARCHAR,               -- 'lobby'|'running'|'finished'
  started_at    TIMESTAMPTZ NULL,
  finished_at   TIMESTAMPTZ NULL
)

-- Natijalar (o'yinchi-sessiya)
game_results (
  id           UUID PK,
  session_id   UUID FK -> game_sessions,
  user_id      UUID FK -> users,
  score        NUMERIC,
  correct_cnt  INT,
  rank         INT,
  UNIQUE(session_id, user_id)
)

-- Javoblar logi (analitika / anti-cheat audit)
answers_log (
  id          UUID PK,
  session_id  UUID FK,
  user_id     UUID FK,
  question_id UUID FK,
  given       JSONB,
  is_correct  BOOLEAN,
  time_ms     INT,        -- server o'lchagan reaksiya vaqti
  created_at  TIMESTAMPTZ
)

-- 📚 O'rganish: Spaced Repetition (har user-savol uchun SM-2 holati)
srs_cards (
  user_id        UUID FK,
  question_id    UUID FK,
  ease           NUMERIC,     -- SM-2 ease factor (default 2.5)
  interval_days  INT,         -- keyingi takrorgacha kun
  repetitions    INT,         -- ketma-ket to'g'ri javoblar
  due_at         TIMESTAMPTZ, -- qachon qayta ko'rsatish
  last_reviewed  TIMESTAMPTZ,
  PRIMARY KEY (user_id, question_id)
)

-- 📊 Baholash: soha/kategoriya bo'yicha mastery (vaqt bo'yicha o'sadi)
user_mastery (
  user_id      UUID FK,
  category_id  UUID FK,
  mastery      NUMERIC,     -- 0..100 (%) yoki 0..1
  attempts     INT,
  updated_at   TIMESTAMPTZ,
  PRIMARY KEY (user_id, category_id)
)

-- 🏆 Matchmaking reytingi (1v1/duel uchun, subject bo'yicha ELO)
user_rating (
  user_id     UUID FK,
  subject_id  UUID FK,
  rating      INT DEFAULT 1000,   -- ELO
  games       INT,
  PRIMARY KEY (user_id, subject_id)
)

-- 🏆 Ochiq turnirlar
tournaments (
  id          UUID PK,
  title       VARCHAR,
  subject_id  UUID FK,
  mode        VARCHAR,
  starts_at   TIMESTAMPTZ,
  ends_at     TIMESTAMPTZ,
  status      VARCHAR        -- 'upcoming'|'active'|'finished'
)
tournament_entries (
  tournament_id UUID FK,
  user_id       UUID FK,
  score         NUMERIC,
  rank          INT,
  PRIMARY KEY (tournament_id, user_id)
)
```

> **Eski koddagi `users.name` global UNIQUE bug'i** shu yerda hal bo'ladi:
> o'yin ichidagi ism (display name) jonli sessiya holatida saqlanadi va
> faqat **xona ichida** unikal; global akkaunt ismi alohida.

### Jonli holat (in-memory `StateStore`, Go struct'lari)
Hozir xotirada — Go struct + `sync.RWMutex`. `StateStore` interfeysi orqali (keyin Redis).
```go
type Room struct {
    SessionID   string
    Code        string
    Status      string            // "lobby" | "running" | "finished"
    Mode        string
    CurrentIdx  int
    Players     map[string]*Player // userId -> player
    Questions   []*LiveQuestion    // correct SHU YERDA qoladi, client'ga ketmaydi
    mu          sync.RWMutex
}
type Player struct { Name string; Score float64; CorrectCnt int; Connected bool }
type LiveQuestion struct { QuestionID string; Correct any; AskedAt, Deadline int64;
                           Answered map[string]bool } // takror javobni rad etish

// Indekslar: code -> sessionID (tez qidiruv), userID -> sessionID (reconnect uchun)
```
Tashlandiq o'yinlar **idle-timeout** (masalan 3 soat) bilan tozalanadi (goroutine).
Scaling bosqichida shu interfeys Redis HASH/SET + TTL bilan almashtiriladi.

---

## 5. Soha va savol turlari (kengaytiriladigan model)

Har bir savol **turi** (`type`) bir generator/validator strategiyasiga ega.
Yangi soha qo'shish = (a) DB'ga savollar yuklash yoki (b) generator yozish.

### Qo'llab-quvvatlanadigan savol/sinov turlari (to'liq katalog)
Har tur — `QuestionType` strategiyasi: **`Render`** (client payload, `correct`siz) +
**`Validate`** (server baholash). Engine o'zgarmaydi; yangi tur = bitta strategiya.

**1️⃣ Tanlov asosidagi** (server-authoritative oson)
| type | Tavsif | Misol |
|---|---|---|
| `mcq` | 4 variantdan biri | "go → went" |
| `multi_select` | Bir nechta to'g'ri | "Qaysilari tub son?" |
| `true_false` | Ha/Yo'q | "`SELECT` o'zgartiradi?" |
| `match` | Juftlash | so'z ↔ tarjima |
| `ordering` | To'g'ri ketma-ketlik | voqealar/qadamlar tartibi |
| `categorize` | Toifaga ajratish | fe'l/ot/sifat |

**2️⃣ Kiritish asosidagi** (server normalize + fuzzy/qabul-ro'yxati)
| type | Tavsif | Misol |
|---|---|---|
| `fill_blank` | Bo'sh joy | "2 + 2 = __" |
| `type_answer` | Erkin javob | "go ning V2?" yozib |
| `numeric` | Raqamli (tolerance) | "√144 = ?" |
| `cloze` | Ko'p bo'sh joy | grammatik gap |
| `spelling` | To'g'ri yozish | tinglab/ko'rib |

**3️⃣ So'z o'yinlari** (ingliz tili sohasiga ideal)
| type | Tavsif |
|---|---|
| `anagram` | Harflardan so'z tuzish (T-E-N-W → WENT) |
| `word_search` | So'z topish to'ri (asl "word-find" g'oyasi) |
| `hangman` | Harf topib so'zni ochish |

**4️⃣ Media asosidagi** (obyekt-saqlash kerak — §2 infra)
| type | Tavsif |
|---|---|
| `image_choice` | Rasm asosida savol |
| `hotspot` | Rasmda joyni bosish (xarita) |
| `audio` | Tinglab javob (listening/diktant) |

**5️⃣ Soha-maxsus — kod (IT sohasi)**
| type | Tavsif |
|---|---|
| `code_output` | "Bu kod nima chiqaradi?" |
| `code_debug` | Xatoni top |
| `code_fill` | Kod to'ldirish |

> **Bosqichlash:** 1-2 guruh — B1-2 (arzon, katta qamrov); 3-guruh (so'z) + `ordering/categorize` — B3; 4-guruh (media) + 5-guruh (kod) — keyingi bosqichlar (media→saqlash, kod→maxsus validatsiya).

### Sohalar (1-bosqich, hammasi kiritiladi)
1. **Ingliz tili** — mavjud `irregularVerb.json` (110 fe'l) ko'chiriladi + lug'at, grammatika. `mcq`, `match`.
2. **Matematika** — **generatsiyalanadi** (arifmetika, algebra, foiz). Statik bank shart emas; `fill_blank`, `mcq`.
3. **Dasturlash / IT** — statik bank (terminologiya, kod natijasi, algoritm). `mcq`, `true_false`.
4. **Umumiy bilim** — statik bank (tarix, geografiya, fan), admin panel orqali to'ldiriladi. `mcq`.

### Generator interfeysi (Go)
```go
// har provider bir xil shaklda savol qaytaradi (Correct serverda qoladi)
type QuestionProvider interface {
    // statik bankdan yoki algoritmik generatsiya
    GetQuestions(opts QuestionOpts) ([]Question, error)
}
type QuestionOpts struct { CategoryID string; Count int; Difficulty int }

// Reyestr: subjectSlug -> QuestionProvider
var providers = map[string]QuestionProvider{
    "english": EnglishVerbProvider{}, "math": MathProvider{}, /* ... */
}
```
Eski `generateVerb.js` mantig'i (distractor tanlash) **Go'ga `EnglishVerbProvider`
sifatida qayta yoziladi**; `irregularVerb.json` ma'lumot bo'lib seed qilinadi.

---

## 6. O'yin metodlari (modes)

| mode | Ustun | Mexanika |
|---|---|---|
| `classic` | 🏆 | Hamma bir xil savolga sinxron javob beradi. Tezlik + to'g'rilik = ball. |
| `survival` | 🏆 | Xato javob = o'yindan chiqish. Oxirgi qolgan g'olib. |
| `time_attack` | 🏆 | Belgilangan vaqt ichida iloji boricha ko'p to'g'ri javob (har kim o'z tezligida). |
| `team` | 🏆 | O'yinchilar jamoalarga bo'linadi, jamoa balli yig'iladi. |
| `practice` | 📚 | Yakka, raqobatsiz; flashcard + izoh. **SRS** (SM-2) navbati bilan — due savollar takrorlanadi. |
| `assessment` | 📊 | Yakka test → natija + **mastery** yangilanadi (soha/kategoriya foizi). Raqobatsiz, baholash uchun. |

**Raqobat formatlari** (kim o'ynaydi) modes'dan **ortogonal**:
- `human` (host-led xona / matchmaking), `bot` (simulyatsion raqib), `mixed` (inson + bot to'ldiruvchi).
- Bot — `BotPlayer`: qiyinlik darajasiga qarab to'g'ri javob ehtimoli (`p`) va javob vaqti taqsimoti; scheduler bot nomidan javob "yuboradi".

Har metod Go `GameMode` interfeysi: `OnStart`, `OnAnswer`, `OnQuestionEnd`, `OnGameEnd`. Asosiy engine o'zgarmaydi; bot va matchmaking engine ustidagi qatlamlar.

---

## 7. Server-authoritative ball va anti-cheat

Eski koddagi 2 ta kritik teshik yopiladi:

1. **To'g'ri javob client'ga ketmaydi.** `question` payload'ida `correct` **yo'q**. Faqat `prompt` + `options`.
2. **Ballni server hisoblaydi.** Client faqat tanlangan variant indeksini yuboradi. Server:
   - javob deadline ichida kelganini tekshiradi (`AskedAt` `StateStore`da),
   - `Answered` map'da takror javobni rad etadi,
   - to'g'rilikni `StateStore`dagi `Correct` bilan solishtiradi,
   - ballni vaqt bo'yicha hisoblaydi: `score += base + speedBonus(time_ms)`.
3. To'g'ri javob va izoh **savol deadline tugagach** alohida `reveal` event bilan yuboriladi.

Audit uchun har javob `answers_log`'ga yoziladi.

---

## 8. Transport: WebSocket protokoli va REST API

### WebSocket JSON xabar protokoli (`/ws` endpoint)
Har xabar bitta konvert: `{ "type": "...", "data": { ... } }`. Socket.IO event nomlari
o'rniga `type` maydoni ishlatiladi. Hub xabarni tegishli xonaga broadcast qiladi.

**Client → Server** (`type` qiymatlari)
- `room:create { subjectId, categoryId?, mode, questionCount, timePerQ }`
- `room:join { code, displayName }`
- `room:leave`
- `game:start`
- `answer:submit { questionIndex, choice }`   ← faqat xom tanlov

**Server → Client** (`type` qiymatlari)
- `room:state { players, host, config }`
- `game:countdown { secondsLeft }`
- `question:show { index, total, prompt, options, deadlineTs }`  ← correct YO'Q
- `question:reveal { index, correct, explanation, leaderboard }`
- `player:scored { leaderboard }`   (ixtiyoriy, jonli reyting)
- `game:over { finalLeaderboard }`
- `error { code, message }`

> `deadlineTs` — server vaqti (epoch ms). Client shunga qarab progress bar
> chizadi → timer drift yo'qoladi (eski koddagi mustaqil intervallar muammosi).
>
> **Reconnect:** client uzilsa, qayta ulanib `room:resume { sessionId, token }`
> yuboradi; server `StateStore`dan holatni qaytaradi (Socket.IO'siz buni o'zimiz qilamiz).

### REST API (chi router)
- `POST /api/auth/register`, `POST /api/auth/login`, `POST /api/auth/guest`
- `POST /api/auth/telegram` — Telegram `initData` ni HMAC-SHA256 bilan tekshiradi (Bosqich 2)
- `GET  /api/subjects`, `GET /api/subjects/:id/categories`
- `GET  /api/me/history` — o'yin tarixi
- `GET  /api/leaderboard/global` — umumiy reyting
- 📚 `GET /api/me/srs/due`, `POST /api/srs/review { questionId, grade }` — SRS takrorlash
- 📊 `GET /api/me/mastery` — soha/kategoriya bo'yicha daraja
- 🏆 `POST /api/match/queue { subjectId }` — matchmaking navbati (raqib/bot)
- 🏆 `GET /api/tournaments`, `POST /api/tournaments/:id/join`
- **Admin:** `POST /api/admin/questions` (bulk import), `POST /api/admin/subjects` — RBAC (`role=admin`) bilan himoyalangan. Admin UI **asosiy web ilova ichida** (`/admin` marshruti), alohida ilova emas.

---

## 9. Eski koddan nima qayta ishlatiladi

> **Backend butunlay Go'da noldan yoziladi** — eski Node kodi ko'chirilmaydi.
> Faqat *ma'lumot* va *mantiq g'oyalari* olinadi.

| Eski (Node/React) | Yangi |
|---|---|
| `irregularVerb.json` (110 fe'l) | **Ma'lumot olinadi** — `english` sohasiga Postgres'ga seed |
| `generateVerb.js` mantiq | **G'oya olinadi** — `EnglishVerbProvider` Go'da qayta yoziladi (distractor logikasi) |
| `shuffle.js` | Go'da qayta yoziladi (`math/rand` shuffle) |
| React sahifalar (Regis/Room/RunTest) | **UI/UX asos** sifatida qayta ishlanadi (yangi Vite+TS loyihada) |
| Socket abstraksiyasi (`socket.js`) | Yangi native-WebSocket wrapper bilan **almashtiriladi** |
| Joi validatorlar | Go'da `go-playground/validator` bilan **qayta yoziladi** (eski `answerReqValidator` bug'i takrorlanmaydi) |

---

## 10. Repo strukturasi (taklif)

```
quizarena/
├── docker-compose.yml          # postgres + app (Redis'siz)
├── server/                     # Go modul (go.mod)
│   ├── cmd/
│   │   └── server/main.go      # bootstrap: config → db → hub → http/ws
│   ├── internal/               # tashqaridan import qilinmaydigan kod
│   │   ├── config/             # env o'qish (caarlos0/env)
│   │   ├── httpapi/            # chi routerlar (auth, subjects, history, admin)
│   │   │   └── middleware/     # JWT, CORS, ratelimit, logging
│   │   ├── ws/                 # gorilla/websocket + Hub (xonalar, broadcast)
│   │   │   └── protocol.go     # JSON xabar tiplari (§8)
│   │   ├── game/
│   │   │   ├── engine.go       # savol oqimi, deadline, ball
│   │   │   ├── scheduler.go    # time.Timer + goroutine taymer
│   │   │   ├── bot.go          # 🏆 BotPlayer (simulyatsion raqib)
│   │   │   ├── matchmaking.go  # 🏆 1v1 navbat + ELO (keyin)
│   │   │   ├── modes/          # classic, survival, time_attack, team, practice, assessment
│   │   │   └── providers/      # english, math, programming, general
│   │   ├── learn/              # 📚 SRS (SM-2) — srs_cards, due navbati
│   │   ├── assess/             # 📊 mastery hisoblash (user_mastery)
│   │   ├── auth/               # GuestProvider, PasswordProvider, TelegramProvider
│   │   ├── state/              # StateStore interfeys + memory impl (keyin redis)
│   │   ├── store/              # Postgres: sqlc-generatsiya qilingan querylar
│   │   └── seed/               # irregularVerb.json → DB
│   ├── migrations/             # goose/golang-migrate SQL fayllari
│   ├── queries/                # sqlc uchun .sql manba fayllari
│   ├── sqlc.yaml
│   └── *_test.go               # testing + testify (paket yonida)
└── client/                     # TypeScript monorepo (pnpm workspaces)
    ├── packages/core/          # FAQAT MANTIQ: ws/api client, zustand store, tiplar (UI yo'q!)
    ├── packages/ui-web/        # shadcn/ui komponentlar (web + telegram uchun umumiy)
    ├── apps/web/               # Vite React SPA + Tailwind   ┐ core + ui-web ni import qiladi
    ├── apps/telegram/          # Telegram Mini App (@telegram-apps/sdk-react) ┘ (PWA manifest ham shu yerda)
    └── apps/native/            # React Native (Expo) + NativeWind — faqat core'ni import qiladi (UI o'zi)
```

---

## 11. Bosqichli yo'l xaritasi (roadmap)

**Bosqich 0 — Poydevor**
- Go modul (`go.mod`), Docker Compose (**Postgres + app, Redis'siz**), env config.
- Migratsiya (`goose`) + `sqlc` quvuri; `chi` HTTP "health" + `gorilla/ws` "echo" Hub.
- `StateStore` interfeysi (in-memory impl) — scaling uchun keyin almashtiriladigan qatlam boshidanoq.

**Bosqich 1 — 🏆 Raqobat MVP (host-led classic) + Web**
- Auth: **mehmon + akkaunt** (email/parol bcrypt, JWT). Provider interfeysi.
- **Host-led xona** yaratish/qo'shilish (Hub + xona), **Web (Vite+React+TS+Tailwind+shadcn)** UI.
- Server-authoritative classic quiz, ingliz tili sohasi (`irregularVerb.json` seed).
- Reconnect (`room:resume`), `deadlineTs` sinxronizatsiyasi, natija PostgreSQL'ga.
- i18n (uzbek default + ingliz).

**Bosqich 2 — Telegram + ko'p soha**
- **Telegram Mini App** (web build) + `TelegramProvider` auth (`initData` HMAC).
- **Minimal bot** (Go `telego`): `/start` → "O'ynash" tugmasi (Mini App), `/stats`.
- `MathProvider` (generativ), `ProgrammingProvider`, `GeneralProvider` (statik).
- Savol turlari **1-2 guruh** (§5): `mcq, multi_select, true_false, match, fill_blank, type_answer, numeric, cloze, ordering, categorize`. Soha/kategoriya tanlash UI.

**Bosqich 3 — 🏆 Ko'p raqobat metodi + so'z o'yinlari + Mobile (React Native)**
- `survival`, `time_attack`, `team` rejimlari (`GameMode` interfeysi).
- Savol turlari **3-guruh** (§5): `anagram, word_search, hangman` (asl "word-find" g'oyasi).
- **React Native (Expo)** ilova — `core` paketni qayta ishlatadi, UI RN'da (NativeWind). Store/push.

**Bosqich 4 — 📚 O'rganish (SRS) + 📊 Baholash (mastery)**
- `practice` mode + **Spaced Repetition** (SM-2: `srs_cards`, due navbati, izoh/flashcard).
- `assessment` mode + **mastery** kuzatuvi (`user_mastery`, soha/kategoriya foizi, progress).
- Bu ustunlar **yakka** — real-time Hub'siz, REST + lokal holat yetarli.

**Bosqich 5 — 🏆 Raqobat kengaytma (bot, matchmaking, turnir)**
- **Kompyuter/AI raqib** (`BotPlayer`, qiyinlik darajalari) — raqib bo'lmaganda.
- **1v1 / Matchmaking** (navbat + subject bo'yicha **ELO** `user_rating`).
- **Ochiq turnir / reyting** (`tournaments`, asinxron).

**Bosqich 6 — Kengaytma + Production**
- Savol turlari **4-5 guruh** (§5): media (`image_choice, hotspot, audio` → MinIO/S3) + kod (`code_output, code_debug, code_fill`).
- Admin panel (`/admin`, savol/soha boshqaruvi, bulk import, analitika).
- Global leaderboard, profil, o'yin tarixi, statistika.
- Rate limiting, CORS qattiqlash, monitoring, CI/CD.
- (Ixtiyoriy) **haqiqiy AI** — savol/izoh generatsiya, adaptiv qiyinlik, tutor.
- **Scaling (faqat yuk talab qilganda):** `StateStore` → Redis (`go-redis`), Hub orasiga Redis pub/sub (yoki Centrifugo), N instans + sticky LB. Avval yuk testi (k6/`vegeta`).

---

## 12. Ochiq savollar (keyingi qadamda hal qilinadi)
- ✅ ~~Platformalar?~~ → Web (1) → Telegram Mini App (2) → Mobile/PWA (3). Desktop yo'q. (§1.5)
- ✅ ~~Akkaunt majburiymi?~~ → Mehmon + akkaunt + Telegram, uchchalasi. (§1.5)
- ✅ ~~Scaling/multi-instans hozir kerakmi?~~ → Yo'q, kechiktirildi (§1.1, §2, Bosqich 5).
- ✅ ~~Telegram: Mini App vs chat-bot?~~ → Mini App + **minimal bot** (`/start`, `/stats`). (§1.6, Bosqich 2)
- ✅ ~~Mobile: PWA vs React Native?~~ → **React Native (Expo)**; web/Telegram bilan faqat `core` mantiq ulashiladi. (§1.5)
- ✅ ~~UI tili?~~ → **uzbek default + ingliz i18n** (react-i18next). UI matni vs savol kontenti alohida tarjima qatlami.
- ✅ ~~Admin panel alohidami?~~ → **asosiy web ichida** (`/admin`, `role=admin`).
```
