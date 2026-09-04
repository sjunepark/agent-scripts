import { formatCsv } from "./csv.js";

export function renderReport(records, format) {
  if (format === "csv") {
    return formatCsv(records);
  }

  return JSON.stringify(records);
}
