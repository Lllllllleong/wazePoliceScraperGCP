import { describe, expect, it } from 'vitest';
import { isFiltered, NO_FILTERS, pick } from './filters.ts';

const set = (...v: number[]) => new Set(v);

describe('pick', () => {
  it('selects only the clicked values', () => {
    expect(pick(set(1, 2), [3], false)).toEqual(set(3));
  });

  it('clears when the clicked values were the whole selection', () => {
    expect(pick(set(3), [3], false)).toEqual(set());
  });

  it('adds and removes with an additive click', () => {
    expect(pick(set(1), [2], true)).toEqual(set(1, 2));
    expect(pick(set(1, 2), [2], true)).toEqual(set(1));
  });

  it('treats a group as one click', () => {
    expect(pick(set(1), [1, 2, 3], false)).toEqual(set(1, 2, 3));
    expect(pick(set(1, 2, 3), [1, 2, 3], false)).toEqual(set());
    expect(pick(set(1, 9), [1, 2], true)).toEqual(set(1, 2, 9));
  });
});

describe('isFiltered', () => {
  it('is false only with no filters', () => {
    expect(isFiltered(NO_FILTERS)).toBe(false);
    expect(isFiltered({ ...NO_FILTERS, days: [0, 3] })).toBe(true);
    expect(isFiltered({ ...NO_FILTERS, towns: set(1) })).toBe(true);
    expect(isFiltered({ ...NO_FILTERS, inView: true })).toBe(true);
  });
});
