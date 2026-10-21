import { describe, expect, it } from 'vitest';
import { NO_FILTERS, type Filters } from './filters.ts';
import { decodeFilters, encodeFilters, type Limits } from './url.ts';

const limits: Limits = { days: 172, subtypes: 4, roadTypes: 23, roads: 100, towns: 50 };

describe('filters in the URL', () => {
  it('is empty with no filters', () => {
    expect(encodeFilters(NO_FILTERS)).toBe('');
    expect(decodeFilters('', limits)).toEqual(NO_FILTERS);
  });

  it('round-trips every dimension', () => {
    const f: Filters = {
      days: [84, 100],
      cells: new Set([111, 96]),
      subtypes: new Set([1]),
      roadTypes: new Set([3, 6]),
      roads: new Set([0]),
      towns: new Set([12, 3]),
      inView: true,
    };
    const hash = encodeFilters(f);
    expect(hash).toBe('d=84-100&h=96,111&t=1&rt=3,6&r=0&tn=3,12&view=1');
    expect(decodeFilters(`#${hash}`, limits)).toEqual(f);
  });

  it('writes runs as ranges', () => {
    const friday = Array.from({ length: 24 }, (_, h) => 96 + h);
    const f: Filters = { ...NO_FILTERS, cells: new Set([...friday, 130, 140, 141]) };
    expect(encodeFilters(f)).toBe('h=96-119,130,140-141');
    expect(decodeFilters('h=96-119,130,140-141', limits).cells).toEqual(f.cells);
  });

  it('ignores values out of range or malformed', () => {
    const f = decodeFilters('d=170-400&t=1,9,x&r=-1,5&tn=48-60&view=yes', limits);
    expect(f.days).toEqual([170, 171]);
    expect(f.subtypes).toEqual(new Set([1]));
    expect(f.roads).toEqual(new Set([5]));
    expect(f.towns).toEqual(new Set([48, 49]));
    expect(f.inView).toBe(false);
  });

  it('drops a backwards date range', () => {
    expect(decodeFilters('d=10-2', limits).days).toBeNull();
  });
});
