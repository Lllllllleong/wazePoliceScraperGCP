import { describe, expect, it } from 'vitest';
import { daylightSavingDays, makeFile, OCT_4_2025 } from '../testing/fixtures.ts';
import { decodeDataset, extentOf } from './dataset.ts';

describe('decodeDataset', () => {
  it('undoes the delta encoding and assigns days', () => {
    const ds = decodeDataset(makeFile([{ minute: 5 }, { minute: 1500 }, { minute: 1439 }, { minute: 3000 }]));
    expect([...ds.minute]).toEqual([5, 1439, 1500, 3000]);
    expect([...ds.day]).toEqual([0, 0, 1, 2]);
  });

  it('works out local hour and weekday', () => {
    // 3 Nov 2025 was a Monday. 1500 minutes in is Tue 1:00 am.
    const ds = decodeDataset(makeFile([{ minute: 15 * 60 }, { minute: 1500 }]));
    expect([...ds.hour]).toEqual([15, 1]);
    expect([...ds.weekday]).toEqual([0, 1]);
    expect([...ds.cell]).toEqual([15, 24 + 1]);
  });

  it('handles the 23-hour daylight-saving day', () => {
    // Sun 5 Oct 2025: clocks jump from 2 am to 3 am. 150 minutes after
    // midnight is 3:30 am local, not 2:30.
    const file = makeFile([{ minute: 1440 + 150 }, { minute: 1440 + 60 }, { minute: 2820 + 480 }], {
      epoch: OCT_4_2025,
      days: daylightSavingDays(),
    });
    const ds = decodeDataset(file);
    expect([...ds.hour]).toEqual([1, 3, 8]);
    expect([...ds.weekday]).toEqual([6, 6, 0]);
  });

  it('flags days with incomplete coverage', () => {
    const ds = decodeDataset(makeFile([{ minute: 0 }], { epoch: OCT_4_2025, days: daylightSavingDays() }));
    expect(ds.days.map((d) => d.incomplete)).toEqual([false, false, true]);
  });

  it('groups streets into roads and counts totals', () => {
    const ds = decodeDataset(
      makeFile([
        { minute: 1, street: 0, subtype: 1 },
        { minute: 2, street: 1, subtype: 1 },
        { minute: 3, street: 2, roadType: 1 },
      ]),
    );
    expect(ds.roads[0].name).toBe('M31 – Hume Mwy');
    expect([...ds.road]).toEqual([0, 0, 1]);
    expect([...ds.totals.roads.slice(0, 2)]).toEqual([2, 1]);
    expect([...ds.totals.subtypes]).toEqual([1, 2, 0]);
    expect(ds.totals.roadTypes[3]).toBe(2);
  });

  it('finds the extent of a road', () => {
    const ds = decodeDataset(
      makeFile([
        { minute: 1, street: 0, lat: -34.1, lon: 150.8 },
        { minute: 2, street: 1, lat: -34.7, lon: 150.0 },
        { minute: 3, street: 2, lat: -35.3, lon: 149.1 },
      ]),
    );
    expect(extentOf(ds, 'road', 0)).toEqual([150.0, -34.7, 150.8, -34.1]);
    expect(extentOf(ds, 'town', 2)).toBeNull();
  });

  it('rejects columns of the wrong length', () => {
    const file = makeFile([{ minute: 1 }, { minute: 2 }]);
    file.cols.lat.pop();
    expect(() => decodeDataset(file)).toThrow(/lat/);
  });
});
