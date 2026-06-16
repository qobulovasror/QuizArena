import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Dev'da frontend (5173) → Go backend (8099) ga proxy: REST + WebSocket.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": { target: "http://localhost:8099", changeOrigin: true },
      "/ws": { target: "ws://localhost:8099", ws: true },
    },
  },
});
