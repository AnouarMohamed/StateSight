import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 5173,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true
      },
      "/healthz": {
        target: "http://localhost:8080",
        changeOrigin: true
      },
      "/readyz": {
        target: "http://localhost:8080",
        changeOrigin: true
      }
    }
  }
});
