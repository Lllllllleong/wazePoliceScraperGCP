import type { Dataset } from '../data/dataset.ts';
import type { Bbox } from '../geo.ts';
import type { Filters } from './filters.ts';

export interface CrossfilterResult {
  /** Alerts matching every filter. */
  selected: number;
  /**
   * Each chart's counts honour every filter except its own, so a chart shows
   * the alternatives to its current selection.
   */
  perDay: Int32Array;
  cells: Int32Array;
  subtypes: Int32Array;
  roadTypes: Int32Array;
  roads: Int32Array;
  towns: Int32Array;
  /**
   * Per alert: 2 if it matches every filter except the map view (so panning
   * doesn't recolour the map), 1 otherwise.
   */
  mapState: Uint8Array;
}

const DAYS = 1;
const CELLS = 2;
const SUBTYPES = 4;
const ROAD_TYPES = 8;
const ROADS = 16;
const TOWNS = 32;
const VIEW = 64;

/**
 * Applies the filters in one pass. `view` is the visible map area, used only
 * when `filters.inView` is on.
 */
export function crossfilter(ds: Dataset, filters: Filters, view: Bbox | null): CrossfilterResult {
  const n = ds.count;
  const perDay = new Int32Array(ds.days.length);
  const cells = new Int32Array(7 * 24);
  const subtypes = new Int32Array(ds.subtypes.length);
  const roadTypes = new Int32Array(ds.roadTypeCount);
  const roads = new Int32Array(ds.roads.length);
  const towns = new Int32Array(ds.towns.length);
  const mapState = new Uint8Array(n);

  const days = filters.days;
  const [dayFrom, dayTo] = days ?? [0, 0];
  const cellMask = lookup(filters.cells, cells.length);
  const subtypeMask = lookup(filters.subtypes, subtypes.length);
  const roadTypeMask = lookup(filters.roadTypes, roadTypes.length);
  const roadMask = lookup(filters.roads, roads.length);
  const townMask = lookup(filters.towns, towns.length);
  const bounds = filters.inView ? view : null;

  let selected = 0;
  for (let i = 0; i < n; i++) {
    let failed = 0;
    if (days && (ds.day[i] < dayFrom || ds.day[i] > dayTo)) failed |= DAYS;
    if (cellMask && !cellMask[ds.cell[i]]) failed |= CELLS;
    if (subtypeMask && !subtypeMask[ds.subtype[i]]) failed |= SUBTYPES;
    if (roadTypeMask && !roadTypeMask[ds.roadType[i]]) failed |= ROAD_TYPES;
    if (roadMask && !roadMask[ds.road[i]]) failed |= ROADS;
    if (townMask && !townMask[ds.town[i]]) failed |= TOWNS;
    if (bounds && (ds.lon[i] < bounds[0] || ds.lon[i] > bounds[2] || ds.lat[i] < bounds[1] || ds.lat[i] > bounds[3])) {
      failed |= VIEW;
    }

    mapState[i] = (failed & ~VIEW) === 0 ? 2 : 1;

    if (failed === 0) {
      selected++;
      perDay[ds.day[i]]++;
      cells[ds.cell[i]]++;
      subtypes[ds.subtype[i]]++;
      roadTypes[ds.roadType[i]]++;
      roads[ds.road[i]]++;
      towns[ds.town[i]]++;
    } else if ((failed & (failed - 1)) === 0) {
      // Failed exactly one filter: still counts in that filter's own chart.
      switch (failed) {
        case DAYS:
          perDay[ds.day[i]]++;
          break;
        case CELLS:
          cells[ds.cell[i]]++;
          break;
        case SUBTYPES:
          subtypes[ds.subtype[i]]++;
          break;
        case ROAD_TYPES:
          roadTypes[ds.roadType[i]]++;
          break;
        case ROADS:
          roads[ds.road[i]]++;
          break;
        case TOWNS:
          towns[ds.town[i]]++;
          break;
      }
    }
  }

  return { selected, perDay, cells, subtypes, roadTypes, roads, towns, mapState };
}

/** A membership table for a set filter, or null when the filter is off. */
function lookup(values: ReadonlySet<number>, size: number): Uint8Array | null {
  if (values.size === 0) return null;
  const table = new Uint8Array(size);
  for (const v of values) if (v >= 0 && v < size) table[v] = 1;
  return table;
}
