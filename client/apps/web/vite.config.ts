import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { fileURLToPath, URL } from "node:url";

// Dev'da frontend (5173) → Go backend (8099) ga proxy: REST + WebSocket.
export default defineConfig({
  plugins: [react()],
  resolve: {
    // @core → umumiy mantiq paketi (packages/core/src) — manbadan iste'mol.
    // Umumiy paket apps/web'dan tashqarida bo'lgani uchun, uning bare-importlari
    // (zustand, i18next, ...) ham apps/web/node_modules'dan hal qilinadi (dedup).
    alias: {
      "@core": fileURLToPath(new URL("../../packages/core/src", import.meta.url)),
      zustand: fileURLToPath(new URL("./node_modules/zustand", import.meta.url)),
      i18next: fileURLToPath(new URL("./node_modules/i18next", import.meta.url)),
      "react-i18next": fileURLToPath(new URL("./node_modules/react-i18next", import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/api": { target: "http://localhost:8099", changeOrigin: true },
      "/ws": { target: "ws://localhost:8099", ws: true },
    },
  },
});
