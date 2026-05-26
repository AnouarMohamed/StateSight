import type { EvidenceRecord } from "./api";

export type EvidenceMetadata = Record<string, unknown>;

export function parseEvidenceMetadata(record: EvidenceRecord): EvidenceMetadata {
  try {
    const parsed = JSON.parse(record.metadata) as unknown;
    return parsed !== null && typeof parsed === "object" && !Array.isArray(parsed) ? (parsed as EvidenceMetadata) : {};
  } catch {
    return {};
  }
}

export function metadataText(metadata: EvidenceMetadata, key: string) {
  const value = metadata[key];
  return typeof value === "string" && value.length > 0 ? value : undefined;
}

export function metadataBoolean(metadata: EvidenceMetadata, key: string) {
  const value = metadata[key];
  return typeof value === "boolean" ? value : undefined;
}

export function evidenceSourceLabel(source: string) {
  switch (source) {
    case "git":
      return "Desired state";
    case "kubectl":
      return "Live observation";
    case "synthetic":
      return "Synthetic observation";
    case "managedFields":
      return "Field ownership";
    default:
      return source;
  }
}

export function actorLabel(actor: string) {
  return actor === "not-attributed" ? "No actor observed" : actor;
}
