import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        "ops-bg": "#0c0a09",
        "ops-panel": "#0c0a09",
        "ops-panel-alt": "#131211",
        "ops-shell": "#0c0a09",
        "ops-shell-hover": "#1a1918",
        "ops-border": "#292524",
        "ops-border-muted": "#1a1918",
        "ops-text": "#e7e5e4",
        "ops-muted": "#a8a29e",
        "ops-dim": "#78716c",
        "ops-accent": "#f97316",
        "ops-accent-soft": "#28150d",
        "ops-action": "#ea580c",
        "ops-action-hover": "#f97316",
        "ops-action-border": "#f97316",
        "ops-good": "#4ade80",
        "ops-good-soft": "#101d15",
        "ops-good-border": "#166534",
        "ops-warn": "#fbbf24",
        "ops-warn-soft": "#221a0c",
        "ops-warn-border": "#854d0e",
        "ops-bad": "#fb7185",
        "ops-bad-soft": "#241013",
        "ops-bad-border": "#9f1239"
      }
    }
  },
  plugins: []
};

export default config;
