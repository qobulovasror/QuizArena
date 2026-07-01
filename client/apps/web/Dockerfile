# --- build bosqichi: Node bilan statik fayllarni yig'ish ---
# MUHIM: build konteksti = client/ (monorepo ildizi), chunki apps/web
# @core alias orqali ../../packages/core/src'ni iste'mol qiladi (vite.config.ts).
FROM node:20-alpine AS build
WORKDIR /app

# 1) Umumiy paket manbasi (apps/web build vaqtida @core shu yerdan hal bo'ladi)
COPY packages/core ./packages/core

# 2) Avval faqat lock fayllar — layer cache samarali bo'lishi uchun
COPY apps/web/package.json apps/web/package-lock.json ./apps/web/
RUN cd apps/web && npm ci

# 3) Qolgan manba va build
COPY apps/web ./apps/web
RUN cd apps/web && npm run build

# --- runtime bosqichi: nginx bilan tarqatish ---
FROM nginx:1.27-alpine
COPY apps/web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/apps/web/dist /usr/share/nginx/html
EXPOSE 80
