const integer = new Intl.NumberFormat('en-AU');

export function fmtInt(n: number): string {
  return integer.format(n);
}

/** 0.1203 → "12.0%" */
export function fmtPercent(share: number): string {
  return `${(share * 100).toFixed(1)}%`;
}

/** 0 → "12 am", 15 → "3 pm" */
export function fmtHour(hour: number): string {
  return `${hour % 12 || 12} ${hour < 12 ? 'am' : 'pm'}`;
}

/** 176 → "2 h 56 min" */
export function fmtDuration(minutes: number): string {
  if (minutes < 60) return `${minutes} min`;
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return m ? `${h} h ${m} min` : `${h} h`;
}

const POINTS = ['N', 'NNE', 'NE', 'ENE', 'E', 'ESE', 'SE', 'SSE', 'S', 'SSW', 'SW', 'WSW', 'W', 'WNW', 'NW', 'NNW'];

/** 249 → "WSW" */
export function compass(degrees: number): string {
  const d = ((degrees % 360) + 360) % 360;
  return POINTS[Math.round(d / 22.5) % 16];
}

const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

/** "2025-12-19" → "19 Dec 2025" (or "19 Dec" without the year). */
export function fmtDate(date: string, withYear = true): string {
  const [y, m, d] = date.split('-').map(Number);
  return withYear ? `${d} ${MONTHS[m - 1]} ${y}` : `${d} ${MONTHS[m - 1]}`;
}

/** A date range, dropping the repeated year: "19 Dec 2025 – 4 Jan 2026", "3 – 9 Nov 2025". */
export function fmtDateRange(from: string, to: string): string {
  if (from === to) return fmtDate(from);
  const sameYear = from.slice(0, 4) === to.slice(0, 4);
  return `${fmtDate(from, !sameYear)} – ${fmtDate(to)}`;
}

const dateTimeParts = new Map<string, Intl.DateTimeFormat>();

/** An instant in the given zone: "Mon 13 Oct 2025, 3:55 pm". */
export function fmtDateTime(ms: number, timeZone: string): string {
  let f = dateTimeParts.get(timeZone);
  if (!f) {
    f = new Intl.DateTimeFormat('en-AU', {
      timeZone,
      weekday: 'short',
      day: 'numeric',
      month: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: '2-digit',
      hourCycle: 'h23',
    });
    dateTimeParts.set(timeZone, f);
  }
  const p = Object.fromEntries(f.formatToParts(ms).map((x) => [x.type, x.value]));
  const hour = Number(p.hour);
  return `${p.weekday} ${Number(p.day)} ${MONTHS[Number(p.month) - 1]} ${p.year}, ${hour % 12 || 12}:${p.minute} ${hour < 12 ? 'am' : 'pm'}`;
}
