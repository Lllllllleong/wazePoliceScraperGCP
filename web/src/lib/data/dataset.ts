import type { Bbox } from '../geo.ts';
import { groupRoads, type RoadGroup } from './roads.ts';
import type { DatasetFile } from './schema.ts';

/** Days the scraper observed for less than this share of the time are flagged as incomplete. */
export const INCOMPLETE_COVERAGE = 0.95;

export interface Day {
  index: number;
  /** Local date, YYYY-MM-DD. */
  date: string;
  /** Minute offset of local midnight from the dataset epoch. */
  start: number;
  minutes: number;
  /** 0 = Monday … 6 = Sunday. */
  weekday: number;
  /** Share of the day the scraper was observing, 0–1. */
  coverage: number;
  incomplete: boolean;
}

/** The decoded dataset: one typed array per field, indexed by alert. */
export interface Dataset {
  count: number;
  timezone: string;
  /** Unix ms of local midnight on the first day. */
  epochMs: number;
  days: Day[];

  subtypes: string[];
  streets: string[];
  towns: string[];
  roads: RoadGroup[];
  /** Length of arrays indexed by road type code. */
  roadTypeCount: number;

  /** Minutes since the epoch when the alert was reported. */
  minute: Int32Array;
  day: Uint16Array;
  /** Local hour of day. */
  hour: Uint8Array;
  weekday: Uint8Array;
  /** weekday * 24 + hour */
  cell: Uint8Array;
  /** Minutes from report until last seen. */
  duration: Uint16Array;
  lat: Float64Array;
  lon: Float64Array;
  subtype: Uint8Array;
  street: Uint16Array;
  road: Uint16Array;
  town: Uint16Array;
  roadType: Uint8Array;
  reliability: Uint8Array;
  confidence: Uint8Array;
  thumbsUp: Uint16Array;
  /** Degrees, -1 when unknown. */
  heading: Int16Array;

  /** Counts over the whole dataset, to compare filtered counts against. */
  totals: {
    subtypes: Int32Array;
    roadTypes: Int32Array;
    roads: Int32Array;
    towns: Int32Array;
  };
}

export function decodeDataset(file: DatasetFile): Dataset {
  const n = file.count;
  for (const [name, col] of Object.entries(file.cols)) {
    if (col.length !== n) throw new Error(`Column "${name}" has ${col.length} rows, expected ${n}`);
  }

  const epochMs = file.epoch * 1000;
  const days: Day[] = file.days.map((d, index) => ({
    index,
    date: d.date,
    start: d.start,
    minutes: d.minutes,
    weekday: weekdayOf(d.date),
    coverage: d.observed / d.minutes,
    incomplete: d.observed / d.minutes < INCOMPLETE_COVERAGE,
  }));

  const minute = new Int32Array(n);
  const day = new Uint16Array(n);
  const hour = new Uint8Array(n);
  const weekday = new Uint8Array(n);
  const cell = new Uint8Array(n);
  const localHour = new Intl.DateTimeFormat('en-AU', { timeZone: file.timezone, hour: 'numeric', hourCycle: 'h23' });

  // Rows are sorted by report time, so the day index only moves forward.
  for (let i = 0, t = 0, d = 0; i < n; i++) {
    t += file.cols.publish[i];
    minute[i] = t;
    while (d + 1 < days.length && t >= days[d + 1].start) d++;
    const today = days[d];
    // Daylight-saving changeover days aren't 24 hours long, so ask Intl for those.
    const h = today.minutes === 1440 ? Math.floor((t - today.start) / 60) : Number(localHour.format(epochMs + t * 60_000));
    const clamped = Math.min(23, Math.max(0, h));
    day[i] = d;
    hour[i] = clamped;
    weekday[i] = today.weekday;
    cell[i] = today.weekday * 24 + clamped;
  }

  const lat = Float64Array.from(file.cols.lat, (v) => v / 1e5);
  const lon = Float64Array.from(file.cols.lon, (v) => v / 1e5);
  const subtype = Uint8Array.from(file.cols.subtype);
  const street = Uint16Array.from(file.cols.street);
  const town = Uint16Array.from(file.cols.city);
  const roadType = Uint8Array.from(file.cols.roadType);

  const streetCounts = tally(street, file.dict.street.length);
  const { roads, streetToRoad } = groupRoads(file.dict.street, streetCounts);
  const road = street.map((s) => streetToRoad[s]);

  let roadTypeCount = 1;
  for (let i = 0; i < n; i++) roadTypeCount = Math.max(roadTypeCount, roadType[i] + 1);

  return {
    count: n,
    timezone: file.timezone,
    epochMs,
    days,
    subtypes: file.dict.subtype,
    streets: file.dict.street,
    towns: file.dict.city,
    roads,
    roadTypeCount,
    minute,
    day,
    hour,
    weekday,
    cell,
    duration: Uint16Array.from(file.cols.duration, (v) => Math.min(v, 65535)),
    lat,
    lon,
    subtype,
    street,
    road,
    town,
    roadType,
    reliability: Uint8Array.from(file.cols.reliability),
    confidence: Uint8Array.from(file.cols.confidence),
    thumbsUp: Uint16Array.from(file.cols.thumbsUp),
    heading: Int16Array.from(file.cols.heading),
    totals: {
      subtypes: tally(subtype, file.dict.subtype.length),
      roadTypes: tally(roadType, roadTypeCount),
      roads: tally(road, roads.length),
      towns: tally(town, file.dict.city.length),
    },
  };
}

/** The bounding box of every alert on a road or in a town, or null if there are none. */
export function extentOf(ds: Dataset, column: 'road' | 'town', id: number): Bbox | null {
  const values = ds[column];
  let w = Infinity, s = Infinity, e = -Infinity, n = -Infinity;
  for (let i = 0; i < ds.count; i++) {
    if (values[i] !== id) continue;
    w = Math.min(w, ds.lon[i]);
    e = Math.max(e, ds.lon[i]);
    s = Math.min(s, ds.lat[i]);
    n = Math.max(n, ds.lat[i]);
  }
  return w === Infinity ? null : [w, s, e, n];
}

function tally(values: ArrayLike<number>, size: number): Int32Array {
  const counts = new Int32Array(size);
  for (let i = 0; i < values.length; i++) counts[values[i]]++;
  return counts;
}

/** "2025-12-19" → 4 (Friday), with Monday as 0. */
function weekdayOf(date: string): number {
  const [y, m, d] = date.split('-').map(Number);
  return (new Date(Date.UTC(y, m - 1, d)).getUTCDay() + 6) % 7;
}
