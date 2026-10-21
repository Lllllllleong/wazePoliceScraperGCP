/**
 * Static dataset format, written by scripts/build-data.ts and loaded by the app.
 *
 * The file is columnar: row i is the i-th entry of every array in `cols`.
 * Rows are sorted by publish time, and `cols.publish` is delta-encoded, so
 * the absolute value for row i is the running sum of entries 0..i.
 */

export const DATASET_VERSION = 1;

/** Where the app loads the dataset from. The version is in the name so it can be cached forever. */
export const DATASET_PATH = `data/alerts.v${DATASET_VERSION}.json.gz`;

/** Day boundaries and hour-of-day are all in Canberra/Sydney local time. */
export const TIMEZONE = 'Australia/Sydney';

export interface DatasetFile {
  version: typeof DATASET_VERSION;
  generatedAt: string;
  timezone: string;
  /** Unix seconds of local midnight on the first day; minute offsets count from here. */
  epoch: number;
  count: number;
  days: DayCoverage[];
  dict: {
    subtype: string[];
    street: string[];
    city: string[];
  };
  cols: {
    /** Minutes since `epoch` when the alert was reported, delta-encoded. */
    publish: number[];
    /** Minutes from report until the scraper last saw the alert. */
    duration: number[];
    /** Degrees × 1e5. */
    lat: number[];
    lon: number[];
    subtype: number[];
    street: number[];
    city: number[];
    /** Waze road type code (see ROAD_TYPE_LABELS). 0 when unknown. */
    roadType: number[];
    /** Final observed values, 0–10. */
    reliability: number[];
    confidence: number[];
    thumbsUp: number[];
    /** Reporter heading in degrees (Waze `magvar`), -1 when unknown. */
    heading: number[];
  };
}

export interface DayCoverage {
  /** Local date, YYYY-MM-DD. */
  date: string;
  /** Minute offset of local midnight. */
  start: number;
  /** Length of the day in minutes (1380 on the daylight-saving changeover). */
  minutes: number;
  /** Minutes the scraper was observing at least one alert. */
  observed: number;
  /** Unobserved stretches of at least GAP_MINUTES, as [start, end) minute offsets. */
  gaps: [number, number][];
}
