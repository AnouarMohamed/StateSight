import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        "ops-bg": "#0b1017",
        "ops-panel": "#121a25",
        "ops-border": "#1f2a3a",
        "ops-text": "#dbe6f4",
        "ops-muted": "#8ba0b7",
        "ops-accent": "#4aa3ff",
        "ops-good": "#42b883",
        "ops-warn": "#f2b134",
        "ops-bad": "#f76c6c"
      },
      boxShadow: {
        panel: "0 12px 28px rgba(0, 0, 0, 0.35)"
      }
    }
  },
  plugins: []
};

export default config;
