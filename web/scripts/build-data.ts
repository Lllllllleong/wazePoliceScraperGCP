/**
 * Builds the static dataset the dashboard loads from the daily archives the
 * archive service wrote to GCS.
 *
 * Usage (from web/):
 *   gcloud storage cp "gs://wazepolicescrapergcp-archive/*.jsonl" <dir>
 *   npm run build:data -- <dir> [out-file]
 *
 * The collected data is final, so the output is committed and this only needs
 * rerunning if the format changes (bump DATASET_VERSION when it does).
 *
 * Only the fields the dashboard uses are kept. In particular the raw Waze
 * payloads are dropped, since they carry reporter usernames and comments.
 */
import { createReadStream, mkdirSync, readdirSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { createInterface } from 'node:readline';
import { gzipSync } from 'node:zlib';
import { DATASET_PATH, DATASET_VERSION, TIMEZONE } from '../src/lib/data/schema.ts';
import type { DatasetFile, DayCoverage } from '../src/lib/data/schema.ts';

/** Unobserved stretches at least this long are reported as coverage gaps. */
const GAP_MINUTES = 10;
const MINUTE = 60_000;

const localParts = new Intl.DateTimeFormat('en-CA', {
  timeZone: TIMEZONE,
  year: 'numeric',
  month: '2-digit',
  day: '2-digit',
  hour: '2-digit',
  minute: '2-digit',
  hourCycle: 'h23',
});

/** One line of an archive file (a serialized PoliceAlert). */
interface ArchiveRecord {
  UUID: string;
  Type: string;
  Subtype?: string;
  Street?: string;
  City?: string;
  LocationGeo?: { latitude: number; longitude: number } | null;
  Reliability?: number;
  Confidence?: number;
  PublishTime: string;
  ScrapeTime: string;
  ExpireTime: string;
  NThumbsUpLast?: number;
  RawDataLast?: string;
}

/** The parts of the raw Waze payload we keep. */
interface RawAlert {
  reliability?: number;
  confidence?: number;
  roadType?: number;
  magvar?: number;
}

interface Alert {
  publish: number;
  firstSeen: number;
  lastSeen: number;
  lat: number;
  lon: number;
  subtype: string;
  street: string;
  city: string;
  roadType: number;
  reliability: number;
  confidence: number;
  thumbsUp: number;
  heading: number;
}

const [inDir, outArg] = process.argv.slice(2);
if (!inDir) {
  console.error('usage: node scripts/build-data.ts <archive-dir> [out-file]');
  process.exit(1);
}
const outFile = resolve(outArg ?? join(import.meta.dirname, '../public', DATASET_PATH));

const dates = readdirSync(inDir)
  .filter((f) => /^\d{4}-\d{2}-\d{2}\.jsonl$/.test(f))
  .map((f) => f.slice(0, 10))
  .sort();
if (dates.length === 0) {
  console.error(`no YYYY-MM-DD.jsonl files in ${inDir}`);
  process.exit(1);
}

// --- Read and dedupe -------------------------------------------------------
// An alert active across midnight appears in both days' archives, so keep one
// record per UUID: the latest state, with the earliest first-seen time.

const byId = new Map<string, Alert>();
let linesRead = 0;
let skipped = 0;

for (const date of dates) {
  const lines = createInterface({ input: createReadStream(join(inDir, `${date}.jsonl`)), crlfDelay: Infinity });
  for await (const line of lines) {
    if (!line.trim()) continue;
    linesRead++;
    const rec = JSON.parse(line) as ArchiveRecord;
    if (rec.Type !== 'POLICE' || !rec.LocationGeo) {
      skipped++;
      continue;
    }

    const firstSeen = parseTime(rec.ScrapeTime);
    const lastSeen = parseTime(rec.ExpireTime);
    const prev = byId.get(rec.UUID);
    if (prev && prev.lastSeen >= lastSeen) {
      prev.firstSeen = Math.min(prev.firstSeen, firstSeen);
      continue;
    }

    const raw = parseRaw(rec.RawDataLast);
    byId.set(rec.UUID, {
      publish: parseTime(rec.PublishTime),
      firstSeen: prev ? Math.min(prev.firstSeen, firstSeen) : firstSeen,
      lastSeen,
      lat: rec.LocationGeo.latitude,
      lon: rec.LocationGeo.longitude,
      subtype: rec.Subtype ?? '',
      street: rec.Street ?? '',
      city: rec.City ?? '',
      roadType: raw.roadType ?? 0,
      // The top-level values are from the first scrape; the raw payload has the final ones.
      reliability: raw.reliability ?? rec.Reliability ?? 0,
      confidence: raw.confidence ?? rec.Confidence ?? 0,
      thumbsUp: rec.NThumbsUpLast ?? 0,
      heading: raw.magvar ?? -1,
    });
  }
}

// --- Day boundaries and coverage ------------------------------------------

const epochMs = localMidnight(dates[0]);
const endMs = localMidnight(nextDate(dates[dates.length - 1]));
const totalMinutes = (endMs - epochMs) / MINUTE;

// Difference array over every minute in range: >0 means the scraper was
// observing at least one alert during that minute.
const diff = new Int32Array(totalMinutes + 1);
for (const a of byId.values()) {
  const from = Math.max(0, Math.floor((a.firstSeen - epochMs) / MINUTE));
  const to = Math.min(totalMinutes, Math.floor((a.lastSeen - epochMs) / MINUTE) + 1);
  if (from >= to) continue;
  diff[from]++;
  diff[to]--;
}
const observed = new Uint8Array(totalMinutes);
for (let m = 0, active = 0; m < totalMinutes; m++) {
  active += diff[m];
  observed[m] = active > 0 ? 1 : 0;
}

const days: DayCoverage[] = dates.map((date) => {
  const start = (localMidnight(date) - epochMs) / MINUTE;
  const end = (localMidnight(nextDate(date)) - epochMs) / MINUTE;
  let seen = 0;
  const gaps: [number, number][] = [];
  for (let m = start; m < end; m++) {
    if (observed[m]) {
      seen++;
      continue;
    }
    let runEnd = m;
    while (runEnd < end && !observed[runEnd]) runEnd++;
    if (runEnd - m >= GAP_MINUTES) gaps.push([m, runEnd]);
    m = runEnd - 1;
  }
  return { date, start, minutes: end - start, observed: seen, gaps };
});

// --- Encode ----------------------------------------------------------------

const alerts = [...byId.values()]
  .filter((a) => a.publish >= epochMs && a.publish < endMs)
  .sort((a, b) => a.publish - b.publish);
const outOfRange = byId.size - alerts.length;

const subtypes = dictionary(alerts.map((a) => a.subtype));
const streets = dictionary(alerts.map((a) => a.street));
const cities = dictionary(alerts.map((a) => a.city));

const cols: DatasetFile['cols'] = {
  publish: [],
  duration: [],
  lat: [],
  lon: [],
  subtype: [],
  street: [],
  city: [],
  roadType: [],
  reliability: [],
  confidence: [],
  thumbsUp: [],
  heading: [],
};
let prevPublish = 0;
for (const a of alerts) {
  const publish = Math.floor((a.publish - epochMs) / MINUTE);
  cols.publish.push(publish - prevPublish);
  prevPublish = publish;
  cols.duration.push(Math.max(0, Math.round((a.lastSeen - a.publish) / MINUTE)));
  cols.lat.push(Math.round(a.lat * 1e5));
  cols.lon.push(Math.round(a.lon * 1e5));
  cols.subtype.push(subtypes.index.get(a.subtype)!);
  cols.street.push(streets.index.get(a.street)!);
  cols.city.push(cities.index.get(a.city)!);
  cols.roadType.push(a.roadType);
  cols.reliability.push(a.reliability);
  cols.confidence.push(a.confidence);
  cols.thumbsUp.push(a.thumbsUp);
  cols.heading.push(a.heading);
}

const dataset: DatasetFile = {
  version: DATASET_VERSION,
  generatedAt: new Date().toISOString(),
  timezone: TIMEZONE,
  epoch: epochMs / 1000,
  count: alerts.length,
  days,
  dict: { subtype: subtypes.values, street: streets.values, city: cities.values },
  cols,
};

const json = Buffer.from(JSON.stringify(dataset));
const gz = gzipSync(json, { level: 9 });
mkdirSync(dirname(outFile), { recursive: true });
writeFileSync(outFile, gz);

const gapDays = days.filter((d) => d.gaps.length > 0);
console.log(`days            ${dates[0]} → ${dates[dates.length - 1]} (${dates.length})`);
console.log(`lines read      ${linesRead}`);
console.log(`skipped         ${skipped} (not POLICE or no location)`);
console.log(`unique alerts   ${byId.size} (${linesRead - skipped - byId.size} duplicates merged)`);
console.log(`out of range    ${outOfRange} (published outside the day range)`);
console.log(`rows written    ${alerts.length}`);
console.log(`dictionaries    ${subtypes.values.length} subtypes, ${streets.values.length} streets, ${cities.values.length} cities`);
console.log(`days with gaps  ${gapDays.length}`);
for (const d of gapDays) {
  const spans = d.gaps.map(([s, e]) => `${clock(s - d.start)}–${clock(e - d.start)}`).join(', ');
  console.log(`  ${d.date}  ${((d.observed / d.minutes) * 100).toFixed(1)}% observed  ${spans}`);
}
console.log(`size            ${(json.length / 1048576).toFixed(1)} MB json, ${(gz.length / 1048576).toFixed(2)} MB gzipped`);
console.log(`wrote           ${outFile}`);

// --- Helpers ---------------------------------------------------------------

function parseTime(iso: string): number {
  const t = Date.parse(iso);
  if (Number.isNaN(t)) throw new Error(`bad timestamp: ${iso}`);
  return t;
}

function parseRaw(raw: string | undefined): RawAlert {
  if (!raw) return {};
  try {
    return JSON.parse(raw) as RawAlert;
  } catch {
    return {};
  }
}

/** Distinct values, most frequent first so common values get short indices. */
function dictionary(values: string[]) {
  const counts = new Map<string, number>();
  for (const v of values) counts.set(v, (counts.get(v) ?? 0) + 1);
  const sorted = [...counts.keys()].sort((a, b) => counts.get(b)! - counts.get(a)!);
  return { values: sorted, index: new Map(sorted.map((v, i) => [v, i])) };
}

/** Offset of TIMEZONE from UTC at the given instant, in ms. */
function tzOffset(utcMs: number): number {
  const p = Object.fromEntries(localParts.formatToParts(utcMs).map((x) => [x.type, x.value]));
  const asUtc = Date.UTC(+p.year, +p.month - 1, +p.day, +p.hour, +p.minute);
  return asUtc - Math.floor(utcMs / MINUTE) * MINUTE;
}

/** UTC ms of local midnight on a YYYY-MM-DD date in TIMEZONE. */
function localMidnight(date: string): number {
  const [y, m, d] = date.split('-').map(Number);
  const guess = Date.UTC(y, m - 1, d);
  return guess - tzOffset(guess - tzOffset(guess));
}

function nextDate(date: string): string {
  const [y, m, d] = date.split('-').map(Number);
  return new Date(Date.UTC(y, m - 1, d + 1)).toISOString().slice(0, 10);
}

function clock(minuteOfDay: number): string {
  const h = Math.floor(minuteOfDay / 60);
  return `${String(h).padStart(2, '0')}:${String(minuteOfDay % 60).padStart(2, '0')}`;
}
