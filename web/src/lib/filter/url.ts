import { NO_FILTERS, type Filters, type SetDimension } from './filters.ts';

/** How many values each dimension has, so links from elsewhere can't select nonsense. */
export interface Limits {
  days: number;
  subtypes: number;
  roadTypes: number;
  roads: number;
  towns: number;
}

const SET_PARAMS: [SetDimension, string][] = [
  ['cells', 'h'],
  ['subtypes', 't'],
  ['roadTypes', 'rt'],
  ['roads', 'r'],
  ['towns', 'tn'],
];

/** Filters as a URL fragment (without the leading "#"); empty when nothing is filtered. */
export function encodeFilters(f: Filters): string {
  const params = new URLSearchParams();
  if (f.days) params.set('d', `${f.days[0]}-${f.days[1]}`);
  for (const [dim, key] of SET_PARAMS) {
    if (f[dim].size) params.set(key, compactList(f[dim]));
  }
  if (f.inView) params.set('view', '1');
  // Commas are safe in a fragment and much easier to read than %2C.
  return params.toString().replace(/%2C/g, ',');
}

export function decodeFilters(fragment: string, limits: Limits): Filters {
  const params = new URLSearchParams(fragment.replace(/^#/, ''));
  const size: Record<SetDimension, number> = {
    cells: 7 * 24,
    subtypes: limits.subtypes,
    roadTypes: limits.roadTypes,
    roads: limits.roads,
    towns: limits.towns,
  };

  const f: Filters = { ...NO_FILTERS, inView: params.get('view') === '1' };

  const d = /^(\d+)-(\d+)$/.exec(params.get('d') ?? '');
  if (d) {
    const from = Math.min(Number(d[1]), limits.days - 1);
    const to = Math.min(Number(d[2]), limits.days - 1);
    if (from <= to) f.days = [from, to];
  }

  for (const [dim, key] of SET_PARAMS) {
    f[dim] = new Set(expandList(params.get(key) ?? '', size[dim]));
  }
  return f;
}

/** [1, 2, 3, 7, 9, 10] → "1-3,7,9-10", so a whole weekday is "96-119" rather than 24 numbers. */
function compactList(values: ReadonlySet<number>): string {
  const sorted = [...values].sort((a, b) => a - b);
  const parts: string[] = [];
  for (let i = 0; i < sorted.length; i++) {
    const start = sorted[i];
    while (i + 1 < sorted.length && sorted[i + 1] === sorted[i] + 1) i++;
    parts.push(start === sorted[i] ? `${start}` : `${start}-${sorted[i]}`);
  }
  return parts.join(',');
}

function expandList(text: string, size: number): number[] {
  const values: number[] = [];
  for (const part of text.split(',')) {
    const m = /^(\d+)(?:-(\d+))?$/.exec(part);
    if (!m) continue;
    const from = Number(m[1]);
    const to = Math.min(Number(m[2] ?? m[1]), size - 1);
    for (let v = from; v <= to; v++) values.push(v);
  }
  return values;
}
