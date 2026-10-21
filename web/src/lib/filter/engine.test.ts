import { describe, expect, it } from 'vitest';
import { decodeDataset } from '../data/dataset.ts';
import { makeFile } from '../testing/fixtures.ts';
import { crossfilter } from './engine.ts';
import { NO_FILTERS, type Filters } from './filters.ts';

// Day 0 (Mon): two hidden alerts in Goulburn, one camera in Marulan.
// Day 1 (Tue): one hidden alert in Goulburn, far to the east.
const ds = decodeDataset(
  makeFile([
    { minute: 8 * 60, subtype: 1, city: 1 },
    { minute: 9 * 60, subtype: 1, city: 1 },
    { minute: 10 * 60, subtype: 2, city: 2, street: 2, roadType: 1 },
    { minute: 1440 + 8 * 60, subtype: 1, city: 1, lon: 150.9 },
  ]),
);

const filters = (f: Partial<Filters>): Filters => ({ ...NO_FILTERS, ...f });

describe('crossfilter', () => {
  it('selects everything with no filters', () => {
    const r = crossfilter(ds, NO_FILTERS, null);
    expect(r.selected).toBe(4);
    expect([...r.perDay]).toEqual([3, 1, 0]);
    expect([...r.subtypes]).toEqual([0, 3, 1]);
    expect([...r.mapState]).toEqual([2, 2, 2, 2]);
  });

  it("leaves a chart's own counts alone so other choices stay visible", () => {
    const r = crossfilter(ds, filters({ subtypes: new Set([2]) }), null);
    expect(r.selected).toBe(1);
    // The subtype chart still shows every subtype...
    expect([...r.subtypes]).toEqual([0, 3, 1]);
    // ...while the others only count the camera alert.
    expect([...r.perDay]).toEqual([1, 0, 0]);
    expect([...r.towns]).toEqual([0, 0, 1]);
    expect([...r.mapState]).toEqual([1, 1, 2, 1]);
  });

  it('combines filters', () => {
    const r = crossfilter(ds, filters({ days: [0, 0], towns: new Set([1]) }), null);
    expect(r.selected).toBe(2);
    expect([...r.perDay]).toEqual([2, 1, 0]);
    expect([...r.towns]).toEqual([0, 2, 1]);
    expect([...r.subtypes]).toEqual([0, 2, 0]);
  });

  it('filters by hour-of-week cell', () => {
    const r = crossfilter(ds, filters({ cells: new Set([8, 24 + 8]) }), null);
    expect(r.selected).toBe(2);
    expect(r.cells[9]).toBe(1);
  });

  it('applies the map view only when asked, and never greys out the map for it', () => {
    const west = [148, -36, 150, -33] as const;
    expect(crossfilter(ds, NO_FILTERS, west).selected).toBe(4);

    const r = crossfilter(ds, filters({ inView: true }), west);
    expect(r.selected).toBe(3);
    expect([...r.mapState]).toEqual([2, 2, 2, 2]);
  });
});
