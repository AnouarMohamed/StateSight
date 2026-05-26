import type { BadgeTone } from "../lib/badgeTones";

const toneClass: Record<BadgeTone, string> = {
  neutral: "border-ops-border bg-ops-panel-alt text-ops-muted",
  good: "border-ops-good-border bg-ops-good-soft text-ops-good",
  warn: "border-ops-warn-border bg-ops-warn-soft text-ops-warn",
  bad: "border-ops-bad-border bg-ops-bad-soft text-ops-bad"
};

export function Badge({ label, tone = "neutral" }: { label: string; tone?: BadgeTone }) {
  return <span className={`inline-flex max-w-full rounded-md border px-2 py-0.5 text-xs font-medium ${toneClass[tone]}`}>{label}</span>;
}
