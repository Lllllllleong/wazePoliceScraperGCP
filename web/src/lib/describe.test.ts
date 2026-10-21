import { describe, expect, it } from 'vitest';
import { describeCells, listLabel } from './describe.ts';

const day = (d: number) => Array.from({ length: 24 }, (_, h) => d * 24 + h);
const hour = (h: number) => Array.from({ length: 7 }, (_, d) => d * 24 + h);

describe('listLabel', () => {
  it('lists a few and counts the rest', () => {
    expect(listLabel(['Freeway'])).toBe('Freeway');
    expect(listLabel(['Freeway', 'Ramp'])).toBe('Freeway, Ramp');
    expect(listLabel(['Freeway', 'Ramp', 'Street', 'Alley'])).toBe('Freeway, Ramp +2 more');
  });
});

describe('describeCells', () => {
  it('names a single slot', () => {
    expect(describeCells(new Set([4 * 24 + 15]))).toBe('Fri 3 pm');
  });

  it('names whole weekdays', () => {
    expect(describeCells(new Set([...day(4), ...day(5)]))).toBe('Fri, Sat');
  });

  it('names a run of hours across the week', () => {
    expect(describeCells(new Set([...hour(15), ...hour(16), ...hour(17)]))).toBe('3 pm – 6 pm');
    expect(describeCells(new Set([...hour(23)]))).toBe('11 pm – 12 am');
  });

  it('lists scattered hours', () => {
    expect(describeCells(new Set([...hour(8), ...hour(17)]))).toBe('8 am, 5 pm');
  });

  it('counts anything else', () => {
    expect(describeCells(new Set([0, 30, 100]))).toBe('3 hour slots');
  });
});
