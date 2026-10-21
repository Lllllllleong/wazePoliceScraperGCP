import type { FeatureCollection, Polygon } from 'geojson';

/** [west, south, east, north] in degrees. */
export type Bbox = readonly [number, number, number, number];

/**
 * The areas the scraper polled, from configs/bboxes.yaml. Nothing outside
 * them was collected, so empty map there means "not scraped", not "no police".
 */
export const COVERAGE: readonly Bbox[] = [
  [150.38822599217056, -34.254577954626086, 151.00867887302994, -33.937977044844004],
  [149.58926145838367, -34.76915040190209, 150.83016722010242, -34.138639582841435],
  [149.09281124417694, -35.21080621952668, 150.3337170058957, -34.583661538587855],
  [148.80885598970738, -35.4530012424677, 149.42930887056676, -35.14096097196958],
];

export const COVERAGE_BOUNDS: Bbox = [
  Math.min(...COVERAGE.map((b) => b[0])),
  Math.min(...COVERAGE.map((b) => b[1])),
  Math.max(...COVERAGE.map((b) => b[2])),
  Math.max(...COVERAGE.map((b) => b[3])),
];

export function contains(b: Bbox, lon: number, lat: number): boolean {
  return lon >= b[0] && lon <= b[2] && lat >= b[1] && lat <= b[3];
}

/** Web Mercator x in 0..1. */
export function mercatorX(lon: number): number {
  return (lon + 180) / 360;
}

/** Web Mercator y in 0..1 (0 at the north edge). */
export function mercatorY(lat: number): number {
  const s = Math.sin((lat * Math.PI) / 180);
  return 0.5 - Math.log((1 + s) / (1 - s)) / (4 * Math.PI);
}

/** Approximate distance in metres; accurate enough over a few kilometres. */
export function distanceMeters(lat1: number, lon1: number, lat2: number, lon2: number): number {
  const ky = 110_574;
  const kx = 111_320 * Math.cos((((lat1 + lat2) / 2) * Math.PI) / 180);
  return Math.hypot((lon2 - lon1) * kx, (lat2 - lat1) * ky);
}

/**
 * Non-overlapping rectangles that cover `outer` except where `boxes` are,
 * used to shade the area that was never scraped.
 */
export function outsideCoverage(outer: Bbox, boxes: readonly Bbox[] = COVERAGE): Bbox[] {
  const within = (v: number, lo: number, hi: number) => v >= lo && v <= hi;
  const xs = uniqueSorted([outer[0], outer[2], ...boxes.flatMap((b) => [b[0], b[2]])].filter((x) => within(x, outer[0], outer[2])));
  const ys = uniqueSorted([outer[1], outer[3], ...boxes.flatMap((b) => [b[1], b[3]])].filter((y) => within(y, outer[1], outer[3])));

  const rects: Bbox[] = [];
  for (let j = 0; j + 1 < ys.length; j++) {
    const cy = (ys[j] + ys[j + 1]) / 2;
    let runStart: number | null = null;
    for (let i = 0; i + 1 < xs.length; i++) {
      const cx = (xs[i] + xs[i + 1]) / 2;
      const covered = boxes.some((b) => cx > b[0] && cx < b[2] && cy > b[1] && cy < b[3]);
      if (!covered && runStart === null) runStart = xs[i];
      if (covered && runStart !== null) {
        rects.push([runStart, ys[j], xs[i], ys[j + 1]]);
        runStart = null;
      }
    }
    if (runStart !== null) rects.push([runStart, ys[j], xs[xs.length - 1], ys[j + 1]]);
  }
  return rects;
}

export function rectsToGeoJSON(rects: readonly Bbox[]): FeatureCollection<Polygon> {
  return {
    type: 'FeatureCollection',
    features: rects.map(([w, s, e, n]) => ({
      type: 'Feature',
      properties: {},
      geometry: { type: 'Polygon', coordinates: [[[w, s], [e, s], [e, n], [w, n], [w, s]]] },
    })),
  };
}

function uniqueSorted(values: number[]): number[] {
  return [...new Set(values)].sort((a, b) => a - b);
}
