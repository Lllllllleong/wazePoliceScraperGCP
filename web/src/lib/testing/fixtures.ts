import type { DatasetFile, DayCoverage } from '../data/schema.ts';

export interface FixtureRow {
  /** Minutes since the fixture epoch. */
  minute: number;
  lat?: number;
  lon?: number;
  subtype?: number;
  street?: number;
  city?: number;
  roadType?: number;
  duration?: number;
  thumbsUp?: number;
}

/** Local midnight on Mon 3 Nov 2025 in Sydney (UTC+11). */
export const NOV_3_2025 = Date.UTC(2025, 10, 2, 13) / 1000;

/** Local midnight on Sat 4 Oct 2025 in Sydney (UTC+10). Daylight saving starts the next morning. */
export const OCT_4_2025 = Date.UTC(2025, 9, 3, 14) / 1000;

export const STREETS = ['M31 - Hume Mwy', 'M31 - Hume Hwy', 'Main St', ''];
export const CITIES = ['', 'Goulburn', 'Marulan'];
export const SUBTYPES = ['', 'POLICE_HIDING', 'POLICE_WITH_MOBILE_CAMERA'];

/** Three plain 24-hour days from 3 Nov 2025. */
export function plainDays(): DayCoverage[] {
  return ['2025-11-03', '2025-11-04', '2025-11-05'].map((date, i) => ({
    date,
    start: i * 1440,
    minutes: 1440,
    observed: 1440,
    gaps: [],
  }));
}

/** 4–6 Oct 2025, where the 5th is only 23 hours long. */
export function daylightSavingDays(): DayCoverage[] {
  return [
    { date: '2025-10-04', start: 0, minutes: 1440, observed: 1440, gaps: [] },
    { date: '2025-10-05', start: 1440, minutes: 1380, observed: 1380, gaps: [] },
    { date: '2025-10-06', start: 2820, minutes: 1440, observed: 600, gaps: [[2820, 3660]] },
  ];
}

export function makeFile(rows: FixtureRow[], opts: { epoch?: number; days?: DayCoverage[] } = {}): DatasetFile {
  const sorted = [...rows].sort((a, b) => a.minute - b.minute);
  let prev = 0;
  return {
    version: 1,
    generatedAt: '2026-01-01T00:00:00Z',
    timezone: 'Australia/Sydney',
    epoch: opts.epoch ?? NOV_3_2025,
    count: sorted.length,
    days: opts.days ?? plainDays(),
    dict: { subtype: SUBTYPES, street: STREETS, city: CITIES },
    cols: {
      publish: sorted.map((r) => {
        const delta = r.minute - prev;
        prev = r.minute;
        return delta;
      }),
      duration: sorted.map((r) => r.duration ?? 30),
      lat: sorted.map((r) => Math.round((r.lat ?? -34.75) * 1e5)),
      lon: sorted.map((r) => Math.round((r.lon ?? 149.72) * 1e5)),
      subtype: sorted.map((r) => r.subtype ?? 0),
      street: sorted.map((r) => r.street ?? 0),
      city: sorted.map((r) => r.city ?? 0),
      roadType: sorted.map((r) => r.roadType ?? 3),
      reliability: sorted.map(() => 5),
      confidence: sorted.map(() => 0),
      thumbsUp: sorted.map((r) => r.thumbsUp ?? 0),
      heading: sorted.map(() => -1),
    },
  };
}
