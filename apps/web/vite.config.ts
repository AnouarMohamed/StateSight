import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const devAPIProxyTarget = process.env.DEV_API_PROXY_TARGET ?? "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  server: {
    host: true,
    port: 5173,
    proxy: {
      "/api": {
        target: devAPIProxyTarget,
        changeOrigin: true
      },
      "/healthz": {
        target: devAPIProxyTarget,
        changeOrigin: true
      },
      "/readyz": {
        target: devAPIProxyTarget,
        changeOrigin: true
      }
    }
  }
});
