import { describe, expect, it } from 'vitest';
import { contains, distanceMeters, mercatorX, mercatorY, outsideCoverage, type Bbox } from './geo.ts';

describe('outsideCoverage', () => {
  const outer: Bbox = [0, 0, 10, 10];
  const boxes: Bbox[] = [
    [1, 1, 4, 4],
    [3, 3, 6, 6],
  ];
  const rects = outsideCoverage(outer, boxes);
  const shaded = (x: number, y: number) => rects.some((r) => contains(r, x, y));

  it('shades everything outside the boxes and nothing inside', () => {
    for (const [x, y] of [[0.5, 0.5], [5, 2], [2, 5], [9, 9], [6.5, 6.5]]) expect(shaded(x, y)).toBe(true);
    for (const [x, y] of [[2, 2], [3.5, 3.5], [5, 5]]) expect(shaded(x, y)).toBe(false);
  });

  it('does not overlap', () => {
    const area = rects.reduce((a, [w, s, e, n]) => a + (e - w) * (n - s), 0);
    expect(area).toBeCloseTo(100 - 9 - 9 + 1);
  });
});

describe('projection and distance', () => {
  it('maps the origin to the middle of the mercator square', () => {
    expect(mercatorX(0)).toBe(0.5);
    expect(mercatorY(0)).toBeCloseTo(0.5);
    expect(mercatorY(-34)).toBeGreaterThan(0.5);
  });

  it('measures short distances', () => {
    expect(distanceMeters(-34, 150, -34.001, 150)).toBeCloseTo(110.6, 0);
    expect(distanceMeters(-34, 150, -34, 150.001)).toBeCloseTo(92.3, 0);
  });
});
