export function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.valueOf())) {
    return value;
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short"
  }).format(date);
}

export function formatPercent(value: number) {
  return `${Math.round(value * 100)}%`;
}

export function displayValue(value: string) {
  return value.trim().length === 0 ? "Absent" : value;
}
