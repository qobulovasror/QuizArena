# apps/native — React Native (Expo)

> ⚠️ **Skelet / poydevor.** Bu ilova `@quizarena/core` (umumiy mantiq) **web bilan bir xil**
> qayta ishlatilishini ko'rsatadi. **Jonli build/test Expo muhitini talab qiladi** va bu
> repozitoriy CI/sandbox'ida tekshirilmagan — versiyalar va sozlamalarni mahalliy ishga
> tushirishda moslash kerak bo'lishi mumkin.

## Arxitektura (PLAN.md §1.5)

- **Mantiq ulashiladi**: `@core/*` (Zustand store, WS/REST client, protokol tiplari, i18n)
  — `packages/core/src` dan, web bilan **aynan bir xil** kod.
- **UI alohida**: bu yerda React Native komponentlari (`View`/`Text`/`TextInput`), web esa
  shadcn/Tailwind. Faqat ko'rinish platformaga xos.
- **Platforma sozlash**: `App.tsx` da `configureCore({ apiBase, wsBase })` — RN'da relative
  URL yo'q, shuning uchun backend manzili absolute beriladi (`@core/config`).

## Tuzilma

```
apps/native/
├── App.tsx                 # store-driven routing (web App.tsx mantig'i) + configureCore
├── index.ts                # registerRootComponent
├── src/screens/
│   ├── AuthScreen.tsx       # mehmon/akkaunt kirish — @core/api + @core/store
│   └── HomeScreen.tsx       # asosiy ekran (poydevor)
├── babel.config.js          # module-resolver: @core → packages/core/src
├── metro.config.js          # monorepo: workspace'ni kuzatadi, deps'ni hal qiladi
├── app.json, tsconfig.json, package.json
```

## Ishga tushirish (mahalliy)

```bash
cd client/apps/native
npm install            # yoki pnpm install (workspace ildizidan)
# App.tsx dagi apiBase/wsBase ni o'z backend manzilingizga sozlang
#   Android emulyator: http://10.0.2.2:8080 ; iOS simulyator: http://localhost:8080
npx expo start         # so'ng a (Android) / i (iOS)
```

Backend ishlab turishi kerak (`PORT=8080 make run` yoki Docker).

## Qolgan ish (keyingi)

- Lobby / Play / Result / Practice / Assess / Rating / Tournaments ekranlari — web sahifalari
  asosida RN'da. **Store oqimi va API o'zgarmaydi** (faqat ko'rinish qayta yoziladi).
- NativeWind (Tailwind-RN) yoki Tamagui bilan stillarni soddalashtirish.
- Push xabarnomalar, store deep-linking.
