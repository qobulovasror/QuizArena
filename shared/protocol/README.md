# QuizArena — WebSocket JSON protokoli (v1)

Bu hujjat **manba haqiqat** (source of truth). Ikki implementatsiya unga amal qiladi:
- **Server (Go):** `server/internal/ws/protocol.go`
- **Client (TS):** `shared/protocol/messages.ts`

> Maqsad — kelajakda shu shartnomadan **ikkala tomon tiplarini generatsiya** qilish
> (drift yo'qolsin). Hozircha qo'lda sinxron saqlanadi.

---

## 1. Ulanish

- Endpoint: `GET /ws` (WebSocket upgrade).
- **Auth (ixtiyoriy):** akkaunt/Telegram foydalanuvchisi `?token=<JWT>` query yoki
  `Authorization: Bearer` bilan ulanadi. **Mehmon** tokensiz ulanadi va `room:join`
  da `displayName` beradi — kimligi `room:joined` da qaytariladi.
- Ulangach klient odatda `room:create`, `room:join` yoki `room:resume` yuboradi.

## 2. Konvert (envelope)

Har bir xabar — bitta JSON obyekt:

```json
{ "type": "room:join", "data": { "code": "ABC123", "displayName": "Ali" }, "id": "c-1" }
```

| Maydon | Tur | Izoh |
|---|---|---|
| `type` | string | Xabar turi (quyidagi katalog). |
| `data` | object | Turga xos yuk. Bo'sh bo'lishi mumkin. |
| `id` | string? | **Ixtiyoriy.** Client→server korrelyatsiya (javob/ack shu `id` bilan qaytadi). |

Protokol versiyasi: **v1**. O'zgarsa, yangi turlar qo'shiladi yoki `type` prefiksi versiyalanadi.

---

## 3. Client → Server xabarlar

| `type` | `data` | Izoh |
|---|---|---|
| `room:create` | `{ subjectId, categoryId?, mode, opponent?, questionCount, timePerQ, displayName }` | Xona yaratish; yuboruvchi **host** (va o'yinchi) bo'ladi. `mode`=classic/... ; `opponent`=human/bot/mixed. `timePerQ` — soniya. |
| `room:join` | `{ code, displayName }` | Kod bilan qo'shilish. |
| `room:resume` | `{ sessionId, resumeToken }` | Uzilishdan keyin qayta ulanish (token `room:joined` da berilgan). |
| `room:leave` | `{}` | Xonadan chiqish. |
| `game:start` | `{}` | O'yinni boshlash (**faqat host**). |
| `answer:submit` | `{ questionIndex, choice }` | Xom javob. `choice` shakli savol turiga bog'liq (§6). To'g'rilik **client'ga bog'liq emas** — server hal qiladi. |

## 4. Server → Client xabarlar

| `type` | `data` | Izoh |
|---|---|---|
| `room:state` | `{ sessionId, code, host, status, config, players[] }` | Xonaning **to'liq snapshot**i. Har o'zgarishda broadcast. `status`=lobby/running/finished. |
| `room:joined` | `{ sessionId, userId, resumeToken }` | Qo'shilgan/yaratgan klientga **kimligi** + reconnect tokeni. |
| `game:countdown` | `{ secondsLeft }` | Boshlanish sanog'i (5…1). |
| `question:show` | `{ index, total, type, prompt, options?, deadlineTs }` | Savol. **`correct` YO'Q.** `deadlineTs` — server epoch **ms**. |
| `answer:ack` | `{ index, accepted }` | *Ixtiyoriy.* Javob qabul qilingani (to'g'rilikni **oshkor qilmaydi**). |
| `question:reveal` | `{ index, correct, explanation?, leaderboard[], teams? }` | Deadline tugagach: to'g'ri javob + izoh + reyting. `teams` — faqat team rejimi. |
| `player:scored` | `{ leaderboard[] }` | *Ixtiyoriy* jonli reyting yangilanishi. |
| `game:over` | `{ finalLeaderboard[], teams? }` | O'yin tugadi. `teams` — faqat team rejimi. |
| `error` | `{ code, message }` | Xato (§7 kodlari). |

---

## 5. Umumiy tuzilmalar

```
Player           { userId, name, score, connected, isBot?, eliminated?, team? }
RoomConfig       { subjectId, mode, questionCount, timePerQ }
Option           { id, text }              // id — opaque (server shuffle qiladi)
LeaderboardEntry { userId, name, score, correctCnt, rank, eliminated?, team? }   // eliminated: survival, team: team rejimi
TeamStanding     { team, score, correctCnt, rank }    // team rejimi yig'indisi
```

## 6. Savol turiga qarab `choice` / `correct` shakli

`options` opaque `id` lardan iborat (server aralashtiradi) — shu sabab to'g'ri javob
oshkor bo'lmaydi. Javob va to'g'ri javob shakli (1-bosqich `mcq` uchun):

| Tur | `question:show.options` | `answer:submit.choice` | `question:reveal.correct` |
|---|---|---|---|
| `mcq` | `[{id,text}]` | `{ optionId }` | `{ optionId }` |
| `true_false` | — | `{ value: true|false }` | `{ value }` |
| `multi_select` | `[{id,text}]` | `{ optionIds: [] }` | `{ optionIds: [] }` |
| `type_answer`/`fill_blank` | — | `{ text }` | `{ accepted: [] }` |
| `numeric` | — | `{ value }` | `{ value, tolerance? }` |
| `match` | `[{id,text}]` (chap) | `{ pairs: { leftId: rightId } }` | `{ pairs: { leftId: rightId } }` |
| `ordering` | `[{id,text}]` (aralash) | `{ order: [id,…] }` | `{ order: [id,…] }` |
| `categorize` | `[{id,text}]` (elementlar) | `{ assign: { itemId: catId } }` | `{ assign: { itemId: catId } }` |
| `cloze` | — | `{ blanks: [text,…] }` | `{ blanks: [ {accepted:[…]}, … ] }` |

> `match`/`categorize` — to'la moslik (har juft/element to'g'ri); `ordering` — aniq tartib;
> `cloze` — har bo'shliq mos `accepted` ro'yxatida (normalize: trim + lowercase).
> Boshqa turlar (§5 katalog) keyingi bosqichlarda shu jadvalga qo'shiladi.

**Render uchun qo'shimcha (`question:show`):**
- `match` — `options` = chap ustun (juftlanadigan), **`targets`** = o'ng ustun (variantlar). Ikkalasi opaque `id`, server aralashtiradi.
- `categorize` — `options` = elementlar, **`targets`** = toifalar.
- `ordering` — `options` = aralashtirilgan elementlar; client qayta tartiblaydi, javobda `order` id ketma-ketligi.
- `cloze` — `options` yo'q; `prompt` ichida har bo'shliq **`___`** (3 ta pastki chiziq) bilan belgilanadi; client shu markerlar sonicha kiritish maydoni chizadi.
> Server `choice` ni **xom** (`json.RawMessage`/`unknown`) qabul qiladi va tur strategiyasi
> orqali tekshiradi.

---

## 7. Xato kodlari (`error.code`)

| Kod | Ma'no |
|---|---|
| `BAD_REQUEST` | Yuk yaroqsiz/yetishmayapti. |
| `INVALID_MESSAGE` | Konvert/JSON buzuq yoki noma'lum `type`. |
| `ROOM_NOT_FOUND` | Kod/sessiya topilmadi. |
| `ROOM_FULL` | Xona to'la. |
| `NOT_HOST` | Faqat host bajara oladigan amal. |
| `GAME_ALREADY_STARTED` | O'yin allaqachon boshlangan. |
| `ALREADY_ANSWERED` | Shu savolga javob berilgan. |
| `DEADLINE_PASSED` | Javob deadline'dan keyin keldi. |
| `UNAUTHENTICATED` | Token talab qilinadi/yaroqsiz. |
| `INTERNAL` | Server xatosi. |

---

## 8. Hayot sikli (classic)

```
ulanish → room:create/join ──► room:joined (+ resumeToken)
                               room:state (lobby)            ◄─ har o'zgarishda
host: game:start ──► game:countdown (5..1)
   ┌─ har savol uchun ────────────────────────────────────────────┐
   │ server ► question:show (correct YO'Q, deadlineTs bilan)        │
   │ client ► answer:submit { index, choice }                       │
   │ server ► answer:ack (ixtiyoriy) / player:scored (ixtiyoriy)    │
   │ deadline ► server ► question:reveal (correct + leaderboard)    │
   └────────────────────────────────────────────────────────────────┘  × N
server ► game:over { finalLeaderboard }
```

### 8.1 Rejimga xos oqim farqlari

- **`classic` / `team`** — yuqoridagi sinxron oqim. `team`'da o'yinchilar StartGame'da
  2 jamoaga (A/B) balanslab taqsimlanadi; ball individ, lekin `question:reveal` va
  `game:over` qo'shimcha `teams[]` (jamoa yig'indisi) yuboradi. `Player.team` to'ldiriladi.
- **`survival`** — sinxron, lekin xato javob o'yinchini chiqaradi (`eliminated`); bittadan
  kam tirik qolsa o'yin erta tugaydi.
- **`time_attack`** — **per-player oqim**: har o'yinchi o'z sur'atida. `game:countdown`dan
  so'ng har kim **o'z** `question:show`'ini oladi (`deadlineTs` — butun o'yin uchun yagona).
  Har `answer:submit` → `answer:ack` + **darhol keyingi** `question:show` (reveal **yo'q**).
  Savollar tugaganda yoki yagona deadline kelganda `game:over`. Ball = to'g'ri javob soni.

## 9. Reconnect

1. Klient uzilsa, qayta ulanib `room:resume { sessionId, resumeToken }` yuboradi.
2. Server `StateStore` dan holatni tiklaydi:
   - `room:joined` (yangi/eski `userId`) + `room:state`,
   - o'yin ketayotgan bo'lsa — joriy `question:show` (qolgan `deadlineTs` bilan).
3. Token yaroqsiz/sessiya yo'q bo'lsa — `error { ROOM_NOT_FOUND | UNAUTHENTICATED }`.

> `deadlineTs` mutlaq server vaqti (epoch ms) bo'lgani uchun reconnect'dan keyin ham
> progress bar **drift'siz** to'g'ri ko'rsatiladi.
