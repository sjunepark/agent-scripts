export function formatCsv(records) {
  const header = "name,total";
  const rows = records.map(({ name, total }) => `${name},${total}`);

  return [header, ...rows].join("\n");
}
