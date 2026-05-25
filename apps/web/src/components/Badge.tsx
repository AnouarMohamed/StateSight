import type { BadgeTone } from "../lib/badgeTones";

const toneClass: Record<BadgeTone, string> = {
  neutral: "bg-[#1b2838] text-ops-muted",
  good: "bg-[rgba(66,184,131,0.2)] text-ops-good",
  warn: "bg-[rgba(242,177,52,0.2)] text-ops-warn",
  bad: "bg-[rgba(247,108,108,0.2)] text-ops-bad"
};

export function Badge({ label, tone = "neutral" }: { label: string; tone?: BadgeTone }) {
  return <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-medium ${toneClass[tone]}`}>{label}</span>;
}
