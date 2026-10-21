import { fmtHour } from './format.ts';
import { WEEKDAYS } from './labels.ts';

/** "Freeway", "Freeway, Ramp", "Freeway, Ramp +2 more" */
export function listLabel(labels: readonly string[], max = 2): string {
  if (labels.length <= max) return labels.join(', ');
  return `${labels.slice(0, max).join(', ')} +${labels.length - max} more`;
}

const HOURS = Array.from({ length: 24 }, (_, h) => h);
const DAYS = Array.from({ length: 7 }, (_, d) => d);

/**
 * Summarises a set of hour-of-week cells: whole weekdays ("Fri, Sat"),
 * whole hours across the week ("3 pm – 6 pm"), one slot ("Fri 3 pm"), or a count.
 */
export function describeCells(cells: ReadonlySet<number>): string {
  if (cells.size === 1) {
    const [c] = cells;
    return `${WEEKDAYS[Math.floor(c / 24)]} ${fmtHour(c % 24)}`;
  }

  const fullDays = DAYS.filter((d) => HOURS.every((h) => cells.has(d * 24 + h)));
  if (fullDays.length * 24 === cells.size) return listLabel(fullDays.map((d) => WEEKDAYS[d]), 3);

  const fullHours = HOURS.filter((h) => DAYS.every((d) => cells.has(d * 24 + h)));
  if (fullHours.length * 7 === cells.size) {
    const contiguous = fullHours.every((h, i) => i === 0 || h === fullHours[i - 1] + 1);
    if (contiguous) return `${fmtHour(fullHours[0])} – ${fmtHour((fullHours[fullHours.length - 1] + 1) % 24)}`;
    return listLabel(fullHours.map(fmtHour), 3);
  }

  return `${cells.size} hour slots`;
}
