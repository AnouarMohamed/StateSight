export type BadgeTone = "neutral" | "good" | "warn" | "bad";

export function severityTone(severity: string): BadgeTone {
  switch (severity.toLowerCase()) {
    case "high":
      return "bad";
    case "medium":
      return "warn";
    case "low":
      return "good";
    default:
      return "neutral";
  }
}

export function recommendationTone(action: string): BadgeTone {
  switch (action.toLowerCase()) {
    case "reconcile":
      return "bad";
    case "investigate":
      return "warn";
    case "monitor":
      return "good";
    default:
      return "neutral";
  }
}
