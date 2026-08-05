export function formatSom(value: number) {
  return `${new Intl.NumberFormat('ru-KG').format(value)} сом`;
}
