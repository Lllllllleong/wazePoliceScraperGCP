import type { Dataset } from './data/dataset.ts';

export interface SpotStats {
  /** Alerts within the radius, ignoring filters. */
  count: number;
  /** Of those, how many match the current filters. */
  matching: number;
  /** Distinct days with at least one alert. */
  days: number;
  hours: Int32Array;
  subtypes: Int32Array;
}

/** Everything reported within `radiusM` of alert `index`. */
export function spotStats(ds: Dataset, index: number, radiusM: number, mapState?: Uint8Array): SpotStats {
  const lat0 = ds.lat[index];
  const lon0 = ds.lon[index];
  const ky = 110_574;
  const kx = 111_320 * Math.cos((lat0 * Math.PI) / 180);
  const r2 = radiusM * radiusM;
  const dLat = radiusM / ky;

  const hours = new Int32Array(24);
  const subtypes = new Int32Array(ds.subtypes.length);
  const seenDays = new Uint8Array(ds.days.length);
  let count = 0;
  let matching = 0;
  let days = 0;

  for (let i = 0; i < ds.count; i++) {
    const dy = ds.lat[i] - lat0;
    if (dy > dLat || dy < -dLat) continue;
    const x = (ds.lon[i] - lon0) * kx;
    const y = dy * ky;
    if (x * x + y * y > r2) continue;
    count++;
    if (mapState?.[i] === 2) matching++;
    hours[ds.hour[i]]++;
    subtypes[ds.subtype[i]]++;
    if (!seenDays[ds.day[i]]) {
      seenDays[ds.day[i]] = 1;
      days++;
    }
  }
  return { count, matching: mapState ? matching : count, days, hours, subtypes };
}
