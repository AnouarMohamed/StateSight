import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        "ops-bg": "#0d1117",
        "ops-panel": "#161b22",
        "ops-panel-alt": "#0f141a",
        "ops-shell": "#010409",
        "ops-shell-hover": "#171d25",
        "ops-border": "#30363d",
        "ops-border-muted": "#21262d",
        "ops-text": "#e6edf3",
        "ops-muted": "#8b949e",
        "ops-accent": "#4493f8",
        "ops-accent-soft": "#13243d",
        "ops-action": "#238636",
        "ops-action-hover": "#2ea043",
        "ops-action-border": "#328844",
        "ops-good": "#3fb950",
        "ops-good-soft": "#12251a",
        "ops-good-border": "#276738",
        "ops-warn": "#d29922",
        "ops-warn-soft": "#2b210d",
        "ops-warn-border": "#6e5219",
        "ops-bad": "#f85149",
        "ops-bad-soft": "#2d1516",
        "ops-bad-border": "#6e2e32"
      },
      boxShadow: {
        panel: "0 1px 0 rgba(1, 4, 9, 0.24)"
      }
    }
  },
  plugins: []
};

export default config;
